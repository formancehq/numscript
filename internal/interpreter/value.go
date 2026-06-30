package interpreter

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
)

type Value interface {
	value()
	String() string
}

type String string
type Asset string
type Portion big.Rat

// AccountAddress is an account, optionally partitioned by a scope. The scope is
// a separate dimension of the account rather than part of its name, so it is
// modeled as its own field instead of being encoded into the name string.
type AccountAddress struct {
	Name  string
	Scope string
}

type MonetaryInt big.Int
type Monetary struct {
	Amount MonetaryInt
	Asset  Asset
}

func (String) value()         {}
func (AccountAddress) value() {}
func (MonetaryInt) value()    {}
func (Monetary) value()       {}
func (Portion) value()        {}
func (Asset) value()          {}

func NewAccountAddress(src string) (AccountAddress, InterpreterError) {
	if !checkAccountName(src) {
		return AccountAddress{}, InvalidAccountName{Name: src}
	}
	return AccountAddress{Name: src}, nil
}

func NewAsset(src string) (Asset, InterpreterError) {
	if !checkAssetName(src) {
		return Asset(""), InvalidAsset{Name: src}
	}
	return Asset(src), nil
}

// A Value is (de)serialized as a tagged-JSON discriminated union, keyed by
// "type", so the on-wire form is type-explicit and unambiguous (e.g. the string
// "42" and the number 42 are distinguishable), rather than stringly-typed:
//
//	string   -> { "type": "string",   "value": "abc" }
//	number   -> { "type": "number",   "value": "42" }
//	asset    -> { "type": "asset",    "name": "COIN" }
//	account  -> { "type": "account",  "name": "x", "scope": "s" }   // scope optional
//	monetary -> { "type": "monetary", "asset": "COIN", "amount": "100" }
//	portion  -> { "type": "portion",  "value": "1/2" }
const (
	valueTypeString   = "string"
	valueTypeNumber   = "number"
	valueTypeAsset    = "asset"
	valueTypeAccount  = "account"
	valueTypeMonetary = "monetary"
	valueTypePortion  = "portion"
)

// The per-shape tagged-JSON structs below are each shared by their type's
// MarshalJSON and by ParseTaggedValue, so the layout is defined once.
//
//	scalar (string/number/portion) -> { "type": ..., "value": "..." }
//	asset                          -> { "type": "asset",    "name": "COIN" }
//	account                        -> { "type": "account",  "name": "x", "scope": "s" }
//	monetary                       -> { "type": "monetary", "asset": "COIN", "amount": "100" }
type (
	taggedScalar struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	taggedAsset struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	taggedAccount struct {
		Type  string `json:"type"`
		Name  string `json:"name"`
		Scope string `json:"scope,omitempty"`
	}
	taggedMonetary struct {
		Type   string `json:"type"`
		Asset  string `json:"asset"`
		Amount string `json:"amount"`
	}
)

// ParseTaggedValue decodes the tagged-JSON representation of a Value. It reads
// the "type" discriminator (json can't unmarshal directly into the Value
// interface), then decodes into the struct shared with that type's MarshalJSON.
func ParseTaggedValue(data []byte) (Value, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &tag); err != nil {
		return nil, err
	}

	switch tag.Type {
	case valueTypeString:
		var v taggedScalar
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return String(v.Value), nil

	case valueTypeAccount:
		var v taggedAccount
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return AccountAddress{Name: v.Name, Scope: v.Scope}, nil

	case valueTypeAsset:
		var v taggedAsset
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return Asset(v.Name), nil

	case valueTypeNumber:
		var v taggedScalar
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		n, ok := new(big.Int).SetString(v.Value, 10)
		if !ok {
			return nil, fmt.Errorf("invalid number value: %q", v.Value)
		}
		return MonetaryInt(*n), nil

	case valueTypeMonetary:
		var v taggedMonetary
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		n, ok := new(big.Int).SetString(v.Amount, 10)
		if !ok {
			return nil, fmt.Errorf("invalid monetary amount: %q", v.Amount)
		}
		return Monetary{Asset: Asset(v.Asset), Amount: MonetaryInt(*n)}, nil

	case valueTypePortion:
		var v taggedScalar
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		r, ok := new(big.Rat).SetString(v.Value)
		if !ok {
			return nil, fmt.Errorf("invalid portion value: %q", v.Value)
		}
		return Portion(*r), nil

	case "":
		return nil, fmt.Errorf("missing value type")
	default:
		return nil, fmt.Errorf("unknown value type: %q", tag.Type)
	}
}

func (v String) MarshalJSON() ([]byte, error) {
	return json.Marshal(taggedScalar{valueTypeString, string(v)})
}

func (v Asset) MarshalJSON() ([]byte, error) {
	return json.Marshal(taggedAsset{valueTypeAsset, string(v)})
}

func (v AccountAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(taggedAccount{valueTypeAccount, v.Name, v.Scope})
}

func (v MonetaryInt) MarshalJSON() ([]byte, error) {
	bi := big.Int(v)
	return json.Marshal(taggedScalar{valueTypeNumber, bi.String()})
}

func (v Portion) MarshalJSON() ([]byte, error) {
	r := big.Rat(v)
	return json.Marshal(taggedScalar{valueTypePortion, r.String()})
}

func (v Monetary) MarshalJSON() ([]byte, error) {
	return json.Marshal(taggedMonetary{valueTypeMonetary, string(v.Asset), v.Amount.String()})
}

func (v String) String() string {
	return fmt.Sprintf(`"%s"`, string(v))
}

func (v AccountAddress) String() string {
	if v.Scope == "" {
		return fmt.Sprintf(`@%s`, v.Name)
	}
	return fmt.Sprintf(`scoped(%s, "%s")`, v.Name, v.Scope)
}

func (v MonetaryInt) String() string {
	i := big.Int(v)
	return i.String()
}

func (v Monetary) String() string {
	return fmt.Sprintf("%s %s", v.Asset, v.Amount)
}

func (p Portion) String() string {
	r := big.Rat(p)
	return r.String()
}

func (v Asset) String() string {
	return string(v)
}

func expectMonetary(v Value, r parser.Range) (Monetary, InterpreterError) {
	switch v := v.(type) {
	case Monetary:
		return v, nil

	default:
		return Monetary{}, TypeError{Expected: analysis.TypeMonetary, Value: v, Range: r}
	}
}

func expectMonetaryOfAsset(expectedAsset Asset) func(Value, parser.Range) (MonetaryInt, InterpreterError) {
	return func(v Value, r parser.Range) (MonetaryInt, InterpreterError) {
		m, err := expectMonetary(v, r)
		if err != nil {
			return MonetaryInt{}, err
		}

		if m.Asset != expectedAsset {
			return MonetaryInt{}, MismatchedCurrencyError{Expected: string(expectedAsset), Got: string(m.Asset)}
		}

		return m.Amount, nil
	}
}

func expectNumber(v Value, r parser.Range) (MonetaryInt, InterpreterError) {
	switch v := v.(type) {
	case MonetaryInt:
		return v, nil

	default:
		return MonetaryInt{}, TypeError{Expected: analysis.TypeNumber, Value: v, Range: r}
	}
}

func expectString(v Value, r parser.Range) (String, InterpreterError) {
	switch v := v.(type) {
	case String:
		return v, nil

	default:
		return "", TypeError{Expected: analysis.TypeString, Value: v, Range: r}
	}
}

func expectAsset(v Value, r parser.Range) (Asset, InterpreterError) {
	switch v := v.(type) {
	case Asset:
		return v, nil

	default:
		return "", TypeError{Expected: analysis.TypeAsset, Value: v, Range: r}
	}
}

func expectAccount(v Value, r parser.Range) (AccountAddress, InterpreterError) {
	switch v := v.(type) {
	case AccountAddress:
		return v, nil

	default:
		return AccountAddress{}, TypeError{Expected: analysis.TypeAccount, Value: v, Range: r}
	}
}

func expectPortion(v Value, r parser.Range) (*big.Rat, InterpreterError) {
	switch v := v.(type) {
	case Portion:
		return (*big.Rat)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypePortion, Value: v, Range: r}
	}
}

func expectAnything(v Value, _ parser.Range) (Value, InterpreterError) {
	return v, nil
}

func expectOneOf[T any](combinators ...func(v Value, r parser.Range) (T, InterpreterError)) func(v Value, r parser.Range) (T, InterpreterError) {
	return func(v Value, r parser.Range) (T, InterpreterError) {
		if len(combinators) == 0 {
			// this should be unreachable
			panic("Invalid argument: no combinators given")
		}

		var errs []TypeError
		for _, combinator := range combinators {
			out, err := combinator(v, r)
			if err == nil {
				return out, nil
			}

			typeErr, ok := err.(TypeError)
			if !ok {
				var default_ T
				return default_, err
			}
			errs = append(errs, typeErr)
		}

		// e.g. typeErr.map(e => e.Expected).join("|")
		expected := ""
		for index, typeErr := range errs {
			if index != 0 {
				expected += "|"
			}
			expected += typeErr.Expected
		}

		var default_ T
		return default_, TypeError{
			Range:    r,
			Value:    v,
			Expected: expected,
		}
	}
}

func expectMapped[T any, U any](
	combinator func(v Value, r parser.Range) (T, InterpreterError),
	mapper func(value T) U,
) func(v Value, r parser.Range) (U, InterpreterError) {
	return func(v Value, r parser.Range) (U, InterpreterError) {
		out, err := combinator(v, r)
		if err != nil {
			var default_ U
			return default_, err
		}
		mapped := mapper(out)
		return mapped, nil
	}
}

func NewMonetary(asset string, n int64) Monetary {
	return Monetary{
		Asset:  Asset(asset),
		Amount: NewMonetaryInt(n),
	}
}

func NewMonetaryIntBig(n *big.Int) MonetaryInt {
	bi := new(big.Int).Set(n)
	return MonetaryInt(*bi)
}

func NewMonetaryInt(n int64) MonetaryInt {
	bi := big.NewInt(n)
	return MonetaryInt(*bi)
}

func (m MonetaryInt) Add(other MonetaryInt) MonetaryInt {
	bi := big.Int(m)
	otherBi := big.Int(other)

	sum := new(big.Int).Add(&bi, &otherBi)
	return MonetaryInt(*sum)
}

func (m MonetaryInt) Sub(other MonetaryInt) MonetaryInt {
	bi := big.Int(m)
	otherBi := big.Int(other)

	sum := new(big.Int).Sub(&bi, &otherBi)
	return MonetaryInt(*sum)
}

func (asset Asset) GetBaseAndScale() (string, int64) {
	parts := strings.Split(string(asset), "/")
	if len(parts) == 2 {
		scale, err := strconv.ParseInt(parts[1], 10, 64)
		if err == nil {
			return parts[0], scale
		}
		// fallback if parsing fails
		return parts[0], 0
	}
	return string(asset), 0

}
