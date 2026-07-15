package interpreter

import (
	"context"
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
)

// zeroStore backs the runtime.RunState's lazy balance fallback. The interpreter
// fetches every needed balance through its own scope-aware Store and Prewarms it
// into the runtime, treating any un-fetched (account, scope, asset, color) as
// zero — exactly the semantics this store provides.
type zeroStore struct{}

func (zeroStore) GetBalance(account, asset, color string) *big.Int { return new(big.Int) }

// fetchAndPrewarm fetches the not-yet-cached tuples of query from the scope-aware
// Store in one round-trip and seeds them into rs, so later reads hit the cache.
// Shared by the single-key balance reader and the batched pre-execution pass.
func fetchAndPrewarm(ctx context.Context, store Store, rs *runtime.RunState, query BalanceQuery) error {
	var missing BalanceQuery
	for _, item := range query {
		if !rs.Has(item.Account, item.Scope, item.Asset, item.Color) {
			missing = append(missing, item)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	rows, err := store.GetBalances(ctx, missing)
	if err != nil {
		return err
	}
	seed := make(map[runtime.PairKey]*big.Int, len(rows))
	for _, row := range rows {
		seed[runtime.PairKey{Account: row.Account, Scope: row.Scope, Asset: row.Asset, Color: row.Color}] = row.Amount
	}
	rs.Prewarm(seed)
	return nil
}

// evalEnv is the environment for evaluating expressions. It reads metadata
// straight from the Store (cached), but balance reads are injected via getBalance
// because their policy differs by caller: script execution reads running balances
// from rs, while dependency resolution only needs the read recorded. Evaluation
// never touches the funds engine directly.
type evalEnv struct {
	ctx          context.Context
	Store        Store
	FeatureFlags map[string]struct{}
	vars         map[string]Value

	getBalance         func(account AccountAddress, asset Asset) (*big.Int, InterpreterError)
	CachedAccountsMeta InternalAccountsMetadata
}

func newEvalEnv(
	ctx context.Context,
	store Store,
	featureFlags map[string]struct{},
	getBalance func(AccountAddress, Asset) (*big.Int, InterpreterError),
	varDecls *parser.VarDeclarations,
	rawVars map[string]string,
) (evalEnv, InterpreterError) {
	env := evalEnv{
		ctx:                ctx,
		Store:              store,
		FeatureFlags:       featureFlags,
		vars:               map[string]Value{},
		getBalance:         getBalance,
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

// newBalanceGetter builds the balance reader used during evaluation: a lazy,
// write-through fetch over the batched, scope-aware Store into rs, so a mid-script
// balance() sees running balances mutated by funds execution (both share rs).
func newBalanceGetter(ctx context.Context, store Store, rs *runtime.RunState) func(AccountAddress, Asset) (*big.Int, InterpreterError) {
	return func(account AccountAddress, asset Asset) (*big.Int, InterpreterError) {
		color := String("")
		query := BalanceQuery{
			{Account: account.Name, Asset: string(asset), Color: string(color), Scope: account.Scope},
		}
		if err := fetchAndPrewarm(ctx, store, rs, query); err != nil {
			return nil, QueryBalanceError{WrappedError: err}
		}
		return rs.GetAccountBalance(account.Name, account.Scope, string(asset), string(color)), nil
	}
}

// getMetadata is a lazy, cached read of account metadata from the Store.
func (env *evalEnv) getMetadata(account AccountAddress, key string) (string, bool, InterpreterError) {
	if !env.CachedAccountsMeta.has(account, key) {
		rows, err := env.Store.GetAccountsMetadata(env.ctx, MetadataQuery{
			{Account: account.Name, Scope: account.Scope, Keys: []string{key}},
		})
		if err != nil {
			return "", false, QueryMetadataError{WrappedError: err}
		}
		env.CachedAccountsMeta.Merge(rows)
	}
	value, ok := env.CachedAccountsMeta.Get(account, key)
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

			default:
				return nil, unhandledErr(part)
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
			return nil, unhandledErr(expr.Operator)
		}

	case *parser.Prefix:
		switch expr.Operator {
		case parser.PrefixOperatorMinus:
			return unaryNegOp(env, expr.Expr)

		default:
			return nil, unhandledErr(expr.Operator)
		}

	case *parser.FnCall:
		// nil type: not a direct var origin, hence a mid-script call.
		return evaluateFnCall(env, nil, *expr)

	default:
		return nil, unhandledErr(expr)
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

	if !runtime.ValidateColor(string(color)) {
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
