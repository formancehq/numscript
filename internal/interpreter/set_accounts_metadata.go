package interpreter

import (
	"encoding/json"
	"sort"
)

// SetAccountMetadataRow is a single piece of account metadata set by the script
// during execution. Unlike the input metadata (which is opaque and string-valued,
// since its serialization format isn't always known), the set value's type is
// known, so it is carried as a typed Value and serialized in the tagged form.
type SetAccountMetadataRow struct {
	Account string `json:"account"`
	Key     string `json:"key"`
	Value   Value  `json:"value"`
	Scope   string `json:"scope,omitempty"`
}

// SetAccountsMetadata is the account metadata produced by the script (the
// execution result's accountsMeta, and a spec's expect.metadata).
type SetAccountsMetadata []SetAccountMetadataRow

func (r *SetAccountMetadataRow) UnmarshalJSON(data []byte) error {
	var raw struct {
		Account string          `json:"account"`
		Key     string          `json:"key"`
		Value   json.RawMessage `json:"value"`
		Scope   string          `json:"scope"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	value, err := ParseTaggedValue(raw.Value)
	if err != nil {
		return err
	}
	r.Account, r.Key, r.Scope, r.Value = raw.Account, raw.Key, raw.Scope, value
	return nil
}

// CompareSetAccountsMetadata reports whether two lists hold the same rows,
// ignoring order but respecting multiplicity (so [x, x] != [x, y]). Values are
// compared on their canonical source form.
func CompareSetAccountsMetadata(a SetAccountsMetadata, b SetAccountsMetadata) bool {
	if len(a) != len(b) {
		return false
	}
	key := func(r SetAccountMetadataRow) string {
		value := ""
		if r.Value != nil {
			value = r.Value.String()
		}
		return r.Account + "\x00" + r.Key + "\x00" + r.Scope + "\x00" + value
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
// (account, scope, key) for deterministic output.
func (m internalSetAccountsMeta) toRows() SetAccountsMetadata {
	rows := make(SetAccountsMetadata, 0, len(m))
	for k, value := range m {
		rows = append(rows, SetAccountMetadataRow{
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
