package interpreter_test

import (
	"encoding/json"
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

func (c *TestCase) setVarsFromJSON(t *testing.T, str string) {
	var jsonVars map[string]string
	err := json.Unmarshal([]byte(str), &jsonVars)
	require.NoError(t, err)
	c.vars = jsonVars
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
		assert.Equal(t, err, expected.Error)
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

func TestVariables(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
		account $rider
		account $driver
		string 	$description
 		number 	$nb
 		asset 	$ass
	}
	send [$ass 999] (
		source=$rider
		destination=$driver
	)
 	set_tx_meta("description", $description)
 	set_tx_meta("ride", $nb)`)
	tc.vars = map[string]string{
		"rider":       "users:001",
		"driver":      "users:002",
		"description": "midnight ride",
		"nb":          "1",
		"ass":         "EUR/2",
	}
	tc.setBalance("users:001", "EUR/2", 1000)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "EUR/2",
				Amount:      big.NewInt(999),
				Source:      "users:001",
				Destination: "users:002",
			},
		},
		Metadata: map[string]machine.Value{
			"description": machine.String("midnight ride"),
			"ride":        machine.NewMonetaryInt(1),
		},
		Error: nil,
	}
	test(t, tc)
}

func TestVariablesJSON(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
		account $rider
		account $driver
		string 	$description
		number 	$nb
		asset 	$ass
	}
	send [$ass 999] (
		source=$rider
		destination=$driver
	)
	set_tx_meta("description", $description)
	set_tx_meta("ride", $nb)`)
	tc.setVarsFromJSON(t, `{
		"rider": "users:001",
		"driver": "users:002",
		"description": "midnight ride",
		"nb": "1",
 		"ass": "EUR/2"
	}`)
	tc.setBalance("users:001", "EUR/2", 1000)
	tc.expected = CaseResult{

		Postings: []Posting{
			{
				Asset:       "EUR/2",
				Amount:      big.NewInt(999),
				Source:      "users:001",
				Destination: "users:002",
			},
		},
		Metadata: map[string]machine.Value{
			"description": machine.String("midnight ride"),
			"ride":        machine.NewMonetaryInt(1),
		},
		Error: nil,
	}
	test(t, tc)
}

func TestSource(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
		account $balance
		account $payment
		account $seller
	}
	send [GEM 15] (
		source = {
			$balance
			$payment
		}
		destination = $seller
	)`)
	tc.setVarsFromJSON(t, `{
		"balance": "users:001",
		"payment": "payments:001",
		"seller": "users:002"
	}`)
	tc.setBalance("users:001", "GEM", 3)
	tc.setBalance("payments:001", "GEM", 12)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "GEM",
				Amount:      big.NewInt(3),
				Source:      "users:001",
				Destination: "users:002",
			},
			{
				Asset:       "GEM",
				Amount:      big.NewInt(12),
				Source:      "payments:001",
				Destination: "users:002",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestAllocation(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
		account $rider
		account $driver
	}
	send [GEM 15] (
		source = $rider
		destination = {
			80% to $driver
			8% to @a
			12% to @b
		}
	)`)
	tc.setVarsFromJSON(t, `{
		"rider": "users:001",
		"driver": "users:002"
	}`)
	tc.setBalance("users:001", "GEM", 15)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "GEM",
				Amount:      big.NewInt(13),
				Source:      "users:001",
				Destination: "users:002",
			},
			{
				Asset:       "GEM",
				Amount:      big.NewInt(1),
				Source:      "users:001",
				Destination: "a",
			},
			{
				Asset:       "GEM",
				Amount:      big.NewInt(1),
				Source:      "users:001",
				Destination: "b",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestDynamicAllocation(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
		portion $p
	}
	send [GEM 15] (
		source = @a
		destination = {
			80% to @b
			$p to @c
			remaining to @d
		}
	)`)
	tc.setVarsFromJSON(t, `{
		"p": "15%"
	}`)
	tc.setBalance("a", "GEM", 15)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "GEM",
				Amount:      big.NewInt(13),
				Source:      "a",
				Destination: "b",
			},
			{
				Asset:       "GEM",
				Amount:      big.NewInt(2),
				Source:      "a",
				Destination: "c",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

// TODO impl
func TestSendAll(t *testing.T) {
	t.Skip()

	tc := NewTestCase()
	tc.compile(t, `send [USD/2 *] (
		source = @users:001
		destination = @platform
	)`)
	tc.setBalance("users:001", "USD/2", 17)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "USD/2",
				Amount:      big.NewInt(17),
				Source:      "users:001",
				Destination: "platform",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

// TODO impl
func TestSendAllMulti(t *testing.T) {
	t.Skip()

	tc := NewTestCase()
	tc.compile(t, `send [USD/2 *] (
		source = {
		  @users:001:wallet
		  @users:001:credit
		}
		destination = @platform
	)
	`)
	tc.setBalance("users:001:wallet", "USD/2", 19)
	tc.setBalance("users:001:credit", "USD/2", 22)
	tc.expected = CaseResult{

		Postings: []Posting{
			{
				Asset:       "USD/2",
				Amount:      big.NewInt(19),
				Source:      "users:001:wallet",
				Destination: "platform",
			},
			{
				Asset:       "USD/2",
				Amount:      big.NewInt(22),
				Source:      "users:001:credit",
				Destination: "platform",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestInsufficientFunds(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
		account $balance
		account $payment
		account $seller
	}
	send [GEM 16] (
		source = {
			$balance
			$payment
		}
		destination = $seller
	)`)
	tc.setVarsFromJSON(t, `{
		"balance": "users:001",
		"payment": "payments:001",
		"seller": "users:002"
	}`)
	tc.setBalance("users:001", "GEM", 3)
	tc.setBalance("payments:001", "GEM", 12)
	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.MissingFundsErr{
			Missing: *big.NewInt(1),
			Sent:    *big.NewInt(15),
		},
	}
	test(t, tc)
}

func TestWorldSource(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [GEM 15] (
		source = {
			@a
			@world
		}
		destination = @b
	)`)
	tc.setBalance("a", "GEM", 1)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "GEM",
				Amount:      big.NewInt(1),
				Source:      "a",
				Destination: "b",
			},
			{
				Asset:       "GEM",
				Amount:      big.NewInt(14),
				Source:      "world",
				Destination: "b",
			},
		},
		Error: nil,
	}
	test(t, tc)
}
