package interpreter

import (
	"context"
	"math/big"
	"strings"
)

// For each account, list of the needed assets
type BalanceQueryItem struct {
	Account string
	Asset   string
	Color   string
	Scope   string
}

type MetadataQueryItem = struct {
	Account string
	Scope   string
	Keys    []string
}

type BalanceQuery []BalanceQueryItem

type MetadataQuery []MetadataQueryItem

type Store interface {
	// Returns the batched balances for a given batched query.
	//
	// Note: the "Balances" result is expected not to have duplicate entries.
	// Malformed Balances will result it undefined behaviour, and the implementation doesn't guarantee keys are deduped.
	GetBalances(context.Context, BalanceQuery) (Balances, error)
	GetAccountsMetadata(context.Context, MetadataQuery) (AccountsMetadata, error)
}

type StaticStore struct {
	Balances Balances
	Meta     AccountsMetadata
}

func (s StaticStore) GetBalances(_ context.Context, q BalanceQuery) (Balances, error) {
	var output Balances
	for _, item := range q {
		baseAsset, isCatchAll := strings.CutSuffix(item.Asset, "/*")

		if isCatchAll {
			// return every stored asset (of the queried color) under the base asset
			for _, row := range s.Balances {
				if row.Account != item.Account || row.Color != item.Color || row.Scope != item.Scope {
					continue
				}
				if row.Asset == baseAsset || strings.HasPrefix(row.Asset, baseAsset+"/") {
					output = append(output, BalanceRow{
						Account: row.Account,
						Asset:   row.Asset,
						Color:   row.Color,
						Scope:   row.Scope,
						Amount:  new(big.Int).Set(row.Amount),
					})
				}
			}
			continue
		}

		// materialize the queried (account, asset, color, scope), defaulting to a zero balance
		amount := new(big.Int)
		for _, row := range s.Balances {
			if row.Account == item.Account && row.Asset == item.Asset && row.Color == item.Color && row.Scope == item.Scope {
				amount.Set(row.Amount)
				break
			}
		}
		output = append(output, BalanceRow{
			Account: item.Account,
			Asset:   item.Asset,
			Color:   item.Color,
			Scope:   item.Scope,
			Amount:  amount,
		})
	}

	return output, nil
}

func (s StaticStore) GetAccountsMetadata(context.Context, MetadataQuery) (AccountsMetadata, error) {
	if s.Meta == nil {
		s.Meta = AccountsMetadata{}
	}
	return s.Meta, nil
}
