package interpreter

import (
	"github.com/formancehq/numscript/internal/utils"
)

type AccountMetadataRow struct {
	Account string `json:"account"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Scope   string `json:"scope,omitempty"`
}

// AccountsMetadata is the external, serialized representation of account
// metadata. The runtime works with the in-memory InternalAccountsMetadata and
// converts to this at the boundaries (store queries, execution result).
type AccountsMetadata []AccountMetadataRow

// FirstDuplicate returns the first row whose (account, key, scope) key already
// appeared earlier in the list, if any. That triple is the identity of a
// metadata entry and the value is its content, so a repeated key is an
// ambiguous, malformed input.
func (rows AccountsMetadata) FirstDuplicate() (AccountMetadataRow, bool) {
	seen := make(map[[3]string]struct{}, len(rows))
	for _, row := range rows {
		key := [3]string{row.Account, row.Key, row.Scope}
		if _, ok := seen[key]; ok {
			return row, true
		}
		seen[key] = struct{}{}
	}
	return AccountMetadataRow{}, false
}

func (m AccountsMetadata) PrettyPrint() string {
	// the Scope column is dropped automatically when no entry has a scope
	header := []string{"Account", "Scope", "Name", "Value"}

	var rows [][]string
	for _, row := range m {
		rows = append(rows, []string{row.Account, row.Scope, row.Key, row.Value})
	}

	return utils.CsvPrettyOmitEmptyCols(header, rows, true)
}

// CompareAccountsMetadata reports whether two metadata lists hold the same rows,
// ignoring order but respecting multiplicity. A duplicated row in one list must
// be matched by the same number of occurrences in the other, so e.g. [x, x] is
// not considered equal to [x, y].
func CompareAccountsMetadata(a AccountsMetadata, b AccountsMetadata) bool {
	if len(a) != len(b) {
		return false
	}
	// AccountMetadataRow is an all-string (comparable) struct, so it can key the
	// multiset directly.
	counts := make(map[AccountMetadataRow]int, len(a))
	for _, row := range a {
		counts[row]++
	}
	for _, row := range b {
		counts[row]--
		if counts[row] < 0 {
			return false
		}
	}
	// equal lengths + every b row consumed a distinct a row => exact multiset match
	return true
}
