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

func (v String) String() string         { return string(v) }
func (v AccountAddress) String() string { return string(v) }
func (v MonetaryInt) String() string    { i := big.Int(v); return i.String() }
func (v Monetary) String() string       { return fmt.Sprintf("%s %s", v.Asset, v.Amount) }
func (Portion) String() string          { panic("TODO impl") }
func (v Asset) String() string          { return string(v) }

func expectMonetary(v Value) (*Monetary, error) {
	switch v := v.(type) {
	case Monetary:
		return &v, nil

	default:
		return nil, TypeError{Expected: analysis.TypeMonetary, Value: v}
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

func NewMonetaryInt(n int64) MonetaryInt {
	bi := big.NewInt(n)
	return MonetaryInt(*bi)
}
