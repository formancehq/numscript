package numscript_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript"
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

	var getBalancesCalls []numscript.BalanceQuery
	var getMetaCalls []numscript.MetadataQuery

	_, err := parseResult.Run(
		numscript.Store{
			VariablesMap: numscript.VariablesMap{
				"s1": "source1",
			},
			GetBalances: func(bq numscript.BalanceQuery) (numscript.Balances, error) {
				getBalancesCalls = append(getBalancesCalls, bq)
				return numscript.Balances{
					"source1":                    {"COIN": big.NewInt(10)},
					"source2":                    {"COIN": big.NewInt(10)},
					"source3":                    {"COIN": big.NewInt(10)},
					"account_that_needs_balance": {"USD/2": big.NewInt(10)},
				}, nil
			},
			GetAccountsMetadata: func(mq numscript.MetadataQuery) (numscript.Metadata, error) {
				getMetaCalls = append(getMetaCalls, mq)
				return numscript.Metadata{
					"account_that_needs_meta": numscript.AccountMetadata{"k": "source2"},
				}, nil
			},
		},
	)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.MetadataQuery{
			{
				"account_that_needs_meta": {"k"},
			},
		},
		getMetaCalls)

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
		getBalancesCalls)
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

	var getBalancesCalls []numscript.BalanceQuery
	_, err := parseResult.Run(
		numscript.Store{
			GetBalances: func(bq numscript.BalanceQuery) (numscript.Balances, error) {
				getBalancesCalls = append(getBalancesCalls, bq)
				return numscript.Balances{
					"a": {"COIN": big.NewInt(10000)},
					"b": {"COIN": big.NewInt(10000)},
				}, nil
			},
		},
	)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				"a": {"COIN"},
				"b": {"COIN"},
			},
		},
		getBalancesCalls)
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

	var getBalancesCalls []numscript.BalanceQuery
	_, err := parseResult.Run(
		numscript.Store{
			GetBalances: func(bq numscript.BalanceQuery) (numscript.Balances, error) {
				getBalancesCalls = append(getBalancesCalls, bq)
				return numscript.Balances{
					"a": {"COIN": big.NewInt(10000)},
					"b": {"COIN": big.NewInt(10000)},
				}, nil
			},
		},
	)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				"a": {"COIN"},
			},
		},
		getBalancesCalls)
}
