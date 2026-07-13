package compiler

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/typecheck"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// e2eStore is a minimal vm.Store for the end-to-end test.
type e2eStore struct {
	balances map[runtime.PairKey]*big.Int
}

func (s e2eStore) GetBalance(account, asset, color string) *big.Int {
	if v, ok := s.balances[runtime.PairKey{Account: account, Asset: asset, Color: color}]; ok {
		return v
	}
	return new(big.Int)
}

// TestE2E_CompileAssembleRun exercises the whole pipeline: source -> compiler
// (virtual instructions) -> assembler (vm.Program) -> VM execution -> postings.
func TestE2E_CompileAssembleRun(t *testing.T) {
	src := `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	machine := vm.NewVm(program)
	postings, execErr := vm.Exec(machine, nil, store)
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}
	requirePostingsEqual(t, want, postings)
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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(6),
		{Account: "b", Asset: "USD/2", Color: ""}: big.NewInt(10),
		{Account: "c", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	machine := vm.NewVm(program)
	postings, execErr := vm.Exec(machine, nil, store)
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(6)},
		{Source: "b", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(4)},
	}
	requirePostingsEqual(t, want, postings)
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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(3),
		{Account: "b", Asset: "USD/2", Color: ""}: big.NewInt(100),
		{Account: "c", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	machine := vm.NewVm(program)
	postings, execErr := vm.Exec(machine, nil, store)
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(3)},
		{Source: "b", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(5)},
		{Source: "c", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(2)},
	}
	requirePostingsEqual(t, want, postings)
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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	// src only has 4, but 10 is required.
	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(4),
	}}

	machine := vm.NewVm(program)
	_, execErr := vm.Exec(machine, nil, store)
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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{}}

	machine := vm.NewVm(program)
	postings, execErr := vm.Exec(machine, nil, store)
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "world", Destination: "x", Asset: "USD/2", Amount: big.NewInt(30)},
		{Source: "world", Destination: "y", Asset: "USD/2", Amount: big.NewInt(70)},
	}
	requirePostingsEqual(t, want, postings)
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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{}}

	machine := vm.NewVm(program)
	postings, execErr := vm.Exec(machine, nil, store)
	require.Nil(t, execErr)

	// only the remaining 70 is posted; the kept 30 produces no posting
	want := []runtime.Posting{
		{Source: "world", Destination: "y", Asset: "USD/2", Amount: big.NewInt(70)},
	}
	requirePostingsEqual(t, want, postings)
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
	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)
	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)
	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "s1", Asset: "USD/2", Color: ""}: big.NewInt(10),
		{Account: "s2", Asset: "USD/2", Color: ""}: big.NewInt(1000),
	}}
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(machine, nil, store)
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
	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)
	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
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
	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)
	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	postings, execErr := vm.Exec(vm.NewVm(program), nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	postings, execErr := vm.Exec(vm.NewVm(program), nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	postings, execErr := vm.Exec(vm.NewVm(program), nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []runtime.Posting{
		{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
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

	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)

	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	_, execErr := vm.Exec(vm.NewVm(program), nil, store)
	require.IsType(t, vm.AssetMismatchError{}, execErr)
}

func TestE2E_RejectsUnboundVariable(t *testing.T) {
	parsed := parser.Parse(`send [C 10] (source = $undeclared destination = @d)`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToVirtual(parsed.Value)
	require.IsType(t, TypeError{}, cErr)
	require.IsType(t, typecheck.UnboundVariable{}, cErr.(TypeError).Kind)
}

func TestE2E_RejectsTypeMismatch(t *testing.T) {
	parsed := parser.Parse(`vars { string $s } send [C 10] (source = $s destination = @d)`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToVirtual(parsed.Value)
	require.IsType(t, TypeError{}, cErr)
	require.IsType(t, typecheck.TypeMismatch{}, cErr.(TypeError).Kind)
}

func TestE2E_AllotmentDuplicateRemaining(t *testing.T) {
	parsed := parser.Parse(`
		send [USD/2 100] (
			source = @world
			destination = {
				remaining to @a
				remaining to @b
			}
		)
	`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToVirtual(parsed.Value)
	require.IsType(t, DuplicateRemaining{}, cErr)
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

func TestE2E_CapAssetMismatch(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = max [EUR/2 5] from @a
			destination = @dest
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)
	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
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
	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)
	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(machine, nil, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
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

func runE2E(t *testing.T, src string, store e2eStore) []runtime.Posting {
	t.Helper()
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)
	program, aErr := Assemble(compiled.instructions)
	require.NoError(t, aErr)
	machine := vm.NewVm(program)
	postings, execErr := vm.Exec(machine, nil, store)
	require.Nil(t, execErr)
	return postings
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
