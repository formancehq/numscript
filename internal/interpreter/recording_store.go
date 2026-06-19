package interpreter

import (
	"context"
	"math/big"
)

// recordingStore wraps a Store and records all balance and metadata reads,
// preserving the order in which the underlying store returned them.
//
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

	for _, row := range result {
		if r.balanceReads.hasRow(row.Account, row.Asset, row.Color) {
			continue
		}
		amount := new(big.Int)
		if row.Amount != nil {
			amount.Set(row.Amount)
		}
		r.balanceReads = append(r.balanceReads, BalanceRow{
			Account: row.Account,
			Asset:   row.Asset,
			Color:   row.Color,
			Amount:  amount,
		})
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

func (rows Balances) hasRow(account, asset, color string) bool {
	for i := range rows {
		if rows[i].Account == account && rows[i].Asset == asset && rows[i].Color == color {
			return true
		}
	}
	return false
}
