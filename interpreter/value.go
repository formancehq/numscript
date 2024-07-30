package interpreter

import (
	"fmt"
	"math/big"
	"numscript/analysis"
	"numscript/parser"
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
func (Monetary) String() string         { panic("TODO impl") }
func (Portion) String() string          { panic("TODO impl") }
func (Asset) String() string            { panic("TODO impl") }

func expectMonetary(v Value) (*Monetary, error) {
	switch v := v.(type) {
	case Monetary:
		return &v, nil

	default:
		return nil, TypeError{Expected: analysis.TypeMonetary, Value: v}
	}
}

func expectNumber(v Value) (*MonetaryInt, error) {
	switch v := v.(type) {
	case MonetaryInt:
		return &v, nil

	default:
		return nil, TypeError{Expected: analysis.TypeNumber, Value: v}
	}
}

func expectString(v Value) (*String, error) {
	switch v := v.(type) {
	case String:
		return &v, nil

	default:
		return nil, TypeError{Expected: analysis.TypeString, Value: v}
	}
}

func expectAsset(v Value) (*Asset, error) {
	switch v := v.(type) {
	case Asset:
		return &v, nil

	default:
		return nil, TypeError{Expected: analysis.TypeAsset, Value: v}
	}
}

func expectAccount(v Value) (*AccountAddress, error) {
	switch v := v.(type) {
	case AccountAddress:
		return &v, nil

	default:
		return nil, TypeError{Expected: analysis.TypeAccount, Value: v}
	}
}

// TODO delete the following:

func expectAccountLit(literal parser.Literal, vars map[string]Value) (AccountAddress, error) {
	switch literal := literal.(type) {
	case *parser.AccountLiteral:
		return AccountAddress(literal.Name), nil

	case *parser.VariableLiteral:
		value, found := vars[literal.Name]
		if !found {
			panic("var not found")
		}

		account, ok := value.(AccountAddress)
		if !ok {
			fmt.Printf("VALUE =%#v\n", value)
			panic("TODO wrong type for var: " + literal.Name)
		}
		return account, nil

	default:
		panic("TODO invalid type (expected asset)")
	}
}

func expectPortionLit(literal parser.Literal, vars map[string]Value) (Portion, error) {
	switch literal := literal.(type) {
	case *parser.RatioLiteral:
		return Portion(*literal.ToRatio()), nil

	case *parser.VariableLiteral:
		portion, ok := vars[literal.Name].(Portion)
		if !ok {
			panic("TODO ret err")
		}
		return portion, nil

	default:
		panic("TODO invalid type (expected asset)")
	}
}

func NewMonetaryInt(n int64) MonetaryInt {
	bi := big.NewInt(n)
	return MonetaryInt(*bi)
}
