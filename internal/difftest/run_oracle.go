package difftest

import (
	"context"
	"maps"
	"math/big"

	acctmetadata "github.com/formancehq/go-libs/v5/pkg/types/metadata"
	"github.com/formancehq/numscript/internal/gen"
	"github.com/formancehq/numscript/internal/oracle/machine"
	"github.com/formancehq/numscript/internal/oracle/machine/script/compiler"
	"github.com/formancehq/numscript/internal/oracle/machine/vm"
)

func runOracle(ctx context.Context, script string, vars map[string]string, balances map[gen.BalanceKey]*big.Int, metadata map[gen.MetaKey]string) SideResult {
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
	getOrCreateEntry := func(account string) *vm.AccountWithBalances {
		entry, ok := store[account]
		if !ok {
			entry = &vm.AccountWithBalances{
				Account:  vm.Account{Address: account, Metadata: acctmetadata.Metadata{}},
				Balances: map[string]*big.Int{},
			}
			store[account] = entry
		}
		return entry
	}
	for k, amount := range balances {
		getOrCreateEntry(k.Account).Balances[k.Asset] = new(big.Int).Set(amount)
	}
	for k, value := range metadata {
		getOrCreateEntry(k.Account).Metadata[k.Key] = value
	}
	if err := m.ResolveResources(ctx, store); err != nil {
		return SideResult{CompileErr: err.Error()}
	}
	if err := m.ResolveBalances(ctx, store); err != nil {
		return SideResult{CompileErr: err.Error()}
	}

	if err := m.Execute(); err != nil {
		return SideResult{RunErr: err.Error(), MissingFunds: machine.IsInsufficientFundError(err)}
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
