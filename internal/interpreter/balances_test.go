package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterQuery(t *testing.T) {
	fullBalance := Balances{
		"alice": AccountBalance{
			"EUR/2": big.NewInt(1),
			"USD/2": big.NewInt(2),
		},
		"bob": AccountBalance{
			"BTC": big.NewInt(3),
		},
	}

	filteredQuery := fullBalance.filterQuery(BalanceQuery{
		"alice":   []string{"GBP/2", "YEN", "EUR/2"},
		"bob":     []string{"BTC"},
		"charlie": []string{"ETH"},
	})

	require.Equal(t, BalanceQuery{
		"alice":   []string{"GBP/2", "YEN"},
		"charlie": []string{"ETH"},
	}, filteredQuery)
}

func TestCloneBalances(t *testing.T) {
	fullBalance := Balances{
		"alice": AccountBalance{
			"EUR/2": big.NewInt(1),
			"USD/2": big.NewInt(2),
		},
		"bob": AccountBalance{
			"BTC": big.NewInt(3),
		},
	}

	cloned := fullBalance.deepClone()

	fullBalance["alice"]["USD/2"].Set(big.NewInt(42))

	require.Equal(t, big.NewInt(2), cloned["alice"]["USD/2"])
}
