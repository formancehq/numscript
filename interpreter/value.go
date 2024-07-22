package interpreter

import (
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

func expectMonetary(literal parser.Literal) (Monetary, error) {
	switch literal := literal.(type) {
	case *parser.MonetaryLiteral:
		asset, err := expectAsset(literal.Asset)
		if err != nil {
			return Monetary{}, err
		}

		amt, err := expectAmount(literal.Amount)
		if err != nil {
			return Monetary{}, err
		}

		return Monetary{Asset: asset, Amount: amt}, nil

	default:
		panic("TODO invalid type")
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

func expectAsset(literal parser.Literal) (Asset, error) {
	switch literal := literal.(type) {
	case *parser.AssetLiteral:
		return Asset(literal.Asset), nil

	default:
		panic("TODO invalid type")
	}
}
