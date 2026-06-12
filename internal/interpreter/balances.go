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
}
type Balances []BalanceRow

func (b Balances) PrettyPrint() string {
	// TODO show colors
	header := []string{"Account", "Asset", "Balance"}

	rows := make([][]string, 0, len(b))
	for _, entry := range b {
		amount := "0"
		if entry.Amount != nil {
			amount = entry.Amount.String()
		}
		rows = append(rows, []string{entry.Account, entry.Asset, amount})
	}
	return utils.CsvPretty(header, rows, true)
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
