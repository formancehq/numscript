package interpreter

import "math/big"

// An internal representation of the balances. Used to cache balances we get from external store.
// Whereas the external representation (interpreter.Balances) is user-facing and be a stable contract,
// (for example, allowing more columns if we need an higher level of fungibility), this one is used internally by the runtime, and
// could change over time, for example to add more indexes for faster lookups
type InternalBalances map[string][]AccountBalance

// A single balance entry for an account: an (asset, color) pair and its amount.
type AccountBalance struct {
	Asset  string
	Color  string
	Amount *big.Int
}

func FromBalancesRows(b Balances) InternalBalances {
	out := make(InternalBalances, len(b))
	for _, row := range b {
		amount := new(big.Int) // clone so the map doesn't alias the slice's *big.Int
		if row.Amount != nil {
			amount.Set(row.Amount)
		}
		out[row.Account] = append(out[row.Account], AccountBalance{
			Asset:  row.Asset,
			Color:  row.Color,
			Amount: amount,
		})
	}
	return out
}

func (b InternalBalances) DeepClone() InternalBalances {
	cloned := make(InternalBalances, len(b))
	for account, entries := range b {
		clonedEntries := make([]AccountBalance, len(entries))
		for i, e := range entries {
			clonedEntries[i] = AccountBalance{
				Asset:  e.Asset,
				Color:  e.Color,
				Amount: new(big.Int).Set(e.Amount),
			}
		}
		cloned[account] = clonedEntries
	}
	return cloned
}

// Get the (account, asset, color) balance from the cache.
// If it is not present, it writes a zero balance in it and returns it.
func (b InternalBalances) fetchBalance(account AccountAddress, asset Asset, color String) *big.Int {
	acc := string(account)
	for i := range b[acc] {
		entry := &b[acc][i]
		if entry.Asset == string(asset) && entry.Color == string(color) {
			return entry.Amount
		}
	}

	amount := new(big.Int)
	b[acc] = append(b[acc], AccountBalance{
		Asset:  string(asset),
		Color:  string(color),
		Amount: amount,
	})
	return amount
}

// Set assigns amount to the (account, asset, color) balance.
func (b InternalBalances) Set(account string, asset string, color string, amount *big.Int) {
	for i := range b[account] {
		if b[account][i].Asset == asset && b[account][i].Color == color {
			b[account][i].Amount = amount
			return
		}
	}
	b[account] = append(b[account], AccountBalance{
		Asset:  asset,
		Color:  color,
		Amount: amount,
	})
}

func (b InternalBalances) has(account string, asset string, color string) bool {
	for _, entry := range b[account] {
		if entry.Asset == asset && entry.Color == color {
			return true
		}
	}
	return false
}

// given a BalanceQuery, return a new query which only contains needed
// (account, asset, color) tuples (that is, the ones that aren't already cached)
func (b InternalBalances) filterQuery(q BalanceQuery) BalanceQuery {
	filteredQuery := BalanceQuery{}
	for _, item := range q {
		if !b.has(item.Account, item.Asset, item.Color) {
			filteredQuery = append(filteredQuery, item)
		}
	}
	return filteredQuery
}

// Merge the queried balance rows into the cache
func (b InternalBalances) Merge(update []BalanceRow) {
	for _, row := range update {
		b.Set(row.Account, row.Asset, row.Color, row.Amount)
	}
}
