package interpreter

import (
	"math/big"
	"reflect"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

func (st *programState) evaluateExpr(expr parser.ValueExpr) (Value, InterpreterError) {
	switch expr := expr.(type) {
	case *parser.AssetLiteral:
		return Asset(expr.Asset), nil
	case *parser.AccountLiteral:
		return AccountAddress(expr.Name), nil
	case *parser.StringLiteral:
		return String(expr.String), nil
	case *parser.RatioLiteral:
		return Portion(*expr.ToRatio()), nil
	case *parser.NumberLiteral:
		return MonetaryInt(*big.NewInt(int64(expr.Number))), nil
	case *parser.MonetaryLiteral:
		asset, err := evaluateExprAs(st, expr.Asset, expectAsset)
		if err != nil {
			return nil, err
		}

		amount, err := evaluateExprAs(st, expr.Amount, expectNumber)
		if err != nil {
			return nil, err
		}

		return Monetary{Asset: Asset(*asset), Amount: MonetaryInt(*amount)}, nil

	case *parser.Variable:
		value, ok := st.ParsedVars[expr.Name]
		if !ok {
			return nil, UnboundVariableErr{
				Name:  expr.Name,
				Range: expr.Range,
			}
		}
		return value, nil

	case *parser.BinaryInfix:
		switch expr.Operator {
		case parser.InfixOperatorPlus:
			return st.plusOp(expr.Left, expr.Right)

		case parser.InfixOperatorMinus:
			return st.subOp(expr.Left, expr.Right)

		case parser.InfixOperatorEq:
			return st.eqOp(expr.Left, expr.Right)

		case parser.InfixOperatorNeq:
			return st.neqOp(expr.Left, expr.Right)

		case parser.InfixOperatorGt:
			return st.gtOp(expr.Left, expr.Right)

		case parser.InfixOperatorGte:
			return st.gteOp(expr.Left, expr.Right)

		case parser.InfixOperatorLt:
			return st.ltOp(expr.Left, expr.Right)

		case parser.InfixOperatorLte:
			return st.lteOp(expr.Left, expr.Right)

		case parser.InfixOperatorAnd:
			return st.andOp(expr.Left, expr.Right)

		case parser.InfixOperatorOr:
			return st.orOp(expr.Left, expr.Right)

		default:
			utils.NonExhaustiveMatchPanic[any](expr.Operator)
			return nil, nil
		}

	default:
		utils.NonExhaustiveMatchPanic[any](expr)
		return nil, nil
	}
}

func evaluateExprAs[T any](st *programState, expr parser.ValueExpr, expect func(Value, parser.Range) (*T, InterpreterError)) (*T, InterpreterError) {
	value, err := st.evaluateExpr(expr)
	if err != nil {
		return nil, err
	}

	res, err := expect(value, expr.GetRange())
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (st *programState) evaluateExpressions(literals []parser.ValueExpr) ([]Value, InterpreterError) {
	var values []Value
	for _, argLit := range literals {
		value, err := st.evaluateExpr(argLit)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (st *programState) plusOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	leftValue, err := evaluateExprAs(st, left, expectOneOf(
		expectMapped(expectMonetary, func(m Monetary) opAdd {
			return m
		}),

		// while "x.map(identity)" is the same as "x", just writing "expectNumber" would't typecheck
		expectMapped(expectNumber, func(bi big.Int) opAdd {
			return MonetaryInt(bi)
		}),
	))

	if err != nil {
		return nil, err
	}

	return (*leftValue).evalAdd(st, right)
}

func (st *programState) subOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	leftValue, err := evaluateExprAs(st, left, expectOneOf(
		expectMapped(expectMonetary, func(m Monetary) opSub {
			return m
		}),
		expectMapped(expectNumber, func(bi big.Int) opSub {
			return MonetaryInt(bi)
		}),
	))

	if err != nil {
		return nil, err
	}

	return (*leftValue).evalSub(st, right)
}

func (st *programState) numOp(left parser.ValueExpr, right parser.ValueExpr, op func(left *big.Int, right *big.Int) Value) (Value, InterpreterError) {
	parsedLeft, err := evaluateExprAs(st, left, expectNumber)
	if err != nil {
		return nil, err
	}

	parsedRight, err := evaluateExprAs(st, right, expectNumber)
	if err != nil {
		return nil, err
	}

	return op(parsedLeft, parsedRight), nil
}

func (st *programState) eqOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	parsedLeft, err := evaluateExprAs(st, left, expectAnything)
	if err != nil {
		return nil, err
	}

	parsedRight, err := evaluateExprAs(st, right, expectAnything)
	if err != nil {
		return nil, err
	}

	// TODO remove reflect usage
	return Bool(reflect.DeepEqual(parsedLeft, parsedRight)), nil
}

func (st *programState) neqOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	parsedLeft, err := evaluateExprAs(st, left, expectAnything)
	if err != nil {
		return nil, err
	}

	parsedRight, err := evaluateExprAs(st, right, expectAnything)
	if err != nil {
		return nil, err
	}

	// TODO remove reflect usage
	return Bool(!(reflect.DeepEqual(parsedLeft, parsedRight))), nil
}

func (st *programState) ltOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	return st.numOp(left, right, func(left, right *big.Int) Value {
		switch left.Cmp(right) {
		case -1:
			return Bool(true)
		default:
			return Bool(false)
		}
	})
}

func (st *programState) gtOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	return st.numOp(left, right, func(left, right *big.Int) Value {
		switch left.Cmp(right) {
		case 1:
			return Bool(true)
		default:
			return Bool(false)
		}
	})
}

func (st *programState) lteOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	return st.numOp(left, right, func(left, right *big.Int) Value {
		switch left.Cmp(right) {
		case -1, 0:
			return Bool(true)
		default:
			return Bool(false)
		}
	})
}

func (st *programState) gteOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	return st.numOp(left, right, func(left, right *big.Int) Value {
		switch left.Cmp(right) {
		case 1, 0:
			return Bool(true)
		default:
			return Bool(false)
		}
	})
}

func (st *programState) andOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	parsedLeft, err := evaluateExprAs(st, left, expectBool)
	if err != nil {
		return nil, err
	}

	if !*parsedLeft {
		return Bool(false), nil
	}

	parsedRight, err := evaluateExprAs(st, right, expectBool)
	if err != nil {
		return nil, err
	}

	return Bool(*parsedRight), nil
}

func (st *programState) orOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	parsedLeft, err := evaluateExprAs(st, left, expectBool)
	if err != nil {
		return nil, err
	}

	if *parsedLeft {
		return Bool(true), nil
	}

	parsedRight, err := evaluateExprAs(st, right, expectBool)
	if err != nil {
		return nil, err
	}

	return Bool(*parsedRight), nil
}
