package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/utils"
)

// Uncolored builds a ColorBalance holding a single amount under the empty
// color key (the "no color" bucket). It is meant to keep test setup terse.
func Uncolored(amount *big.Int) ColorBalance {
	return ColorBalance{"": amount}
}

func (b Balances) DeepClone() Balances {
	cloned := make(Balances)
	for account, accountBalances := range b {
		clonedAcc := AccountBalance{}
		cloned[account] = clonedAcc
		for asset, colorMap := range accountBalances {
			clonedColors := ColorBalance{}
			clonedAcc[asset] = clonedColors
			for color, amount := range colorMap {
				clonedColors[color] = new(big.Int).Set(amount)
			}
		}
	}
	return cloned
}

// Get the (account, asset, color) balance from Balances.
// If the entry is not present, it will write a big.NewInt(0) in it and return it.
func (b Balances) fetchBalance(account string, asset string, color string) *big.Int {
	accBalance := utils.MapGetOrPutDefault(b, account, func() AccountBalance {
		return AccountBalance{}
	})
	colorMap := utils.MapGetOrPutDefault(accBalance, asset, func() ColorBalance {
		return ColorBalance{}
	})
	return utils.MapGetOrPutDefault(colorMap, color, func() *big.Int {
		return new(big.Int)
	})
}

func (b Balances) has(account string, asset string, color string) bool {
	accountBalances, ok := b[account]
	if !ok {
		return false
	}
	colorMap, ok := accountBalances[asset]
	if !ok {
		return false
	}
	_, ok = colorMap[color]
	return ok
}

// given a BalanceQuery, return a new query which only contains needed (asset, color) pairs
// (that is, the ones that aren't already cached)
func (b Balances) filterQuery(q BalanceQuery) BalanceQuery {
	filteredQuery := BalanceQuery{}
	for accountName, queriedItems := range q {
		filteredItems := utils.Filter(queriedItems, func(item AssetColor) bool {
			return !b.has(accountName, item.Asset, item.Color)
		})

		if len(filteredItems) > 0 {
			filteredQuery[accountName] = filteredItems
		}
	}
	return filteredQuery
}

// Merge balances by adding balances in the "update" arg
func (b Balances) Merge(update Balances) {
	for acc, accBalances := range update {
		cachedAcc := utils.MapGetOrPutDefault(b, acc, func() AccountBalance {
			return AccountBalance{}
		})

		for asset, colorMap := range accBalances {
			cachedColors := utils.MapGetOrPutDefault(cachedAcc, asset, func() ColorBalance {
				return ColorBalance{}
			})
			for color, amt := range colorMap {
				cachedColors[color] = amt
			}
		}
	}
}

func (b Balances) PrettyPrint() string {
	header := []string{"Account", "Asset", "Color", "Balance"}

	var rows [][]string
	for account, accBalances := range b {
		for asset, colorMap := range accBalances {
			for color, balance := range colorMap {
				rows = append(rows, []string{account, asset, color, balance.String()})
			}
		}
	}
	return utils.CsvPretty(header, rows, true)
}

func CompareBalances(b1 Balances, b2 Balances) bool {
	return utils.MapCmp(b1, b2, func(ab1, ab2 AccountBalance) bool {
		return utils.MapCmp(ab1, ab2, func(cm1, cm2 ColorBalance) bool {
			return utils.MapCmp(cm1, cm2, func(a1, a2 *big.Int) bool {
				return a1.Cmp(a2) == 0
			})
		})
	})
}

// Returns whether the first value is a subset of the second one
func CompareBalancesIncluding(b1 Balances, b2 Balances) bool {
	return utils.MapIncludes(b2, b1, func(a2 AccountBalance, a1 AccountBalance) bool {
		return utils.MapIncludes(a2, a1, func(cm2, cm1 ColorBalance) bool {
			return utils.MapIncludes(cm2, cm1, func(x2, x1 *big.Int) bool {
				return x2.Cmp(x1) == 0
			})
		})
	})
}
