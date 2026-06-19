package interpreter

import (
	"slices"

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

func (m AccountsMetadata) PrettyPrint() string {
	header := []string{"Account", "Name", "Value"}

	var rows [][]string
	for _, row := range m {
		rows = append(rows, []string{row.Account, row.Key, row.Value})
	}

	return utils.CsvPretty(header, rows, true)
}

// CompareAccountsMetadata reports whether two metadata lists hold the same set
// of rows, ignoring order.
func CompareAccountsMetadata(a AccountsMetadata, b AccountsMetadata) bool {
	if len(a) != len(b) {
		return false
	}
	for _, row := range a {
		if !slices.Contains(b, row) {
			return false
		}
	}
	return true
}
