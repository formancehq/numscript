package interpreter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
)

// Balances JSON shape.
//
// Canonical write form per (account, asset) entry:
//
//   - a bare integer when only the uncolored bucket is set:
//     "USD/2": 100
//
//   - an array of value-objects when one or more colors are present:
//     "USD/2": [
//     { "amount": 100 },                // uncolored — no "color" field
//     { "color": "RED", "amount": 50 }  // colored
//     ]
//
// Tolerant read accepts three forms per (account, asset) entry:
//  1. JSON number  → single uncolored bucket
//  2. JSON object  → single value-object with optional "color" and required "amount"
//  3. JSON array   → list of value-objects (canonical multi-color form)
//
// Adding orthogonal dimensions in the future (scopes, scales, …) is done by
// extending the value-object schema rather than adding a new nesting level.

type balanceEntry struct {
	Color  string   `json:"color,omitempty"`
	Amount *big.Int `json:"amount"`
}

func (b Balances) MarshalJSON() ([]byte, error) {
	type assetWire = json.RawMessage

	out := make(map[string]map[string]assetWire, len(b))
	for account, accBalances := range b {
		assets := make(map[string]assetWire, len(accBalances))
		for asset, colorMap := range accBalances {
			raw, err := marshalColorBalance(colorMap)
			if err != nil {
				return nil, fmt.Errorf("balances[%q][%q]: %w", account, asset, err)
			}
			assets[asset] = raw
		}
		out[account] = assets
	}
	return json.Marshal(out)
}

func marshalColorBalance(cb ColorBalance) (json.RawMessage, error) {
	if len(cb) == 1 {
		if amount, ok := cb[""]; ok {
			return json.Marshal(amount)
		}
	}

	colors := make([]string, 0, len(cb))
	for c := range cb {
		colors = append(colors, c)
	}
	sort.Strings(colors)

	entries := make([]balanceEntry, len(colors))
	for i, c := range colors {
		entries[i] = balanceEntry{Color: c, Amount: cb[c]}
	}
	return json.Marshal(entries)
}

func (b *Balances) UnmarshalJSON(data []byte) error {
	var raw map[string]map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	out := make(Balances, len(raw))
	for account, assets := range raw {
		accBalance := make(AccountBalance, len(assets))
		for asset, rawValue := range assets {
			cb, err := unmarshalColorBalance(rawValue)
			if err != nil {
				return fmt.Errorf("balances[%q][%q]: %w", account, asset, err)
			}
			accBalance[asset] = cb
		}
		out[account] = accBalance
	}
	*b = out
	return nil
}

func unmarshalColorBalance(data []byte) (ColorBalance, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty balance entry")
	}

	switch trimmed[0] {
	case '[':
		var entries []balanceEntry
		dec := json.NewDecoder(bytes.NewReader(trimmed))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&entries); err != nil {
			return nil, err
		}
		cb := make(ColorBalance, len(entries))
		for i, e := range entries {
			if e.Amount == nil {
				return nil, fmt.Errorf("entry %d: missing \"amount\"", i)
			}
			if _, exists := cb[e.Color]; exists {
				return nil, fmt.Errorf("entry %d: duplicate color %q", i, e.Color)
			}
			cb[e.Color] = e.Amount
		}
		return cb, nil

	case '{':
		var entry balanceEntry
		dec := json.NewDecoder(bytes.NewReader(trimmed))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&entry); err != nil {
			return nil, err
		}
		if entry.Amount == nil {
			return nil, fmt.Errorf("missing \"amount\" field")
		}
		return ColorBalance{entry.Color: entry.Amount}, nil

	default:
		amount := new(big.Int)
		if err := json.Unmarshal(trimmed, amount); err != nil {
			return nil, fmt.Errorf("expected number, value-object, or array; got %s", string(trimmed))
		}
		return ColorBalance{"": amount}, nil
	}
}
