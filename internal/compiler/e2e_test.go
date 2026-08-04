package compiler_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// e2eStore is a minimal vm.Store for the end-to-end test.
type e2eStore struct {
	balances map[runtime.PairKey]*big.Int
	metadata map[string]map[string]string
}

func (s e2eStore) GetBalance(ctx context.Context, account, asset, color string) (*big.Int, error) {
	if v, ok := s.balances[runtime.PairKey{Account: account, Asset: asset, Color: color}]; ok {
		return v, nil
	}
	return new(big.Int), nil
}

func (s e2eStore) GetMetadata(ctx context.Context, account, key string) (string, bool, error) {
	v, ok := s.metadata[account][key]
	return v, ok, nil
}

// TestE2E_CompileAssembleRun exercises the whole pipeline: source -> compiler
// (IR) -> assembler (vm.Program) -> VM execution -> postings.
func TestE2E_CompileAssembleRun(t *testing.T) {
	src := `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}
	requirePostingsEqual(t, want, res.Postings)
}

// TestE2E_Inorder exercises an inorder source { @a @b @c } end-to-end, including
// the early-exit jump: @a has 6, @b has 10, @c has 100; sending 10 pulls 6 from
// @a (cap -> 4), then 4 from @b (cap -> 0 -> jump past @c). @c is never touched.
func TestE2E_Inorder(t *testing.T) {
	src := `
		send [USD/2 10] (
			source = {
				@a
				@b
				@c
			}
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(6),
		{Account: "b", Asset: "USD/2", Color: ""}: big.NewInt(10),
		{Account: "c", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(6)},
		{Source: "b", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(4)},
	}
	requirePostingsEqual(t, want, res.Postings)
}

// TestE2E_InorderWithCap exercises a capped (`max`) source inside an inorder
// end-to-end. @b holds 100 but is capped at 5, so the cap must bind: @a gives 3
// (remaining 10->7), @b gives only 5 (not 7) -> remaining 2, @c gives 2.
func TestE2E_InorderWithCap(t *testing.T) {
	src := `
		send [USD/2 10] (
			source = {
				@a
				max [USD/2 5] from @b
				@c
			}
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(3),
		{Account: "b", Asset: "USD/2", Color: ""}: big.NewInt(100),
		{Account: "c", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(3)},
		{Source: "b", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(5)},
		{Source: "c", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(2)},
	}
	requirePostingsEqual(t, want, res.Postings)
}

// TestE2E_InsufficientFunds checks the failure path: when the source can't cover
// the sent amount, the VM's CheckEnoughFunds must report a MissingFundsError.
func TestE2E_InsufficientFunds(t *testing.T) {
	src := `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	// src only has 4, but 10 is required.
	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(4),
	}}

	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.IsType(t, vm.MissingFundsError{}, execErr)
}

// TestE2E_DestinationInorder exercises a destination-inorder split end-to-end:
// 100 pulled from @world is distributed as `max [USD/2 30] to @x; remaining to
// @y`, so @x must get 30 and @y the remaining 70.
func TestE2E_DestinationInorder(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				max [USD/2 30] to @x
				remaining to @y
			}
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{}}

	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "world", Destination: "x", Asset: "USD/2", Amount: big.NewInt(30)},
		{Source: "world", Destination: "y", Asset: "USD/2", Amount: big.NewInt(70)},
	}
	requirePostingsEqual(t, want, res.Postings)
}

// TestE2E_DestinationKept exercises a `kept` clause: of 100 pulled from @world,
// 30 is kept (refunded, no posting) and the remaining 70 goes to @y.
func TestE2E_DestinationKept(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				max [USD/2 30] kept
				remaining to @y
			}
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{}}

	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.Nil(t, execErr)

	// only the remaining 70 is posted; the kept 30 produces no posting
	want := []runtime.Posting{
		{Source: "world", Destination: "y", Asset: "USD/2", Amount: big.NewInt(70)},
	}
	requirePostingsEqual(t, want, res.Postings)
}

// TestE2E_DestinationAllotment splits the pulled amount by portions. 100 from
// @world with { 1/2 to @a; remaining to @b } => a=50, b=50.
func TestE2E_DestinationAllotment(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				1/2 to @a
				remaining to @b
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "a", Asset: "USD/2", Amount: big.NewInt(50)},
		{Source: "world", Destination: "b", Asset: "USD/2", Amount: big.NewInt(50)},
	}, postings)
}

// TestE2E_DestinationAllotmentThirds exercises the floor-then-distribute-leftover
// rounding: 100 by thirds => 34, 33, 33 (the leftover unit goes to the earliest).
func TestE2E_DestinationAllotmentThirds(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				1/3 to @a
				1/3 to @b
				remaining to @c
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "a", Asset: "USD/2", Amount: big.NewInt(34)},
		{Source: "world", Destination: "b", Asset: "USD/2", Amount: big.NewInt(33)},
		{Source: "world", Destination: "c", Asset: "USD/2", Amount: big.NewInt(33)},
	}, postings)
}

// TestE2E_SourceAllotment splits the requested amount across sub-sources, pulling
// each exactly. 100 with { 1/4 from @s1; remaining from @s2 } => 25 from s1, 75
// from s2.
func TestE2E_SourceAllotment(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = {
				1/4 from @s1
				remaining from @s2
			}
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "s1", Asset: "USD/2", Color: ""}: big.NewInt(1000),
		{Account: "s2", Asset: "USD/2", Color: ""}: big.NewInt(1000),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "s1", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(25)},
		{Source: "s2", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(75)},
	}, postings)
}

// TestE2E_SourceAllotmentThirds checks the rounding split on the source side too:
// 100 by thirds => 34, 33, 33.
func TestE2E_SourceAllotmentThirds(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = {
				1/3 from @a
				1/3 from @b
				remaining from @c
			}
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(1000),
		{Account: "b", Asset: "USD/2", Color: ""}: big.NewInt(1000),
		{Account: "c", Asset: "USD/2", Color: ""}: big.NewInt(1000),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(34)},
		{Source: "b", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(33)},
		{Source: "c", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(33)},
	}, postings)
}

// TestE2E_SourceAllotmentInsufficient: a sub-source must provide its exact share,
// else MissingFunds. s1 only has 10 but its 1/2 share of 100 is 50.
func TestE2E_SourceAllotmentInsufficient(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = {
				1/2 from @s1
				remaining from @s2
			}
			destination = @dest
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "s1", Asset: "USD/2", Color: ""}: big.NewInt(10),
		{Account: "s2", Asset: "USD/2", Color: ""}: big.NewInt(1000),
	}}
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.IsType(t, vm.MissingFundsError{}, execErr)
}

// TestE2E_AllotmentOverSum: portions summing to > 1 must error (leftover < 0).
func TestE2E_AllotmentOverSum(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				2/3 to @a
				2/3 to @b
			}
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	require.IsType(t, vm.InvalidAllotmentSum{}, execErr)
	allotErr := execErr.(vm.InvalidAllotmentSum)
	require.Equal(t, "4/3", allotErr.ActualSum.String())
	require.EqualError(t, allotErr, "invalid allotment: portions must sum to 1, got 4/3")
}

// TestE2E_AllotmentUnderSum: without a `remaining` clause the portions must sum
// to exactly 1, so 1/3 + 1/3 = 2/3 must error.
func TestE2E_AllotmentUnderSum(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				1/3 to @a
				1/3 to @b
			}
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	require.IsType(t, vm.InvalidAllotmentSum{}, execErr)
}

// TestE2E_AllotmentExactNoRemaining: a no-remaining allotment summing to exactly
// 1 is valid.
func TestE2E_AllotmentExactNoRemaining(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				1/4 to @a
				3/4 to @b
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "a", Asset: "USD/2", Amount: big.NewInt(25)},
		{Source: "world", Destination: "b", Asset: "USD/2", Amount: big.NewInt(75)},
	}, postings)
}

// TestE2E_AllotmentRemainingOnly: `{ remaining to @dest }` is 100% (leftover = 1),
// which must remain valid (a `< 1` check would wrongly reject it).
func TestE2E_AllotmentRemainingOnly(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				remaining to @dest
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(100)},
	}, postings)
}

func TestE2E_IntAddition(t *testing.T) {
	src := `
		send [USD/2 4 + 6] (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	res, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, res.Postings)
}

func TestE2E_IntSubtraction(t *testing.T) {
	src := `
		send [USD/2 16 - 6] (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	res, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, res.Postings)
}

func TestE2E_MonetaryAddition(t *testing.T) {
	src := `
		vars {
			monetary $a = [USD/2 3]
			monetary $b = [USD/2 7]
		}
		send $a + $b (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	res, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, res.Postings)
}

func TestE2E_MonetarySubtraction(t *testing.T) {
	src := `
		vars {
			monetary $a = [USD/2 30]
			monetary $b = [USD/2 20]
		}
		send $a - $b (
			source = @src
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
}

func TestE2E_MonetarySubtractionAssetMismatch(t *testing.T) {
	src := `
		vars {
			monetary $a = [USD/2 30]
			monetary $b = [EUR/2 20]
		}
		send $a - $b (
			source = @src
			destination = @dest
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	_, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
	require.IsType(t, vm.AssetMismatchError{}, execErr)
}

func TestE2E_MonetaryAdditionAssetMismatch(t *testing.T) {
	src := `
		vars {
			monetary $a = [USD/2 3]
			monetary $b = [EUR/2 7]
		}
		send $a + $b (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	_, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, store)
	require.IsType(t, vm.AssetMismatchError{}, execErr)
}

func TestE2E_GetAmount(t *testing.T) {
	src := `
		#![feature("experimental-get-amount-function")]
		vars {
			monetary $m = [USD/2 42]
			number $n = get_amount($m)
		}
		send [USD/2 $n] (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	res, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(42)},
	}, res.Postings)
}

func TestE2E_GetAsset(t *testing.T) {
	src := `
		#![feature("experimental-get-asset-function")]
		vars {
			monetary $m = [USD/2 42]
			asset $a = get_asset($m)
		}
		send [$a 10] (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	res, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, res.Postings)
}

func TestE2E_PrefixMinusNumber(t *testing.T) {
	// $neg = -10 (prefix on literal), $pos = -$neg = 10 (prefix on var)
	src := `
		vars {
			number $neg = -10
			number $pos = -$neg
		}
		send [USD/2 $pos] (
			source = @src
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
}

func TestE2E_PrefixMinusMonetary(t *testing.T) {
	// $neg_mon = [USD/2 -10], -$neg_mon = [USD/2 10]
	src := `
		vars {
			monetary $neg_mon = [USD/2 -10]
			monetary $pos_mon = -$neg_mon
		}
		send $pos_mon (
			source = @src
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
}

func TestE2E_Balance(t *testing.T) {
	// $bal = balance(@src, USD/2) reads @src's balance (100), then sends it all
	src := `
		vars {
			monetary $bal = balance(@src, USD/2)
		}
		send $bal (
			source = @src
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(100)},
	}, postings)
}

func TestE2E_AccountInterpolation(t *testing.T) {
	// destination = @users:<$id>:wallet, with $id = "alice"
	src := `
		#![feature("experimental-account-interpolation")]
		vars {
			string $id = "alice"
		}
		send [USD/2 10] (
			source = @world
			destination = @users:$id:wallet
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "users:alice:wallet", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
}

func TestE2E_AccountInterpolationInt(t *testing.T) {
	// destination = @account:<$n>, with $n = 42
	src := `
		#![feature("experimental-account-interpolation")]
		vars {
			number $n = 42
		}
		send [USD/2 10] (
			source = @world
			destination = @account:$n
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "account:42", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
}

func TestE2E_BoundedOverdraft(t *testing.T) {
	src := `
		send [USD/2 42] (
			source = @a allowing overdraft up to [USD/2 5]
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(40),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(42)},
	}, postings)
}

func TestE2E_NestedDestination(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				1/2 to {
					max [USD/2 10] to @x
					remaining to @a
				}
				remaining to @b
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "x", Asset: "USD/2", Amount: big.NewInt(10)},
		{Source: "world", Destination: "a", Asset: "USD/2", Amount: big.NewInt(40)},
		{Source: "world", Destination: "b", Asset: "USD/2", Amount: big.NewInt(50)},
	}, postings)
}

func TestE2E_SendAll(t *testing.T) {
	src := `send [USD/2 *] (source = @a destination = @dest)`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(30),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(30)},
	}, postings)
}

func TestE2E_UncappedBoundedOverdraft(t *testing.T) {
	src := `
		send [USD/2 *] (
			source = @a allowing overdraft up to [USD/2 5]
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(40),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(45)},
	}, postings)
}

func TestE2E_SendAllMultiSource(t *testing.T) {
	// unbounded inorder: pull everything from each source in order and sum it
	src := `
		send [USD/2 *] (
			source = {
				@a
				max [USD/2 5] from @b
				@c
			}
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(10),
		{Account: "b", Asset: "USD/2", Color: ""}: big.NewInt(100),
		{Account: "c", Asset: "USD/2", Color: ""}: big.NewInt(7),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
		{Source: "b", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(5)},
		{Source: "c", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(7)},
	}, postings)
}

func TestE2E_SendAllNegativeOverdraftBoundClamped(t *testing.T) {
	// a negative overdraft bound is clamped to 0 in the unbounded path, so only
	// the positive balance is sent (mirrors the interpreter's NonNeg).
	src := `
		send [COIN *] (
			source = @s allowing overdraft up to [COIN -10]
			destination = @dest
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "s", Asset: "COIN", Color: ""}: big.NewInt(1),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "s", Destination: "dest", Asset: "COIN", Amount: big.NewInt(1)},
	}, postings)
}

func TestE2E_CapAssetMismatch(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = max [EUR/2 5] from @a
			destination = @dest
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	require.IsType(t, vm.AssetMismatchError{}, execErr)
}

func TestE2E_OverdraftAssetMismatch(t *testing.T) {
	src := `
		send [USD/2 42] (
			source = @a allowing overdraft up to [EUR/2 5]
			destination = @dest
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	require.IsType(t, vm.AssetMismatchError{}, execErr)
}

func TestE2E_Save(t *testing.T) {
	// save 30 of @a's 100, so the send-all only takes the remaining 70
	src := `
		save [USD/2 30] from @a
		send [USD/2 *] (source = @a destination = @dest)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(70)},
	}, postings)
}

func TestE2E_InternalVar(t *testing.T) {
	src := `
		vars { account $acc = @src }
		send [USD/2 10] (source = $acc destination = @dest)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
}

func TestE2E_OverdraftFunction(t *testing.T) {
	src := `
		#![feature("experimental-overdraft-function")]
		vars { monetary $od = overdraft(@acc, USD/2) }
		send $od (source = @world destination = @dest)
	`
	// negative balance -> overdraft is the debt
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "acc", Asset: "USD/2", Color: ""}: big.NewInt(-100),
	}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(100)},
	}, postings)

	// positive balance -> overdraft is 0, nothing sent
	postings = runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "acc", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
	requirePostingsEqual(t, []runtime.Posting{}, postings)
}

func TestE2E_BalanceNegativeErrors(t *testing.T) {
	src := `
		vars { monetary $b = balance(@acc, USD/2) }
		send $b (source = @world destination = @dest)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "acc", Asset: "USD/2", Color: ""}: big.NewInt(-1),
	}})
	require.IsType(t, vm.NegativeBalanceError{}, execErr)
}

func TestE2E_DivideByZero(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				1/0 to @a
				remaining kept
			}
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	require.IsType(t, vm.DivideByZeroError{}, execErr)
}

func TestE2E_ColoredSource(t *testing.T) {
	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "COIN", Color: ""}:    big.NewInt(100),
		{Account: "src", Asset: "COIN", Color: "RED"}: big.NewInt(30),
	}}

	got := runE2E(t, `
		#![feature("experimental-asset-colors")]
		send [COIN 10] (
			source = @src \ "RED"
			destination = @dest
		)
	`, store)

	want := []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "COIN", Color: "RED", Amount: big.NewInt(10)},
	}
	requirePostingsEqual(t, want, got)
}

func TestE2E_InvalidColor(t *testing.T) {
	src := `
		#![feature("experimental-asset-colors")]
		send [COIN 10] (
			source = @src \ "not a color"
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{})
	require.IsType(t, vm.InvalidColor{}, execErr)
}

// countingStore is an e2eStore that records how many balances it was asked for.
type countingStore struct {
	e2eStore
	balanceCalls int
}

func (s *countingStore) GetBalance(ctx context.Context, account, asset, color string) (*big.Int, error) {
	s.balanceCalls++
	return s.e2eStore.GetBalance(ctx, account, asset, color)
}

// The compiled world arm has no overdraft operand, which is what makes the pull
// unbounded and therefore free of Store round-trips. numscript_test.go asserts
// the same for the interpreter.
func TestE2E_WorldSourceReadsNoBalance(t *testing.T) {
	src := `send [USD/2 100] (source = @world destination = @dest)`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	store := &countingStore{}
	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(100)},
	}, res.Postings)
	require.Zero(t, store.balanceCalls, "a world source must not read any balance")
}

// the run-time branch, not the literal, is what decides it
func TestE2E_DynamicWorldSourceReadsNoBalance(t *testing.T) {
	src := `
		vars { account $src }
		send [USD/2 100] (source = $src destination = @dest)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	enc, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	vars, err := enc.Encode(map[string]string{"src": "world"})
	require.NoError(t, err)
	store := &countingStore{}
	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, &vars, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(100)},
	}, res.Postings)
	require.Zero(t, store.balanceCalls, "a world source must not read any balance")
}

// A send-all needs a bounded source to know how much "all" is, and @world is
// unbounded. The specs format has no expectation field for this error, so it is
// asserted here; the interpreter's twin is TestInvalidUnboundedWorldInSendAll.
func TestE2E_SendAllFromWorldErrors(t *testing.T) {
	src := `send [USD/2 *] (source = @world destination = @dest)`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{})

	require.Equal(t, vm.InvalidUncappedSource{Account: "world"}, execErr)
}

// same, but world is only known at run time, so the compiler cannot reject it
func TestE2E_SendAllFromDynamicWorldErrors(t *testing.T) {
	src := `
		vars { account $src }
		send [USD/2 *] (source = $src destination = @dest)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	enc, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	vars, err := enc.Encode(map[string]string{"src": "world"})
	require.NoError(t, err)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, &vars, e2eStore{})

	require.Equal(t, vm.InvalidUncappedSource{Account: "world"}, execErr)
}

func runE2E(t *testing.T, src string, store e2eStore) []runtime.Posting {
	t.Helper()
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.Nil(t, execErr)
	return res.Postings
}

func requirePostingsEqual(t *testing.T, want, got []runtime.Posting) {
	t.Helper()
	require.Len(t, got, len(want))
	for i := range want {
		w, g := want[i], got[i]
		require.Equal(t, w.Source, g.Source, "posting[%d].Source", i)
		require.Equal(t, w.Destination, g.Destination, "posting[%d].Destination", i)
		require.Equal(t, w.Asset, g.Asset, "posting[%d].Asset", i)
		require.Equal(t, w.Color, g.Color, "posting[%d].Color", i)
		require.Zero(t, g.Amount.Cmp(w.Amount), "posting[%d].Amount: got %s want %s", i, g.Amount, w.Amount)
	}
}

// --- Allotment rounding, ported from internal/runtime/allotment_test.go -----
// These pinned runtime.MakeAllotment before the split was lowered into pure
// instructions; they now pin the compiler's lowering of it.

// A two-unit shortfall: 1/6,1/6,4/6 of 100 floors to 16,16,66 (sum 98), so the
// first two shares each get one unit back.
func TestE2E_AllotmentLeftoverTwoUnits(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				1/6 to @a
				1/6 to @b
				remaining to @c
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "a", Asset: "USD/2", Amount: big.NewInt(17)},
		{Source: "world", Destination: "b", Asset: "USD/2", Amount: big.NewInt(17)},
		{Source: "world", Destination: "c", Asset: "USD/2", Amount: big.NewInt(66)},
	}, postings)
}

// An odd amount split in half: 7 -> 3,3 (sum 6), leftover unit to the earliest.
func TestE2E_AllotmentHalvesOfOddAmount(t *testing.T) {
	src := `
		send [USD/2 7] (
			source = @world
			destination = {
				1/2 to @a
				remaining to @b
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "a", Asset: "USD/2", Amount: big.NewInt(4)},
		{Source: "world", Destination: "b", Asset: "USD/2", Amount: big.NewInt(3)},
	}, postings)
}

// A single whole share: the lowering emits no fixup blocks at all for n == 1.
func TestE2E_AllotmentSinglePortionWhole(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				remaining to @a
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "a", Asset: "USD/2", Amount: big.NewInt(100)},
	}, postings)
}

// Percentages that divide exactly: no leftover, so no share is adjusted.
func TestE2E_AllotmentPercentagesDivideExactly(t *testing.T) {
	src := `
		send [USD/2 10000] (
			source = @world
			destination = {
				19/100 to @a
				remaining to @b
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "a", Asset: "USD/2", Amount: big.NewInt(1900)},
		{Source: "world", Destination: "b", Asset: "USD/2", Amount: big.NewInt(8100)},
	}, postings)
}

// Sevenths of 1001 floor awkwardly (143 + 286 + 572 = 1001 exactly here), the
// point being that the shares must always sum back to the amount.
func TestE2E_AllotmentPartsSumToAmount(t *testing.T) {
	src := `
		send [USD/2 1001] (
			source = @world
			destination = {
				1/7 to @a
				2/7 to @b
				remaining to @c
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})

	total := new(big.Int)
	for _, p := range postings {
		total.Add(total, p.Amount)
	}
	require.Zero(t, total.Cmp(big.NewInt(1001)), "shares sum to %s, want 1001 (%v)", total, postings)
}

// Beyond int64: ~1e27+1 split in half, the odd unit going to the earliest share.
func TestE2E_AllotmentBeyondInt64(t *testing.T) {
	src := `
		send [USD/2 1000000000000000000000000001] (
			source = @world
			destination = {
				1/2 to @a
				remaining to @b
			}
		)
	`
	postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{}})

	amount, _ := new(big.Int).SetString("1000000000000000000000000001", 10)
	half := new(big.Int).Div(amount, big.NewInt(2))
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "world", Destination: "a", Asset: "USD/2", Amount: new(big.Int).Add(half, big.NewInt(1))},
		{Source: "world", Destination: "b", Asset: "USD/2", Amount: half},
	}, postings)
}

// TestE2E_Oneof runs a compiled `oneof` end to end, which the IR snapshot tests
// cannot: they check the emitted instructions, not that executing them backtracks
// correctly. What matters here is that the single mark_pop at the join is reached
// on every path — the branch that covered the amount jumps straight to it, and the
// last branch falls through to it.
func TestE2E_Oneof(t *testing.T) {
	src := `
		#![feature("experimental-oneof")]
		send [USD/2 10] (
			source = oneof {
				@a
				@b
				@c
			}
			destination = @dest
		)
	`

	t.Run("first branch covers it", func(t *testing.T) {
		postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
			{Account: "a", Asset: "USD/2"}: big.NewInt(10),
			{Account: "b", Asset: "USD/2"}: big.NewInt(10),
			{Account: "c", Asset: "USD/2"}: big.NewInt(10),
		}})
		requirePostingsEqual(t, []runtime.Posting{
			{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
		}, postings)
	})

	t.Run("backtracks past a short branch", func(t *testing.T) {
		// @a can only cover 3 of the 10, so its partial funding is discarded whole
		// rather than combined with @b's
		postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
			{Account: "a", Asset: "USD/2"}: big.NewInt(3),
			{Account: "b", Asset: "USD/2"}: big.NewInt(10),
			{Account: "c", Asset: "USD/2"}: big.NewInt(10),
		}})
		requirePostingsEqual(t, []runtime.Posting{
			{Source: "b", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
		}, postings)
	})

	t.Run("backtracks twice, to the last branch", func(t *testing.T) {
		postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
			{Account: "a", Asset: "USD/2"}: big.NewInt(3),
			{Account: "b", Asset: "USD/2"}: big.NewInt(9),
			{Account: "c", Asset: "USD/2"}: big.NewInt(10),
		}})
		requirePostingsEqual(t, []runtime.Posting{
			{Source: "c", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
		}, postings)
	})

	t.Run("no branch covers it", func(t *testing.T) {
		// the last branch is not rewound, so check_enough_funds reports what it got
		parsed := parser.Parse(src)
		require.Empty(t, parsed.Errors)
		_, program, cErr := compiler.Compile(parsed.Value, nil)
		require.Nil(t, cErr)
		_, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, e2eStore{
			balances: map[runtime.PairKey]*big.Int{
				{Account: "a", Asset: "USD/2"}: big.NewInt(3),
				{Account: "b", Asset: "USD/2"}: big.NewInt(4),
				{Account: "c", Asset: "USD/2"}: big.NewInt(5),
			},
		})
		require.IsType(t, vm.MissingFundsError{}, execErr)
	})
}

// A oneof nested inside another must keep the two regions independent: the inner
// backtrack may not discard what the outer branch already pulled.
func TestE2E_OneofNested(t *testing.T) {
	src := `
		#![feature("experimental-oneof")]
		send [USD/2 10] (
			source = oneof {
				{
					@a
					oneof { @b @c }
				}
				@d
			}
			destination = @dest
		)
	`

	t.Run("inner backtrack keeps the outer branch's funds", func(t *testing.T) {
		// @a gives 4, so the inner oneof needs 6: @b has only 5 -> rewind -> @c
		// covers it. @a's 4 must survive the inner rewind.
		postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
			{Account: "a", Asset: "USD/2"}: big.NewInt(4),
			{Account: "b", Asset: "USD/2"}: big.NewInt(5),
			{Account: "c", Asset: "USD/2"}: big.NewInt(6),
			{Account: "d", Asset: "USD/2"}: big.NewInt(10),
		}})
		requirePostingsEqual(t, []runtime.Posting{
			{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(4)},
			{Source: "c", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(6)},
		}, postings)
	})

	t.Run("outer backtrack discards the whole inner region too", func(t *testing.T) {
		// the inorder branch tops out at 4+6=10... but with @c at 5 it reaches 9,
		// so the outer oneof rewinds @a and @c together and takes @d instead
		postings := runE2E(t, src, e2eStore{balances: map[runtime.PairKey]*big.Int{
			{Account: "a", Asset: "USD/2"}: big.NewInt(4),
			{Account: "b", Asset: "USD/2"}: big.NewInt(3),
			{Account: "c", Asset: "USD/2"}: big.NewInt(5),
			{Account: "d", Asset: "USD/2"}: big.NewInt(10),
		}})
		requirePostingsEqual(t, []runtime.Posting{
			{Source: "d", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
		}, postings)
	})
}
