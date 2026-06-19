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
	// the Scope column is shown only when at least one entry has a scope
	hasScope := slices.ContainsFunc(m, func(row AccountMetadataRow) bool {
		return row.Scope != ""
	})

	var header []string
	if hasScope {
		header = []string{"Account", "Scope", "Name", "Value"}
	} else {
		header = []string{"Account", "Name", "Value"}
	}

	var rows [][]string
	for _, row := range m {
		if hasScope {
			rows = append(rows, []string{row.Account, row.Scope, row.Key, row.Value})
		} else {
			rows = append(rows, []string{row.Account, row.Key, row.Value})
		}
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
