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
