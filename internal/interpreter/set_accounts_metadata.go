package interpreter

import (
	"sort"
)

// SetAccountMetadataRow is a single piece of account metadata set by the script
// during execution. The value is carried as the rendered text produced by
// MetaString, matching the input metadata: the wire format is untyped, so a
// value's type is not part of what gets stored or asserted.
type SetAccountMetadataRow struct {
	Account string `json:"account"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Scope   string `json:"scope,omitempty"`
}

// SetAccountsMetadata is the account metadata produced by the script (the
// execution result's accountsMeta, and a spec's expect.metadata).
type SetAccountsMetadata []SetAccountMetadataRow

// CompareSetAccountsMetadata reports whether two lists hold the same rows,
// ignoring order but respecting multiplicity (so [x, x] != [x, y]).
func CompareSetAccountsMetadata(a SetAccountsMetadata, b SetAccountsMetadata) bool {
	if len(a) != len(b) {
		return false
	}
	key := func(r SetAccountMetadataRow) string {
		return r.Account + "\x00" + r.Key + "\x00" + r.Scope + "\x00" + r.Value
	}
	counts := make(map[string]int, len(a))
	for _, r := range a {
		counts[key(r)]++
	}
	for _, r := range b {
		k := key(r)
		counts[k]--
		if counts[k] < 0 {
			return false
		}
	}
	return true
}

// internalSetAccountsMeta is the in-memory store of metadata set during
// execution, keyed for upserts.
type internalSetAccountsMeta map[metadataKey]Value

func (m internalSetAccountsMeta) Set(account, scope, key string, value Value) {
	m[metadataKey{Account: account, Scope: scope, Key: key}] = value
}

// toRows flattens the set metadata into the external representation, sorted by
// (account, scope, key) for deterministic output. Values are rendered here: the
// runtime keeps them typed, the wire format does not.
func (m internalSetAccountsMeta) toRows() SetAccountsMetadata {
	rows := make(SetAccountsMetadata, 0, len(m))
	for k, value := range m {
		rows = append(rows, SetAccountMetadataRow{
			Account: k.Account,
			Scope:   k.Scope,
			Key:     k.Key,
			Value:   MetaString(value),
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
