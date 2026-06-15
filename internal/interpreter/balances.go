package interpreter

import (
	"math/big"
	"slices"

	"github.com/formancehq/numscript/internal/utils"
)

type BalanceRow struct {
	Account string   `json:"account"`
	Asset   string   `json:"asset"`
	Amount  *big.Int `json:"amount"`
	Color   string   `json:"color,omitempty"`
}
type Balances []BalanceRow

// FirstDuplicate returns the first row whose (account, asset, color) key already
// appeared earlier in the list, if any. That triple is the identity of a balance
// entry and the amount is its value, so a repeated key is an ambiguous,
// malformed input.
func (rows Balances) FirstDuplicate() (BalanceRow, bool) {
	seen := make(map[[3]string]struct{}, len(rows))
	for _, row := range rows {
		key := [3]string{row.Account, row.Asset, row.Color}
		if _, ok := seen[key]; ok {
			return row, true
		}
		seen[key] = struct{}{}
	}
	return BalanceRow{}, false
}

func (rows Balances) PrettyPrint() string {
	// the Color column is shown only when at least one entry has a color
	hasColor := slices.ContainsFunc(rows, func(row BalanceRow) bool {
		return row.Color != ""
	})

	var header []string
	if hasColor {
		header = []string{"Account", "Asset", "Color", "Balance"}
	} else {
		header = []string{"Account", "Asset", "Balance"}
	}

	var tableRows [][]string
	for _, row := range rows {
		var amount string
		if row.Amount != nil {
			amount = row.Amount.String()
		}
		if hasColor {
			tableRows = append(tableRows, []string{row.Account, row.Asset, row.Color, amount})
		} else {
			tableRows = append(tableRows, []string{row.Account, row.Asset, amount})
		}
	}
	return utils.CsvPretty(header, tableRows, true)
}

// findRow returns the amount for a given (account, asset, color), if present.
func findRow(rows Balances, account, asset, color string) (*big.Int, bool) {
	for i := range rows {
		if rows[i].Account == account && rows[i].Asset == asset && rows[i].Color == color {
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
	return CompareBalancesIncluding(b1, b2)
}

// Returns whether the first value is a subset of the second one.
func CompareBalancesIncluding(b1 Balances, b2 Balances) bool {
	for _, entry := range b1 {
		amount2, ok := findRow(b2, entry.Account, entry.Asset, entry.Color)
		if !ok || !amountsEqual(entry.Amount, amount2) {
			return false
		}
	}
	return true
}
