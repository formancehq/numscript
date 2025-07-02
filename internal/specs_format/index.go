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
	It                   string                        `json:"it"`
	Balances             interpreter.Balances          `json:"balances,omitempty"`
	Vars                 interpreter.VariablesMap      `json:"vars,omitempty"`
	Meta                 interpreter.AccountsMetadata  `json:"accountsMeta,omitempty"`
	ExpectedPostings     []interpreter.Posting         `json:"expectedPostings"`
	ExpectedTxMeta       *interpreter.AccountsMetadata `json:"expectedTxMeta,omitempty"`
	ExpectedAccountsMeta *map[string]string            `json:"expectedAccountsMeta,omitempty"`
	ExpectMissingFunds   bool                          `json:"expectMissingFunds,omitempty"`
}

type TestCaseResult struct {
	It                   string                        `json:"it"`
	Pass                 bool                          `json:"pass"`
	Balances             interpreter.Balances          `json:"balances"`
	Vars                 interpreter.VariablesMap      `json:"vars"`
	Meta                 interpreter.AccountsMetadata  `json:"accountsMeta"`
	ExpectedPostings     []interpreter.Posting         `json:"expectedPostings"`
	ActualPostings       []interpreter.Posting         `json:"actualPostings"`
	ExpectedTxMeta       *map[string]string            `json:"expectedTxMeta,omitempty"`
	ActualTxMeta         *map[string]string            `json:"actualTxMeta,omitempty"`
	ExpectedAccountsMeta *interpreter.AccountsMetadata `json:"expectedAccountsMeta,omitempty"`
	ActualAccountsMeta   *interpreter.AccountsMetadata `json:"actualAccountsMeta,omitempty"`
}

type SpecsResult struct {
	// Invariants: total==passing+failing
	Total   uint `json:"total"`
	Passing uint `json:"passing"`
	Failing uint `json:"failing"`
	Cases   []TestCaseResult
}

func Check(program parser.Program, specs Specs) SpecsResult {
	specsResult := SpecsResult{}

	for _, testCase := range specs.TestCases {
		// TODO merge balances, vars, meta
		meta := mergeAccountsMeta(specs.Meta, testCase.Meta)
		balances := mergeBalances(specs.Balances, testCase.Balances)
		vars := mergeVars(specs.Vars, testCase.Vars)

		specsResult.Total += 1

		result, err := interpreter.RunProgram(
			context.Background(),
			program,
			vars,
			interpreter.StaticStore{
				Meta:     meta,
				Balances: balances,
			}, nil)

		var pass bool
		var actualPostings []interpreter.Posting

		// TODO recover err on missing funds
		if err != nil {
			if _, ok := err.(interpreter.MissingFundsErr); ok {

				pass = testCase.ExpectedPostings == nil
				actualPostings = nil
			} else {
				panic(err)
			}

		} else {
			pass = reflect.DeepEqual(result.Postings, testCase.ExpectedPostings)
			actualPostings = result.Postings
		}

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
			ActualPostings:   actualPostings,
		})
	}

	return specsResult
}

func mergeVars(v1 interpreter.VariablesMap, v2 interpreter.VariablesMap) interpreter.VariablesMap {
	out := interpreter.VariablesMap{}
	for k, v := range v1 {
		out[k] = v
	}
	for k, v := range v2 {
		out[k] = v
	}
	return out
}

func mergeAccountsMeta(m1 interpreter.AccountsMetadata, m2 interpreter.AccountsMetadata) interpreter.AccountsMetadata {
	out := m1.DeepClone()
	out.Merge(m2)
	return out
}

func mergeBalances(b1 interpreter.Balances, b2 interpreter.Balances) interpreter.Balances {
	out := b1.DeepClone()
	out.Merge(b2)
	return out
}
