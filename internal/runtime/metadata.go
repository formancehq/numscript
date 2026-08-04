package runtime

import (
	"github.com/formancehq/numscript/internal/utils"
)

// AccountMetadataEntry is one piece of account metadata set during execution.
// Scope is a second dimension of the account, like the balance PairKey: a flat
// list rather than a nested map, since JSON object keys must be strings and
// folding scope into the account key would need an encoding hack.
type AccountMetadataEntry struct {
	Account string `json:"account"`
	Scope   string `json:"scope,omitempty"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

// AccountsMetadata is the account metadata produced by the script (the
// execution result's accountsMeta).
type AccountsMetadata []AccountMetadataEntry

func (m AccountsMetadata) PrettyPrint() string {
	header := []string{"Account", "Scope", "Key", "Value"}

	rows := make([][]string, 0, len(m))
	for _, e := range m {
		rows = append(rows, []string{e.Account, e.Scope, e.Key, e.Value})
	}

	return utils.CsvPretty(header, rows, true)
}
