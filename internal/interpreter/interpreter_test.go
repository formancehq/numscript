package interpreter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"

	"github.com/formancehq/numscript/internal/flags"
	machine "github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/specs_format"

	"testing"

	"github.com/formancehq/numscript/internal/parser"

	"github.com/stretchr/testify/require"
)

const scriptsFolder = "testdata/script-tests"

func TestScripts(t *testing.T) {
	rawSpecs, err := specs_format.ReadSpecsFiles([]string{scriptsFolder})
	require.Nil(t, err)

	var buf bytes.Buffer
	buf.WriteByte('\n')
	ok := specs_format.RunSpecs(&buf, &buf, rawSpecs)
	if !ok {
		t.Log(buf.String())
		t.Fail()
	}
}

type TestCase struct {
	source   string
	program  *parser.Program
	vars     map[string]string
	meta     machine.AccountsMetadata
	balances map[string]map[string]*big.Int
	expected CaseResult
}

func NewTestCase() TestCase {
	return TestCase{
		vars:     make(map[string]string),
		meta:     machine.AccountsMetadata{},
		balances: make(map[string]map[string]*big.Int),
		expected: CaseResult{
			Error: nil,
		},
	}
}

// returns a version of the error in which the range is normalized
// to golang's default value
func removeRange(e machine.InterpreterError) machine.InterpreterError {
	switch e := e.(type) {
	case machine.MissingFundsErr:
		e.Range = parser.Range{}
		return e
	case machine.TypeError:
		e.Range = parser.Range{}
		return e
	case machine.InvalidTypeErr:
		e.Range = parser.Range{}
		return e
	case machine.NegativeAmountErr:
		e.Range = parser.Range{}
		return e
	default:
		return e
	}
}

func (c *TestCase) setVarsFromJSON(t *testing.T, str string) {
	var jsonVars map[string]string
	err := json.Unmarshal([]byte(str), &jsonVars)
	require.NoError(t, err)
	c.vars = jsonVars
}

func (tc *TestCase) compile(t *testing.T, src string) string {
	t.Parallel()

	tc.source = src
	parsed := parser.Parse(src)
	if len(parsed.Errors) != 0 {
		t.Errorf("Got parsing errors: %v\n", parsed.Errors)
	}
	tc.program = &parsed.Value
	return src
}

func (c *TestCase) setBalance(account string, asset string, amount int64) {
	if _, ok := c.balances[account]; !ok {
		c.balances[account] = make(map[string]*big.Int)
	}
	c.balances[account][asset] = big.NewInt(amount)
}

func test(t *testing.T, testCase TestCase) {
	testWithFeatureFlag(t, testCase, "")
}

// A version of test() which tests code under a feature flag
// if the feature flag is the empty string, it behaves as test()
// otherwise, it tests the program under that feature flag and also tests that
// the same script, without the flag, yields the ExperimentalFeature{} error
func testWithFeatureFlag(t *testing.T, testCase TestCase, flagName string) {
	prog := testCase.program

	require.NotNil(t, prog)

	featureFlags := map[string]struct{}{}
	if flagName != "" {
		featureFlags[flagName] = struct{}{}

		_, err := machine.RunProgram(
			context.Background(),
			*prog,
			testCase.vars,
			machine.StaticStore{
				testCase.balances,
				testCase.meta,
			},
			nil,
		)

		require.Equal(t, machine.ExperimentalFeature{
			FlagName: flagName,
		}, removeRange(err))
	}

	_, err := machine.RunProgram(
		context.Background(),
		*prog,
		testCase.vars,
		machine.StaticStore{
			testCase.balances,
			testCase.meta,
		},
		featureFlags,
	)

	expected := testCase.expected
	if expected.Error != nil {
		require.Equal(t, removeRange(expected.Error), removeRange(err))
	} else {
		require.NoError(t, err)
	}
}

func TestStaticStore(t *testing.T) {
	store := machine.StaticStore{
		Balances: machine.Balances{
			"a": machine.AccountBalance{
				"USD/2": big.NewInt(10),
				"EUR/2": big.NewInt(1),
			},
			"b": machine.AccountBalance{
				"USD/2": big.NewInt(10),
				"COIN":  big.NewInt(11),
			},
		},
	}

	q1, _ := store.GetBalances(context.TODO(), machine.BalanceQuery{
		"a": []string{"USD/2"},
	})
	require.Equal(t, machine.Balances{
		"a": machine.AccountBalance{
			"USD/2": big.NewInt(10),
		},
	}, q1)

	q2, _ := store.GetBalances(context.TODO(), machine.BalanceQuery{
		"b": []string{"USD/2", "COIN"},
	})
	require.Equal(t, machine.Balances{
		"b": machine.AccountBalance{
			"USD/2": big.NewInt(10),
			"COIN":  big.NewInt(11),
		},
	}, q2)
}

type CaseResult struct {
	Postings        []machine.Posting
	TxMetadata      map[string]machine.Value
	AccountMetadata machine.AccountsMetadata
	Error           machine.InterpreterError
}

type Posting = machine.Posting

func TestBadPortionSyntax(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `vars {
		portion $por
	}
	send [COIN 3] (
		source = @world
		destination = {
			$por to @a
			remaining kept
		}
	)
	`)
	tc.setVarsFromJSON(t, `{
		"por": "not a portion"
	}`)
	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.BadPortionParsingErr{
			Source: "not a portion",
			Reason: "invalid format",
		},
	}
	test(t, tc)
}

func TestInvalidAllotInSendAll(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [USD/2 *] (
		source = {
			1/2 from @a
			2/3 from @b
		}
		destination = @dest
	)`)
	tc.expected = CaseResult{
		Error: machine.InvalidAllotmentInSendAll{},
	}
	test(t, tc)
}

func TestDivByZero(t *testing.T) {
	tc := NewTestCase()
	src := tc.compile(t, `set_tx_meta("k", 3/0)`)
	tc.expected = CaseResult{
		Error: machine.DivideByZero{
			Numerator: big.NewInt(3),
			Range:     parser.RangeOfIndexed(src, "3/0", 0),
		},
	}
	test(t, tc)
}

func TestInvalidUnboundedWorldInSendAll(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [USD/2 *] (
		source = @world
		destination = @dest
	)`)
	tc.expected = CaseResult{
		Error: machine.InvalidUnboundedInSendAll{Name: "world"},
	}
	test(t, tc)
}

func TestInvalidUnboundedInSendAll(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [USD/2 *] (
		source = @a allowing unbounded overdraft
		destination = @dest
	)`)
	tc.expected = CaseResult{
		Error: machine.InvalidUnboundedInSendAll{Name: "a"},
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
			Asset:     "GEM",
			Needed:    *big.NewInt(16),
			Available: *big.NewInt(15),
		},
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
			Asset:     "COIN",
			Needed:    *big.NewInt(50),
			Available: *big.NewInt(10),
		},
	}
	test(t, tc)
}

func TestInvalidSourceAllotmentSum(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [COIN 100] (
		source = {
			42% from @world
		}
		destination = @dest
	)`)

	tc.expected = CaseResult{
		Error: machine.InvalidAllotmentSum{
			ActualSum: *big.NewRat(42, 100),
		},
	}
	test(t, tc)
}

func TestInvalidDestinationAllotmentSum(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [COIN 100] (
		source = @world
		destination = {
			1/4 to @x
		}
	)`)

	tc.expected = CaseResult{
		Error: machine.InvalidAllotmentSum{
			ActualSum: *big.NewRat(1, 4),
		},
	}
	test(t, tc)
}

func TestSourceAllotmentInvalidAmt(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `send [COIN 100] (
		source = {
			// a doesn't have enough amount
			10% from @a

			// world has, but the computation has already failed
			remaining from @world
		}
		destination = @d
	)`)
	tc.setBalance("a", "COIN", 1)
	tc.expected = CaseResult{
		Error: machine.MissingFundsErr{
			Asset:     "COIN",
			Needed:    *big.NewInt(10),
			Available: *big.NewInt(1),
		},
	}
	test(t, tc)
}

func TestNegativeBalance(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
	vars {
		monetary $balance = balance(@a, EUR/2)
	}

	send $balance (
		source = @world
		destination = @dest
	)`)
	tc.setBalance("a", "EUR/2", -100)
	tc.expected = CaseResult{
		Error: machine.NegativeBalanceError{
			Account: "a",
			Amount:  *big.NewInt(-100),
		},
	}
	test(t, tc)
}

func TestNegativeBalanceLiteral(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
	send [EUR/2 -100] (
		source = @world
		destination = @dest
	)`)
	tc.expected = CaseResult{
		Error: machine.NegativeAmountErr{
			Amount: machine.MonetaryInt(*big.NewInt(-100)),
		},
	}
	test(t, tc)
}

// TODO TestVariablesParsing, TestSetVarsFromJSON, TestResolveResources, TestResolveBalances, TestMachine

func TestOverdraftBadCurrency(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
send [COIN 100] (
 source = @users:1234 allowing overdraft up to [WRONGCURR 100]
 destination = @dest
)
`)
	tc.expected = CaseResult{
		Error: machine.MismatchedCurrencyError{
			Expected: "COIN",
			Got:      "WRONGCURR",
		},
	}
	test(t, tc)
}

func TestOverdraftWhenNotEnoughFunds(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
send [COIN 100] (
 source = @users:1234 allowing overdraft up to [COIN 10]
 destination = @dest
)
`)

	tc.setBalance("users:1234", "COIN", 1)

	tc.expected = CaseResult{
		Error: machine.MissingFundsErr{
			Asset:     "COIN",
			Needed:    *big.NewInt(100),
			Available: *big.NewInt(11),
		},
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
				Name:  "unbound_var",
				Range: parser.RangeOfIndexed(tc.source, "$unbound_var", 0),
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

	t.Run("bad currency type in max (source)", func(t *testing.T) {
		tc := NewTestCase()
		tc.compile(t, `
			send [EUR/2 1] (
				source = max [USD/2 10] from @world
				destination = @b
			)
		`)
		tc.expected = CaseResult{
			Error: machine.MismatchedCurrencyError{
				Expected: "EUR/2",
				Got:      "USD/2",
			},
		}
		test(t, tc)
	})
}

func TestTrackBalancesTricky(t *testing.T) {
	t.Skip()

	tc := NewTestCase()
	tc.setBalance("src", "COIN", 5)
	tc.compile(t, `
	send [COIN 25] ( // send 10 + 15
		source= {
			max [COIN 10] from @world
			@src // src only has 5 before the program starts
		}
		destination = {
			max [COIN 10] to @src
			remaining to @dest // but @src needs to send 15 here
		}
	)
	`)
	tc.expected = CaseResult{
		Postings: []machine.Posting{
			{
				"world",
				"src",
				big.NewInt(10),
				"GEM",
			},
			{
				"src",
				"dest",
				big.NewInt(15),
				"GEM",
			},
		},
	}
	test(t, tc)
}

func TestInvalidNumberLiteral(t *testing.T) {
	script := `
 		vars { number $amt }

 		send [$amt USD/2] (
 			source = @world
 			destination = @dest
 		)
	`

	tc := NewTestCase()
	tc.compile(t, script)

	tc.vars = map[string]string{
		"amt": "not a number",
	}

	tc.expected = CaseResult{
		Postings: []Posting{},
		Error:    machine.InvalidNumberLiteral{Range: parser.Range{}, Source: "not a number"},
	}
	test(t, tc)
}

func TestSaveFromAccount(t *testing.T) {

	t.Run("save causes failure", func(t *testing.T) {
		script := `
 			save [USD/2 1] from @alice

 			send [USD/2 30] (
 			   source = @alice
 			   destination = @bob
 			)`
		tc := NewTestCase()
		tc.compile(t, script)
		tc.setBalance("alice", "USD/2", 30)
		tc.expected = CaseResult{
			Postings: []Posting{},
			Error: machine.MissingFundsErr{
				Asset:     "USD/2",
				Needed:    *big.NewInt(30),
				Available: *big.NewInt(29),
			},
		}
		test(t, tc)
	})

	t.Run("negative amount", func(t *testing.T) {
		script := `
	
			save [USD -100] from @A`
		tc := NewTestCase()
		tc.compile(t, script)
		tc.setBalance("A", "USD", -100)
		tc.expected = CaseResult{
			Postings: []Posting{},
			Error: machine.NegativeAmountErr{
				Amount: machine.NewMonetaryInt(-100),
			},
		}
		test(t, tc)
	})
}

func TestAddNumbersInvalidRightType(t *testing.T) {
	script := `
 		set_tx_meta("k", 1 + "not a number")
	`

	tc := NewTestCase()
	tc.compile(t, script)

	tc.expected = CaseResult{
		Error: machine.TypeError{
			Expected: "number",
			Value:    machine.String("not a number"),
		},
	}
	test(t, tc)
}

func TestAddMonetariesDifferentCurrencies(t *testing.T) {
	script := `
 		send [USD/2 1] + [EUR/2 2] (
 			source = @world
 			destination = @dest
 		)
	`

	tc := NewTestCase()
	tc.compile(t, script)

	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.MismatchedCurrencyError{
			Expected: "USD/2",
			Got:      "EUR/2",
		},
	}
	test(t, tc)
}

func TestAddInvalidLeftType(t *testing.T) {
	script := `
 		set_tx_meta("k", EUR/2 + EUR/3)
	`

	tc := NewTestCase()
	tc.compile(t, script)

	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.TypeError{
			Expected: "monetary|number",
			Value:    machine.Asset("EUR/2"),
		},
	}
	test(t, tc)
}

func TestOneofAllFailing(t *testing.T) {
	tc := NewTestCase()
	tc.compile(t, `
	send [GEM 1] (
		source = oneof {
			@empty1
			@empty2
			@empty3
		}
		destination = @dest
	)
	`)
	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.MissingFundsErr{
			Asset:     "GEM",
			Needed:    *big.NewInt(1),
			Available: *big.NewInt(0),
		},
	}
	testWithFeatureFlag(t, tc, flags.ExperimentalOneofFeatureFlag)
}

func TestInvalidAccount(t *testing.T) {
	script := `
		vars {
			account $acc
		}
 		set_tx_meta("k", $acc)
	`

	tc := NewTestCase()
	tc.setVarsFromJSON(t, `
		{
			"acc": "!invalid acc.."
		}
	`)

	tc.compile(t, script)

	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.InvalidAccountName{
			Name: "!invalid acc..",
		},
	}
	test(t, tc)
}

func TestInvalidInterpAccount(t *testing.T) {
	script := `
		vars {
			string $status
		}
 		set_tx_meta("k", @user:$status)
	`

	tc := NewTestCase()
	tc.setVarsFromJSON(t, `
		{
			"status": "!invalid acc.."
		}
	`)

	tc.compile(t, script)

	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.InvalidAccountName{
			Name: "user:!invalid acc..",
		},
	}
	testWithFeatureFlag(t, tc, flags.ExperimentalAccountInterpolationFlag)
}

func TestAccountInvalidString(t *testing.T) {
	script := `
		vars {
			monetary $m
		}
 		set_tx_meta("k", @acc:$m)
	`

	tc := NewTestCase()
	tc.setVarsFromJSON(t, `
		{
			"m": "USD/2 10"
		}
	`)

	tc.compile(t, script)

	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.CannotCastToString{
			Range: parser.RangeOfIndexed(script, "@acc:$m", 0),
			Value: machine.Monetary{
				Amount: machine.NewMonetaryInt(10),
				Asset:  machine.Asset("USD/2"),
			},
		},
	}
	testWithFeatureFlag(t, tc, flags.ExperimentalAccountInterpolationFlag)
}

func TestInvalidNestedMetaCall(t *testing.T) {
	script := `
		vars {
			number $x = 1 + meta(@acc, "k")
		}
	`

	tc := NewTestCase()
	tc.meta = machine.AccountsMetadata{
		"acc": {
			"k": "42",
		},
	}
	tc.compile(t, script)

	tc.expected = CaseResult{
		Error: machine.InvalidNestedMeta{},
	}

	testWithFeatureFlag(t, tc, flags.ExperimentalMidScriptFunctionCall)
}

func TestColorRestrictBalanceWhenMissingFunds(t *testing.T) {
	script := `
 		send [COIN 20] (
			source = @acc \ "RED"
			destination = @dest
		)
	`

	tc := NewTestCase()
	tc.setBalance("acc", "COIN", 100)
	tc.setBalance("acc", "COIN_RED", 1)
	tc.compile(t, script)

	tc.expected = CaseResult{
		Postings: []Posting{},
		Error: machine.MissingFundsErr{
			Needed:    *big.NewInt(20),
			Available: *big.NewInt(1),
			Asset:     "COIN",
		},
	}
	testWithFeatureFlag(t, tc, flags.ExperimentalAssetColors)
}

func TestSafeMaxWithdraft(t *testing.T) {
	require.Equal(t, big.NewInt(0), machine.CalculateMaxSafeWithdraw(
		big.NewInt(0),
		big.NewInt(0),
	))

	require.Equal(t, big.NewInt(200), machine.CalculateMaxSafeWithdraw(
		big.NewInt(100),
		big.NewInt(100),
	))

	require.Equal(t, big.NewInt(105), machine.CalculateMaxSafeWithdraw(
		big.NewInt(100),
		big.NewInt(5),
	))

	require.Equal(t, big.NewInt(0), machine.CalculateMaxSafeWithdraw(
		big.NewInt(-10),
		big.NewInt(0),
	))

	require.Equal(t, big.NewInt(0), machine.CalculateMaxSafeWithdraw(
		big.NewInt(-10),
		big.NewInt(5),
	))

	require.Equal(t, big.NewInt(0), machine.CalculateMaxSafeWithdraw(
		big.NewInt(-10),
		big.NewInt(10),
	))

	require.Equal(t, big.NewInt(1), machine.CalculateMaxSafeWithdraw(
		big.NewInt(-10),
		big.NewInt(11),
	))
}

// TODO this should be a fuzz test instead
func TestSafeWithdraft(t *testing.T) {
	t.Run("with zero overdraft, only take what's available", func(t *testing.T) {
		t.Run("balance > 0 allows you to take what's available", func(t *testing.T) {
			require.Equal(t, big.NewInt(10), machine.CalculateSafeWithdraw(
				big.NewInt(100),
				big.NewInt(0),
				big.NewInt(10),
			))
			require.Equal(t, big.NewInt(10), machine.CalculateSafeWithdraw(
				big.NewInt(10),
				big.NewInt(0),
				big.NewInt(10),
			))
			require.Equal(t, big.NewInt(1), machine.CalculateSafeWithdraw(
				big.NewInt(1),
				big.NewInt(0),
				big.NewInt(10),
			))
			require.Equal(t, big.NewInt(0), machine.CalculateSafeWithdraw(
				big.NewInt(10),
				big.NewInt(0),
				big.NewInt(0),
			))

			// not enough balance:
			require.Equal(t, big.NewInt(10), machine.CalculateSafeWithdraw(
				big.NewInt(10),
				big.NewInt(0),
				big.NewInt(100),
			))

		})

		t.Run("balance == 0 doesn't let you take anything", func(t *testing.T) {
			require.Equal(t, big.NewInt(0), machine.CalculateSafeWithdraw(
				big.NewInt(0),
				big.NewInt(0),
				big.NewInt(0),
			))
		})

		t.Run("balance < 0 doesn't let you take anything if there's no overdraft", func(t *testing.T) {
			require.Equal(t, big.NewInt(0), machine.CalculateSafeWithdraw(
				big.NewInt(-100),
				big.NewInt(0),
				big.NewInt(10),
			))
			require.Equal(t, big.NewInt(0), machine.CalculateSafeWithdraw(
				big.NewInt(0),
				big.NewInt(0),
				big.NewInt(10),
			))
			require.Equal(t, big.NewInt(0), machine.CalculateSafeWithdraw(
				big.NewInt(-1),
				big.NewInt(0),
				big.NewInt(10),
			))
			require.Equal(t, big.NewInt(0), machine.CalculateSafeWithdraw(
				big.NewInt(-10),
				big.NewInt(0),
				big.NewInt(0),
			))
		})
	})

	t.Run("when overdraft is not zero, you can go over your balance", func(t *testing.T) {
		t.Run("if we have enough balance>=requestedAmount, overdraft is ignored matter", func(t *testing.T) {
			require.Equal(t, big.NewInt(10), machine.CalculateSafeWithdraw(
				big.NewInt(100),
				big.NewInt(100),
				big.NewInt(10),
			))
			require.Equal(t, big.NewInt(100), machine.CalculateSafeWithdraw(
				big.NewInt(100),
				big.NewInt(42),
				big.NewInt(100),
			))
		})

		t.Run("if we have zero balance, overdraft allows us to withdraw", func(t *testing.T) {
			require.Equal(t, big.NewInt(10), machine.CalculateSafeWithdraw(
				big.NewInt(0),
				big.NewInt(100),
				big.NewInt(10),
			))
		})

	})

}
