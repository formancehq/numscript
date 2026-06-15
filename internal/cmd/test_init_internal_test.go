package cmd

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestMakeSpecsFileRetryExhaustion(t *testing.T) {
	parseResult := parser.Parse(`
		send [USD/2 100] (
			 source = @alice
			 destination = @bob
		)
 `)
	require.Empty(t, parseResult.Errors)

	// With a zero default balance the program keeps failing with
	// MissingFundsErr; a zero retry budget must surface a clear error
	// instead of recursing.
	_, err := makeSpecsFile(
		parseResult.Value,
		map[string]string{},
		map[string]struct{}{},
		big.NewInt(0),
		0,
	)

	require.ErrorContains(t, err, "exceeded the maximum number of retries")
}
