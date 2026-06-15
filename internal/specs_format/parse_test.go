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
  "balances": [
    { "account": "alice", "asset": "EUR", "amount": 200 }
  ],
  "variables": {
    "amt": "200"
  },
  "testCases": [
    {
      "it": "d1",
      "balances": [
        { "account": "bob", "asset": "EUR", "amount": 42 }
      ],
      "expect.postings": [
        {
          "source": "src",
          "destination": "dest",
          "asset": "EUR",
          "amount": 100
        }
      ]
    }
  ]
}

	`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(raw), &specs)
	require.Nil(t, err)

	require.Equal(t, specs_format.Specs{
		Balances: interpreter.Balances{
			{Account: "alice", Asset: "EUR", Amount: big.NewInt(200)},
		},
		Vars: interpreter.VariablesMap{
			"amt": "200",
		},
		TestCases: []specs_format.TestCase{
			{
				It: "d1",
				Balances: interpreter.Balances{
					{Account: "bob", Asset: "EUR", Amount: big.NewInt(42)},
				},
				ExpectPostings: []interpreter.Posting{
					{
						Source:      "src",
						Destination: "dest",
						Asset:       "EUR",
						Amount:      big.NewInt(100),
					},
				},
			},
		},
	}, specs)

}
