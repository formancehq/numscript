package interpreter

import (
	"fmt"
	"math/big"
	"numscript/analysis"
)

type Value interface {
	value()
	String() string
}

type String string
type Asset string
type Portion big.Rat
type AccountAddress string
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

func (v MonetaryInt) MarshalJSON() ([]byte, error) {
	bigInt := big.Int(v)
	s := fmt.Sprintf(`"%s"`, bigInt.String())
	return []byte(s), nil
}

func (v Portion) MarshalJSON() ([]byte, error) {
	r := big.Rat(v)
	s := fmt.Sprintf(`"%s"`, r.String())
	return []byte(s), nil
}

func (v Monetary) MarshalJSON() ([]byte, error) {
	m := fmt.Sprintf("\"%s %s\"", v.Asset, v.Amount.String())
	return []byte(m), nil
}

func (v String) String() string {
	return string(v)
}

func (v AccountAddress) String() string {
	return string(v)
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

func expectMonetary(v Value) (*Monetary, error) {
	switch v := v.(type) {
	case Monetary:
		return &v, nil

	default:
		return nil, TypeError{Expected: analysis.TypeMonetary, Value: v}
	}
}

func expectMonetaryOfAsset(expectedAsset string) func(v Value) (*big.Int, error) {
	return func(v Value) (*big.Int, error) {
		m, err := expectMonetary(v)
		if err != nil {
			return nil, err
		}

		asset := string(m.Asset)

		if asset != expectedAsset {
			return nil, MismatchedCurrencyError{Expected: expectedAsset, Got: asset}
		}

		i := big.Int(m.Amount)
		return &i, nil
	}
}

func expectNumber(v Value) (*big.Int, error) {
	switch v := v.(type) {
	case MonetaryInt:
		return (*big.Int)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypeNumber, Value: v}
	}
}

func expectString(v Value) (*string, error) {
	switch v := v.(type) {
	case String:
		return (*string)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypeString, Value: v}
	}
}

func expectAsset(v Value) (*string, error) {
	switch v := v.(type) {
	case Asset:
		return (*string)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypeAsset, Value: v}
	}
}

func expectAccount(v Value) (*string, error) {
	switch v := v.(type) {
	case AccountAddress:
		return (*string)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypeAccount, Value: v}
	}
}

func expectPortion(v Value) (*big.Rat, error) {
	switch v := v.(type) {
	case Portion:
		return (*big.Rat)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypePortion, Value: v}
	}
}

func expectAnything(v Value) (*Value, error) {
	return &v, nil
}

func NewMonetaryInt(n int64) MonetaryInt {
	bi := big.NewInt(n)
	return MonetaryInt(*bi)
}
