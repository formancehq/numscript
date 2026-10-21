package difftest

import (
	"context"
	"fmt"
	"maps"
	"math/big"
	"strings"

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

	// SetVarsFromJSON clears the map it is given, and vars is shared with
	// runNew's call, so pass a copy: a mismatch report needs the real values.
	if err := m.SetVarsFromJSON(maps.Clone(vars)); err != nil {
		return SideResult{ResolveErr: err.Error()}
	}
	// Not vm.EmptyStore: its GetBalances returns no rows at all, where
	// vm.StaticStore{} materializes a zero balance per queried key. credit() only
	// updates a balance whose map entry already exists, so EmptyStore would
	// silently drop all in-script `world -> accN` funding.
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
		return SideResult{ResolveErr: err.Error()}
	}
	if err := m.ResolveBalances(ctx, store); err != nil {
		return SideResult{ResolveErr: err.Error()}
	}

	if err := m.Execute(); err != nil {
		return SideResult{
			RunErr:            err.Error(),
			MissingFunds:      machine.IsInsufficientFundError(err),
			NegativeMaxReject: isNegativeMaxReject(err),
		}
	}

	// The legacy machine still emits zero-amount postings, the rewrite does not
	// (internal/oracle/DIVERGENCES.md).
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

	txMeta := make(map[string]string, len(m.TxMeta))
	for k, v := range m.TxMeta {
		txMeta[k] = normalizeOracleMetaValue(v)
	}
	accountMeta := map[string]string{}
	for account, kv := range m.AccountsMeta {
		// The machine keys accounts as "@name"; the rewrite reports the bare
		// address.
		addr := strings.TrimPrefix(string(account), "@")
		for k, v := range kv {
			accountMeta[metaKey(addr, k)] = normalizeOracleMetaValue(v)
		}
	}

	return SideResult{Postings: postings, TxMeta: txMeta, AccountMeta: accountMeta}
}

// isNegativeMaxReject reports whether err is ledger's OP_TAKE_MAX guard
// rejecting a negative `max` clause, source- or destination-side
// (oracle/DIVERGENCES.md #4). That guard (vm/machine.go's OP_TAKE_MAX case)
// returns a bare fmt.Errorf, not one of the vendored machine.Err* types, so it
// is matched by the fixed message it always produces rather than by type.
func isNegativeMaxReject(err error) bool {
	return strings.Contains(err.Error(), "cannot send a monetary with a negative amount")
}

// normalizeOracleMetaValue renders one legacy-machine metadata value the way
// the rewrite reports the same value, so the two can be compared as strings.
// The machine keeps typed values and its default formatting differs (a String
// prints with quotes, a Monetary as "[COIN 5]"), so this cannot be %v.
func normalizeOracleMetaValue(v machine.Value) string {
	switch x := v.(type) {
	case machine.String:
		return string(x)
	case machine.AccountAddress:
		return string(x)
	case machine.Asset:
		return string(x)
	case *machine.MonetaryInt:
		return (*big.Int)(x).String()
	case machine.Monetary:
		return string(x.Asset) + " " + (*big.Int)(x.Amount).String()
	case machine.Portion:
		// internal/gen never presets a "remaining" portion as metadata (only an
		// "n/d" literal, internal/gen/extra.go), and the rewrite has no such
		// concept for a bare portion value either.
		if x.Remaining {
			panic("normalizeOracleMetaValue: unexpected remaining portion in metadata")
		}
		// Matches interpreter.Portion.String(), also a bare big.Rat rendering.
		return x.Specific.String()
	default:
		// Every other value type internal/gen can write as metadata is covered
		// above. Guessing a %v rendering for anything else would risk a spurious
		// mismatch that looks like a real divergence.
		panic(fmt.Sprintf("normalizeOracleMetaValue: unhandled machine.Value type %T", v))
	}
}
