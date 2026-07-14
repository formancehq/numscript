package compiler_test

import (
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

// scriptsBlacklist lists spec files the compiler+VM can't run yet: unimplemented
// core features (variables, metadata, ...) and everything under a feature flag
// (all in experimental/). Delete entries as features land, until it's empty.
var scriptsBlacklist = []string{
	// feature-flagged (experimental) — not core numscript
	"experimental/account-interpolation/account-interp.num",
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
	"experimental/get-amount-function/get-amount-function.num",
	"experimental/get-asset-function/get-asset-function.num",
	"experimental/oneof/oneof-all-failing.num",
	"experimental/oneof/oneof-destination-first-clause.num",
	"experimental/oneof/oneof-destination-remaining-clause.num",
	"experimental/oneof/oneof-destination-second-clause.num",
	"experimental/oneof/oneof-in-send-all.num",
	"experimental/oneof/oneof-in-source-send-first-branch.num",
	"experimental/oneof/oneof-in-source.num",
	"experimental/oneof/oneof-singleton.num",
	"experimental/oneof/update-balances-with-oneof.num",
	"experimental/overdraft-function/overdraft-function-use-case-remove-debt.num",
	"experimental/overdraft-function/overdraft-function-when-negative.num",
	"experimental/overdraft-function/overdraft-function-when-positive.num",
	"experimental/overdraft-function/overdraft-function-when-zero.num",
	"experimental/overdraft-function/reach-zero.num",

	// unimplemented core features
	"feature-flag-syntax.num",
	"overdraft-when-negative-balance-in-send-all.num",
	"overdraft-when-negative-ovedraft-in-send-all.num",
	"send-all-destinatio-allot-complex.num",
	"send-all-multi.num",
	"send-allt-max-in-src.num",
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

	enc, program, cErr := compiler.Compile(parsed.Value)
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
		res, execErr := vm.Exec(machine, &vars, store)

		if tc.ExpectMissingFunds {
			require.IsType(t, vm.MissingFundsError{}, execErr, "case %q", tc.It)
			continue
		}
		require.Nil(t, execErr, "case %q: unexpected error: %v", tc.It, execErr)

		if tc.ExpectPostings != nil {
			requirePostingsEqual(t, tc.ExpectPostings, res.Postings)
		}
		if tc.ExpectTxMeta != nil {
			require.Equal(t, tc.ExpectTxMeta, res.Metadata, "case %q: tx metadata", tc.It)
		}
		if tc.ExpectAccountsMeta != nil {
			require.Equal(t, tc.ExpectAccountsMeta, res.AccountsMetadata, "case %q: account metadata", tc.It)
		}
	}
}

func scriptStore(balancesOuter, balancesInner interpreter.Balances, metaOuter, metaInner interpreter.AccountsMetadata) e2eStore {
	m := map[runtime.PairKey]*big.Int{}
	for _, b := range append(append(interpreter.Balances{}, balancesOuter...), balancesInner...) {
		m[runtime.PairKey{Account: b.Account, Asset: b.Asset, Color: b.Color}] = b.Amount
	}

	meta := map[string]map[string]string{}
	for _, src := range []interpreter.AccountsMetadata{metaOuter, metaInner} {
		for acc, kv := range src {
			if meta[acc] == nil {
				meta[acc] = map[string]string{}
			}
			for k, v := range kv {
				meta[acc][k] = v
			}
		}
	}

	return e2eStore{balances: m, metadata: meta}
}
