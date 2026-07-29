package interpreter

import (
	"encoding/json"
	"sort"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
)

// MetaValueToValue rebuilds the typed Value behind a metadata entry the VM
// produced. The VM stores metadata stringified, plus the type it was stringified
// from; this is the inverse of that stringification, so a compiled run can report
// the same typed values a tree-walking run does.
func MetaValueToValue(mv runtime.MetaValue) (Value, InterpreterError) {
	var typ string
	switch mv.Typ {
	case runtime.MetaValueStr:
		typ = analysis.TypeString
	case runtime.MetaValueAccount:
		typ = analysis.TypeAccount
	case runtime.MetaValueAsset:
		typ = analysis.TypeAsset
	case runtime.MetaValueInt:
		typ = analysis.TypeNumber
	case runtime.MetaValuePortion:
		typ = analysis.TypePortion
	case runtime.MetaValueMonetary:
		typ = analysis.TypeMonetary
	default:
		return nil, InvalidTypeErr{Name: mv.Typ.String()}
	}
	return parseVar(typ, mv.Value, parser.Range{})
}

// MetadataFromVM converts the VM's transaction metadata into the typed Metadata
// the execution result exposes.
func MetadataFromVM(m map[string]runtime.MetaValue) (Metadata, InterpreterError) {
	if m == nil {
		return nil, nil
	}
	out := make(Metadata, len(m))
	for key, mv := range m {
		value, err := MetaValueToValue(mv)
		if err != nil {
			return nil, err
		}
		out[key] = value
	}
	return out, nil
}

// SetAccountsMetadataFromVM converts the VM's account metadata into the typed
// row form, sorted by (account, key) to match internalSetAccountsMeta.toRows.
// Scope is left empty: the VM has no scopes yet.
func SetAccountsMetadataFromVM(m runtime.AccountsMetadata) (SetAccountsMetadata, InterpreterError) {
	if m == nil {
		return nil, nil
	}
	rows := make(SetAccountsMetadata, 0, len(m))
	for account, accMeta := range m {
		for key, mv := range accMeta {
			value, err := MetaValueToValue(mv)
			if err != nil {
				return nil, err
			}
			rows = append(rows, SetAccountMetadataRow{Account: account, Key: key, Value: value})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Account != rows[j].Account {
			return rows[i].Account < rows[j].Account
		}
		return rows[i].Key < rows[j].Key
	})
	return rows, nil
}

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
