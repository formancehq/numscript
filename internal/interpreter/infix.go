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
		Amount: m.Amount.Add(m2.Amount),
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
