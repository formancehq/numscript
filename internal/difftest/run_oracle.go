package difftest

import (
	"context"
	"maps"
	"math/big"

	"github.com/formancehq/numscript/internal/oracle/machine/script/compiler"
	"github.com/formancehq/numscript/internal/oracle/machine/vm"
)

func runOracle(ctx context.Context, script string, vars map[string]string) SideResult {
	p, err := compiler.Compile(script)
	if err != nil {
		return SideResult{CompileErr: err.Error()}
	}

	m := vm.NewMachine(*p)

	// SetVarsFromJSON mutates (clears) the map it's given, and vars here is
	// shared with runNew's call — pass it a copy so both callers, and
	// whoever reports the original vars on a mismatch, see the real value.
	if err := m.SetVarsFromJSON(maps.Clone(vars)); err != nil {
		return SideResult{CompileErr: err.Error()}
	}
	// vm.EmptyStore's GetBalances always returns nothing at all (not even
	// zero-value rows), unlike vm.StaticStore{} (an empty map), which
	// materializes a zero balance for every queried key. The oracle's own
	// credit() only updates a balance if the map entry already exists, so
	// with EmptyStore all in-script seed funding (world -> accN, credited
	// by an earlier statement in the same script) would be silently
	// dropped, causing spurious "missing balance" errors later on.
	store := vm.StaticStore{}
	if err := m.ResolveResources(ctx, store); err != nil {
		return SideResult{CompileErr: err.Error()}
	}
	if err := m.ResolveBalances(ctx, store); err != nil {
		return SideResult{CompileErr: err.Error()}
	}

	if err := m.Execute(); err != nil {
		return SideResult{RunErr: err.Error()}
	}

	// The legacy machine, unlike the rewrite, still emits zero-amount
	// postings (see differences-with-machine.md) — filter them out here so
	// the comparison isn't polluted by this one documented, intentional
	// difference.
	postings := make([]Posting, 0, len(m.Postings))
	for _, p := range m.Postings {
		amount := (*big.Int)(p.Amount)
		if amount.Sign() == 0 {
			continue
		}
		postings = append(postings, Posting{
			Source:      p.Source,
			Destination: p.Destination,
			Asset:       p.Asset,
			Amount:      amount,
		})
	}

	return SideResult{Postings: postings}
}
