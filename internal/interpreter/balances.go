package interpreter

import (
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/utils"
)

func (b Balances) DeepClone() Balances {
	cloned := make(Balances)
	for account, accountBalances := range b {
		for asset, amount := range accountBalances {
			utils.NestedMapGetOrPutDefault(cloned, account, asset, func() *big.Int {
				return new(big.Int).Set(amount)
			})
		}
	}
	return cloned
}

func coloredAsset(asset Asset, color String) string {
	strAsset := string(asset)

	if color == "" {
		return strAsset
	}

	// note: 1 <= len(parts) <= 2
	parts := strings.Split(strAsset, "/")

	coloredAsset := parts[0] + "_" + string(color)
	if len(parts) > 1 {
		coloredAsset += "/" + parts[1]
	}
	return coloredAsset
}

// Get the (account, asset) tuple from the Balances
// if the tuple is not present, it will write a big.NewInt(0) in it and return it
func (b Balances) fetchBalance(account AccountAddress, asset Asset, color String) *big.Int {
	return utils.NestedMapGetOrPutDefault(b,
		string(account),
		coloredAsset(asset, color),
		func() *big.Int {
			return new(big.Int)
		})
}

func (b Balances) has(account string, asset string) bool {
	accountBalances := utils.MapGetOrPutDefault(b, account, func() AccountBalance {
		return AccountBalance{}
	})

	_, ok := accountBalances[asset]
	return ok
}

// given a BalanceQuery, return a new query which only contains needed (asset, account) pairs
// (that is, the ones that aren't already cached)
func (b Balances) filterQuery(q BalanceQuery) BalanceQuery {
	filteredQuery := BalanceQuery{}
	for _, item := range q {
		if !b.has(item.Account, item.Asset) {
			filteredQuery = append(filteredQuery, item)
		}
	}
	return filteredQuery
}

// Merge balances by adding balances in the "update" arg
func (b Balances) Merge(update Balances) {
	// merge queried balance
	for acc, accBalances := range update {
		cachedAcc := utils.MapGetOrPutDefault(b, acc, func() AccountBalance {
			return AccountBalance{}
		})

		for curr, amt := range accBalances {
			cachedAcc[curr] = amt
		}
	}
}

func (b Balances) PrettyPrint() string {
	header := []string{"Account", "Asset", "Balance"}

	var rows [][]string
	for account, accBalances := range b {
		for asset, balance := range accBalances {
			row := []string{account, asset, balance.String()}
			rows = append(rows, row)
		}
	}
	return utils.CsvPretty(header, rows, true)
}

func CompareBalances(b1 Balances, b2 Balances) bool {
	return utils.Map2Cmp(b1, b2, func(ab1, ab2 *big.Int) bool {
		return ab1.Cmp(ab2) == 0
	})
}

// Returns whether the first value is a subset of the second one
func CompareBalancesIncluding(b1 Balances, b2 Balances) bool {
	return utils.MapIncludes(b2, b1, func(a2 AccountBalance, a1 AccountBalance) bool {
		return utils.MapIncludes(a2, a1, func(a2 *big.Int, a1 *big.Int) bool {
			return a2.Cmp(a1) == 0
		})
	})
}
