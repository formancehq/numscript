package interpreter

// metadataKey identifies a single account-metadata entry in the in-memory cache.
type metadataKey struct {
	Account string
	Scope   string
	Key     string
}

// InternalAccountsMetadata is the read-side in-memory cache of the (opaque,
// string-valued) input account metadata. Whereas the external representation
// (interpreter.AccountsMetadata) is the user-facing serialized contract, this
// one is used internally by the runtime and may change over time. The set/output
// side is handled separately by internalSetAccountsMeta (typed values).
type InternalAccountsMetadata map[metadataKey]string

// FromAccountsMetadataRows builds the read cache from the external rows.
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
