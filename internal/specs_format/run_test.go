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
				"variables": { "source": "src", "amount": "42" },
				"balances": { "src": { "USD": 9999 } },
				"expect.postings": [
					{ "source": "src", "destination": "dest", "asset": "USD", "amount": 42 }
				]
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out, err := specs_format.Check(exampleProgram.Value, specs)
	require.Nil(t, err)

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
				Meta:             interpreter.AccountsMetadata{},
				FailedAssertions: nil,

				Postings: []interpreter.Posting{
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
		"variables": { "source": "src", "amount": "42" },
		"balances": { "src": { "USD": 10 } },
		"testCases": [
			{
				"variables": { "amount": "1" },
				"balances": {
					"src": { "EUR": 2 },
					"dest": { "USD": 1 }
				},
				"it": "t1",
				"expect.postings": [
					{ "source": "src", "destination": "dest", "asset": "USD", "amount": 1 }
				]
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out, err := specs_format.Check(exampleProgram.Value, specs)
	require.Nil(t, err)

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
				FailedAssertions: nil,
				Postings: []interpreter.Posting{
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
				"variables": { "source": "src", "amount": "42" },
				"balances": { "src": { "USD": 1 } },
				"expect.error.missingFunds": false,
				"expect.postings": null
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out, err := specs_format.Check(exampleProgram.Value, specs)
	require.Nil(t, err)

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
				FailedAssertions: []specs_format.AssertionMismatch[any]{
					{
						Assertion: "expect.error.missingFunds",
						Expected:  false,
						Got:       true,
					},
				},
				// ExpectedPostings: nil,
				// ActualPostings:   nil,
			},
		},
	}, out)

}

func TestRunWithMissingBalanceWhenExpectedPostings(t *testing.T) {
	j := `{
		"testCases": [
			{
				"it": "t1",
				"variables": { "source": "src", "amount": "42" },
				"balances": { "src": { "USD": 1 } },
				"expect.postings": [
					{ "source": "src", "destination": "dest", "asset": "USD", "amount": 1 }
				]
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out, err := specs_format.Check(exampleProgram.Value, specs)
	require.Nil(t, err)

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
				FailedAssertions: []specs_format.AssertionMismatch[any]{
					{
						Assertion: "expect.error.missingFunds",
						Got:       true,
						Expected:  false,
					},
				},
			},
		},
	}, out)

}

func TestNullPostingsIsNoop(t *testing.T) {
	exampleProgram := parser.Parse(``)

	j := `{
		"testCases": [
			{
				"it": "t1",
				"variables": { "source": "src", "amount": "42" },
				"balances": { "src": { "USD": 1 } },
				"expect.postings": null
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out, err := specs_format.Check(exampleProgram.Value, specs)
	require.Nil(t, err)

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
				FailedAssertions: nil,
				Postings:         []interpreter.Posting{},
			},
		},
	}, out)

}

func TestNegativeAmt(t *testing.T) {
	exampleProgram := parser.Parse(`
		vars { number $amt }
		send [USD $amt] (
			source = @world
			destination = @dest
		)
	`)

	j := `{
		"testCases": [
			{
				"it": "t1",
				"variables": { "amt": "-100" },
				"expect.error.negativeAmount": true
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out, err := specs_format.Check(exampleProgram.Value, specs)
	require.Nil(t, err)

	require.Equal(t, specs_format.SpecsResult{
		Total:   1,
		Failing: 0,
		Passing: 1,
		Cases: []specs_format.TestCaseResult{
			{
				It:   "t1",
				Pass: true,
				Vars: interpreter.VariablesMap{
					"amt": "-100",
				},
				Balances:         interpreter.Balances{},
				Meta:             interpreter.AccountsMetadata{},
				FailedAssertions: nil,
			},
		},
	}, out)

}

func TestSkip(t *testing.T) {
	j := `{
		"testCases": [
			{
				"it": "t1",
				"skip": true
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out, err := specs_format.Check(exampleProgram.Value, specs)
	require.Nil(t, err)

	require.Equal(t, specs_format.SpecsResult{
		Total:   1,
		Failing: 0,
		Passing: 0,
		Skipped: 1,
		Cases: []specs_format.TestCaseResult{
			{
				It:      "t1",
				Pass:    false,
				Skipped: true,
			},
		},
	}, out)

}

func TestFocus(t *testing.T) {
	j := `{
		"testCases": [
			{
				"it": "t1",
				"variables": { "source": "src", "amount": "10" },
				"balances": { "src": { "USD": 9999 } },
				"expect.postings": [
					{ "source": "src", "destination": "dest", "asset": "USD", "amount": 42 }
				]
			},
			{
				"it": "t2",
				"focus": true,
				"variables": { "source": "src", "amount": "42" },
				"balances": { "src": { "USD": 9999 } },
				"expect.postings": [
					{ "source": "src", "destination": "dest", "asset": "USD", "amount": 42 }
				]
			}
		]
	}`

	var specs specs_format.Specs
	err := json.Unmarshal([]byte(j), &specs)
	require.Nil(t, err)

	out, err := specs_format.Check(exampleProgram.Value, specs)
	require.Nil(t, err)

	require.Equal(t, specs_format.SpecsResult{
		Total:   2,
		Failing: 0,
		Passing: 1,
		Skipped: 1,
		Cases: []specs_format.TestCaseResult{
			{
				It:      "t1",
				Skipped: true,
			},

			{
				It:   "t2",
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
				Meta:             interpreter.AccountsMetadata{},
				FailedAssertions: nil,

				Postings: []interpreter.Posting{
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
