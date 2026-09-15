// Package oracle_test contains a handful of hand-picked smoke tests for the
// vendored legacy "machine" interpreter under internal/oracle/machine. These
// exist purely to confirm the vendoring (import-path rewrite + the small
// ledger.* replacement types) didn't subtly break anything, independent of
// any generator or comparison harness. They are not meant to be a
// replacement for the oracle's own (not copied) test suite.
package oracle_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/formancehq/numscript/internal/oracle/machine/script/compiler"
	"github.com/formancehq/numscript/internal/oracle/machine/vm"
)

func compileAndRun(t *testing.T, script string, store vm.StaticStore) (*vm.Machine, error) {
	t.Helper()

	p, err := compiler.Compile(script)
	require.NoError(t, err)

	m := vm.NewMachine(*p)

	err = m.ResolveResources(context.Background(), store)
	require.NoError(t, err)

	err = m.ResolveBalances(context.Background(), store)
	require.NoError(t, err)

	return m, m.Execute()
}

func TestOracleSimpleSend(t *testing.T) {
	store := vm.StaticStore{
		"alice": {
			Account:  vm.Account{Address: "alice"},
			Balances: map[string]*big.Int{"USD/2": big.NewInt(100)},
		},
	}

	m, err := compileAndRun(t, `
send [USD/2 100] (
  source = @alice
  destination = @bob
)
`, store)
	require.NoError(t, err)

	require.Len(t, m.Postings, 1)
	require.Equal(t, "alice", m.Postings[0].Source)
	require.Equal(t, "bob", m.Postings[0].Destination)
	require.Equal(t, "USD/2", m.Postings[0].Asset)
	require.Equal(t, big.NewInt(100), (*big.Int)(m.Postings[0].Amount))
}

func TestOracleInsufficientFunds(t *testing.T) {
	store := vm.StaticStore{
		"alice": {
			Account:  vm.Account{Address: "alice"},
			Balances: map[string]*big.Int{"USD/2": big.NewInt(10)},
		},
	}

	_, err := compileAndRun(t, `
send [USD/2 100] (
  source = @alice
  destination = @bob
)
`, store)
	require.Error(t, err)
}

func TestOracleZeroPostingIsEmitted(t *testing.T) {
	// Unlike the rewrite (see differences-with-machine.md), the legacy
	// machine DOES emit zero-amount postings. This is asserted here so the
	// difference is pinned down at the oracle level, not just documented.
	store := vm.StaticStore{}

	m, err := compileAndRun(t, `
send [USD/2 0] (
  source = @alice
  destination = @bob
)
`, store)
	require.NoError(t, err)

	require.Len(t, m.Postings, 1)
	require.Equal(t, big.NewInt(0), (*big.Int)(m.Postings[0].Amount))
}

// compileAndRunWithVars is compileAndRun plus variable bindings.
func compileAndRunWithVars(t *testing.T, script string, vars map[string]string, store vm.StaticStore) (*vm.Machine, error) {
	t.Helper()

	p, err := compiler.Compile(script)
	require.NoError(t, err)

	m := vm.NewMachine(*p)

	require.NoError(t, m.SetVarsFromJSON(vars))

	err = m.ResolveResources(context.Background(), store)
	require.NoError(t, err)

	err = m.ResolveBalances(context.Background(), store)
	require.NoError(t, err)

	return m, m.Execute()
}

func requirePosting(t *testing.T, p vm.Posting, src, dst, asset string, amount int64) {
	t.Helper()
	require.Equal(t, src, p.Source)
	require.Equal(t, dst, p.Destination)
	require.Equal(t, asset, p.Asset)
	require.Equal(t, big.NewInt(amount), (*big.Int)(p.Amount))
}

// The three balance() resource collisions from ledger 52da79031: two distinct
// resources could resolve to the same account, to the same asset, or want two
// assets of one account. Keying balance requests by account address dropped
// all but one, leaving a resource with a nil amount that panicked on first use.

func TestOracleBalanceVarsOnAliasedAccountResources(t *testing.T) {
	store := vm.StaticStore{
		"src": {
			Account:  vm.Account{Address: "src"},
			Balances: map[string]*big.Int{"USD": big.NewInt(10)},
		},
	}

	m, err := compileAndRunWithVars(t, `vars {
	account $acc
	monetary $a = balance($acc, USD)
	monetary $b = balance(@src, USD)
}

send $a (
	source = @world
	destination = @dst1
)

send $b (
	source = @world
	destination = @dst2
)`, map[string]string{"acc": "src"}, store)
	require.NoError(t, err)

	require.Len(t, m.Postings, 2)
	requirePosting(t, m.Postings[0], "world", "dst1", "USD", 10)
	requirePosting(t, m.Postings[1], "world", "dst2", "USD", 10)
}

func TestOracleBalanceVarsOnAliasedAssetResources(t *testing.T) {
	store := vm.StaticStore{
		"src": {
			Account:  vm.Account{Address: "src"},
			Balances: map[string]*big.Int{"USD": big.NewInt(10)},
		},
	}

	m, err := compileAndRunWithVars(t, `vars {
	asset $ass
	monetary $a = balance(@src, $ass)
	monetary $b = balance(@src, USD)
}

send $a (
	source = @world
	destination = @dst1
)

send $b (
	source = @world
	destination = @dst2
)`, map[string]string{"ass": "USD"}, store)
	require.NoError(t, err)

	require.Len(t, m.Postings, 2)
	requirePosting(t, m.Postings[0], "world", "dst1", "USD", 10)
	requirePosting(t, m.Postings[1], "world", "dst2", "USD", 10)
}

func TestOracleBalanceVarsOnSameAccountDifferentAssets(t *testing.T) {
	store := vm.StaticStore{
		"src": {
			Account: vm.Account{Address: "src"},
			Balances: map[string]*big.Int{
				"USD": big.NewInt(10),
				"EUR": big.NewInt(20),
			},
		},
	}

	m, err := compileAndRunWithVars(t, `vars {
	monetary $a = balance(@src, USD)
	monetary $b = balance(@src, EUR)
}

send $a (
	source = @world
	destination = @dst1
)

send $b (
	source = @world
	destination = @dst2
)`, map[string]string{}, store)
	require.NoError(t, err)

	require.Len(t, m.Postings, 2)
	requirePosting(t, m.Postings[0], "world", "dst1", "USD", 10)
	requirePosting(t, m.Postings[1], "world", "dst2", "EUR", 20)
}
