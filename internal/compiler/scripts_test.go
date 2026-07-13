package compiler

import (
	"encoding/json"
	"math/big"
	"path/filepath"
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
var scriptsBlacklist = map[string]bool{
	"add-monetaries-same-currency.num":                      true,
	"add-numbers.num":                                       true,
	"allocate-dont-take-too-much.num":                       true,
	"allocation.num":                                        true,
	"ask-balance-twice.num":                                 true,
	"balance-not-found.num":                                 true,
	"balance-simple.num":                                    true,
	"balance.num":                                           true,
	"big-int-monetary.num":                                  true,
	"big-int.num":                                           true,
	"bigint-literal.num":                                    true,
	"cascading-sources.num":                                 true,
	"dynamic-allocation.num":                                true,
	"feature-flag-syntax.num":                               true,
	"insufficient-funds.num":                                true,
	"metadata.num":                                          true,
	"minus-infix-monetary.num":                              true,
	"minus-infix-number.num":                                true,
	"minus-prefix-number.num":                               true,
	"neg-max-dest.num":                                      true,
	"negative-max-send-all.num":                             true,
	"negative-max.num":                                      true,
	"nested-remaining-complex.num":                          true,
	"ovedrafts-playground-example.num":                      true,
	"overdraft-not-enough-funds.num":                        true,
	"overdraft-when-enough-funds.num":                       true,
	"overdraft-when-negative-balance-in-send-all.num":       true,
	"overdraft-when-negative-balance.num":                   true,
	"overdraft-when-negative-ovedraft-in-send-all.num":      true,
	"overdraft-when-not-enough-funds.num":                   true,
	"override-account-meta.num":                             true,
	"portion-syntax.num":                                    true,
	"save/save-from-account__multi-postings.num":            true,
	"save/save-from-account__save-a-different-asset.num":    true,
	"save/save-from-account__save-all-negative-balance.num": true,
	"save/save-from-account__save-all.num":                  true,
	"save/save-from-account__save-causes-failure.num":       true,
	"save/save-from-account__save-more-than-balance.num":    true,
	"save/save-from-account__simple.num":                    true,
	"save/save-from-account__with-asset-var.num":            true,
	"save/save-from-account__with-monetary-var.num":         true,
	"send-all-destinatio-allot-complex.num":                 true,
	"send-all-destinatio-allot.num":                         true,
	"send-all-multi.num":                                    true,
	"send-all-variable.num":                                 true,
	"send-all-when-negative-with-overdraft.num":             true,
	"send-all-when-negative.num":                            true,
	"send-all.num":                                          true,
	"send-allt-max-in-src.num":                              true,
	"set-account-meta.num":                                  true,
	"set-tx-meta.num":                                       true,
	"source-complex.num":                                    true,
	"source.num":                                            true,
	"sub-monetaries.num":                                    true,
	"sub-numbers.num":                                       true,
	"use-balance-twice.num":                                 true,
	"use-different-assets-with-same-source-account.num":     true,
	"variable-asset.num":                                    true,
	"variable-balance__1.num":                               true,
	"variable-balance__2.num":                               true,
	"variable-balance__3.num":                               true,
	"variable-balance__4.num":                               true,
	"variable-balance__5.num":                               true,
	"variable-portion-part.num":                             true,
	"variables-json.num":                                    true,
	"variables.num":                                         true,
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
			if scriptsBlacklist[rel] {
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
