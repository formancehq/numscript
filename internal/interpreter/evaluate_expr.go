package interpreter

import (
	"math/big"

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
	case *parser.PercentageLiteral:
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

		case parser.InfixOperatorDiv:
			return st.divOp(expr.Range, expr.Left, expr.Right)

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

func (st *programState) divOp(rng parser.Range, left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	leftValue, err := evaluateExprAs(st, left, expectNumber)
	if err != nil {
		return nil, err
	}

	rightValue, err := evaluateExprAs(st, right, expectNumber)
	if err != nil {
		return nil, err
	}

	if rightValue.Cmp(big.NewInt(0)) == 0 {
		return nil, DivideByZero{
			Range:     rng,
			Numerator: leftValue,
		}
	}

	rat := new(big.Rat).SetFrac(leftValue, rightValue)

	return Portion(*rat), nil
}
