package specs_format_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/stretchr/testify/require"
)

var exampleProgram = parser.Parse(`
	vars {
		account $source
		number $amount
	}

	send [USD $amount] (
		source = $source
		destination = @dest
	)
`)

func TestRunSpecsSimple(t *testing.T) {
	j := `{
		"testCases": [
			{
				"it": "t1",
				"vars": { "source": "src", "amount": "42" },
				"balances": { "src": { "USD": 9999 } },
				"expectedPostings": [
					{ "source": "src", "destination": "dest", "asset": "USD", "amount": 42 }
				]
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out := specs_format.Check(exampleProgram.Value, specs)
	require.Equal(t, specs_format.SpecsResult{
		Total:   1,
		Failing: 0,
		Passing: 1,
		Cases: []specs_format.TestCaseResult{
			{
				It:   "t1",
				Pass: true,
				Vars: interpreter.VariablesMap{
					"source": "src",
					"amount": "42",
				},
				Balances: interpreter.Balances{
					"src": interpreter.AccountBalance{
						"USD": big.NewInt(9999),
					},
				},
				Meta: interpreter.AccountsMetadata{},
				ExpectedPostings: []interpreter.Posting{
					{
						Source:      "src",
						Destination: "dest",
						Asset:       "USD",
						Amount:      big.NewInt(42),
					},
				},
				ActualPostings: []interpreter.Posting{
					{
						Source:      "src",
						Destination: "dest",
						Asset:       "USD",
						Amount:      big.NewInt(42),
					},
				},
			},
		},
	}, out)

}

func TestRunSpecsMergeOuter(t *testing.T) {
	j := `{
		"vars": { "source": "src", "amount": "42" },
		"balances": { "src": { "USD": 10 } },
		"testCases": [
			{
				"vars": { "amount": "1" },
				"balances": {
					"src": { "EUR": 2 },
					"dest": { "USD": 1 }
				},
				"it": "t1",
				"expectedPostings": [
					{ "source": "src", "destination": "dest", "asset": "USD", "amount": 1 }
				]
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out := specs_format.Check(exampleProgram.Value, specs)
	require.Equal(t, specs_format.SpecsResult{
		Total:   1,
		Failing: 0,
		Passing: 1,
		Cases: []specs_format.TestCaseResult{
			{
				It:   "t1",
				Pass: true,
				Vars: interpreter.VariablesMap{
					"source": "src",
					"amount": "1",
				},
				Meta: interpreter.AccountsMetadata{},
				Balances: interpreter.Balances{
					"src": interpreter.AccountBalance{
						"USD": big.NewInt(10),
						"EUR": big.NewInt(2),
					},
					"dest": interpreter.AccountBalance{
						"USD": big.NewInt(1),
					},
				},
				ExpectedPostings: []interpreter.Posting{
					{
						Source:      "src",
						Destination: "dest",
						Asset:       "USD",
						Amount:      big.NewInt(1),
					},
				},
				ActualPostings: []interpreter.Posting{
					{
						Source:      "src",
						Destination: "dest",
						Asset:       "USD",
						Amount:      big.NewInt(1),
					},
				},
			},
		},
	}, out)

}

func TestRunWithMissingBalance(t *testing.T) {
	j := `{
		"testCases": [
			{
				"it": "t1",
				"vars": { "source": "src", "amount": "42" },
				"balances": { "src": { "USD": 1 } },
				"expectedPostings": null
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out := specs_format.Check(exampleProgram.Value, specs)
	require.Equal(t, specs_format.SpecsResult{
		Total:   1,
		Failing: 0,
		Passing: 1,
		Cases: []specs_format.TestCaseResult{
			{
				It:   "t1",
				Pass: true,
				Vars: interpreter.VariablesMap{
					"source": "src",
					"amount": "42",
				},
				Balances: interpreter.Balances{
					"src": interpreter.AccountBalance{
						"USD": big.NewInt(1),
					},
				},
				Meta:             interpreter.AccountsMetadata{},
				ExpectedPostings: nil,
				ActualPostings:   nil,
			},
		},
	}, out)

}

func TestRunWithMissingBalanceWhenExpectedPostings(t *testing.T) {
	j := `{
		"testCases": [
			{
				"it": "t1",
				"vars": { "source": "src", "amount": "42" },
				"balances": { "src": { "USD": 1 } },
				"expectedPostings": [
					{ "source": "src", "destination": "dest", "asset": "USD", "amount": 1 }
				]
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out := specs_format.Check(exampleProgram.Value, specs)
	require.Equal(t, specs_format.SpecsResult{
		Total:   1,
		Failing: 1,
		Passing: 0,
		Cases: []specs_format.TestCaseResult{
			{
				It:   "t1",
				Pass: false,
				Vars: interpreter.VariablesMap{
					"source": "src",
					"amount": "42",
				},
				Balances: interpreter.Balances{
					"src": interpreter.AccountBalance{
						"USD": big.NewInt(1),
					},
				},
				Meta: interpreter.AccountsMetadata{},
				ExpectedPostings: []interpreter.Posting{
					{
						Source:      "src",
						Destination: "dest",
						Asset:       "USD",
						Amount:      big.NewInt(1),
					},
				},
				ActualPostings: nil,
			},
		},
	}, out)

}

func TestNoPostingsIsNotNullPostings(t *testing.T) {
	exampleProgram := parser.Parse(``)

	j := `{
		"testCases": [
			{
				"it": "t1",
				"vars": { "source": "src", "amount": "42" },
				"balances": { "src": { "USD": 1 } },
				"expectedPostings": null
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out := specs_format.Check(exampleProgram.Value, specs)
	require.Equal(t, specs_format.SpecsResult{
		Total:   1,
		Failing: 1,
		Passing: 0,
		Cases: []specs_format.TestCaseResult{
			{
				It:   "t1",
				Pass: false,
				Vars: interpreter.VariablesMap{
					"source": "src",
					"amount": "42",
				},
				Balances: interpreter.Balances{
					"src": interpreter.AccountBalance{
						"USD": big.NewInt(1),
					},
				},
				Meta:             interpreter.AccountsMetadata{},
				ExpectedPostings: nil,
				ActualPostings:   []interpreter.Posting{},
			},
		},
	}, out)

}
