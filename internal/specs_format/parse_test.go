package specs_format_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/stretchr/testify/require"
)

func TestParseSpecs(t *testing.T) {

	raw := `
		{
			"it": "d1",
			"balances": {
				"alice": { "EUR": 200 }
			},
			"vars": {
				"amt": "200"
			},
			"expectedPostings": [
				{
					"source": "src",
					"destination": "dest",
					"asset": "EUR",
					"amount": 100
				}
			]
		}
	`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(raw), &specs)
	require.Nil(t, err)

	require.Equal(t, specs_format.Specs{
		It: "d1",
		Balances: interpreter.Balances{
			"alice": {
				"EUR": big.NewInt(200),
			},
		},
		Vars: interpreter.VariablesMap{
			"amt": "200",
		},
		ExpectedPostings: []interpreter.Posting{
			{
				Source:      "src",
				Destination: "dest",
				Asset:       "EUR",
				Amount:      big.NewInt(100),
			},
		},
	}, specs)

}
