package interpreter

import (
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/utils"
)

func (b Balances) fetchAccountBalances(account string) AccountBalance {
	return defaultMapGet(b, account, func() AccountBalance {
		return AccountBalance{}
	})
}

func (b Balances) DeepClone() Balances {
	cloned := make(Balances)
	for account, accountBalances := range b {
		for asset, amount := range accountBalances {
			clonedAccountBalances := cloned.fetchAccountBalances(account)
			defaultMapGet(clonedAccountBalances, asset, func() *big.Int {
				return new(big.Int).Set(amount)
			})
		}
	}
	return cloned
}

func coloredAsset(asset string, color *string) string {
	if color == nil || *color == "" {
		return asset
	}

	// note: 1 <= len(parts) <= 2
	parts := strings.Split(asset, "/")

	coloredAsset := parts[0] + "_" + *color
	if len(parts) > 1 {
		coloredAsset += "/" + parts[1]
	}
	return coloredAsset
}

// Get the (account, asset) tuple from the Balances
// if the tuple is not present, it will write a big.NewInt(0) in it and return it
func (b Balances) fetchBalance(account string, uncoloredAsset string, color string) *big.Int {
	accountBalances := b.fetchAccountBalances(account)

	return defaultMapGet(accountBalances, coloredAsset(uncoloredAsset, &color), func() *big.Int {
		return new(big.Int)
	})
}

func (b Balances) has(account string, asset string) bool {
	accountBalances := defaultMapGet(b, account, func() AccountBalance {
		return AccountBalance{}
	})

	_, ok := accountBalances[asset]
	return ok
}

// given a BalanceQuery, return a new query which only contains needed (asset, account) pairs
// (that is, the ones that aren't already cached)
func (b Balances) filterQuery(q BalanceQuery) BalanceQuery {
	filteredQuery := BalanceQuery{}
	for accountName, queriedCurrencies := range q {
		filteredCurrencies := utils.Filter(queriedCurrencies, func(currency string) bool {
			return !b.has(accountName, currency)
		})

		if len(filteredCurrencies) > 0 {
			filteredQuery[accountName] = filteredCurrencies
		}

	}
	return filteredQuery
}

// Merge balances by adding balances in the "update" arg
func (b Balances) Merge(update Balances) {
	// merge queried balance
	for acc, accBalances := range update {
		cachedAcc := defaultMapGet(b, acc, func() AccountBalance {
			return AccountBalance{}
		})

		for curr, amt := range accBalances {
			cachedAcc[curr] = amt
		}
	}
}
