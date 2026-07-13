package compiler

import (
	"encoding/json"
	"math/big"
	"path/filepath"
	"slices"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/formancehq/numscript/internal/vm"

	"github.com/stretchr/testify/require"
)

const scriptsFolder = "../interpreter/testdata/script-tests"

// scriptsBlacklist lists spec files the compiler+VM can't run yet (unimplemented
// core features: variables, save, tx/account metadata, ...). Feature-flag specs
// are skipped separately. Delete entries as features land, until it's empty.
var scriptsBlacklist = []string{
	"add-monetaries-same-currency.num",
	"add-numbers.num",
	"allocate-dont-take-too-much.num",
	"allocation.num",
	"ask-balance-twice.num",
	"balance-not-found.num",
	"balance-simple.num",
	"balance.num",
	"big-int-monetary.num",
	"big-int.num",
	"bigint-literal.num",
	"cascading-sources.num",
	"dynamic-allocation.num",
	"feature-flag-syntax.num",
	"insufficient-funds.num",
	"metadata.num",
	"minus-infix-monetary.num",
	"minus-infix-number.num",
	"minus-prefix-number.num",
	"neg-max-dest.num",
	"negative-max-send-all.num",
	"negative-max.num",
	"nested-remaining-complex.num",
	"ovedrafts-playground-example.num",
	"overdraft-not-enough-funds.num",
	"overdraft-when-enough-funds.num",
	"overdraft-when-negative-balance-in-send-all.num",
	"overdraft-when-negative-balance.num",
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
	"use-balance-twice.num",
	"use-different-assets-with-same-source-account.num",
	"variable-asset.num",
	"variable-balance__1.num",
	"variable-balance__2.num",
	"variable-balance__3.num",
	"variable-balance__4.num",
	"variable-balance__5.num",
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
			var specs specs_format.Specs
			require.NoError(t, json.Unmarshal(rawSpec.SpecsFileContent, &specs))

			if len(specs.FeatureFlags) != 0 {
				t.Skip("feature-flag spec (core numscript only)")
			}
			if slices.Contains(scriptsBlacklist, rel) {
				t.Skip("blacklisted: not supported yet")
			}

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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)
	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

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
