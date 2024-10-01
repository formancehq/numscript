package numscript_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/stretchr/testify/require"
)

func TestGetBalancesInorder(t *testing.T) {
	parseResult := numscript.Parse(`vars {
	account $s1
	account $s2 = meta(@account_that_needs_meta, "k")
	number $b = balance(@account_that_needs_balance, USD/2)
}

send [COIN 100] (
	source = {
		$s1
		$s2
		@source3
		@world
	}
  	destination = @dest
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{
		StaticStore: interpreter.StaticStore{
			Balances: interpreter.Balances{},
			Meta:     interpreter.Metadata{"account_that_needs_meta": {"k": "source2"}},
		},
	}
	_, err := parseResult.Run(context.Background(), numscript.VariablesMap{
		"s1": "source1",
	},
		&store,
	)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.MetadataQuery{
			{
				"account_that_needs_meta": {"k"},
			},
		},
		store.GetMetadataCalls)

	require.Equal(t,
		[]numscript.BalanceQuery{
			// TODO maybe those calls can be batched together
			{
				// this is required by the balance() call
				"account_that_needs_balance": {"USD/2"},
			},
			{
				// this is defined in the variables
				"source1": {"COIN"},

				// this is defined in account metadata
				"source2": {"COIN"},

				// this appears as literal
				"source3": {"COIN"},
			},
		},
		store.GetBalancesCalls)
}

func TestGetBalancesAllotment(t *testing.T) {
	parseResult := numscript.Parse(`send [COIN 100] (
	source = {
		1/2 from @a
		remaining from @b
	}
  	destination = @dest
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{
		StaticStore: interpreter.StaticStore{
			Balances: interpreter.Balances{
				"a": {"COIN": big.NewInt(10000)},
				"b": {"COIN": big.NewInt(10000)},
			},
		},
	}

	_, err := parseResult.Run(context.Background(),
		numscript.VariablesMap{},
		&store,
	)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				"a": {"COIN"},
				"b": {"COIN"},
			},
		},
		store.GetBalancesCalls)
}

func TestGetBalancesOverdraft(t *testing.T) {
	parseResult := numscript.Parse(`send [COIN 100] (
	source = {
		@a allowing overdraft up to [COIN 10]
		@b allowing unbounded overdraft
	}
  	destination = @dest
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{}

	_, err := parseResult.Run(context.Background(), interpreter.VariablesMap{}, &store)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				"a": {"COIN"},
			},
		},
		store.GetBalancesCalls)
}

type ObservableStore struct {
	StaticStore      interpreter.StaticStore
	GetBalancesCalls []numscript.BalanceQuery
	GetMetadataCalls []numscript.MetadataQuery
}

func (os *ObservableStore) GetBalances(ctx context.Context, q interpreter.BalanceQuery) (interpreter.Balances, error) {
	os.GetBalancesCalls = append(os.GetBalancesCalls, q)
	return os.StaticStore.GetBalances(ctx, q)

}

func (os *ObservableStore) GetAccountsMetadata(ctx context.Context, q interpreter.MetadataQuery) (interpreter.Metadata, error) {
	os.GetMetadataCalls = append(os.GetMetadataCalls, q)
	return os.StaticStore.GetAccountsMetadata(ctx, q)
}
