package interpreter

// metadataKey identifies a single account-metadata entry. Used as the key of
// the write-side internalSetAccountsMeta.
type metadataKey struct {
	Account string
	Scope   string
	Key     string
}

// An internal representation of the account metadata. Used to cache metadata we get from external store.
// Whereas the external representation (interpreter.AccountsMetadata) is user-facing and a stable contract,
// this one is used internally by the runtime, and could change over time, for example to add more indexes for faster lookups.
// It is keyed by the (account, scope) pair, holding that account's (key -> value) entries.
type InternalAccountsMetadata map[AccountAddress]map[string]string

// Get the (account, key) metadata value from the cache.
func (m InternalAccountsMetadata) Get(account AccountAddress, key string) (string, bool) {
	value, ok := m[account][key]
	return value, ok
}

func (m InternalAccountsMetadata) has(account AccountAddress, key string) bool {
	_, ok := m[account][key]
	return ok
}

// Set assigns value to the (account, key) metadata entry.
func (m InternalAccountsMetadata) Set(account AccountAddress, key string, value string) {
	entries := m[account]
	if entries == nil {
		entries = map[string]string{}
		m[account] = entries
	}
	entries[key] = value
}

// Merge the queried metadata rows into the cache.
func (m InternalAccountsMetadata) Merge(rows AccountsMetadata) {
	for _, row := range rows {
		m.Set(AccountAddress{Name: row.Account, Scope: row.Scope}, row.Key, row.Value)
	}
}
