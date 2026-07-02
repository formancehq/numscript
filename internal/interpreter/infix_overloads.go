package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
)

type opAdd interface {
	evalAdd(env expressionEnv, other parser.ValueExpr) (Value, InterpreterError)
}

var _ opAdd = (*MonetaryInt)(nil)
var _ opAdd = (*Monetary)(nil)

func (m MonetaryInt) evalAdd(env expressionEnv, other parser.ValueExpr) (Value, InterpreterError) {
	leftBi := big.Int(m)
	m2, err := evaluateExprAs(env, other, expectNumber)
	if err != nil {
		return nil, err
	}

	rightBi := big.Int(m2)
	sum := new(big.Int).Add(&leftBi, &rightBi)
	return MonetaryInt(*sum), nil
}

func (m Monetary) evalAdd(env expressionEnv, other parser.ValueExpr) (Value, InterpreterError) {
	m2, err := evaluateExprAs(env, other, expectMonetary)
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
	evalSub(env expressionEnv, other parser.ValueExpr) (Value, InterpreterError)
}

var _ opSub = (*MonetaryInt)(nil)
var _ opSub = (*Monetary)(nil)

func (m MonetaryInt) evalSub(env expressionEnv, other parser.ValueExpr) (Value, InterpreterError) {
	leftBi := big.Int(m)
	m2, err := evaluateExprAs(env, other, expectNumber)
	if err != nil {
		return nil, err
	}
	rightBi := big.Int(m2)
	sum := new(big.Int).Sub(&leftBi, &rightBi)
	return MonetaryInt(*sum), nil
}

func (m Monetary) evalSub(env expressionEnv, other parser.ValueExpr) (Value, InterpreterError) {
	m2, err := evaluateExprAs(env, other, expectMonetary)
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

type opNeg interface {
	evalNeg(env expressionEnv) (Value, InterpreterError)
}

var _ opNeg = (*MonetaryInt)(nil)
var _ opNeg = (*Monetary)(nil)

func (m MonetaryInt) evalNeg(env expressionEnv) (Value, InterpreterError) {
	m1 := big.Int(m)
	neg := new(big.Int).Neg(&m1)
	return MonetaryInt(*neg), nil
}

func (m Monetary) evalNeg(env expressionEnv) (Value, InterpreterError) {
	m1 := big.Int(m.Amount)
	neg := new(big.Int).Neg(&m1)
	return Monetary{
		Asset:  m.Asset,
		Amount: MonetaryInt(*neg),
	}, nil

}
