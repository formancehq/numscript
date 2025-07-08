package specs_format

import (
	"context"
	"reflect"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
)

// --- Specs:
type Specs struct {
	FeatureFlags []string                     `json:"featureFlags,omitempty"`
	Balances     interpreter.Balances         `json:"balances,omitempty"`
	Vars         interpreter.VariablesMap     `json:"vars,omitempty"`
	Meta         interpreter.AccountsMetadata `json:"accountsMeta,omitempty"`
	TestCases    []TestCase                   `json:"testCases,omitempty"`
}

type TestCase struct {
	It       string                       `json:"it"`
	Balances interpreter.Balances         `json:"balances,omitempty"`
	Vars     interpreter.VariablesMap     `json:"vars,omitempty"`
	Meta     interpreter.AccountsMetadata `json:"accountsMeta,omitempty"`

	// Expectations
	ExpectedPostings     []interpreter.Posting        `json:"expect.postings"`
	ExpectedTxMeta       map[string]string            `json:"expect.txMeta,omitempty"`
	ExpectedAccountsMeta interpreter.AccountsMetadata `json:"expect.accountsMeta,omitempty"`
	ExpectMissingFunds   bool                         `json:"expect.missingFunds,omitempty"`
}

type TestCaseResult struct {
	It       string                       `json:"it"`
	Pass     bool                         `json:"pass"`
	Balances interpreter.Balances         `json:"balances"`
	Vars     interpreter.VariablesMap     `json:"vars"`
	Meta     interpreter.AccountsMetadata `json:"accountsMeta"`

	// Assertions
	FailedAssertions []AssertionMismatch[any] `json:"failedAssertions"`
}

type SpecsResult struct {
	// Invariants: total==passing+failing
	Total   uint `json:"total"`
	Passing uint `json:"passing"`
	Failing uint `json:"failing"`
	Cases   []TestCaseResult
}

func runAssertion(failedAssertions []AssertionMismatch[any], assertion string, expected any, got any) []AssertionMismatch[any] {
	eq := reflect.DeepEqual(expected, got)
	if !eq {
		return append(failedAssertions, AssertionMismatch[any]{
			Assertion: assertion,
			Expected:  expected,
			Got:       got,
		})
	}

	return failedAssertions
}

func Check(program parser.Program, specs Specs) (SpecsResult, interpreter.InterpreterError) {
	specsResult := SpecsResult{}

	for _, testCase := range specs.TestCases {
		// TODO merge balances, vars, meta
		meta := mergeAccountsMeta(specs.Meta, testCase.Meta)
		balances := mergeBalances(specs.Balances, testCase.Balances)
		vars := mergeVars(specs.Vars, testCase.Vars)

		specsResult.Total += 1

		featureFlags := make(map[string]struct{})
		for _, flag := range specs.FeatureFlags {
			featureFlags[flag] = struct{}{}
		}

		result, err := interpreter.RunProgram(
			context.Background(),
			program,
			vars,
			interpreter.StaticStore{
				Meta:     meta,
				Balances: balances,
			},
			featureFlags,
		)

		var failedAssertions []AssertionMismatch[any]

		// TODO recover err on missing funds
		if err != nil {
			_, ok := err.(interpreter.MissingFundsErr)
			if !ok {
				return SpecsResult{}, err
			}

			if !testCase.ExpectMissingFunds {
				failedAssertions = append(failedAssertions, AssertionMismatch[any]{
					Assertion: "expect.missingFunds",
					Expected:  false,
					Got:       true,
				})
			}

		} else {

			if testCase.ExpectMissingFunds {
				failedAssertions = append(failedAssertions, AssertionMismatch[any]{
					Assertion: "expect.missingFunds",
					Expected:  true,
					Got:       false,
				})
			}

			if testCase.ExpectedPostings != nil {
				failedAssertions = runAssertion(failedAssertions,
					"expect.postings",
					testCase.ExpectedPostings,
					result.Postings,
				)
			}

			if testCase.ExpectedTxMeta != nil {
				metadata := map[string]string{}
				for k, v := range result.Metadata {
					metadata[k] = v.String()
				}
				failedAssertions = runAssertion(failedAssertions,
					"expect.txMeta",
					testCase.ExpectedTxMeta,
					metadata,
				)
			}

			if testCase.ExpectedAccountsMeta != nil {
				failedAssertions = runAssertion(failedAssertions,
					"expect.accountsMeta",
					testCase.ExpectedAccountsMeta,
					result.AccountsMetadata,
				)
			}

		}

		pass := len(failedAssertions) == 0
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
			FailedAssertions: failedAssertions,
		})
	}

	return specsResult, nil
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

type AssertionMismatch[T any] struct {
	Assertion string `json:"assertion"`
	Expected  T      `json:"expected,omitempty"`
	Got       T      `json:"got,omitempty"`
}
