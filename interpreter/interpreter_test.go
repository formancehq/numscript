package interpreter_test

import (
	"errors"
	"math/big"
	machine "numscript/interpreter"

	"numscript/parser"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	program  *parser.Program
	vars     map[string]string
	meta     map[string]machine.Metadata
	balances map[string]map[string]*big.Int
	expected CaseResult
}

func NewTestCase() TestCase {
	return TestCase{
		vars:     make(map[string]string),
		meta:     make(map[string]machine.Metadata),
		balances: make(map[string]map[string]*big.Int),
		expected: CaseResult{
			Postings: []machine.Posting{},
			Metadata: make(map[string]machine.Value),
			Error:    nil,
		},
	}
}

func (tc *TestCase) compile(t *testing.T, src string) {
	parsed := parser.Parse(src)
	if len(parsed.Errors) != 0 {
		t.Errorf("Got parsing errors: %v\n", parsed.Errors)
	}
	tc.program = &parsed.Value
}

func (c *TestCase) setBalance(account string, asset string, amount int64) {
	if _, ok := c.balances[account]; !ok {
		c.balances[account] = make(map[string]*big.Int)
	}
	c.balances[account][asset] = big.NewInt(amount)
}

func test(t *testing.T, testCase TestCase) {
	prog := testCase.program

	require.NotNil(t, prog)

	store := machine.StaticStore{}
	for account, balances := range testCase.balances {
		store[account] = &machine.AccountWithBalances{
			Balances: func() map[string]*big.Int {
				ret := make(map[string]*big.Int)
				for asset, balance := range balances {
					ret[asset] = (*big.Int)(balance)
				}
				return ret
			}(),
		}
	}

	execResult, err := machine.RunProgram(*prog, testCase.vars, store)

	expected := testCase.expected
	if expected.Error != nil {
		require.True(t, errors.Is(err, expected.Error), "got wrong error, want: %v, got: %v", expected.Error, err)
		if expected.ErrorContains != "" {
			require.ErrorContains(t, err, expected.ErrorContains)
		}
	} else {
		require.NoError(t, err)
	}
	if err != nil {
		return
	}

	if expected.Postings == nil {
		expected.Postings = make([]Posting, 0)
	}
	if expected.Metadata == nil {
		expected.Metadata = make(map[string]machine.Value)
	}

	assert.Equalf(t, expected.Postings, execResult.Postings, "unexpected postings output: %v", execResult.Postings)
	assert.Equalf(t, expected.Metadata, execResult.TxMeta, "unexpected metadata output: %v", execResult.TxMeta)

}

type CaseResult struct {
	Postings      []machine.Posting
	Metadata      map[string]machine.Value
	Error         error
	ErrorContains string
}

type Posting = machine.Posting

func TestSend(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [EUR/2 100] (
		source=@alice
		destination=@bob
	)`)
	tc.setBalance("alice", "EUR/2", 100)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "EUR/2",
				Amount:      big.NewInt(100),
				Source:      "alice",
				Destination: "bob",
			},
		},
		Error: nil,
	}
	test(t, tc)
}
