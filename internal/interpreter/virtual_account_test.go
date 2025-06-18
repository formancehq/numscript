package interpreter_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/stretchr/testify/require"
)

func TestVirtualAccountReceiveAndThenPull(t *testing.T) {

	vacc := interpreter.NewVirtualAccount()

	postings := vacc.Receive("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("src"),
		Amount:  big.NewInt(10),
	})
	require.Empty(t, postings)

	postings = vacc.Pull("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("dest"),
		Amount:  big.NewInt(10),
	})
	require.Equal(t, []Posting{
		{
			Source:      "src",
			Destination: "dest",
			Amount:      big.NewInt(10),
			Asset:       "USD",
		},
	}, postings)
}

func TestVirtualAccountReceiveAndThenPullPartialAmount(t *testing.T) {
	vacc := interpreter.NewVirtualAccount()

	postings := vacc.Receive("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("src"),
		Amount:  big.NewInt(10),
	})
	require.Empty(t, postings)

	postings = vacc.Pull("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("dest"),
		Amount:  big.NewInt(1), // <- we're only pulling 1 out of 10
	})
	require.Equal(t, []Posting{
		{
			Source:      "src",
			Destination: "dest",
			Amount:      big.NewInt(1),
			Asset:       "USD",
		},
	}, postings)
}

func TestVirtualAccountPullFirst(t *testing.T) {
	// <v> -> @dest (10 USD)
	// @src -> <v> (10 USD)
	// => [@src, @dest, 10 USD]

	vacc := interpreter.NewVirtualAccount()

	// Now we pull first. Note the unbounded overdraft
	postings := vacc.Pull("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("dest"),
		Amount:  big.NewInt(10),
	})
	// As there are no funds, no postings are emitted (yet)
	require.Empty(t, postings)

	// Now we that we're sending funds to the account, the postings of the previous ".Pull()" are emitted
	postings = vacc.Receive("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("src"),
		Amount:  big.NewInt(10),
	})
	require.Equal(t, []Posting{
		{
			Source:      "src",
			Destination: "dest",
			Amount:      big.NewInt(10),
			Asset:       "USD",
		},
	}, postings)
}

func TestVirtualAccountPullFirstMixed(t *testing.T) {
	vacc := interpreter.NewVirtualAccount()

	// 1 USD of debt
	vacc.Pull("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("lender"),
		Amount:  big.NewInt(1),
	})

	// 10 USD of credits
	postings := vacc.Receive("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("src"),
		Amount:  big.NewInt(10),
	})
	require.Equal(t, []Posting{
		{
			Source:      "src",
			Destination: "lender",
			Amount:      big.NewInt(1),
			Asset:       "USD",
		},
	}, postings)

	// pull the rest
	postings = vacc.Pull("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("dest"),
		Amount:  big.NewInt(100),
	})
	require.Equal(t, []Posting{
		{
			Source:      "src",
			Destination: "dest",
			Amount:      big.NewInt(9),
			Asset:       "USD",
		},
	}, postings)
}

func TestVirtualAccountTransitiveWhenNotOverdraft(t *testing.T) {
	amt := big.NewInt(10)

	// @src -> $v0 (10 USD)
	// $v0 -> $v1 (10 USD)
	// $v1 -> @dest (10 USD)
	// => [{@src, @dest, 10}]

	v0 := interpreter.NewVirtualAccount()
	v0.Dbg = "v0"

	v1 := interpreter.NewVirtualAccount()
	v0.Dbg = "v1"

	// @src -> $v0 (10 USD)
	require.Empty(t, v0.Receive("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("src"),
		Amount:  amt,
	}))

	// $v0 -> $v1
	require.Empty(t, v1.Receive("USD", interpreter.Sender{
		Account: v0,
		Amount:  amt,
	}))

	// $v1 -> @dest (10 USD)
	// => [{@src, @dest, 10}]

	require.Equal(t, []Posting{
		{"src", "dest", amt, "USD"},
	},
		v1.Pull("USD", interpreter.Sender{
			Account: interpreter.AccountAddress("dest"),
			Amount:  amt,
		}))
}

func TestVirtualAccountTransitiveWhenOverdraft(t *testing.T) {
	amt := big.NewInt(10)

	// $v0 -> $v1 (10 USD)
	// @src -> $v0 (10 USD)
	// $v1 -> @dest (10 USD)
	// => [{@src, @dest, 10}]

	v0 := interpreter.NewVirtualAccount()
	v1 := interpreter.NewVirtualAccount()

	// $v0 -> $v1
	require.Empty(t, v1.Receive("USD", interpreter.Sender{
		Account: v0,
		Amount:  amt,
	}))
	// @src -> $v0 (10 USD)
	require.Empty(t, v0.Receive("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("src"),
		Amount:  amt,
	}))

	// $v1 -> @dest (10 USD)
	// => [{@src, @dest, 10}]
	require.Equal(t, []Posting{
		{"src", "dest", amt, "USD"},
	}, v1.Pull("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("dest"),
		Amount:  amt,
	}))
}

func TestVirtualAccountTransitiveWhenOverdraftAndPayLast(t *testing.T) {
	amt := big.NewInt(10)

	// $v0 -> $v1 (10 USD)
	// $v1 -> @dest (10 USD)
	// @src -> $v0 (10 USD)
	// => [{@src, @dest, 10}]

	v0 := interpreter.NewVirtualAccount()
	v1 := interpreter.NewVirtualAccount()

	// $v0 -> $v1
	require.Empty(t, v1.Receive("USD", interpreter.Sender{
		Account: v0,
		Amount:  amt,
	}))

	// $v1 -> @dest (10 USD)
	require.Empty(t, v1.Pull("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("dest"),
		Amount:  amt,
	}))

	// @src -> $v0 (10 USD)
	// => [{@src, @dest, 10}]
	require.Equal(t, []Posting{
		{"src", "dest", amt, "USD"},
	}, v0.Receive("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("src"),
		Amount:  amt,
	}))
}

func TestVirtualAccountTransitiveTwoSteps(t *testing.T) {
	amt := big.NewInt(10)

	//amt=10USD
	// $v0 -> $v1
	// $v1 -> $v2
	// $v2 -> @dest

	// @src -> $v0
	// => [{@src, @dest, 10}]

	v0 := interpreter.NewVirtualAccount()
	v1 := interpreter.NewVirtualAccount()
	v2 := interpreter.NewVirtualAccount()

	// $v0 -> $v1
	require.Empty(t, v1.Receive("USD", interpreter.Sender{
		Account: v0,
		Amount:  amt,
	}))
	// $v1 -> $v2
	require.Empty(t, v2.Receive("USD", interpreter.Sender{
		Account: v1,
		Amount:  amt,
	}))

	// $v2 -> @dest
	require.Empty(t, v2.Pull("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("dest"),
		Amount:  amt,
	}))

	// @src -> $v0
	// => [{@src, @dest, 10}]
	require.Equal(t, []Posting{
		{"src", "dest", amt, "USD"},
	}, v0.Receive("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("src"),
		Amount:  amt,
	}))
}

func TestVirtualAccountTransitiveTwoStepsPayFirst(t *testing.T) {
	amt := big.NewInt(10)

	//amt=10USD
	// @src -> $v0
	// $v0 -> $v1
	// $v1 -> $v2
	// $v2 -> @dest
	// => [{@src, @dest, 10}]

	v0 := interpreter.NewVirtualAccount()
	v1 := interpreter.NewVirtualAccount()
	v2 := interpreter.NewVirtualAccount()

	// @src -> $v0
	require.Empty(t, v0.Receive("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("src"),
		Amount:  amt,
	}))

	// $v0 -> $v1
	require.Empty(t, v1.Receive("USD", interpreter.Sender{
		Account: v0,
		Amount:  amt,
	}))

	// $v1 -> $v2
	require.Empty(t, v2.Receive("USD", interpreter.Sender{
		Account: v1,
		Amount:  amt,
	}))

	// $v2 -> @dest
	require.Equal(t, []Posting{
		{"src", "dest", amt, "USD"},
	}, v2.Pull("USD", interpreter.Sender{
		Account: interpreter.AccountAddress("dest"),
		Amount:  amt,
	}))

}

func TestCommutativeOrder(t *testing.T) {
	amt := big.NewInt(10)

	//amt=10USD
	// @src -> $v0
	// $v0 -> $v1
	// $v1 -> $v2
	// $v2 -> @dest
	// => [{@src, @dest, 10}]

	var v0 interpreter.VirtualAccount
	var v1 interpreter.VirtualAccount
	var v2 interpreter.VirtualAccount

	ops := []func() []Posting{
		func() []Posting {
			return v0.Receive("USD", interpreter.Sender{
				Account: interpreter.AccountAddress("src"),
				Amount:  amt,
			})
		},
		func() []Posting {
			return v1.Receive("USD", interpreter.Sender{
				Account: v0,
				Amount:  amt,
			})
		},
		func() []Posting {
			return v2.Receive("USD", interpreter.Sender{
				Account: v1,
				Amount:  amt,
			})
		},
		func() []Posting {
			return v2.Pull("USD", interpreter.Sender{
				Account: interpreter.AccountAddress("dest"),
				Amount:  amt,
			})
		},
	}
	permutations := permute(len(ops))

	for _, permutation := range permutations {
		t.Run(fmt.Sprintf("permutation [%v]", permutation), func(t *testing.T) {
			v0 = interpreter.NewVirtualAccount().WithDbg("v0")
			v1 = interpreter.NewVirtualAccount().WithDbg("v1")
			v2 = interpreter.NewVirtualAccount().WithDbg("v2")

			for permIndex, index := range permutation {
				op := ops[index]
				postings := op()
				isLast := permIndex == len(ops)-1
				if isLast {
					require.Equal(t, []Posting{
						{"src", "dest", amt, "USD"},
					}, postings)
				} else {
					require.Empty(t, postings)

				}

			}

		})
	}
}

// gpt-generated (I was too lazy to write that)
func permute(n int) [][]int {
	var res [][]int
	used := make([]bool, n)
	var path []int

	var backtrack func()
	backtrack = func() {
		if len(path) == n {
			perm := make([]int, n)
			copy(perm, path)
			res = append(res, perm)
			return
		}
		for i := 0; i < n; i++ {
			if used[i] {
				continue
			}
			used[i] = true
			path = append(path, i)
			backtrack()
			path = path[:len(path)-1]
			used[i] = false
		}
	}

	backtrack()
	return res
}
