package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
)

func stringToAccount(
	s *programState,
	rng parser.Range,
	args []Value,
) (Value, InterpreterError) {
	if !s.StringToAccountFunctionFeatureFlag {
		return nil, ExperimentalFeature{FlagName: ExperimentalStringToAccountFunctionFeatureFlag}
	}

	// TODO more precise location
	p := NewArgsParser(args)
	accountStr := parseArg(p, rng, expectString)
	err := p.parse()
	if err != nil {
		return nil, err
	}

	return AccountAddress(*accountStr), nil
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

func overdraft(
	s *programState,
	r parser.Range,
	args []Value,
) (*Monetary, InterpreterError) {
	if !s.OverdraftFunctionFeatureFlag {
		return nil, ExperimentalFeature{FlagName: ExperimentalOverdraftFunctionFeatureFlag}
	}

	// TODO more precise args range location
	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	asset := parseArg(p, r, expectAsset)
	err := p.parse()
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

// Utility function to get the balance
func getBalance(
	s *programState,
	account string,
	asset string,
) (*big.Int, InterpreterError) {
	s.batchQuery(account, asset)
	fetchBalanceErr := s.runBalancesQuery()
	if fetchBalanceErr != nil {
		return nil, QueryBalanceError{WrappedError: fetchBalanceErr}
	}
	balance := s.getCachedBalance(account, asset)
	return balance, nil

}
