package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

func (st *programState) evaluateLit(literal parser.ValueExpr) (Value, InterpreterError) {
	switch literal := literal.(type) {
	case *parser.AssetLiteral:
		return Asset(literal.Asset), nil
	case *parser.AccountLiteral:
		return AccountAddress(literal.Name), nil
	case *parser.StringLiteral:
		return String(literal.String), nil
	case *parser.RatioLiteral:
		return Portion(*literal.ToRatio()), nil
	case *parser.NumberLiteral:
		return MonetaryInt(*big.NewInt(int64(literal.Number))), nil
	case *parser.MonetaryLiteral:
		asset, err := evaluateLitExpecting(st, literal.Asset, expectAsset)
		if err != nil {
			return nil, err
		}

		amount, err := evaluateLitExpecting(st, literal.Amount, expectNumber)
		if err != nil {
			return nil, err
		}

		return Monetary{Asset: Asset(*asset), Amount: MonetaryInt(*amount)}, nil

	case *parser.Variable:
		value, ok := st.ParsedVars[literal.Name]
		if !ok {
			return nil, UnboundVariableErr{
				Name:  literal.Name,
				Range: literal.Range,
			}
		}
		return value, nil
	default:
		utils.NonExhaustiveMatchPanic[any](literal)
		return nil, nil
	}
}

func evaluateLitExpecting[T any](st *programState, literal parser.ValueExpr, expect func(Value, parser.Range) (*T, InterpreterError)) (*T, InterpreterError) {
	value, err := st.evaluateLit(literal)
	if err != nil {
		return nil, err
	}

	res, err := expect(value, literal.GetRange())
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (st *programState) evaluateLiterals(literals []parser.ValueExpr) ([]Value, InterpreterError) {
	var values []Value
	for _, argLit := range literals {
		value, err := st.evaluateLit(argLit)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
