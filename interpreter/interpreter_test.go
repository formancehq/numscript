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

	execResult, err := machine.RunProgram(*prog, testCase.vars, testCase.balances, testCase.meta)
	expected := testCase.expected
	if expected.Error != nil {
		assert.Equal(t, err, expected.Error)
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
	Postings []machine.Posting
	Metadata map[string]machine.Value
	Error    error
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

func TestSendAll(t *testing.T) {
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

func TestSendAlltMaxWhenNoAmount(t *testing.T) {
	// False negative: BigInt(0) comparation is failing
	t.Skip()

	tc := NewTestCase()
	tc.compile(t, `send [USD/2 *] (
		source = max [USD/2 5] from @src
		destination = @dest
	)
	`)
	tc.setBalance("src1", "USD/2", 0)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Source:      "src",
				Destination: "dest",
				Amount:      big.NewInt(0),
				Asset:       "USD/2",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestSendAlltMaxInSrc(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [USD/2 *] (
		source = {
		  max [USD/2 5] from @src1
		  @src2
		}
		destination = @dest
	)
	`)
	tc.setBalance("src1", "USD/2", 100)
	tc.setBalance("src2", "USD/2", 200)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "USD/2",
				Amount:      big.NewInt(5),
				Source:      "src1",
				Destination: "dest",
			},
			{
				Asset:       "USD/2",
				Amount:      big.NewInt(200),
				Source:      "src2",
				Destination: "dest",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestSendAlltMaxInDest(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [USD/2 *] (
		source = @src
		destination = {
			max [USD/2 10] to @d1
			remaining to @d2
		}
	)
	`)
	tc.setBalance("src", "USD/2", 100)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "USD/2",
				Amount:      big.NewInt(10),
				Source:      "src",
				Destination: "d1",
			},
			{
				Asset:       "USD/2",
				Amount:      big.NewInt(90),
				Source:      "src",
				Destination: "d2",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestSendAllMulti(t *testing.T) {
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

func TestNoEmptyPostings(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [GEM 2] (
		source = @world
		destination = {
			90% to @a
			10% to @b
		}
	)`)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "GEM",
				Amount:      big.NewInt(2),
				Source:      "world",
				Destination: "a",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestEmptyPostings(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [GEM *] (
		source = @foo
		destination = @bar
	)`)
	tc.setBalance("foo", "GEM", 0)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Source:      "foo",
				Destination: "bar",
				Amount:      big.NewInt(0),
				Asset:       "GEM",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestAllocateDontTakeTooMuch(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [CREDIT 200] (
		source = {
			@users:001
			@users:002
		}
		destination = {
			1/2 to @foo
			1/2 to @bar
		}
	)`)
	tc.setBalance("users:001", "CREDIT", 100)
	tc.setBalance("users:002", "CREDIT", 110)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "CREDIT",
				Amount:      big.NewInt(100),
				Source:      "users:001",
				Destination: "foo",
			},
			{
				Asset:       "CREDIT",
				Amount:      big.NewInt(100),
				Source:      "users:002",
				Destination: "bar",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestMetadata(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
		account $sale
		account $seller = meta($sale, "seller")
		portion $commission = meta($seller, "commission")
	}
	send [EUR/2 100] (
		source = $sale
		destination = {
			remaining to $seller
			$commission to @platform
		}
	)`)
	tc.setVarsFromJSON(t, `{
		"sale": "sales:042"
	}`)
	tc.meta = map[string]machine.Metadata{
		"sales:042": {
			"seller": "users:053",
		},
		"users:053": {
			"commission": "12.5%",
		},
	}
	tc.setBalance("sales:042", "EUR/2", 2500)
	tc.setBalance("users:053", "EUR/2", 500)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "EUR/2",
				Amount:      big.NewInt(88),
				Source:      "sales:042",
				Destination: "users:053",
			},
			{
				Asset:       "EUR/2",
				Amount:      big.NewInt(12),
				Source:      "sales:042",
				Destination: "platform",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestTrackBalances(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
	send [COIN 50] (
		source = @world
		destination = @a
	)
	send [COIN 100] (
		source = @a
		destination = @b
	)`)
	tc.setBalance("a", "COIN", 50)
	tc.expected = CaseResult{

		Postings: []Posting{
			{
				Asset:       "COIN",
				Amount:      big.NewInt(50),
				Source:      "world",
				Destination: "a",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(100),
				Source:      "a",
				Destination: "b",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestTrackBalances2(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
	send [COIN 50] (
		source = @a
		destination = @z
	)
	send [COIN 50] (
		source = @a
		destination = @z
	)`)
	tc.setBalance("a", "COIN", 60)
	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.MissingFundsErr{
			Missing: *big.NewInt(40),
			Sent:    *big.NewInt(10),
		},
	}
	test(t, tc)
}

func TestTrackBalances3(t *testing.T) {
	// TODO unskip this
	t.Skip()

	tc := NewTestCase()
	tc.compile(t, `send [COIN *] (
		source = @foo
		destination = {
			max [COIN 1000] to @bar
			remaining kept
		}
	)
	send [COIN *] (
		source = @foo
		destination = @bar
	)`)
	tc.setBalance("foo", "COIN", 2000)
	tc.expected = CaseResult{

		Postings: []Posting{
			{
				Asset:       "COIN",
				Amount:      big.NewInt(1000),
				Source:      "foo",
				Destination: "bar",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(1000),
				Source:      "foo",
				Destination: "bar",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestSourceAllotment(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [COIN 100] (
		source = {
			60% from @a
			35.5% from @b
			4.5% from @c
		}
		destination = @d
	)`)
	tc.setBalance("a", "COIN", 100)
	tc.setBalance("b", "COIN", 100)
	tc.setBalance("c", "COIN", 100)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "COIN",
				Amount:      big.NewInt(61),
				Source:      "a",
				Destination: "d",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(35),
				Source:      "b",
				Destination: "d",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(4),
				Source:      "c",
				Destination: "d",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestSourceOverlapping(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [COIN 99] (
		source = {
			15% from {
				@b
				@a
			}
			30% from @a
			remaining from @a
		}
		destination = @world
	)`)
	tc.setBalance("a", "COIN", 99)
	tc.setBalance("b", "COIN", 3)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "COIN",
				Amount:      big.NewInt(3),
				Source:      "b",
				Destination: "world",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(96),
				Source:      "a",
				Destination: "world",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestCappedWhenMoreThanBalance(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
	send [COIN 100] (
		source = {
			max [COIN 200] from @world
			@src
		}
		destination = @platform
	)
	`)
	tc.setBalance("src", "COIN", 1000)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "COIN",
				Amount:      big.NewInt(100),
				Source:      "world",
				Destination: "platform",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestCappedWhenLessThanNeeded(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
	send [COIN 100] (
		source = {
			max [COIN 40] from @src1
			@src2
		}
		destination = @platform
	)
	`)
	tc.setBalance("src1", "COIN", 1000)
	tc.setBalance("src2", "COIN", 1000)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "COIN",
				Amount:      big.NewInt(40),
				Source:      "src1",
				Destination: "platform",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(60),
				Source:      "src2",
				Destination: "platform",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestSourceComplex(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
		monetary $max
	}
	send [COIN 200] (
		source = {
			50% from {
				max [COIN 4] from @a
				@b
				@c
			}
			remaining from max $max from @d
		}
		destination = @platform
	)`)
	tc.setVarsFromJSON(t, `{
		"max": "COIN 120"
	}`)
	tc.setBalance("a", "COIN", 1000)
	tc.setBalance("b", "COIN", 40)
	tc.setBalance("c", "COIN", 1000)
	tc.setBalance("d", "COIN", 1000)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "COIN",
				Amount:      big.NewInt(4),
				Source:      "a",
				Destination: "platform",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(40),
				Source:      "b",
				Destination: "platform",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(56),
				Source:      "c",
				Destination: "platform",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(100),
				Source:      "d",
				Destination: "platform",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestKeptInorder(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [COIN 100] (
		source = @world
		destination = {
			max [COIN 10] kept
			remaining to @dest
		}
	)`)

	tc.expected = CaseResult{
		Postings: []Posting{
			// 10 COIN are kept
			{
				Asset:       "COIN",
				Amount:      big.NewInt(90),
				Source:      "world",
				Destination: "dest",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestDestinationComplex(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [COIN 100] (
		source = @world
		destination = {
			20% to @a
			20% kept
			60% to {
				max [COIN 10] to @b
				remaining to @c
			}
		}
	)`)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "COIN",
				Amount:      big.NewInt(20),
				Source:      "world",
				Destination: "a",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(10),
				Source:      "world",
				Destination: "b",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(50),
				Source:      "world",
				Destination: "c",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

// TODO TestNeededBalances, TestSetTxMeta, TestSetAccountMeta

func TestBalance(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
	vars {
		monetary $balance = balance(@a, EUR/2)
	}

	send $balance (
		source = @world
		destination = @dest
	)`)
	tc.setBalance("a", "EUR/2", 123)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "EUR/2",
				Amount:      big.NewInt(123),
				Source:      "world",
				Destination: "dest",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestInoderDestination(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [COIN 100] (
		source = @world
		destination = {
			max [COIN 20] to @dest1
			remaining to @dest2
		}
	)`)
	tc.setBalance("a", "COIN", 123)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "COIN",
				Amount:      big.NewInt(20),
				Source:      "world",
				Destination: "dest1",
			},
			{
				Asset:       "COIN",
				Amount:      big.NewInt(80),
				Source:      "world",
				Destination: "dest2",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

func TestVariableBalance(t *testing.T) {
	script := `
		vars {
		  monetary $initial = balance(@A, USD/2)
		}
		send [USD/2 100] (
		  source = {
			@A
			@C
		  }
		  destination = {
			max $initial to @B
			remaining to @D
		  }
		)`

	t.Run("1", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, script)
		tc.setBalance("A", "USD/2", 40)
		tc.setBalance("C", "USD/2", 90)
		tc.expected = CaseResult{
			Postings: []Posting{
				{
					Asset:       "USD/2",
					Amount:      big.NewInt(40),
					Source:      "A",
					Destination: "B",
				},
				{
					Asset:       "USD/2",
					Amount:      big.NewInt(60),
					Source:      "C",
					Destination: "D",
				},
			},
			Error: nil,
		}
		test(t, tc)
	})

	t.Run("2", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, script)
		tc.setBalance("A", "USD/2", 400)
		tc.setBalance("C", "USD/2", 90)
		tc.expected = CaseResult{
			Postings: []Posting{
				{
					Asset:       "USD/2",
					Amount:      big.NewInt(100),
					Source:      "A",
					Destination: "B",
				},
			},
			Error: nil,
		}
		test(t, tc)
	})

	script = `
		vars {
		  account $acc
		  monetary $initial = balance($acc, USD/2)
		}
		send [USD/2 100] (
		  source = {
			$acc
			@C
		  }
		  destination = {
			max $initial to @B
			remaining to @D
		  }
		)`

	t.Run("3", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, script)
		tc.setBalance("A", "USD/2", 40)
		tc.setBalance("C", "USD/2", 90)
		tc.setVarsFromJSON(t, `{"acc": "A"}`)
		tc.expected = CaseResult{
			Postings: []Posting{
				{
					Asset:       "USD/2",
					Amount:      big.NewInt(40),
					Source:      "A",
					Destination: "B",
				},
				{
					Asset:       "USD/2",
					Amount:      big.NewInt(60),
					Source:      "C",
					Destination: "D",
				},
			},
			Error: nil,
		}
		test(t, tc)
	})

	t.Run("4", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, script)
		tc.setBalance("A", "USD/2", 400)
		tc.setBalance("C", "USD/2", 90)
		tc.setVarsFromJSON(t, `{"acc": "A"}`)
		tc.expected = CaseResult{
			Postings: []Posting{
				{
					Asset:       "USD/2",
					Amount:      big.NewInt(100),
					Source:      "A",
					Destination: "B",
				},
			},
			Error: nil,
		}
		test(t, tc)
	})

	t.Run("5", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `
		vars {
			monetary $max = balance(@maxAcc, COIN)
		}
		send [COIN 200] (
			source = {
				50% from {
					max [COIN 4] from @a
					@b
					@c
				}
				remaining from max $max from @d
			}
			destination = @platform
		)`)
		tc.setBalance("maxAcc", "COIN", 120)
		tc.setBalance("a", "COIN", 1000)
		tc.setBalance("b", "COIN", 40)
		tc.setBalance("c", "COIN", 1000)
		tc.setBalance("d", "COIN", 1000)
		tc.expected = CaseResult{
			Postings: []Posting{
				{
					Asset:       "COIN",
					Amount:      big.NewInt(4),
					Source:      "a",
					Destination: "platform",
				},
				{
					Asset:       "COIN",
					Amount:      big.NewInt(40),
					Source:      "b",
					Destination: "platform",
				},
				{
					Asset:       "COIN",
					Amount:      big.NewInt(56),
					Source:      "c",
					Destination: "platform",
				},
				{
					Asset:       "COIN",
					Amount:      big.NewInt(100),
					Source:      "d",
					Destination: "platform",
				},
			},
			Error: nil,
		}
		test(t, tc)
	})

	t.Run("send negative monetary", func(t *testing.T) {
		tc := NewTestCase()
		script = `
		vars {
		  monetary $amount = balance(@world, USD/2)
		}
		send $amount (
		  source = @A
		  destination = @B
		)`
		tc.compile(t, script)
		tc.setBalance("world", "USD/2", -40)
		tc.expected = CaseResult{
			Error: machine.NegativeAmountErr{Amount: machine.NewMonetaryInt(-40)},
		}
		test(t, tc)
	})
}

// TODO TestVariablesParsing, TestSetVarsFromJSON, TestResolveResources, TestResolveBalances, TestMachine

// TODO
// func TestVariablesErrors(t *testing.T) {
// 	tc := NewTestCase()
// 	tc.compile(t, `vars {
// 		monetary $mon
// 	}
// 	send $mon (
// 		source = @alice
// 		destination = @bob
// 	)`)
// 	tc.setBalance("alice", "COIN", 10)
// 	tc.vars = map[string]string{
// 		"mon": "COIN -1",
// 	}
// 	tc.expected = CaseResult{
// 		Postings:      []Posting{},
// 		Error:         &machine.ErrInvalidVars{},
// 		ErrorContains: "negative amount",
// 	}
// 	test(t, tc)
// }

func TestVariableAsset(t *testing.T) {
	// TODO unskip
	t.Skip()

	script := `
 		vars {
 			asset $ass
 			monetary $bal = balance(@alice, $ass)
 		}

 		send [$ass 15] (
 			source = {
 				@alice
 				@bob
 			}
 			destination = @swap
 		)

 		send [$ass *] (
 			source = @swap
 			destination = {
 				max $bal to @alice_2
 				remaining to @bob_2
 			}
 		)`

	tc := NewTestCase()
	tc.compile(t, script)
	tc.vars = map[string]string{
		"ass": "USD",
	}
	tc.setBalance("alice", "USD", 10)
	tc.setBalance("bob", "USD", 10)
	tc.expected = CaseResult{
		Postings: []Posting{
			{
				Asset:       "USD",
				Amount:      big.NewInt(10),
				Source:      "alice",
				Destination: "swap",
			},
			{
				Asset:       "USD",
				Amount:      big.NewInt(5),
				Source:      "bob",
				Destination: "swap",
			},
			{
				Asset:       "USD",
				Amount:      big.NewInt(10),
				Source:      "swap",
				Destination: "alice_2",
			},
			{
				Asset:       "USD",
				Amount:      big.NewInt(5),
				Source:      "swap",
				Destination: "bob_2",
			},
		},
		Error: nil,
	}
	test(t, tc)
}

// TODO TestSaveFromAccount

func TestUseDifferentAssetsWithSameSourceAccount(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
	account $a_account
}
send [A 100] (
	source = $a_account allowing unbounded overdraft
	destination = @account1
)
send [B 100] (
	source = @world
	destination = @account2
)`)
	tc.setBalance("account1", "A", 100)
	tc.setBalance("account2", "B", 100)
	tc.setVarsFromJSON(t, `{"a_account": "world"}`)
	tc.expected = CaseResult{

		Postings: []Posting{{
			Source:      "world",
			Destination: "account1",
			Amount:      big.NewInt(100),
			Asset:       "A",
		}, {
			Source:      "world",
			Destination: "account2",
			Amount:      big.NewInt(100),
			Asset:       "B",
		}},
	}
	test(t, tc)
}

func TestMaxWithUnboundedOverdraft(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
send [COIN 100] (
	source = {
		max [COIN 10] from @account1 allowing unbounded overdraft
		@account2
	}
	destination = @world
)`)
	tc.setBalance("account1", "COIN", 10000)
	tc.setBalance("account2", "COIN", 10000)
	tc.expected = CaseResult{
		Postings: []Posting{{
			Source:      "account1",
			Destination: "world",
			Amount:      big.NewInt(10),
			Asset:       "COIN",
		}, {
			Source:      "account2",
			Destination: "world",
			Amount:      big.NewInt(90),
			Asset:       "COIN",
		}},
	}
	test(t, tc)
}

func TestErrors(t *testing.T) {
	t.Run("wrong type for send literal", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `
	send @bad:type (
		source = @a
		destination = @b
	)`)
		tc.expected = CaseResult{
			Error: machine.TypeError{
				Expected: "monetary",
				Value:    machine.AccountAddress("bad:type"),
			},
		}
		test(t, tc)
	})

	t.Run("wrong type for account literal", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `
	vars {
		number $var_src
	}

	send [COIN 10] (
		source = {
			1/2 from @world
			remaining from {
				@empty
				max [COIN 100] from $var_src
			}
		}
		destination = @b
	)`)
		tc.setVarsFromJSON(t, `{"var_src": "42"}`)

		tc.expected = CaseResult{
			Error: machine.TypeError{
				Expected: "account",
				Value:    machine.NewMonetaryInt(42),
			},
		}
		test(t, tc)
	})

	t.Run("wrong type for account cap", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `
	vars {
		string $v
	}

	send [COIN 10] (
		source = max $v from @src
		destination = @b
	)`)
		tc.setVarsFromJSON(t, `{"v": "abc"}`)

		tc.expected = CaseResult{
			Error: machine.TypeError{
				Expected: "monetary",
				Value:    machine.String("abc"),
			},
		}
		test(t, tc)
	})

	t.Run("unbound variable", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `
	send $unbound_var (
		source = @a
		destination = @b
	)`)

		tc.expected = CaseResult{
			Error: machine.UnboundVariableErr{
				Name: "unbound_var",
			},
		}
		test(t, tc)
	})

	t.Run("missing variable from json", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `
	vars {
		monetary $x
	}

	send $x (
		source = @a
		destination = @b
	)`)

		tc.expected = CaseResult{
			Error: machine.MissingVariableErr{
				Name: "x",
			},
		}
		test(t, tc)
	})

	t.Run("unbound fn", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `unbound_fn(1, 2)`)

		tc.expected = CaseResult{
			Error: machine.UnboundFunctionErr{
				Name: "unbound_fn",
			},
		}
		test(t, tc)
	})

	t.Run("unbound fn (origin)", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `
			vars {
				number $x = unbound_fn(1, 2)
			}
		`)

		tc.expected = CaseResult{
			Error: machine.UnboundFunctionErr{
				Name: "unbound_fn",
			},
		}
		test(t, tc)
	})

	t.Run("wrong fn arity", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `set_tx_meta()`)

		tc.expected = CaseResult{
			Error: machine.BadArityErr{
				ExpectedArity:  2,
				GivenArguments: 0,
			},
		}
		test(t, tc)
	})

	t.Run("wrong fn type", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `set_tx_meta(@key_wrong_type, "value")`)
		tc.expected = CaseResult{
			Error: machine.TypeError{
				Expected: "string",
				Value:    machine.AccountAddress("key_wrong_type"),
			},
		}
		test(t, tc)
	})

	t.Run("invalid variable type", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `
			vars {
				invalidt $x
			}
		`)
		tc.setVarsFromJSON(t, `{"x": "42"}`)
		tc.expected = CaseResult{
			Error: machine.InvalidTypeErr{
				Name: "invalidt",
			},
		}
		test(t, tc)
	})
}
