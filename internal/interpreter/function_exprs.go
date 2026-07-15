package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
)

func evaluateFnCall(env *evalEnv, type_ *string, fnCall parser.FnCall) (Value, InterpreterError) {
	if type_ == nil {
		if err := env.checkFeatureFlag(flags.ExperimentalMidScriptFunctionCall); err != nil {
			return nil, err
		}
	}

	args, err := evaluateExpressions(env, fnCall.Args)
	if err != nil {
		return nil, err
	}

	switch fnCall.Caller.Name {
	case analysis.FnVarOriginMeta:
		if type_ == nil {
			return nil, InvalidNestedMeta{}
		}

		rawValue, err := meta(env, fnCall.Range, args)
		if err != nil {
			return nil, err
		}
		return parseVar(*type_, rawValue, fnCall.Range)

	case analysis.FnVarOriginBalance:
		monetary, err := balance(env, fnCall.Range, args)
		if err != nil {
			return nil, err
		}
		return monetary, nil

	case analysis.FnVarOriginOverdraft:
		monetary, err := overdraft(env, fnCall.Range, args)
		if err != nil {
			return nil, err
		}
		return monetary, nil

	case analysis.FnVarOriginGetAsset:
		return getAsset(env, fnCall.Range, args)
	case analysis.FnVarOriginGetAmount:
		return getAmount(env, fnCall.Range, args)
	case analysis.FnVarOriginScoped:
		return scoped(env, fnCall.Range, args)

	default:
		return nil, UnboundFunctionErr{Name: fnCall.Caller.Name}
	}
}

func overdraft(
	env *evalEnv,
	r parser.Range,
	args []Value,
) (Monetary, InterpreterError) {
	err := env.checkFeatureFlag(flags.ExperimentalOverdraftFunctionFeatureFlag)
	if err != nil {
		return Monetary{}, err
	}

	// TODO more precise args range location
	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	asset := parseArg(p, r, expectAsset)
	err = p.parse()
	if err != nil {
		return Monetary{}, err
	}

	// overdraft call doesn't handle colors
	balance_, err := env.getBalance(account, asset)
	if err != nil {
		return Monetary{}, err
	}

	balanceIsPositive := balance_.Cmp(big.NewInt(0)) == 1
	if balanceIsPositive {
		return Monetary{
			Amount: NewMonetaryInt(0),
			Asset:  asset,
		}, nil
	}

	overdraft := new(big.Int).Neg(balance_)
	return Monetary{
		Amount: MonetaryInt(*overdraft),
		Asset:  asset,
	}, nil
}

func meta(
	env *evalEnv,
	rng parser.Range,
	args []Value,
) (string, InterpreterError) {
	// TODO more precise location
	p := NewArgsParser(args)
	account := parseArg(p, rng, expectAccount)
	key := parseArg(p, rng, expectString)
	err := p.parse()
	if err != nil {
		return "", err
	}

	value, ok, err := env.getMetadata(account, string(key))
	if err != nil {
		return "", err
	}

	if !ok {
		return "", MetadataNotFound{
			Account: account.Name,
			Scope:   account.Scope,
			Key:     string(key),
			Range:   rng,
		}
	}

	return value, nil
}

func balance(
	env *evalEnv,
	r parser.Range,
	args []Value,
) (Monetary, InterpreterError) {
	// TODO more precise args range location
	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	asset := parseArg(p, r, expectAsset)
	err := p.parse()
	if err != nil {
		return Monetary{}, err
	}

	// body

	// balance call doesn't handle colors
	balance, err := env.getBalance(account, asset)
	if err != nil {
		return Monetary{}, err
	}

	if balance.Cmp(big.NewInt(0)) == -1 {
		return Monetary{}, NegativeBalanceError{
			Account: account.Name,
			Scope:   account.Scope,
			Amount:  *balance,
		}
	}

	balanceCopy := new(big.Int).Set(balance)

	m := Monetary{
		Asset:  Asset(asset),
		Amount: MonetaryInt(*balanceCopy),
	}
	return m, nil
}

func getAsset(
	env *evalEnv,
	r parser.Range,
	args []Value,
) (Value, InterpreterError) {
	err := env.checkFeatureFlag(flags.ExperimentalGetAssetFunctionFeatureFlag)
	if err != nil {
		return nil, err
	}

	p := NewArgsParser(args)
	mon := parseArg(p, r, expectMonetary)
	err = p.parse()
	if err != nil {
		return nil, err
	}

	return mon.Asset, nil
}

func getAmount(
	env *evalEnv,
	r parser.Range,
	args []Value,
) (Value, InterpreterError) {
	err := env.checkFeatureFlag(flags.ExperimentalGetAmountFunctionFeatureFlag)
	if err != nil {
		return nil, err
	}

	p := NewArgsParser(args)
	mon := parseArg(p, r, expectMonetary)
	err = p.parse()
	if err != nil {
		return nil, err
	}

	return mon.Amount, nil
}

func scoped(
	env *evalEnv,
	r parser.Range,
	args []Value,
) (Value, InterpreterError) {
	err := env.checkFeatureFlag(flags.ExperimentalScopedFunction)
	if err != nil {
		return nil, err
	}

	p := NewArgsParser(args)
	acc := parseArg(p, r, expectAccount)
	scope := parseArg(p, r, expectString)
	err = p.parse()

	scopeStr := string(scope)

	// Precondition: scope is valid idenfitier
	if err != nil {
		return nil, err
	}

	if !checkScopeName(scopeStr) {
		return nil, InvalidScope{Scope: scopeStr}
	}

	return AccountAddress{Name: acc.Name, Scope: scopeStr}, nil
}
