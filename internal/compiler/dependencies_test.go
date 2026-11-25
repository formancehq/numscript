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
	"github.com/formancehq/numscript/internal/funds"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// TestIRDependencies checks ir.ResolveDependencies on every spec case against
// two bounds: it is never larger than interpreter.ResolveDependencies, and it
// covers every balance, metadata and posting a successful VM run touches.
func TestIRDependencies(t *testing.T) {
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

			parsed := parser.Parse(rawSpec.NumscriptContent)
			require.Empty(t, parsed.Errors)

			featureFlags := map[string]struct{}{}
			for _, flag := range specs.FeatureFlags {
				featureFlags[flag] = struct{}{}
			}
			enc, instrs, cErr := compiler.CompileIR(parsed.Value, featureFlags)
			require.Nil(t, cErr)
			program, err := ir.Assemble(instrs)
			require.NoError(t, err)

			for _, tc := range specs.TestCases {
				caseVars := map[string]string{}
				maps.Copy(caseVars, specs.Vars)
				maps.Copy(caseVars, tc.Vars)
				vars, encErr := enc.Encode(caseVars)
				require.NoError(t, encErr, "case %q: encode vars", tc.It)

				balances := specs_format.MergeBalances(specs.Balances, tc.Balances)
				store := scriptStore(balances, specs.Meta, tc.Meta)

				irDeps, irErr := ir.ResolveDependencies(context.Background(), instrs, &vars, store)

				interpDeps, interpErr := interpreter.ResolveDependencies(context.Background(),
					interpreter.StaticStore{Balances: balances, Meta: mergeMeta(specs.Meta, tc.Meta)},
					caseVars, parsed.Value)

				rec := &recordingVMStore{inner: store, observed: emptyDeps()}
				res, execErr := vm.Exec(context.Background(), vm.NewVm(program), &vars, rec)

				if execErr == nil {
					require.NoError(t, irErr, "case %q: the run succeeds", tc.It)
					observeResult(rec.observed, res.Postings, res.Metadata, res.AccountsMetadata)
					requireSubset(t, rec.observed, irDeps, "case %q: run ⊆ ir", tc.It)
				}

				switch {
				case irErr == nil && interpErr == nil:
					requireSubset(t, irDeps, fromInterpreterDeps(interpDeps), "case %q: ir ⊆ interpreter", tc.It)
				case irErr != nil && interpErr != nil:
				default:
					t.Logf("case %q: ir err = %v, interpreter err = %v, run err = %v", tc.It, irErr, interpErr, execErr)
				}
			}
		})
	}
}

type recordingVMStore struct {
	inner    vm.Store
	observed ir.Dependencies
}

func (s *recordingVMStore) GetBalance(ctx context.Context, account, scope, asset, color string) (*big.Int, error) {
	s.observed.AccountsReads[ir.AccountDependency{Account: account, Scope: scope, Asset: asset, Color: color}] = struct{}{}
	return s.inner.GetBalance(ctx, account, scope, asset, color)
}

func (s *recordingVMStore) GetMetadata(ctx context.Context, account, scope, key string) (string, bool, error) {
	s.observed.MetaReads[ir.MetaDependency{Account: account, Scope: scope, Key: key}] = struct{}{}
	return s.inner.GetMetadata(ctx, account, scope, key)
}

func emptyDeps() ir.Dependencies {
	return ir.Dependencies{
		AccountsReads:  map[ir.AccountDependency]struct{}{},
		AccountsWrites: map[ir.AccountDependency]struct{}{},
		MetaReads:      map[ir.MetaDependency]struct{}{},
		MetaWrites:     map[ir.MetaDependency]struct{}{},
		TxMetaWrites:   map[string]struct{}{},
	}
}

func observeResult(d ir.Dependencies, postings []funds.Posting, txMeta map[string]string, accountsMeta funds.AccountsMetadata) {
	for _, p := range postings {
		d.AccountsWrites[ir.AccountDependency{Account: p.Source, Scope: p.SourceScope, Asset: p.Asset, Color: p.Color}] = struct{}{}
		d.AccountsWrites[ir.AccountDependency{Account: p.Destination, Scope: p.DestinationScope, Asset: p.Asset, Color: p.Color}] = struct{}{}
	}
	for k := range txMeta {
		d.TxMetaWrites[k] = struct{}{}
	}
	for _, m := range accountsMeta {
		d.MetaWrites[ir.MetaDependency{Account: m.Account, Scope: m.Scope, Key: m.Key}] = struct{}{}
	}
}

func requireSubset(t *testing.T, sub, super ir.Dependencies, msgAndArgs ...any) {
	t.Helper()
	missing := emptyDeps()
	diff := func(sub, super, out map[ir.AccountDependency]struct{}) {
		for k := range sub {
			if _, ok := super[k]; !ok {
				out[k] = struct{}{}
			}
		}
	}
	diff(sub.AccountsReads, super.AccountsReads, missing.AccountsReads)
	diff(sub.AccountsWrites, super.AccountsWrites, missing.AccountsWrites)
	for k := range sub.MetaReads {
		if _, ok := super.MetaReads[k]; !ok {
			missing.MetaReads[k] = struct{}{}
		}
	}
	for k := range sub.MetaWrites {
		if _, ok := super.MetaWrites[k]; !ok {
			missing.MetaWrites[k] = struct{}{}
		}
	}
	for k := range sub.TxMetaWrites {
		if _, ok := super.TxMetaWrites[k]; !ok {
			missing.TxMetaWrites[k] = struct{}{}
		}
	}
	require.Equal(t, emptyDeps(), missing, msgAndArgs...)
}

func mergeMeta(outer, inner interpreter.AccountsMetadata) interpreter.AccountsMetadata {
	type key struct{ account, scope, key string }
	merged := map[key]interpreter.AccountMetadataRow{}
	var order []key
	for _, rows := range []interpreter.AccountsMetadata{outer, inner} {
		for _, row := range rows {
			k := key{row.Account, row.Scope, row.Key}
			if _, ok := merged[k]; !ok {
				order = append(order, k)
			}
			merged[k] = row
		}
	}
	out := make(interpreter.AccountsMetadata, 0, len(order))
	for _, k := range order {
		out = append(out, merged[k])
	}
	return out
}

func fromInterpreterDeps(d interpreter.ResolvedDependencies) ir.Dependencies {
	out := emptyDeps()
	for k := range d.AccountsReads {
		out.AccountsReads[ir.AccountDependency(k)] = struct{}{}
	}
	for k := range d.AccountsWrites {
		out.AccountsWrites[ir.AccountDependency(k)] = struct{}{}
	}
	for k := range d.MetaReads {
		out.MetaReads[ir.MetaDependency(k)] = struct{}{}
	}
	for k := range d.MetaWrites {
		out.MetaWrites[ir.MetaDependency(k)] = struct{}{}
	}
	maps.Copy(out.TxMetaWrites, d.TxMetaWrites)
	return out
}
