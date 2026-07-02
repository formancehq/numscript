package interpreter

import (
	"context"
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

type evalEnv struct {
	ctx          context.Context
	Store        Store
	FeatureFlags map[string]struct{}

	vars               map[string]Value
	CachedBalances     InternalBalances
	CachedAccountsMeta InternalAccountsMetadata
}

func newEvalEnv(ctx context.Context, store Store, featureFlags map[string]struct{}, varDecls *parser.VarDeclarations, rawVars map[string]string) (evalEnv, InterpreterError) {
	env := evalEnv{
		ctx:                ctx,
		Store:              store,
		FeatureFlags:       featureFlags,
		vars:               map[string]Value{},
		CachedBalances:     InternalBalances{},
		CachedAccountsMeta: InternalAccountsMetadata{},
	}
	if err := bindVars(&env, varDecls, rawVars); err != nil {
		return evalEnv{}, err
	}
	return env, nil
}

func (env *evalEnv) checkFeatureFlag(flag string) InterpreterError {
	// a nil set enables every feature (e.g. dependency resolution)
	if env.FeatureFlags == nil {
		return nil
	}
	if _, ok := env.FeatureFlags[flag]; ok {
		return nil
	}
	return ExperimentalFeature{FlagName: flag}
}

func (env *evalEnv) getBalance(account AccountAddress, asset Asset) (*big.Int, InterpreterError) {
	color := String("")
	if account.Name != "world" && !env.CachedBalances.has(account, string(asset), string(color)) {
		rows, err := env.Store.GetBalances(env.ctx, BalanceQuery{
			{Account: account.Name, Asset: string(asset), Color: string(color), Scope: account.Scope},
		})
		if err != nil {
			return nil, QueryBalanceError{WrappedError: err}
		}
		env.CachedBalances.Merge(rows)
	}
	return env.CachedBalances.fetchBalance(account, asset, color), nil
}

func (env *evalEnv) getMetadata(account AccountAddress, key string) (string, bool, InterpreterError) {
	rows, err := env.Store.GetAccountsMetadata(env.ctx, MetadataQuery{
		{Account: account.Name, Scope: account.Scope, Keys: []string{key}},
	})
	if err != nil {
		return "", false, QueryMetadataError{WrappedError: err}
	}
	env.CachedAccountsMeta = FromAccountsMetadataRows(rows)

	value, ok := env.CachedAccountsMeta.Get(account.Name, account.Scope, key)
	return value, ok, nil
}

func bindVars(env *evalEnv, varDecls *parser.VarDeclarations, rawVars map[string]string) InterpreterError {
	if varDecls == nil {
		return nil
	}
	for _, decl := range varDecls.Declarations {
		var value Value
		var err InterpreterError
		if decl.Origin == nil {
			raw, ok := rawVars[decl.Name.Name]
			if !ok {
				return MissingVariableErr{Name: decl.Name.Name}
			}
			value, err = parseVar(decl.Type.Name, raw, decl.Type.Range)
		} else {
			value, err = evaluateVarOrigin(env, decl.Type.Name, *decl.Origin)
		}
		if err != nil {
			return err
		}
		env.vars[decl.Name.Name] = value
	}
	return nil
}

func evaluateExpr(env *evalEnv, expr parser.ValueExpr) (Value, InterpreterError) {
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
				err := env.checkFeatureFlag(flags.ExperimentalAccountInterpolationFlag)
				if err != nil {
					return nil, err
				}

				value, err := evaluateExpr(env, part)
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
		name := strings.Join(parts, "")
		return NewAccountAddress(name)

	case *parser.StringLiteral:
		return String(expr.String), nil
	case *parser.PercentageLiteral:
		return Portion(*expr.ToRatio()), nil
	case *parser.NumberLiteral:
		return MonetaryInt(*expr.Number), nil
	case *parser.MonetaryLiteral:
		asset, err := evaluateExprAs(env, expr.Asset, expectAsset)
		if err != nil {
			return nil, err
		}

		amount, err := evaluateExprAs(env, expr.Amount, expectNumber)
		if err != nil {
			return nil, err
		}

		return Monetary{Asset: asset, Amount: amount}, nil

	case *parser.Variable:
		value := env.vars[expr.Name]
		if value == nil {
			return nil, UnboundVariableErr{
				Name:  expr.Name,
				Range: expr.Range,
			}
		}
		return value, nil

	case *parser.BinaryInfix:
		switch expr.Operator {
		case parser.InfixOperatorPlus:
			return plusOp(env, expr.Left, expr.Right)

		case parser.InfixOperatorMinus:
			return subOp(env, expr.Left, expr.Right)

		case parser.InfixOperatorDiv:
			return divOp(env, expr.Range, expr.Left, expr.Right)

		default:
			utils.NonExhaustiveMatchPanic[any](expr.Operator)
			return nil, nil
		}

	case *parser.Prefix:
		switch expr.Operator {
		case parser.PrefixOperatorMinus:
			return unaryNegOp(env, expr.Expr)

		default:
			utils.NonExhaustiveMatchPanic[any](expr.Operator)
			return nil, nil
		}

	case *parser.FnCall:
		// nil type: not a direct var origin, hence a mid-script call.
		return evaluateFnCall(env, nil, *expr)

	default:
		utils.NonExhaustiveMatchPanic[any](expr)
		return nil, nil
	}
}

func evaluateOptExprAs[T any](env *evalEnv, expr parser.ValueExpr, expect func(Value, parser.Range) (T, InterpreterError)) (T, InterpreterError) {
	var t T
	if expr == nil {
		return t, nil
	}
	return evaluateExprAs(env, expr, expect)
}

func evaluateExprAs[T any](env *evalEnv, expr parser.ValueExpr, expect func(Value, parser.Range) (T, InterpreterError)) (T, InterpreterError) {
	var default_ T
	value, err := evaluateExpr(env, expr)
	if err != nil {
		return default_, err
	}

	res, err := expect(value, expr.GetRange())
	if err != nil {
		return default_, err
	}

	return res, nil
}

func evaluateExpressions(env *evalEnv, literals []parser.ValueExpr) ([]Value, InterpreterError) {
	var values []Value
	for _, argLit := range literals {
		value, err := evaluateExpr(env, argLit)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (s *programState) evaluateColor(colorExpr parser.ValueExpr) (String, InterpreterError) {
	color, err := evaluateOptExprAs(&s.evalEnv, colorExpr, expectString)
	if err != nil {
		return "", err
	}

	isValidColor := colorRe.Match([]byte(string(color)))
	if !isValidColor {
		return "", InvalidColor{
			Range: colorExpr.GetRange(),
			Color: string(color),
		}
	}

	return color, nil
}

func plusOp(env *evalEnv, left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {

	leftValue, err := evaluateExprAs(env, left, expectOneOf(
		expectMapped(expectMonetary, func(m Monetary) opAdd {
			return m
		}),

		// while "x.map(identity)" is the same as "x", just writing "expectNumber" would't typecheck
		expectMapped(expectNumber, func(bi MonetaryInt) opAdd {
			return bi
		}),
	))

	if err != nil {
		return nil, err
	}

	return leftValue.evalAdd(env, right)
}

func subOp(env *evalEnv, left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	leftValue, err := evaluateExprAs(env, left, expectOneOf(
		expectMapped(expectMonetary, func(m Monetary) opSub {
			return m
		}),
		expectMapped(expectNumber, func(bi MonetaryInt) opSub {
			return bi
		}),
	))

	if err != nil {
		return nil, err
	}

	return leftValue.evalSub(env, right)
}

func divOp(env *evalEnv, rng parser.Range, left parser.ValueExpr, right parser.ValueExpr) (Value, InterpreterError) {
	leftValue, err := evaluateExprAs(env, left, expectNumber)
	if err != nil {
		return nil, err
	}

	rightValue, err := evaluateExprAs(env, right, expectNumber)
	if err != nil {
		return nil, err
	}

	rightBi := (*big.Int)(&rightValue)
	leftBi := (*big.Int)(&leftValue)
	if rightBi.Cmp(big.NewInt(0)) == 0 {
		return nil, DivideByZero{
			Range:     rng,
			Numerator: leftBi,
		}
	}

	rat := new(big.Rat).SetFrac(leftBi, rightBi)

	return Portion(*rat), nil
}

func unaryNegOp(env *evalEnv, expr parser.ValueExpr) (Value, InterpreterError) {
	evExpr, err := evaluateExprAs(env, expr, expectOneOf(
		expectMapped(expectMonetary, func(m Monetary) opNeg {
			return m
		}),

		// while "x.map(identity)" is the same as "x", just writing "expectNumber" would't typecheck
		expectMapped(expectNumber, func(bi MonetaryInt) opNeg {
			return bi
		}),
	))

	if err != nil {
		return nil, err
	}

	return evExpr.evalNeg(env)
}

func castToString(v Value, rng parser.Range) (string, InterpreterError) {
	switch v := v.(type) {
	case AccountAddress:
		if v.Scope != "" {
			return "", CannotCastScopedAccountToString{Account: v.Name, Scope: v.Scope, Range: rng}
		}
		return v.Name, nil
	case String:
		return string(v), nil
	case MonetaryInt:
		return v.String(), nil

	default:
		// No asset nor ratio can be implicitly cast to string
		return "", CannotCastToString{Value: v, Range: rng}
	}
}
