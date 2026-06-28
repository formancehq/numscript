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
