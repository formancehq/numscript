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
