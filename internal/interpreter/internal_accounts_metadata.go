package interpreter

import "sort"

// metadataKey identifies a single account-metadata entry in the in-memory cache.
type metadataKey struct {
	Account string
	Scope   string
	Key     string
}

// InternalAccountsMetadata is the in-memory representation of account metadata,
// keyed for O(1) lookups. Whereas the external representation
// (interpreter.AccountsMetadata) is the user-facing, serialized contract, this
// one is used internally by the runtime and may change over time.
type InternalAccountsMetadata map[metadataKey]string

// FromAccountsMetadataRows builds the in-memory cache from the external rows.
func FromAccountsMetadataRows(rows AccountsMetadata) InternalAccountsMetadata {
	out := make(InternalAccountsMetadata, len(rows))
	for _, row := range rows {
		out[metadataKey{Account: row.Account, Scope: row.Scope, Key: row.Key}] = row.Value
	}
	return out
}

// Get returns the value for a given (account, scope, key), if present.
func (m InternalAccountsMetadata) Get(account, scope, key string) (string, bool) {
	value, ok := m[metadataKey{Account: account, Scope: scope, Key: key}]
	return value, ok
}

// Set assigns the value for a given (account, scope, key).
func (m InternalAccountsMetadata) Set(account, scope, key, value string) {
	m[metadataKey{Account: account, Scope: scope, Key: key}] = value
}

// toRows flattens the cache back into the external representation, sorted by
// (account, scope, key) for deterministic output.
func (m InternalAccountsMetadata) toRows() AccountsMetadata {
	rows := make(AccountsMetadata, 0, len(m))
	for k, value := range m {
		rows = append(rows, AccountMetadataRow{
			Account: k.Account,
			Scope:   k.Scope,
			Key:     k.Key,
			Value:   value,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Account != rows[j].Account {
			return rows[i].Account < rows[j].Account
		}
		if rows[i].Scope != rows[j].Scope {
			return rows[i].Scope < rows[j].Scope
		}
		return rows[i].Key < rows[j].Key
	})
	return rows
}
