package compiler_test

import (
	"context"
	"encoding/json"
	"maps"
	"math/big"
	"path/filepath"
	"slices"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/formancehq/numscript/internal/vm"

	"github.com/stretchr/testify/require"
)

const scriptsFolder = "../interpreter/testdata/script-tests"

// scriptsBlacklist lists spec files the compiler+VM can't run yet. What's left is
// the three experimental features the compiler has no lowering for: scopes,
// colors and scaling. Delete entries as features land, until it's empty.
var scriptsBlacklist = []string{
	"experimental/scoped-function/allotment.num",
	"experimental/scoped-function/balance.num",
	"experimental/scoped-function/capped.num",
	"experimental/scoped-function/color-and-scope.num",
	"experimental/scoped-function/overdraft.num",
	"experimental/scoped-function/read-account-meta.num",
	"experimental/scoped-function/save.num",
	"experimental/scoped-function/set-account-meta.num",
	"experimental/scoped-function/simple.num",
	"experimental/asset-colors/color-inorder-send-all.num",
	"experimental/asset-colors/color-inorder.num",
	"experimental/asset-colors/color-restrict-balance-when-missing-funds.num",
	"experimental/asset-colors/color-restrict-balance.num",
	"experimental/asset-colors/color-restriction-in-send-all.num",
	"experimental/asset-colors/color-send-overdrat.num",
	"experimental/asset-colors/color-send.num",
	"experimental/asset-colors/color-with-asset-precision.num",
	"experimental/asset-colors/empty-color.num",
	"experimental/asset-colors/no-double-spending-in-colored-send-all.num",
	"experimental/asset-colors/no-double-spending-in-colored-send.num",
	"experimental/asset-scaling/no-solution.num",
	"experimental/asset-scaling/scaling-all-allotment.num",
	"experimental/asset-scaling/scaling-allotment.num",
	"experimental/asset-scaling/scaling-kept.num",
	"experimental/asset-scaling/scaling-send-all.num",
	"experimental/asset-scaling/scaling-with-oneof.num",
	"experimental/asset-scaling/scaling.num",
	"experimental/asset-scaling/update-swap-account-balance.num",
}

func TestCompilerScripts(t *testing.T) {
	rawSpecs, err := specs_format.ReadSpecsFiles([]string{scriptsFolder})
	require.NoError(t, err)

	for _, rawSpec := range rawSpecs {
		rel, err := filepath.Rel(scriptsFolder, rawSpec.NumscriptPath)
		require.NoError(t, err)

		t.Run(rel, func(t *testing.T) {
			if slices.Contains(scriptsBlacklist, rel) {
				t.Skip("blacklisted: not supported yet")
			}

			var specs specs_format.Specs
			require.NoError(t, json.Unmarshal(rawSpec.SpecsFileContent, &specs))

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic: %v", r)
				}
			}()

			runScriptSpec(t, specs, rawSpec.NumscriptContent)
		})
	}
}

func runScriptSpec(t *testing.T, specs specs_format.Specs, src string) {
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	featureFlags := make(map[string]struct{}, len(specs.FeatureFlags))
	for _, flag := range specs.FeatureFlags {
		featureFlags[flag] = struct{}{}
	}

	enc, program, cErr := compiler.Compile(parsed.Value, featureFlags)
	require.Nil(t, cErr)

	for _, tc := range specs.TestCases {
		if tc.Skip {
			continue
		}
		caseVars := map[string]string{}
		maps.Copy(caseVars, specs.Vars)
		maps.Copy(caseVars, tc.Vars)
		vars, encErr := enc.Encode(caseVars)
		require.NoError(t, encErr, "case %q: encode vars", tc.It)

		machine := vm.NewVm(program)
		store := scriptStore(specs.Balances, tc.Balances, specs.Meta, tc.Meta)
		res, execErr := vm.Exec(context.Background(), machine, &vars, store)

		if tc.ExpectMissingFunds {
			require.IsType(t, vm.MissingFundsError{}, execErr, "case %q", tc.It)
			continue
		}
		require.Nil(t, execErr, "case %q: unexpected error: %v", tc.It, execErr)

		if tc.ExpectPostings != nil {
			got := make([]interpreter.Posting, len(res.Postings))
			for i, p := range res.Postings {
				got[i] = interpreter.Posting{
					Source:      p.Source,
					Destination: p.Destination,
					Amount:      p.Amount,
					Asset:       p.Asset,
					Color:       p.Color,
				}
			}
			require.Equal(t, tc.ExpectPostings, got, "case %q", tc.It)
		}

		if tc.ExpectTxMeta != nil {
			gotMeta, err := interpreter.MetadataFromVM(res.Metadata)
			require.NoError(t, err, "case %q: tx metadata", tc.It)
			require.True(t,
				specs_format.CheckTxMeta(tc.ExpectTxMeta, gotMeta),
				"case %q: expect.txMetadata: expected %v, got %v", tc.It, tc.ExpectTxMeta, gotMeta)
		}

		if tc.ExpectAccountsMeta != nil {
			gotMeta, err := interpreter.SetAccountsMetadataFromVM(res.AccountsMetadata)
			require.NoError(t, err, "case %q: account metadata", tc.It)
			require.True(t,
				interpreter.CompareSetAccountsMetadata(tc.ExpectAccountsMeta, gotMeta),
				"case %q: expect.metadata: expected %v, got %v", tc.It, tc.ExpectAccountsMeta, gotMeta)
		}

		// Assertions this runner can't evaluate yet. Failing loudly beats skipping
		// them silently, which reads as coverage the spec isn't actually getting.
		require.Nil(t, tc.ExpectEndBalances, "case %q: expect.endBalances not supported by the compiler runner", tc.It)
		require.Nil(t, tc.ExpectEndBalancesInclude, "case %q: expect.endBalances.include not supported by the compiler runner", tc.It)
		require.Nil(t, tc.ExpectMovements, "case %q: expect.movements not supported by the compiler runner", tc.It)
		require.False(t, tc.ExpectNegativeAmount, "case %q: expect.error.negativeAmount not supported by the compiler runner", tc.It)
	}
}

func scriptStore(balancesOuter, balancesInner interpreter.Balances, metaOuter, metaInner interpreter.AccountsMetadata) e2eStore {
	m := map[runtime.PairKey]*big.Int{}
	for _, b := range append(append(interpreter.Balances{}, balancesOuter...), balancesInner...) {
		m[runtime.PairKey{Account: b.Account, Asset: b.Asset, Color: b.Color}] = b.Amount
	}

	meta := map[string]map[string]string{}
	for _, src := range []interpreter.AccountsMetadata{metaOuter, metaInner} {
		for _, row := range src {
			if meta[row.Account] == nil {
				meta[row.Account] = map[string]string{}
			}
			meta[row.Account][row.Key] = row.Value
		}
	}

	return e2eStore{balances: m, metadata: meta}
}
