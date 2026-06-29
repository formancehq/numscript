package compiler

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// e2eStore is a minimal vm.Store for the end-to-end test.
type e2eStore struct {
	balances map[runtime.PairKey]*big.Int
}

// zeroBalance is returned for unknown (account, asset, color) triples. It is
// shared and never mutated by the runtime (cachedBalance copies the value), so a
// single instance is safe — and keeps the store off the allocation path so
// benchmarks measure the engine, not the harness.
var zeroBalance = new(big.Int)

func (s e2eStore) GetBalance(account, asset, color string) *big.Int {
	if v, ok := s.balances[runtime.PairKey{Account: account, Asset: asset, Color: color}]; ok {
		return v
	}
	return zeroBalance
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

// TestE2E_OptimizedMatches runs each script through the OPTIONAL peephole pass
// (compile -> optimize -> assemble -> run) and asserts the postings are identical
// to the unoptimized pipeline, and that the pass actually removed instructions.
func TestE2E_OptimizedMatches(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		store e2eStore
		want  []runtime.Posting
	}{
		{
			name: "simple",
			src:  `send [USD/2 10] (source = @src destination = @dest)`,
			store: e2eStore{balances: map[runtime.PairKey]*big.Int{
				{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
			}},
			want: []runtime.Posting{{Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)}},
		},
		{
			name: "capped-inorder",
			src:  `send [USD/2 10] (source = { @a max [USD/2 5] from @b @c } destination = @dest)`,
			store: e2eStore{balances: map[runtime.PairKey]*big.Int{
				{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(3),
				{Account: "b", Asset: "USD/2", Color: ""}: big.NewInt(100),
				{Account: "c", Asset: "USD/2", Color: ""}: big.NewInt(100),
			}},
			want: []runtime.Posting{
				{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(3)},
				{Source: "b", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(5)},
				{Source: "c", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(2)},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed := parser.Parse(tc.src)
			require.Empty(t, parsed.Errors)
			compiled, cErr := compileProgramToVirtual(parsed.Value)
			require.Nil(t, cErr)

			opted := optimize(compiled.instructions, defaultPeepholes())
			require.Less(t, len(opted), len(compiled.instructions), "peephole should remove instructions")

			program, aErr := Assemble(opted)
			require.NoError(t, aErr)

			postings, execErr := vm.Exec(vm.NewVm(program), nil, tc.store)
			require.Nil(t, execErr)
			requirePostingsEqual(t, tc.want, postings)
		})
	}
}

// TestE2E_ReusedVMStaysCorrect runs the same Vm many times (reusing its runstate,
// which recycles big.Ints across runs via the freelist). It guards against pool
// aliasing/corruption: every run must yield identical, correct postings, and
// varying the store between runs must be reflected.
func TestE2E_ReusedVMStaysCorrect(t *testing.T) {
	src := `
		send [USD/2 10] (
			source = { @a max [USD/2 5] from @b @c }
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
	want := []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(3)},
		{Source: "b", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(5)},
		{Source: "c", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(2)},
	}

	machine := vm.NewVm(program)
	for run := 0; run < 50; run++ {
		postings, execErr := vm.Exec(machine, nil, store)
		require.Nil(t, execErr, "run %d", run)
		requirePostingsEqual(t, want, postings)
	}

	// A different store on the same Vm must be reflected (no stale cached state).
	store2 := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(10),
	}}
	postings, execErr := vm.Exec(machine, nil, store2)
	require.Nil(t, execErr)
	requirePostingsEqual(t, []runtime.Posting{
		{Source: "a", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10)},
	}, postings)
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
