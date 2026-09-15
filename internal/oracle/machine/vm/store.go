package vm

import (
	"context"
	"math/big"

	"github.com/formancehq/go-libs/v5/pkg/types/metadata"
)

// BalanceQuery is a map of account/asset
type BalanceQuery map[string][]string

// Balances is a map of account/asset/balance
type Balances map[string]map[string]*big.Int

type Store interface {
	GetBalances(ctx context.Context, query BalanceQuery) (Balances, error)
	GetAccount(ctx context.Context, address string) (*Account, error)
}

type emptyStore struct{}

func (e *emptyStore) GetBalances(context.Context, BalanceQuery) (Balances, error) {
	return Balances{}, nil
}

func (e *emptyStore) GetAccount(_ context.Context, address string) (*Account, error) {
	return &Account{
		Address:  address,
		Metadata: metadata.Metadata{},
	}, nil
}

var _ Store = (*emptyStore)(nil)

var EmptyStore = &emptyStore{}

type AccountWithBalances struct {
	Account
	Balances map[string]*big.Int
}

type StaticStore map[string]*AccountWithBalances

func (s StaticStore) GetBalances(_ context.Context, query BalanceQuery) (Balances, error) {
	ret := Balances{}
	for accountAddress, assets := range query {
		// ret[accountAddress] must be initialized once, before the assets
		// loop below: allocating a fresh map on every asset iteration (as
		// this used to do) silently discards every asset entry but the
		// last one when an account has more than one asset queried.
		ret[accountAddress] = make(map[string]*big.Int)
		account, ok := s[accountAddress]
		for _, asset := range assets {
			if !ok {
				ret[accountAddress][asset] = new(big.Int)
				continue
			}
			balance, ok := account.Balances[asset]
			if !ok {
				ret[accountAddress][asset] = new(big.Int)
				continue
			}
			ret[accountAddress][asset] = balance
		}
	}

	return ret, nil
}

func (s StaticStore) GetAccount(_ context.Context, address string) (*Account, error) {
	account, ok := s[address]
	if !ok {
		return &Account{
			Address:  address,
			Metadata: metadata.Metadata{},
		}, nil
	}

	return &account.Account, nil
}

var _ Store = StaticStore{}
