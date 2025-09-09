package specs_format

import (
	"context"
	"math/big"
	"reflect"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

// --- Specs:
type Specs struct {
	FeatureFlags []string                     `json:"featureFlags,omitempty"`
	Balances     interpreter.Balances         `json:"balances,omitempty"`
	Vars         interpreter.VariablesMap     `json:"variables,omitempty"`
	Meta         interpreter.AccountsMetadata `json:"metadata,omitempty"`
	TestCases    []TestCase                   `json:"testCases,omitempty"`
}

type TestCase struct {
	It       string                       `json:"it"`
	Balances interpreter.Balances         `json:"balances,omitempty"`
	Vars     interpreter.VariablesMap     `json:"variables,omitempty"`
	Meta     interpreter.AccountsMetadata `json:"metadata,omitempty"`

	// Expectations
	ExpectMissingFunds   bool `json:"expect.error.missingFunds,omitempty"`
	ExpectNegativeAmount bool `json:"expect.error.negativeAmount,omitempty"`

	ExpectPostings           []interpreter.Posting        `json:"expect.postings,omitempty"`
	ExpectTxMeta             map[string]string            `json:"expect.txMetadata,omitempty"`
	ExpectAccountsMeta       interpreter.AccountsMetadata `json:"expect.metadata,omitempty"`
	ExpectEndBalances        interpreter.Balances         `json:"expect.endBalances,omitempty"`
	ExpectEndBalancesInclude interpreter.Balances         `json:"expect.endBalances.include,omitempty"`
	ExpectMovements          Movements                    `json:"expect.movements,omitempty"`
}

type TestCaseResult struct {
	It       string                       `json:"it"`
	Pass     bool                         `json:"pass"`
	Balances interpreter.Balances         `json:"balances"`
	Vars     interpreter.VariablesMap     `json:"variables"`
	Meta     interpreter.AccountsMetadata `json:"metadata"`

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

func runAssertion[T any](failedAssertions []AssertionMismatch[any], assertion string, expected T, got T, cmp func(T, T) bool) []AssertionMismatch[any] {
	eq := cmp(expected, got)
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

		if err != nil {
			switch err.(type) {
			case interpreter.MissingFundsErr:
				if !testCase.ExpectMissingFunds {
					failedAssertions = append(failedAssertions, AssertionMismatch[any]{
						Assertion: "expect.error.missingFunds",
						Expected:  false,
						Got:       true,
					})
				}
			case interpreter.NegativeAmountErr:
				if !testCase.ExpectNegativeAmount {
					failedAssertions = append(failedAssertions, AssertionMismatch[any]{
						Assertion: "expect.error.negativeAmount",
						Expected:  false,
						Got:       true,
					})
				}
			default:
				return SpecsResult{}, err
			}
		} else {

			if testCase.ExpectMissingFunds {
				failedAssertions = append(failedAssertions, AssertionMismatch[any]{
					Assertion: "expect.error.missingFunds",
					Expected:  true,
					Got:       false,
				})
			}

			if testCase.ExpectNegativeAmount {
				failedAssertions = append(failedAssertions, AssertionMismatch[any]{
					Assertion: "expect.error.negativeAmount",
					Expected:  true,
					Got:       false,
				})
			}

			if testCase.ExpectPostings != nil {
				failedAssertions = runAssertion[any](failedAssertions,
					"expect.postings",
					testCase.ExpectPostings,
					result.Postings,
					reflect.DeepEqual,
				)
			}

			if testCase.ExpectTxMeta != nil {
				metadata := map[string]string{}
				for k, v := range result.Metadata {
					metadata[k] = v.String()
				}
				failedAssertions = runAssertion[any](failedAssertions,
					"expect.txMeta",
					testCase.ExpectTxMeta,
					metadata,
					reflect.DeepEqual,
				)
			}

			if testCase.ExpectAccountsMeta != nil {
				failedAssertions = runAssertion[any](failedAssertions,
					"expect.accountsMeta",
					testCase.ExpectAccountsMeta,
					result.AccountsMetadata,
					reflect.DeepEqual,
				)
			}

			if testCase.ExpectEndBalances != nil {
				failedAssertions = runAssertion(failedAssertions,
					"expect.endBalances",
					testCase.ExpectEndBalances,
					getBalances(result.Postings, balances),
					interpreter.CompareBalances,
				)
			}

			if testCase.ExpectEndBalancesInclude != nil {
				failedAssertions = runAssertion(failedAssertions,
					"expect.endBalances.include",
					testCase.ExpectEndBalancesInclude,
					getBalances(result.Postings, balances),
					interpreter.CompareBalancesIncluding,
				)
			}

			if testCase.ExpectMovements != nil {
				failedAssertions = runAssertion[any](failedAssertions,
					"expect.movements",
					testCase.ExpectMovements,
					getMovements(result.Postings),
					reflect.DeepEqual,
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

// TODO test
type Movements = map[string]map[string]map[string]*big.Int

func getMovements(postings []interpreter.Posting) Movements {
	m := Movements{}

	for _, posting := range postings {
		assetsMap := utils.NestedMapGetOrPutDefault(m, posting.Source, posting.Destination, func() map[string]*big.Int {
			return map[string]*big.Int{}
		})

		amt := utils.MapGetOrPutDefault(assetsMap, posting.Asset, func() *big.Int {
			return new(big.Int)
		})

		amt.Add(amt, posting.Amount)
	}

	return m
}

func getBalances(postings []interpreter.Posting, initialBalances interpreter.Balances) interpreter.Balances {
	balances := initialBalances.DeepClone()
	for _, posting := range postings {
		sourceBalance := utils.NestedMapGetOrPutDefault(balances, posting.Source, posting.Asset, func() *big.Int {
			return new(big.Int)
		})
		sourceBalance.Sub(sourceBalance, posting.Amount)

		destinationBalance := utils.NestedMapGetOrPutDefault(balances, posting.Destination, posting.Asset, func() *big.Int {
			return new(big.Int)
		})
		destinationBalance.Add(destinationBalance, posting.Amount)
	}

	return balances
}
