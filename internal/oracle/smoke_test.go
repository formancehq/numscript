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
	// Unlike the rewrite, the legacy
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

// TestKeptComplex, adapted from ledger's machine_kept_test.go, with the
// expectations changed. DELIBERATE DIVERGENCE FROM LEDGER, see DIVERGENCES.md
// #2: ledger does not consume a kept funding, so it returns to the pool and
// funds later destinations; the oracle consumes it, matching the interpreter.
//
// Ledger asserts 6 postings, all 11 kept GEM coming off @baz (the last source):
//
//	foo->arst 2, foo->thing 18, bar->thing 24, bar->qux 16, baz->qux 4, baz->quz 25
//
// Here @foo keeps the 6 it kept inside the max-clause and @baz keeps only the
// outer 5%, so @thing and @qux are funded differently. Totals per destination
// are identical either way: arst 2, thing 42, qux 20, quz 25.
func TestOracleKeptComplex(t *testing.T) {
	store := vm.StaticStore{
		"foo": {Account: vm.Account{Address: "foo"}, Balances: map[string]*big.Int{"GEM": big.NewInt(20)}},
		"bar": {Account: vm.Account{Address: "bar"}, Balances: map[string]*big.Int{"GEM": big.NewInt(40)}},
		"baz": {Account: vm.Account{Address: "baz"}, Balances: map[string]*big.Int{"GEM": big.NewInt(40)}},
	}

	m, err := compileAndRun(t, `send [GEM 100] (
	source = {
		@foo
		@bar
		@baz
	}
	destination = {
		50% to {
			max [GEM 8] to {
				50% kept
				25% to @arst
				25% kept
			}
			remaining to @thing
		}
		20% to @qux
		5% kept
		remaining to @quz
	}
)`, store)
	require.NoError(t, err)

	require.Len(t, m.Postings, 6)
	requirePosting(t, m.Postings[0], "foo", "arst", "GEM", 2)
	requirePosting(t, m.Postings[1], "foo", "thing", "GEM", 12)
	requirePosting(t, m.Postings[2], "bar", "thing", "GEM", 30)
	requirePosting(t, m.Postings[3], "bar", "qux", "GEM", 10)
	requirePosting(t, m.Postings[4], "baz", "qux", "GEM", 10)
	requirePosting(t, m.Postings[5], "baz", "quz", "GEM", 25)
}

// `save <monetary expression>` used to reserve only the expression's left
// operand: VisitSaveFromAccount asked VisitExpr for an address, and on that
// path no arithmetic operator is emitted and the address returned is the left
// operand's. Fixed upstream by ledger PR #2063, which the vendored copy
// carries; this pins that the vendoring did not lose it.
func TestOracleSaveMonetaryExpression(t *testing.T) {
	for _, tc := range []struct {
		name string
		save string
		// alice starts with 100; the send then moves whatever is unreserved.
		wantSent int64
	}{
		{"subtraction", `save [COIN 50] - [COIN 40] from @alice`, 90},
		{"addition", `save [COIN 90] + [COIN 5] from @alice`, 5},
		{"plain monetary", `save [COIN 50] from @alice`, 50},
		{"save all", `save [COIN *] from @alice`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := vm.StaticStore{
				"alice": {
					Account:  vm.Account{Address: "alice"},
					Balances: map[string]*big.Int{"COIN": big.NewInt(100)},
				},
			}

			m, err := compileAndRun(t, tc.save+`

send [COIN *] (
	source = @alice
	destination = @bob
)`, store)
			require.NoError(t, err)

			require.Len(t, m.Postings, 1)
			requirePosting(t, m.Postings[0], "alice", "bob", "COIN", tc.wantSent)
		})
	}
}

// OP_SAVE subtracted the saved amount from the tracked balance without
// checking its sign, so a negative INFLATED the balance: with @alice holding
// 100, `save [USD 10] - [USD 20]` let her send 110. Only reachable once save
// started evaluating its expression (see TestOracleSaveMonetaryExpression);
// before that the right operand was dropped and the amount was never
// negative. Fixed upstream by ledger PR #2068, which the vendored copy
// carries; this pins that the vendoring did not lose it.
func TestOracleSaveNegativeAmountRejected(t *testing.T) {
	store := vm.StaticStore{
		"alice": {
			Account:  vm.Account{Address: "alice"},
			Balances: map[string]*big.Int{"USD": big.NewInt(100)},
		},
	}

	_, err := compileAndRun(t, `save [USD 10] - [USD 20] from @alice

send [USD *] (
	source = @alice
	destination = @bob
)`, store)

	require.ErrorContains(t, err, "tried to save a negative amount: [USD -10]")
}

// DELIBERATE DIVERGENCE FROM LEDGER, see DIVERGENCES.md #3: ledger stores
// balance - saved as-is, so a save past the balance leaves it negative and a
// later bounded overdraft has less room. The oracle floors at zero, matching
// numscript's runSaveStatement.
//
// Ledger: acc0 is at -800 after the save, the overdraft of 1000 leaves 200 of
// room and the send of 250 fails with insufficient funds. Here it moves 250.
func TestOracleSaveFloorsAtZero(t *testing.T) {
	store := vm.StaticStore{
		"acc0": {
			Account:  vm.Account{Address: "acc0"},
			Balances: map[string]*big.Int{"COIN": big.NewInt(100)},
		},
	}

	m, err := compileAndRun(t, `save [COIN 900] from @acc0

send [COIN 250] (
	source = @acc0 allowing overdraft up to [COIN 1000]
	destination = @acc1
)`, store)
	require.NoError(t, err)

	require.Len(t, m.Postings, 1)
	requirePosting(t, m.Postings[0], "acc0", "acc1", "COIN", 250)
}

// The same floor from the other side of zero: a save of nothing on a negative
// balance raises it to zero. Ledger leaves -50 and moves 50; here 100.
func TestOracleSaveOnNegativeBalanceFloorsAtZero(t *testing.T) {
	store := vm.StaticStore{
		"acc0": {
			Account:  vm.Account{Address: "acc0"},
			Balances: map[string]*big.Int{"COIN": big.NewInt(-50)},
		},
	}

	m, err := compileAndRun(t, `save [COIN 0] from @acc0

send [COIN *] (
	source = @acc0 allowing overdraft up to [COIN 100]
	destination = @acc1
)`, store)
	require.NoError(t, err)

	require.Len(t, m.Postings, 1)
	requirePosting(t, m.Postings[0], "acc0", "acc1", "COIN", 100)
}
