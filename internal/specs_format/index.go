package specs_format

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"sort"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

// --- Specs:
type Specs struct {
	Schema       string                       `json:"$schema,omitempty"`
	FeatureFlags []string                     `json:"featureFlags,omitempty"`
	Balances     interpreter.Balances         `json:"balances,omitempty"`
	Vars         interpreter.VariablesMap     `json:"variables,omitempty"`
	Meta         interpreter.AccountsMetadata `json:"metadata,omitempty"`
	TestCases    []TestCase                   `json:"testCases,omitempty"`
}

type TestCase struct {
	It string `json:"it"`

	// Preconditions
	Balances interpreter.Balances         `json:"balances,omitempty"`
	Vars     interpreter.VariablesMap     `json:"variables,omitempty"`
	Meta     interpreter.AccountsMetadata `json:"metadata,omitempty"`

	// Select tests
	Focus bool `json:"focus,omitempty"`
	Skip  bool `json:"skip,omitempty"`

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
	Skipped  bool                         `json:"skipped"`
	It       string                       `json:"it"`
	Pass     bool                         `json:"pass"`
	Balances interpreter.Balances         `json:"balances"`
	Vars     interpreter.VariablesMap     `json:"variables"`
	Meta     interpreter.AccountsMetadata `json:"metadata"`

	// Output:
	Postings []interpreter.Posting `json:"postings"`

	// Assertions
	FailedAssertions []AssertionMismatch[any] `json:"failedAssertions"`
}

type SpecsResult struct {
	// Invariants: total==passing+failing
	Total   uint `json:"total"`
	Passing uint `json:"passing"`
	Failing uint `json:"failing"`
	Skipped uint `json:"skipped"`
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
	// we need a first pass to know whether there is at least on test in focus mode
	hasFocusedTest := slices.ContainsFunc(specs.TestCases, func(t TestCase) bool {
		return t.Focus
	})

	for _, testCase := range specs.TestCases {
		shouldSkip := testCase.Skip || (hasFocusedTest && !testCase.Focus)
		if shouldSkip {
			specsResult.Skipped += 1
			specsResult.Cases = append(specsResult.Cases, TestCaseResult{
				It:      testCase.It,
				Skipped: true,
			})
			continue
		}

		meta := mergeAccountsMeta(specs.Meta, testCase.Meta)
		mergedBalances := mergeBalances(specs.Balances, testCase.Balances)

		vars := mergeVars(specs.Vars, testCase.Vars)

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
				Balances: mergedBalances,
			},
			featureFlags,
		)

		balances := mergedBalances

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

		var postings []interpreter.Posting
		if result != nil {
			postings = result.Postings
		}

		specsResult.Cases = append(specsResult.Cases, TestCaseResult{
			It:               testCase.It,
			Pass:             pass,
			Meta:             meta,
			Balances:         mergedBalances,
			Vars:             vars,
			FailedAssertions: failedAssertions,
			Postings:         postings,
		})
	}

	specsResult.Total = specsResult.Failing + specsResult.Passing + specsResult.Skipped

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

// validateSpecs rejects a malformed specs file before any test case is run. A
// balance list must not contain the same (account, asset, color) key twice: it's
// the map key and the amount is its value, so duplicates are ambiguous. A key
// shared across the outer and a test case's inner list is fine, since
// mergeBalances lets the inner entry override the outer one.
func validateSpecs(specs Specs) error {
	if dup, ok := specs.Balances.FirstDuplicate(); ok {
		return duplicateBalanceErr(dup)
	}
	for _, testCase := range specs.TestCases {
		if dup, ok := testCase.Balances.FirstDuplicate(); ok {
			return duplicateBalanceErr(dup)
		}
	}
	return nil
}

func duplicateBalanceErr(dup interpreter.BalanceRow) error {
	key := fmt.Sprintf("account=%q asset=%q", dup.Account, dup.Asset)
	if dup.Color != "" {
		key += fmt.Sprintf(" color=%q", dup.Color)
	}
	return fmt.Errorf("balances must not contain duplicate entries: duplicate entry for %s", key)
}

// Merge two balance inputs, deduping by (account, asset, color).
// Entries in "inner" override matching entries in "outer".
func mergeBalances(outer interpreter.Balances, inner interpreter.Balances) interpreter.Balances {
	merged := interpreter.Balances{}
	indexByKey := map[string]int{}

	addAll := func(items interpreter.Balances) {
		for _, item := range items {
			key := item.Account + "\x00" + item.Asset + "\x00" + item.Color
			if i, ok := indexByKey[key]; ok {
				merged[i] = item
			} else {
				indexByKey[key] = len(merged)
				merged = append(merged, item)
			}
		}
	}

	addAll(outer)
	addAll(inner)
	return merged
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
	// Working set keyed by account for O(1)-ish lookups.
	balances := map[string][]interpreter.AccountBalance{}

	getOrCreate := func(account, asset, color string) *big.Int {
		entries := balances[account]
		for i := range entries {
			if entries[i].Asset == asset && entries[i].Color == color {
				return entries[i].Amount
			}
		}
		amount := new(big.Int)
		balances[account] = append(entries, interpreter.AccountBalance{
			Asset:  asset,
			Color:  color,
			Amount: amount,
		})
		return amount
	}

	// Seed from the initial balances. CLONE each amount (Set, not pointer copy)
	// so the Sub/Add below never mutate the caller's *big.Int values.
	for _, row := range initialBalances {
		dst := getOrCreate(row.Account, row.Asset, row.Color)
		if row.Amount != nil {
			dst.Set(row.Amount)
		}
	}

	for _, posting := range postings {
		sourceBalance := getOrCreate(posting.Source, posting.Asset, posting.Color)
		sourceBalance.Sub(sourceBalance, posting.Amount)

		destinationBalance := getOrCreate(posting.Destination, posting.Asset, posting.Color)
		destinationBalance.Add(destinationBalance, posting.Amount)
	}

	// Flatten back to []BalanceRow, sorted for deterministic output.
	out := make(interpreter.Balances, 0)
	accounts := make([]string, 0, len(balances))
	for account := range balances {
		accounts = append(accounts, account)
	}
	sort.Strings(accounts)
	for _, account := range accounts {
		entries := balances[account]
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Asset != entries[j].Asset {
				return entries[i].Asset < entries[j].Asset
			}
			return entries[i].Color < entries[j].Color
		})
		for _, e := range entries {
			out = append(out, interpreter.BalanceRow{
				Account: account,
				Asset:   e.Asset,
				Color:   e.Color,
				Amount:  e.Amount,
			})
		}
	}
	return out
}
