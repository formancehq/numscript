package interpreter

import (
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

func (st *programState) evaluateExpr(expr parser.ValueExpr) (Value, InterpreterError) {
	switch expr := expr.(type) {
	case *parser.AssetLiteral:
		return Asset(expr.Asset), nil
	case *parser.AccountInterpLiteral:
		var parts []string
		for _, part := range expr.Parts {
			switch part := part.(type) {
			case parser.AccountTextPart:
				parts = append(parts, part.Name)
			case *parser.Variable:
				value, err := st.evaluateExpr(part)
				if err != nil {
					return nil, err
				}
				strValue, err := castToString(value, expr.Range)
				if err != nil {
					return nil, err
				}
				parts = append(parts, strValue)
			}
		}
		name := strings.Join(parts, ":")
		// TODO validate valid names
		return AccountAddress(name), nil

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

	// TypeError
	case *parser.BinaryInfix:

		switch expr.Operator {
		case parser.InfixOperatorPlus:
			return st.plusOp(expr.Left, expr.Right)

		case parser.InfixOperatorMinus:
			return st.subOp(expr.Left, expr.Right)

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

func castToString(v Value, rng parser.Range) (string, InterpreterError) {
	switch v := v.(type) {
	case AccountAddress:
		return v.String(), nil
	case String:
		return v.String(), nil
	case MonetaryInt:
		return v.String(), nil

	default:
		// No asset nor ratio can be implicitly cast to string
		return "", CannotCastToString{Value: v, Range: rng}
	}
}
