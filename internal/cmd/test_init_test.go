package cmd_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/cmd"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/stretchr/testify/require"
)

func TestMakeSpecsFileRetryForMissingFunds(t *testing.T) {
	out, err := cmd.MakeSpecsFile(`
		send [USD/2 10000] (
			 source = @alice
			 destination = @bob
		)
 `)

	require.Nil(t, err)
	require.Equal(t, interpreter.Balances{
		"alice": interpreter.AccountBalance{
			"USD/2": big.NewInt(10000),
		},
	}, out.Balances)
}

func TestUnusedVars(t *testing.T) {
	out, err := cmd.MakeSpecsFile(`
		vars { monetary $m }
 `)

	require.Nil(t, err)
	require.Equal(t, interpreter.VariablesMap{"m": "USD/2 100"}, out.Vars)
}

func TestMakeSpecsFileRetryForMissingFeatureFlags(t *testing.T) {
	out, err := cmd.MakeSpecsFile(`
		send [USD/2 10000] (
			 source = oneof { @world }
			 destination = @bob
		)
 `)

	require.Nil(t, err)
	require.Equal(t, []string{"experimental-oneof"}, out.FeatureFlags)
}

func TestMakeSpecsFileFeatureFlagRetryKeepsDefaultBalance(t *testing.T) {
	// The amount is lower than the default balance (100), so the
	// feature-flag retry must keep using the default balance instead
	// of falling back to a zero balance.
	out, err := cmd.MakeSpecsFile(`
		send [USD/2 42] (
			 source = oneof { @alice }
			 destination = @bob
		)
 `)

	require.Nil(t, err)
	require.Equal(t, []string{"experimental-oneof"}, out.FeatureFlags)
	require.Equal(t, interpreter.Balances{
		"alice": interpreter.AccountBalance{
			"USD/2": big.NewInt(100),
		},
	}, out.Balances)
}
