package specs_format

import (
	"context"
	"reflect"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
)

// --- Specs:
type Specs struct {
	Balances  interpreter.Balances         `json:"balances,omitempty"`
	Vars      interpreter.VariablesMap     `json:"vars,omitempty"`
	Meta      interpreter.AccountsMetadata `json:"accountsMeta,omitempty"`
	TestCases []TestCase                   `json:"testCases,omitempty"`
}

type TestCase struct {
	It               string                       `json:"it"`
	Balances         interpreter.Balances         `json:"balances,omitempty"`
	Vars             interpreter.VariablesMap     `json:"vars,omitempty"`
	Meta             interpreter.AccountsMetadata `json:"accountsMeta,omitempty"`
	ExpectedPostings []interpreter.Posting        `json:"expectedPostings"`
	// TODO expected tx meta, accountsMeta
}

type TestCaseResult struct {
	It               string                       `json:"it"`
	Pass             bool                         `json:"pass"`
	Balances         interpreter.Balances         `json:"balances"`
	Vars             interpreter.VariablesMap     `json:"vars"`
	Meta             interpreter.AccountsMetadata `json:"accountsMeta"`
	ExpectedPostings []interpreter.Posting        `json:"expectedPostings"`
	ActualPostings   []interpreter.Posting        `json:"actualPostings"`

	// TODO expected tx meta, accountsMeta
}

type SpecsResult struct {
	// Invariants: total==passing+failing
	Total   uint `json:"total"`
	Passing uint `json:"passing"`
	Failing uint `json:"failing"`
	Cases   []TestCaseResult
}

func Run(program parser.Program, specs Specs) SpecsResult {
	specsResult := SpecsResult{}

	for _, testCase := range specs.TestCases {
		// TODO merge balances, vars, meta
		meta := specs.Meta
		balances := specs.Balances
		vars := specs.Vars

		specsResult.Total += 1

		result, err := interpreter.RunProgram(
			context.Background(),
			program,
			specs.Vars,
			interpreter.StaticStore{
				// TODO merge balance, meta
				Meta:     meta,
				Balances: balances,
			}, nil)

		// TODO recover err on missing funds
		if err != nil {
			panic(err)
		}

		pass := reflect.DeepEqual(result.Postings, testCase.ExpectedPostings)
		if pass {
			specsResult.Passing += 1
		} else {
			specsResult.Failing += 1
		}

		specsResult.Cases = append(specsResult.Cases, TestCaseResult{
			It:               testCase.It,
			Pass:             pass,
			Meta:             meta,
			Balances:         balances,
			Vars:             vars,
			ExpectedPostings: testCase.ExpectedPostings,
			ActualPostings:   result.Postings,
		})
	}

	return specsResult
}
