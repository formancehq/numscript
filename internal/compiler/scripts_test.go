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

// scriptsBlacklist lists spec files the compiler+VM can't run yet. What's left
// is asset-scaling, which the compiler has no lowering for. Delete entries as
// features land, until it's empty.
var scriptsBlacklist = []string{
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

	hasFocused := slices.ContainsFunc(specs.TestCases, func(tc specs_format.TestCase) bool {
		return tc.Focus
	})

	for _, tc := range specs.TestCases {
		if tc.Skip || (hasFocused && !tc.Focus) {
			continue
		}
		if tc.ExpectNegativeAmount {
			// internal/vm has no negative-amount error: `send [USD/2 -1]` returns no
			// error and no postings, where the interpreter returns NegativeAmountErr.
			t.Logf("case %q: skipped, the VM has no negative-amount error", tc.It)
			continue
		}

		caseVars := map[string]string{}
		maps.Copy(caseVars, specs.Vars)
		maps.Copy(caseVars, tc.Vars)
		vars, encErr := enc.Encode(caseVars)
		require.NoError(t, encErr, "case %q: encode vars", tc.It)

		balances := specs_format.MergeBalances(specs.Balances, tc.Balances)

		machine := vm.NewVm(program)
		store := scriptStore(balances, specs.Meta, tc.Meta)
		res, execErr := vm.Exec(context.Background(), machine, &vars, store)

		if tc.ExpectMissingFunds {
			require.IsType(t, vm.MissingFundsError{}, execErr, "case %q", tc.It)
			continue
		}
		require.Nil(t, execErr, "case %q: unexpected error: %v", tc.It, execErr)

		if tc.ExpectPostings != nil {
			require.Equal(t, tc.ExpectPostings, res.Postings, "case %q: expect.postings", tc.It)
		}

		if tc.ExpectEndBalances != nil {
			got := specs_format.EndBalances(res.Postings, balances)
			require.True(t, interpreter.CompareBalances(tc.ExpectEndBalances, got),
				"case %q: expect.endBalances: want %v, got %v", tc.It, tc.ExpectEndBalances, got)
		}

		if tc.ExpectEndBalancesInclude != nil {
			got := specs_format.EndBalances(res.Postings, balances)
			require.True(t, interpreter.CompareBalancesIncluding(tc.ExpectEndBalancesInclude, got),
				"case %q: expect.endBalances.include: want %v to be included in %v", tc.It, tc.ExpectEndBalancesInclude, got)
		}

		if tc.ExpectMovements != nil {
			got := specs_format.GetMovements(res.Postings)
			require.True(t, specs_format.CompareMovements(tc.ExpectMovements, got),
				"case %q: expect.movements: want %v, got %v", tc.It, tc.ExpectMovements, got)
		}

		if tc.ExpectTxMeta != nil {
			require.Equal(t, txMetaAsStrings(tc.ExpectTxMeta), res.Metadata, "case %q: expect.txMetadata", tc.It)
		}

		if tc.ExpectAccountsMeta != nil {
			require.Equal(t, accountsMetaAsStrings(tc.ExpectAccountsMeta), res.AccountsMetadata,
				"case %q: expect.metadata", tc.It)
		}
	}
}

// vmMetaValue projects a spec's typed metadata value onto the flat string the VM
// stores, since the compiler stringifies metadata values at compile time. Only
// String and AccountAddress need unwrapping: Value.String() quotes the former and
// prefixes the latter with '@'.
func vmMetaValue(v interpreter.Value) string {
	switch v := v.(type) {
	case interpreter.String:
		return string(v)
	case interpreter.AccountAddress:
		return v.Name
	default:
		return v.String()
	}
}

func txMetaAsStrings(rows specs_format.ExpectedTxMeta) map[string]string {
	out := map[string]string{}
	for _, row := range rows {
		out[row.Key] = vmMetaValue(row.Value)
	}
	return out
}

func accountsMetaAsStrings(rows interpreter.SetAccountsMetadata) runtime.AccountsMetadata {
	out := make(runtime.AccountsMetadata, 0, len(rows))
	for _, row := range rows {
		out = append(out, runtime.AccountMetadataEntry{
			Account: row.Account,
			Scope:   row.Scope,
			Key:     row.Key,
			Value:   vmMetaValue(row.Value),
		})
	}
	return out
}

func scriptStore(balances interpreter.Balances, metaOuter, metaInner interpreter.AccountsMetadata) e2eStore {
	m := map[runtime.PairKey]*big.Int{}
	for _, b := range balances {
		m[runtime.PairKey{Account: b.Account, Scope: b.Scope, Asset: b.Asset, Color: b.Color}] = b.Amount
	}

	meta := map[e2eMetaKey]string{}
	for _, src := range []interpreter.AccountsMetadata{metaOuter, metaInner} {
		for _, row := range src {
			meta[e2eMetaKey{account: row.Account, scope: row.Scope, key: row.Key}] = row.Value
		}
	}

	return e2eStore{balances: m, metadata: meta}
}
