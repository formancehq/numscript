package numscript_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript"
	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	t.Skip()

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
					"source1":                    {"COIN": big.NewInt(1000)},
					"source2":                    {"COIN": big.NewInt(1000)},
					"source3":                    {"COIN": big.NewInt(1000)},
					"account_that_needs_balance": {"USD/2": big.NewInt(1000)},
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
			{
				// this is required by the balance() call
				"account_that_needs_balance": {"USD/2"},
			},

			// TODO batch the following calls

			{
				// this is defined in the variables
				"source1": {"COIN"},
			},

			{
				// this is defined in account metadata
				"source2": {"COIN"},
			},

			{
				// this appears as literal
				"source3": {"COIN"},
			},
		},
		getBalancesCalls)
}
