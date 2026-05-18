package interpreter

import (
	"context"
	"math/big"
)

// recordingStore wraps a Store and records all balance and metadata reads.
// It is used by ResolveDependencies to discover which data a script depends on.
type recordingStore struct {
	inner         Store
	balanceReads  Balances
	metadataReads AccountsMetadata
}

func newRecordingStore(inner Store) *recordingStore {
	return &recordingStore{
		inner:         inner,
		balanceReads:  Balances{},
		metadataReads: AccountsMetadata{},
	}
}

func (r *recordingStore) GetBalances(ctx context.Context, query BalanceQuery) (Balances, error) {
	result, err := r.inner.GetBalances(ctx, query)
	if err != nil {
		return nil, err
	}

	for account, assets := range result {
		if _, ok := r.balanceReads[account]; !ok {
			r.balanceReads[account] = AccountBalance{}
		}
		for asset, balance := range assets {
			r.balanceReads[account][asset] = new(big.Int).Set(balance)
		}
	}

	return result, nil
}

func (r *recordingStore) GetAccountsMetadata(ctx context.Context, query MetadataQuery) (AccountsMetadata, error) {
	result, err := r.inner.GetAccountsMetadata(ctx, query)
	if err != nil {
		return nil, err
	}

	for account, meta := range result {
		if _, ok := r.metadataReads[account]; !ok {
			r.metadataReads[account] = AccountMetadata{}
		}
		for key, value := range meta {
			r.metadataReads[account][key] = value
		}
	}

	return result, nil
}
