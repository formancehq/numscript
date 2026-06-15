package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterQuery(t *testing.T) {
	fullBalance := InternalBalances{
		"alice": {
			{Asset: "EUR/2", Amount: big.NewInt(1)},
			{Asset: "USD/2", Amount: big.NewInt(2)},
		},
		"bob": {
			{Asset: "BTC", Amount: big.NewInt(3)},
		},
	}

	filteredQuery := fullBalance.filterQuery(BalanceQuery{
		{Account: "alice", Asset: "GBP/2"},
		{Account: "alice", Asset: "YEN"},
		{Account: "alice", Asset: "EUR/2"},
		{Account: "bob", Asset: "BTC"},
		{Account: "charlie", Asset: "ETH"},
	})

	require.Equal(t, BalanceQuery{
		{Account: "alice", Asset: "GBP/2"},
		{Account: "alice", Asset: "YEN"},
		{Account: "charlie", Asset: "ETH"},
	}, filteredQuery)
}

func TestBalancesFirstDuplicate(t *testing.T) {
	// no duplicate: same account/asset but different color are distinct keys
	_, ok := Balances{
		{Account: "alice", Asset: "USD/2", Amount: big.NewInt(1)},
		{Account: "alice", Asset: "EUR/2", Amount: big.NewInt(2)},
		{Account: "alice", Asset: "USD/2", Color: "X", Amount: big.NewInt(3)},
		{Account: "bob", Asset: "USD/2", Amount: big.NewInt(4)},
	}.FirstDuplicate()
	require.False(t, ok)

	// duplicate (account, asset, color), even with a different amount
	dup, ok := Balances{
		{Account: "alice", Asset: "USD/2", Amount: big.NewInt(1)},
		{Account: "alice", Asset: "USD/2", Amount: big.NewInt(99)},
	}.FirstDuplicate()
	require.True(t, ok)
	require.Equal(t, BalanceRow{Account: "alice", Asset: "USD/2", Amount: big.NewInt(99)}, dup)
}

func TestCloneBalances(t *testing.T) {
	fullBalance := InternalBalances{
		"alice": {
			{Asset: "EUR/2", Amount: big.NewInt(1)},
			{Asset: "USD/2", Amount: big.NewInt(2)},
		},
		"bob": {
			{Asset: "BTC", Amount: big.NewInt(3)},
		},
	}

	cloned := fullBalance.DeepClone()

	// USD/2 is the second entry for alice (index 1).
	fullBalance["alice"][1].Amount.Set(big.NewInt(42))

	require.Equal(t, big.NewInt(2), cloned["alice"][1].Amount)
}
