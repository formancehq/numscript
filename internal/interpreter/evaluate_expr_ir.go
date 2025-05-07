package interpreter

import (
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

func (s *irEmitterState) evaluateExpr(expr parser.ValueExpr) (Value, InterpreterError) {
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
				err := s.checkFeatureFlag(flags.ExperimentalAccountInterpolationFlag)
				if err != nil {
					return nil, err
				}

				value, err := s.evaluateExpr(part)
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
		return NewAccountAddress(name)

	case *parser.StringLiteral:
		return String(expr.String), nil
	case *parser.PercentageLiteral:
		return Portion(*expr.ToRatio()), nil
	case *parser.NumberLiteral:
		return MonetaryInt(*big.NewInt(int64(expr.Number))), nil
	case *parser.MonetaryLiteral:
		asset, err := evaluateExprAs__IR(s, expr.Asset, expectAsset)
		if err != nil {
			return nil, err
		}

		amount, err := evaluateExprAs__IR(s, expr.Amount, expectNumber)
		if err != nil {
			return nil, err
		}

		return Monetary{Asset: Asset(*asset), Amount: MonetaryInt(*amount)}, nil

	case *parser.Variable:
		value, ok := s.vars[expr.Name]
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
			return s.plusOp(expr.Left, expr.Right)

		case parser.InfixOperatorMinus:
			return s.subOp(expr.Left, expr.Right)

		case parser.InfixOperatorDiv:
			return s.divOp(expr.Range, expr.Left, expr.Right)

		default:
			utils.NonExhaustiveMatchPanic[any](expr.Operator)
			return nil, nil
		}

	case *parser.FnCall:
		if !s.varOriginPosition {
			err := s.checkFeatureFlag(flags.ExperimentalMidScriptFunctionCall)
			if err != nil {
				return nil, err
			}
		}

		return s.handleFnCall(nil, *expr)

	default:
		utils.NonExhaustiveMatchPanic[any](expr)
		return nil, nil
	}
}

func evaluateOptExprAs__IR[T any](st *irEmitterState, expr parser.ValueExpr, expect func(Value, parser.Range) (*T, InterpreterError)) (*T, InterpreterError) {
	if expr == nil {
		return nil, nil
	}
	return evaluateExprAs__IR(st, expr, expect)
}

func evaluateExprAs__IR[T any](st *irEmitterState, expr parser.ValueExpr, expect func(Value, parser.Range) (*T, InterpreterError)) (*T, InterpreterError) {
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

func (st *irEmitterState) evaluateExpressions(literals []parser.ValueExpr) ([]Value, InterpreterError) {
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

func (s *irEmitterState) evaluateColor(colorExpr parser.ValueExpr) (*string, InterpreterError) {
	color, err := evaluateOptExprAs__IR(s, colorExpr, expectString)
	if err != nil {
		return nil, err
	}
	if color == nil {
		return nil, nil
	}

	isValidColor := colorRe.Match([]byte(*color))
	if !isValidColor {
		return nil, InvalidColor{
			Range: colorExpr.GetRange(),
			Color: *color,
		}
	}

	return color, nil
}

func (s *irEmitterState) plusOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	leftValue, err := evaluateExprAs__IR(s, left, expectOneOf(
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

	return (*leftValue).evalAdd(s, right)
}

func (st *irEmitterState) subOp(left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	leftValue, err := evaluateExprAs__IR(st, left, expectOneOf(
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

func (st *irEmitterState) divOp(rng parser.Range, left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	leftValue, err := evaluateExprAs__IR(st, left, expectNumber)
	if err != nil {
		return nil, err
	}

	rightValue, err := evaluateExprAs__IR(st, right, expectNumber)
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
