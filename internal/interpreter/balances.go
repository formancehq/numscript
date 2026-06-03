package interpreter

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/utils"
)

// Uncolored builds a ColorBalance holding a single amount under the empty
// color key (the "no color" bucket). It is meant to keep test setup terse.
func Uncolored(amount *big.Int) ColorBalance {
	return ColorBalance{"": amount}
}

// UnmarshalJSON accepts two JSON shapes for a Balances value:
//
//  1. Flat (uncolored shorthand):
//     {"alice": {"USD/2": 100, "EUR/2": -42}}
//     Every asset gets a single entry under the "" (no color) bucket.
//
//  2. Colored (full form):
//     {"alice": {"USD/2": {"": 100, "GRANTS": 50}}}
//
// Shapes can be mixed across assets within the same document.
func (b *Balances) UnmarshalJSON(data []byte) error {
	raw := map[string]map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	out := Balances{}
	for account, assets := range raw {
		accB := AccountBalance{}
		out[account] = accB
		for asset, rawVal := range assets {
			colorMap, err := decodeColorBalance(rawVal)
			if err != nil {
				return fmt.Errorf("balances[%s][%s]: %w", account, asset, err)
			}
			accB[asset] = colorMap
		}
	}
	*b = out
	return nil
}

func decodeColorBalance(data json.RawMessage) (ColorBalance, error) {
	// Try the shorthand: a single number meaning the uncolored bucket.
	var amount json.Number
	if err := json.Unmarshal(data, &amount); err == nil {
		n, ok := new(big.Int).SetString(amount.String(), 10)
		if !ok {
			return nil, fmt.Errorf("invalid integer amount %q", amount.String())
		}
		return ColorBalance{"": n}, nil
	}

	// Otherwise expect a {color: amount} object.
	raw := map[string]json.Number{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("expected integer or {color: amount} object, got %s", string(data))
	}
	out := ColorBalance{}
	for color, amt := range raw {
		n, ok := new(big.Int).SetString(amt.String(), 10)
		if !ok {
			return nil, fmt.Errorf("color %q: invalid integer amount %q", color, amt.String())
		}
		out[color] = n
	}
	return out, nil
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
