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
