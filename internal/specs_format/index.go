package specs_format

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"sort"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
)

// TxMetadataRow is a single transaction metadata entry. Like SetAccountMetadataRow,
// the value's type is known, so it is carried as a typed Value written in the tagged
// value format (e.g. {"type":"account","name":"x"}).
type TxMetadataRow struct {
	Key   string            `json:"key"`
	Value interpreter.Value `json:"value"`
}

// ExpectedTxMeta is a test case's expected transaction metadata: a list of rows,
// mirroring expect.metadata. Comparison ignores order (see compareTxMeta).
type ExpectedTxMeta []TxMetadataRow

func (r *TxMetadataRow) UnmarshalJSON(data []byte) error {
	var raw struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	value, err := interpreter.ParseTaggedValue(raw.Value)
	if err != nil {
		return err
	}
	r.Key, r.Value = raw.Key, value
	return nil
}

// compareTxMeta reports whether two lists hold the same rows, ignoring order but
// respecting multiplicity (so [x, x] != [x, y]). Values are compared on their
// canonical source form, so a string "42" and the number 42 are not conflated.
func compareTxMeta(a ExpectedTxMeta, b ExpectedTxMeta) bool {
	if len(a) != len(b) {
		return false
	}
	key := func(r TxMetadataRow) string {
		value := ""
		if r.Value != nil {
			value = r.Value.String()
		}
		return r.Key + "\x00" + value
	}
	counts := make(map[string]int, len(a))
	for _, r := range a {
		counts[key(r)]++
	}
	for _, r := range b {
		k := key(r)
		counts[k]--
		if counts[k] < 0 {
			return false
		}
	}
	return true
}

// txMetaToRows flattens the interpreter's (map-based) transaction metadata into the
// row form used by expect.txMetadata, so the two can be compared.
func txMetaToRows(m interpreter.Metadata) ExpectedTxMeta {
	rows := make(ExpectedTxMeta, 0, len(m))
	for k, v := range m {
		rows = append(rows, TxMetadataRow{Key: k, Value: v})
	}
	return rows
}

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

	ExpectPostings           []interpreter.Posting           `json:"expect.postings,omitempty"`
	ExpectTxMeta             ExpectedTxMeta                  `json:"expect.txMetadata,omitempty"`
	ExpectAccountsMeta       interpreter.SetAccountsMetadata `json:"expect.metadata,omitempty"`
	ExpectEndBalances        interpreter.Balances            `json:"expect.endBalances,omitempty"`
	ExpectEndBalancesInclude interpreter.Balances            `json:"expect.endBalances.include,omitempty"`
	ExpectMovements          Movements                       `json:"expect.movements,omitempty"`
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
				failedAssertions = runAssertion(failedAssertions,
					"expect.txMetadata",
					testCase.ExpectTxMeta,
					txMetaToRows(result.Metadata),
					compareTxMeta,
				)
			}

			if testCase.ExpectAccountsMeta != nil {
				failedAssertions = runAssertion(failedAssertions,
					"expect.metadata",
					testCase.ExpectAccountsMeta,
					result.AccountsMetadata,
					interpreter.CompareSetAccountsMetadata,
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
				failedAssertions = runAssertion(failedAssertions,
					"expect.movements",
					testCase.ExpectMovements,
					getMovements(result.Postings),
					compareMovements,
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

// Merge two account-metadata inputs, deduping by (account, key, scope).
// Entries in "inner" override matching entries in "outer".
func mergeAccountsMeta(outer interpreter.AccountsMetadata, inner interpreter.AccountsMetadata) interpreter.AccountsMetadata {
	merged := interpreter.AccountsMetadata{}
	indexByKey := map[string]int{}

	addAll := func(items interpreter.AccountsMetadata) {
		for _, item := range items {
			key := item.Account + "\x00" + item.Key + "\x00" + item.Scope
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

// validateSpecs rejects a malformed specs file before any test case is run. A
// balance list must not contain the same (account, asset, color) key twice: it's
// the map key and the amount is its value, so duplicates are ambiguous. A key
// shared across the outer and a test case's inner list is fine, since
// mergeBalances lets the inner entry override the outer one.
func validateSpecs(specs Specs) error {
	if dup, ok := specs.Balances.FirstDuplicate(); ok {
		return duplicateBalanceErr(dup)
	}
	if dup, ok := specs.Meta.FirstDuplicate(); ok {
		return duplicateAccountMetaErr(dup)
	}
	for _, testCase := range specs.TestCases {
		if dup, ok := testCase.Balances.FirstDuplicate(); ok {
			return duplicateBalanceErr(dup)
		}
		if dup, ok := testCase.Meta.FirstDuplicate(); ok {
			return duplicateAccountMetaErr(dup)
		}
	}
	return nil
}

func duplicateBalanceErr(dup interpreter.BalanceRow) error {
	key := fmt.Sprintf("account=%q asset=%q", dup.Account, dup.Asset)
	if dup.Color != "" {
		key += fmt.Sprintf(" color=%q", dup.Color)
	}
	if dup.Scope != "" {
		key += fmt.Sprintf(" scope=%q", dup.Scope)
	}
	return fmt.Errorf("balances must not contain duplicate entries: duplicate entry for %s", key)
}

func duplicateAccountMetaErr(dup interpreter.AccountMetadataRow) error {
	key := fmt.Sprintf("account=%q key=%q", dup.Account, dup.Key)
	if dup.Scope != "" {
		key += fmt.Sprintf(" scope=%q", dup.Scope)
	}
	return fmt.Errorf("metadata must not contain duplicate entries: duplicate entry for %s", key)
}

// Merge two balance inputs, deduping by (account, asset, color).
// Entries in "inner" override matching entries in "outer".
func mergeBalances(outer interpreter.Balances, inner interpreter.Balances) interpreter.Balances {
	merged := interpreter.Balances{}
	indexByKey := map[string]int{}

	addAll := func(items interpreter.Balances) {
		for _, item := range items {
			key := item.Account + "\x00" + item.Asset + "\x00" + item.Color + "\x00" + item.Scope
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

type Movement struct {
	Source           string   `json:"source"`
	SourceScope      string   `json:"sourceScope,omitempty"`
	Destination      string   `json:"destination"`
	DestinationScope string   `json:"destinationScope,omitempty"`
	Asset            string   `json:"asset"`
	Amount           *big.Int `json:"amount"`
	Color            string   `json:"color,omitempty"`
}

type Movements = []Movement

// Compare movements as a set: order does not matter.
// Each (source, sourceScope, destination, destinationScope, asset, color) tuple
// is unique within a Movements list, so we match on that tuple and compare amounts.
func compareMovements(expected Movements, got Movements) bool {
	if len(expected) != len(got) {
		return false
	}

	// multiset comparison, respecting multiplicity (so [x, x] != [x, y]): the
	// amount is part of the key, so a row matches only an identical row.
	key := func(m Movement) string {
		amount := "0"
		if m.Amount != nil {
			amount = m.Amount.String()
		}
		return m.Source + "\x00" + m.SourceScope + "\x00" + m.Destination + "\x00" + m.DestinationScope + "\x00" + m.Asset + "\x00" + m.Color + "\x00" + amount
	}

	counts := make(map[string]int, len(expected))
	for _, m := range expected {
		counts[key(m)]++
	}
	for _, m := range got {
		k := key(m)
		counts[k]--
		if counts[k] < 0 {
			return false
		}
	}
	return true
}

func getMovements(postings []interpreter.Posting) Movements {
	movements := Movements{}

	for _, posting := range postings {
		found := false
		for i := range movements {
			m := &movements[i]
			if m.Source == posting.Source &&
				m.SourceScope == posting.SourceScope &&
				m.Destination == posting.Destination &&
				m.DestinationScope == posting.DestinationScope &&
				m.Asset == posting.Asset &&
				m.Color == posting.Color {
				m.Amount = new(big.Int).Add(m.Amount, posting.Amount)
				found = true
				break
			}
		}

		if !found {
			movements = append(movements, Movement{
				Source:           posting.Source,
				SourceScope:      posting.SourceScope,
				Destination:      posting.Destination,
				DestinationScope: posting.DestinationScope,
				Asset:            posting.Asset,
				Color:            posting.Color,
				Amount:           new(big.Int).Set(posting.Amount),
			})
		}
	}

	return movements
}

func getBalances(postings []interpreter.Posting, initialBalances interpreter.Balances) interpreter.Balances {
	// Working set keyed by (account, scope) for O(1)-ish lookups.
	balances := map[interpreter.AccountAddress][]interpreter.AccountBalance{}

	getOrCreate := func(account, asset, scope, color string) *big.Int {
		key := interpreter.AccountAddress{Name: account, Scope: scope}
		entries := balances[key]
		for i := range entries {
			if entries[i].Asset == asset && entries[i].Color == color {
				return entries[i].Amount
			}
		}
		amount := new(big.Int)
		balances[key] = append(entries, interpreter.AccountBalance{
			Asset:  asset,
			Color:  color,
			Amount: amount,
		})
		return amount
	}

	// Seed from the initial balances. CLONE each amount (Set, not pointer copy)
	// so the Sub/Add below never mutate the caller's *big.Int values.
	for _, row := range initialBalances {
		dst := getOrCreate(row.Account, row.Asset, row.Scope, row.Color)
		if row.Amount != nil {
			dst.Set(row.Amount)
		}
	}

	for _, posting := range postings {
		sourceBalance := getOrCreate(posting.Source, posting.Asset, posting.SourceScope, posting.Color)
		sourceBalance.Sub(sourceBalance, posting.Amount)

		destinationBalance := getOrCreate(posting.Destination, posting.Asset, posting.DestinationScope, posting.Color)
		destinationBalance.Add(destinationBalance, posting.Amount)
	}

	// Flatten back to []BalanceRow, sorted for deterministic output.
	out := make(interpreter.Balances, 0)
	accounts := make([]interpreter.AccountAddress, 0, len(balances))
	for account := range balances {
		accounts = append(accounts, account)
	}
	sort.Slice(accounts, func(i, j int) bool {
		if accounts[i].Name != accounts[j].Name {
			return accounts[i].Name < accounts[j].Name
		}
		return accounts[i].Scope < accounts[j].Scope
	})
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
				Account: account.Name,
				Asset:   e.Asset,
				Scope:   account.Scope,
				Color:   e.Color,
				Amount:  e.Amount,
			})
		}
	}
	return out
}
