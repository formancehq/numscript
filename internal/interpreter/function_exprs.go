package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
)

func overdraft(
	s *programState,
	r parser.Range,
	args []Value,
) (*Monetary, InterpreterError) {
	err := s.checkFeatureFlag(flags.ExperimentalOverdraftFunctionFeatureFlag)
	if err != nil {
		return nil, err
	}

	// TODO more precise args range location
	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	asset := parseArg(p, r, expectAsset)
	err = p.parse()
	if err != nil {
		return nil, err
	}

	balance_, err := getBalance(s, *account, *asset)
	if err != nil {
		return nil, err
	}

	balanceIsPositive := balance_.Cmp(big.NewInt(0)) == 1
	if balanceIsPositive {
		return &Monetary{
			Amount: NewMonetaryInt(0),
			Asset:  Asset(*asset),
		}, nil
	}

	overdraft := new(big.Int).Neg(balance_)
	return &Monetary{
		Amount: MonetaryInt(*overdraft),
		Asset:  Asset(*asset),
	}, nil
}

func meta(
	s *programState,
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

	meta, fetchMetaErr := s.Store.GetAccountsMetadata(s.ctx, MetadataQuery{
		*account: []string{*key},
	})
	if fetchMetaErr != nil {
		return "", QueryMetadataError{WrappedError: fetchMetaErr}
	}
	s.CachedAccountsMeta = meta

	// body
	accountMeta := s.CachedAccountsMeta[*account]
	value, ok := accountMeta[*key]

	if !ok {
		return "", MetadataNotFound{Account: *account, Key: *key, Range: rng}
	}

	return value, nil
}

func balance(
	s *programState,
	r parser.Range,
	args []Value,
) (*Monetary, InterpreterError) {
	// TODO more precise args range location
	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	asset := parseArg(p, r, expectAsset)
	err := p.parse()
	if err != nil {
		return nil, err
	}

	// body

	balance, err := getBalance(s, *account, *asset)
	if err != nil {
		return nil, err
	}

	if balance.Cmp(big.NewInt(0)) == -1 {
		return nil, NegativeBalanceError{
			Account: *account,
			Amount:  *balance,
		}
	}

	balanceCopy := new(big.Int).Set(balance)

	m := Monetary{
		Asset:  Asset(*asset),
		Amount: MonetaryInt(*balanceCopy),
	}
	return &m, nil
}

func getAsset(
	s *programState,
	r parser.Range,
	args []Value,
) (Value, InterpreterError) {
	err := s.checkFeatureFlag(flags.ExperimentalGetAssetFunctionFeatureFlag)
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
