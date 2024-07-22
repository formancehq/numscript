package interpreter

import (
	"fmt"
	"math/big"
	"numscript/parser"
)

type Value interface{ value() }

type String string
type Asset string
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
func (Asset) value()          {}

func expectMonetary(literal parser.Literal, vars map[string]Value) (Monetary, error) {
	switch literal := literal.(type) {
	case *parser.MonetaryLiteral:
		asset, err := expectAsset(literal.Asset, vars)
		if err != nil {
			return Monetary{}, err
		}

		amt, err := expectAmount(literal.Amount)
		if err != nil {
			return Monetary{}, err
		}

		return Monetary{Asset: asset, Amount: amt}, nil

	case *parser.VariableLiteral:
		v := vars[literal.Name]
		fmt.Printf("V: %s\n", v)
		panic("TODO parse var lit")

	default:
		panic("TODO invalid type (expected monetary)")
	}
}

func expectAmount(literal parser.Literal) (MonetaryInt, error) {
	switch literal := literal.(type) {
	case *parser.NumberLiteral:
		return MonetaryInt(*big.NewInt(int64(literal.Number))), nil

	default:
		panic("TODO invalid type")
	}
}

func expectAsset(literal parser.Literal, vars map[string]Value) (Asset, error) {
	switch literal := literal.(type) {
	case *parser.AssetLiteral:
		return Asset(literal.Asset), nil

	case *parser.VariableLiteral:
		asset, ok := vars[literal.Name].(Asset)
		if !ok {
			panic("TODO ret err")
		}
		return asset, nil

	default:
		panic("TODO invalid type (expected asset)")
	}
}

func expectAccount(literal parser.Literal, vars map[string]Value) (AccountAddress, error) {
	switch literal := literal.(type) {
	case *parser.AccountLiteral:
		return AccountAddress(literal.Name), nil

	case *parser.VariableLiteral:
		account, ok := vars[literal.Name].(AccountAddress)
		if !ok {
			panic("TODO ret err")
		}
		return account, nil

	default:
		panic("TODO invalid type (expected asset)")
	}
}

func expectString(literal parser.Literal, vars map[string]Value) (String, error) {
	switch literal := literal.(type) {
	case *parser.StringLiteral:
		return String(literal.String), nil

	case *parser.VariableLiteral:
		account, ok := vars[literal.Name].(String)
		if !ok {
			panic("TODO ret err")
		}
		return account, nil

	default:
		panic("TODO invalid type (expected asset)")
	}
}

func expectAnything(literal parser.Literal, vars map[string]Value) (Value, error) {
	switch literal := literal.(type) {
	case *parser.VariableLiteral:
		value, ok := vars[literal.Name]
		if !ok {
			panic("TODO ret err")
		}
		return value, nil
	case *parser.StringLiteral:
		return String(literal.String), nil

	default:
		panic("TODO invalid type (expected asset)")
	}
}

func NewMonetaryInt(n int64) MonetaryInt {
	bi := big.NewInt(n)
	return MonetaryInt(*bi)
}
