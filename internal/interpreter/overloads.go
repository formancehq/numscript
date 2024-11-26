package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
)

type opAdd interface {
	evalAdd(st *programState, other parser.ValueExpr) (Value, InterpreterError)
}

var _ opAdd = (*MonetaryInt)(nil)
var _ opAdd = (*Monetary)(nil)

func (m MonetaryInt) evalAdd(st *programState, other parser.ValueExpr) (Value, InterpreterError) {
	m1 := big.Int(m)
	m2, err := evaluateExprAs(st, other, expectNumber)
	if err != nil {
		return nil, err
	}

	sum := new(big.Int).Add(&m1, m2)
	return MonetaryInt(*sum), nil
}

func (m Monetary) evalAdd(st *programState, other parser.ValueExpr) (Value, InterpreterError) {
	b2, err := evaluateExprAs(st, other, expectMonetaryOfAsset(string(m.Asset)))
	if err != nil {
		return nil, err
	}
	b1 := big.Int(m.Amount)
	sum := new(big.Int).Add(&b1, b2)
	return Monetary{
		Asset:  m.Asset,
		Amount: MonetaryInt(*sum),
	}, nil

}

type opSub interface {
	evalSub(st *programState, other parser.ValueExpr) (Value, InterpreterError)
}

var _ opSub = (*MonetaryInt)(nil)
var _ opSub = (*Monetary)(nil)

func (m MonetaryInt) evalSub(st *programState, other parser.ValueExpr) (Value, InterpreterError) {
	m1 := big.Int(m)
	m2, err := evaluateExprAs(st, other, expectNumber)
	if err != nil {
		return nil, err
	}
	sum := new(big.Int).Sub(&m1, m2)
	return MonetaryInt(*sum), nil
}

func (m Monetary) evalSub(st *programState, other parser.ValueExpr) (Value, InterpreterError) {
	m2, err := evaluateExprAs(st, other, expectMonetary)
	if err != nil {
		return nil, err
	}

	if m.Asset != m2.Asset {
		return nil, MismatchedCurrencyError{
			Expected: m.Asset.String(),
			Got:      m2.Asset.String(),
		}
	}

	return Monetary{
		Asset:  m.Asset,
		Amount: m.Amount.Sub(m2.Amount),
	}, nil

}

type opCmp interface {
	evalCmp(st *programState, other parser.ValueExpr) (*int, InterpreterError)
}

var _ opCmp = (*MonetaryInt)(nil)
var _ opCmp = (*Monetary)(nil)

func (m MonetaryInt) evalCmp(st *programState, other parser.ValueExpr) (*int, InterpreterError) {
	b2, err := evaluateExprAs(st, other, expectNumber)
	if err != nil {
		return nil, err
	}

	b1 := big.Int(m)

	cmp := b1.Cmp(b2)
	return &cmp, nil
}

func (m Monetary) evalCmp(st *programState, other parser.ValueExpr) (*int, InterpreterError) {
	b2, err := evaluateExprAs(st, other, expectMonetaryOfAsset(string(m.Asset)))
	if err != nil {
		return nil, err
	}

	b1 := big.Int(m.Amount)

	cmp := b1.Cmp(b2)
	return &cmp, nil
}

func (st *programState) evaluateExprAsCmp(expr parser.ValueExpr) (*opCmp, InterpreterError) {
	exprCmp, err := evaluateExprAs(st, expr, expectOneOf(
		expectMapped(expectMonetary, func(m Monetary) opCmp {
			return m
		}),

		// while "x.map(identity)" is the same as "x", just writing "expectNumber" would't typecheck
		expectMapped(expectNumber, func(bi big.Int) opCmp {
			return MonetaryInt(bi)
		}),
	))

	if err != nil {
		return nil, err
	}

	return exprCmp, nil
}
