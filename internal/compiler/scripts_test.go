package compiler_test

import (
	"encoding/json"
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
	"experimental/mid-script-function-call/expr-in-var-origin.num",
	"experimental/mid-script-function-call/midscript-balance-after-decrease.num",
	"experimental/mid-script-function-call/midscript-balance.num",
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
	"add-numbers.num",
	"allocate-dont-take-too-much.num",
	"allocation.num",
	"big-int-monetary.num",
	"big-int.num",
	"bigint-literal.num",
	"cascading-sources.num",
	"dynamic-allocation.num",
	"feature-flag-syntax.num",
	"insufficient-funds.num",
	"metadata.num",
	"neg-max-dest.num",
	"nested-remaining-complex.num",
	"ovedrafts-playground-example.num",
	"overdraft-not-enough-funds.num",
	"overdraft-when-enough-funds.num",
	"overdraft-when-negative-balance-in-send-all.num",
	"overdraft-when-negative-ovedraft-in-send-all.num",
	"overdraft-when-not-enough-funds.num",
	"override-account-meta.num",
	"portion-syntax.num",
	"save/save-from-account__with-asset-var.num",
	"save/save-from-account__with-monetary-var.num",
	"send-all-destinatio-allot-complex.num",
	"send-all-destinatio-allot.num",
	"send-all-multi.num",
	"send-all-variable.num",
	"send-all-when-negative-with-overdraft.num",
	"send-all-when-negative.num",
	"send-all.num",
	"send-allt-max-in-src.num",
	"set-account-meta.num",
	"set-tx-meta.num",
	"source-complex.num",
	"source.num",
	"sub-monetaries.num",
	"sub-numbers.num",
	"use-different-assets-with-same-source-account.num",
	"variable-asset.num",
	"variable-balance__3.num",
	"variable-balance__4.num",
	"variable-portion-part.num",
	"variables-json.num",
	"variables.num",
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

	program, cErr := compiler.Compile(parsed.Value)
	require.Nil(t, cErr)

	for _, tc := range specs.TestCases {
		if tc.Skip {
			continue
		}
		if tc.ExpectTxMeta != nil || tc.ExpectAccountsMeta != nil {
			t.Fatalf("case %q: metadata assertions not supported by the VM", tc.It)
		}

		store := scriptStore(specs.Balances, tc.Balances)
		postings, execErr := vm.Exec(vm.NewVm(program), nil, store)

		if tc.ExpectMissingFunds {
			require.IsType(t, vm.MissingFundsError{}, execErr, "case %q", tc.It)
			continue
		}
		require.Nil(t, execErr, "case %q: unexpected error: %v", tc.It, execErr)

		if tc.ExpectPostings != nil {
			requirePostingsEqual(t, tc.ExpectPostings, postings)
		}
	}
}

func scriptStore(outer, inner interpreter.Balances) e2eStore {
	m := map[runtime.PairKey]*big.Int{}
	for _, b := range append(append(interpreter.Balances{}, outer...), inner...) {
		m[runtime.PairKey{Account: b.Account, Asset: b.Asset, Color: b.Color}] = b.Amount
	}
	return e2eStore{balances: m}
}
