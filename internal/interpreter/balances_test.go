package interpreter

import (
	"math/big"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestPrettyPrintBalance(t *testing.T) {
	fullBalance := Balances{
		{Account: "alice", Asset: "EUR/2", Amount: big.NewInt(1)},
		{Account: "alice", Asset: "USD/1234", Amount: big.NewInt(999999)},
		{Account: "bob", Asset: "BTC", Amount: big.NewInt(3)},
	}

	snaps.MatchSnapshot(t, fullBalance.PrettyPrint())
}

func TestCmpMaps(t *testing.T) {
	b1 := Balances{
		{Account: "alice", Asset: "EUR", Amount: big.NewInt(100)},
	}

	b2 := Balances{
		{Account: "alice", Asset: "EUR", Amount: big.NewInt(42)},
	}

	require.Equal(t, false, CompareBalances(b1, b2))
}

func TestCompareBalancesMultiplicity(t *testing.T) {
	x := BalanceRow{Account: "alice", Asset: "EUR", Amount: big.NewInt(1)}
	y := BalanceRow{Account: "bob", Asset: "EUR", Amount: big.NewInt(1)}

	// [x, x] must not equal [x, y] just because each x is "contained" in the other
	require.False(t, CompareBalances(Balances{x, x}, Balances{x, y}))
	require.False(t, CompareBalances(Balances{x, y}, Balances{x, x}))

	// order-independent and multiplicity-exact equality still holds
	require.True(t, CompareBalances(Balances{x, y}, Balances{y, x}))
	require.True(t, CompareBalances(Balances{x, x}, Balances{x, x}))
}

func TestCmpMapsIncluding(t *testing.T) {
	t.Run("including (subset)", func(t *testing.T) {
		b2 := Balances{
			{Account: "alice", Asset: "EUR", Amount: big.NewInt(100)},
			{Account: "bob", Asset: "EUR", Amount: big.NewInt(100)},
		}

		b1 := Balances{
			{Account: "alice", Asset: "EUR", Amount: big.NewInt(100)},
		}

		require.Equal(t, true, CompareBalancesIncluding(b1, b2))
	})

	t.Run("different value", func(t *testing.T) {
		b2 := Balances{
			{Account: "alice", Asset: "EUR", Amount: big.NewInt(100)},
			{Account: "bob", Asset: "EUR", Amount: big.NewInt(100)},
		}

		b1 := Balances{
			{Account: "alice", Asset: "EUR", Amount: big.NewInt(0)},
		}

		require.Equal(t, false, CompareBalancesIncluding(b1, b2))
	})

	t.Run("extra value", func(t *testing.T) {
		b2 := Balances{
			{Account: "alice", Asset: "EUR", Amount: big.NewInt(100)},
			{Account: "bob", Asset: "EUR", Amount: big.NewInt(100)},
		}

		b1 := Balances{
			{Account: "alice", Asset: "EUR", Amount: big.NewInt(100)},
			{Account: "extra-value", Asset: "EUR", Amount: big.NewInt(100)},
		}

		require.Equal(t, false, CompareBalancesIncluding(b1, b2))
	})
}
