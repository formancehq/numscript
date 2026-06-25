package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/utils"
)

type BalanceRow struct {
	Account string   `json:"account"`
	Asset   string   `json:"asset"`
	Amount  *big.Int `json:"amount"`
	Color   string   `json:"color,omitempty"`
	Scope   string   `json:"scope,omitempty"`
}
type Balances []BalanceRow

// FirstDuplicate returns the first row whose (account, asset, color) key already
// appeared earlier in the list, if any. That triple is the identity of a balance
// entry and the amount is its value, so a repeated key is an ambiguous,
// malformed input.
func (rows Balances) FirstDuplicate() (BalanceRow, bool) {
	seen := make(map[[4]string]struct{}, len(rows))
	for _, row := range rows {
		key := [4]string{row.Account, row.Asset, row.Color, row.Scope}
		if _, ok := seen[key]; ok {
			return row, true
		}
		seen[key] = struct{}{}
	}
	return BalanceRow{}, false
}

func (rows Balances) PrettyPrint() string {
	// the optional columns (scope, color) are dropped automatically when no entry
	// populates them
	header := []string{"Account", "Scope", "Asset", "Color", "Balance"}

	var tableRows [][]string
	for _, row := range rows {
		var amount string
		if row.Amount != nil {
			amount = row.Amount.String()
		}
		tableRows = append(tableRows, []string{row.Account, row.Scope, row.Asset, row.Color, amount})
	}
	return utils.CsvPrettyOmitEmptyCols(header, tableRows, true)
}

// findRow returns the amount for a given (account, asset, color, scope), if present.
func findRow(rows Balances, account, asset, color, scope string) (*big.Int, bool) {
	for i := range rows {
		if rows[i].Account == account && rows[i].Asset == asset && rows[i].Color == color && rows[i].Scope == scope {
			return rows[i].Amount, true
		}
	}
	return nil, false
}

// amountsEqual treats a nil *big.Int as zero, so it never panics.
func amountsEqual(a, b *big.Int) bool {
	if a == nil {
		a = new(big.Int)
	}
	if b == nil {
		b = new(big.Int)
	}
	return a.Cmp(b) == 0
}

func CompareBalances(b1 Balances, b2 Balances) bool {
	if len(b1) != len(b2) {
		return false
	}
	// multiset comparison, respecting multiplicity: a duplicated row in b1 must be
	// matched by the same number of occurrences in b2 (so [x, x] != [x, y]). A
	// plain subset check would wrongly report equality there.
	type rowKey struct{ account, asset, color, scope, amount string }
	mk := func(r BalanceRow) rowKey {
		amount := "0" // amountsEqual treats nil as zero
		if r.Amount != nil {
			amount = r.Amount.String()
		}
		return rowKey{r.Account, r.Asset, r.Color, r.Scope, amount}
	}
	counts := make(map[rowKey]int, len(b1))
	for _, r := range b1 {
		counts[mk(r)]++
	}
	for _, r := range b2 {
		k := mk(r)
		counts[k]--
		if counts[k] < 0 {
			return false
		}
	}
	return true
}

// Returns whether the first value is a subset of the second one.
func CompareBalancesIncluding(b1 Balances, b2 Balances) bool {
	for _, entry := range b1 {
		amount2, ok := findRow(b2, entry.Account, entry.Asset, entry.Color, entry.Scope)
		if !ok || !amountsEqual(entry.Amount, amount2) {
			return false
		}
	}
	return true
}
