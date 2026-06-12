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
}

type BalanceQuery []BalanceQueryItem

// For each account, list of the needed keys
type MetadataQuery map[string][]string

type Store interface {
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
				if row.Account != item.Account || row.Color != item.Color {
					continue
				}
				if row.Asset == baseAsset || strings.HasPrefix(row.Asset, baseAsset+"/") {
					output = append(output, BalanceRow{
						Account: row.Account,
						Asset:   row.Asset,
						Color:   row.Color,
						Amount:  new(big.Int).Set(row.Amount),
					})
				}
			}
			continue
		}

		// materialize the queried (account, asset, color), defaulting to a zero balance
		amount := new(big.Int)
		for _, row := range s.Balances {
			if row.Account == item.Account && row.Asset == item.Asset && row.Color == item.Color {
				amount.Set(row.Amount)
				break
			}
		}
		output = append(output, BalanceRow{
			Account: item.Account,
			Asset:   item.Asset,
			Color:   item.Color,
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
