package compiler_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/funds"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// compileBoth compiles src plain and optimized.
func compileBoth(t *testing.T, src string) (plain, optimized vm.Program) {
	t.Helper()
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, plain, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	_, optimized, cErr = compiler.CompileWithOptions(parsed.Value, nil, compiler.Options{Optimize: true})
	require.Nil(t, cErr)
	return plain, optimized
}

// The fused world send reads no balance either, and posts the same.
func TestOptimizedWorldSourceReadsNoBalance(t *testing.T) {
	_, optimized := compileBoth(t, `send [USD/2 100] (source = @world destination = @dest)`)

	store := &countingStore{}
	res, execErr := vm.Exec(context.Background(), vm.NewVm(optimized), nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []funds.Posting{
		{Source: "world", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(100)},
	}, res.Postings)
	require.Zero(t, store.balanceCalls, "a world source must not read any balance")
}

// A capped unbounded source is still fused, and still short: the check sits
// between the clamp and the posting, as it did between the pull and the send.
func TestOptimizedCappedUnboundedSourceStillChecksFunds(t *testing.T) {
	plain, optimized := compileBoth(t, `send [USD/2 42] (source = max [USD/2 5] from @world destination = @dest)`)

	want := vm.MissingFundsError{Asset: "USD/2", Got: big.NewInt(5), Needed: big.NewInt(42)}
	for name, program := range map[string]vm.Program{"plain": plain, "optimized": optimized} {
		t.Run(name, func(t *testing.T) {
			_, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, e2eStore{})
			require.Equal(t, want, execErr)
		})
	}
}

// Which error surfaces is part of the behavior: the funds check comes before
// the destination is validated, so a short cap wins over a malformed
// destination, and a sufficient one lets the destination's own error through.
func TestOptimizedKeepsTheOrderOfChecks(t *testing.T) {
	src := `
		#![feature("experimental-account-interpolation")]
		vars {
			monetary $cap
			string $suffix
		}
		send [USD/2 42] (
			source = max $cap from @world
			destination = @users:$suffix
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	for name, opts := range map[string]compiler.Options{"plain": {}, "optimized": {Optimize: true}} {
		t.Run(name, func(t *testing.T) {
			enc, program, cErr := compiler.CompileWithOptions(parsed.Value, nil, opts)
			require.Nil(t, cErr)

			t.Run("short cap wins over the malformed destination", func(t *testing.T) {
				vars, err := enc.Encode(map[string]string{"cap": "USD/2 5", "suffix": "not valid"})
				require.NoError(t, err)
				_, execErr := vm.Exec(context.Background(), vm.NewVm(program), &vars, e2eStore{})
				require.Equal(t, vm.MissingFundsError{Asset: "USD/2", Got: big.NewInt(5), Needed: big.NewInt(42)}, execErr)
			})

			t.Run("sufficient cap lets the destination fail", func(t *testing.T) {
				vars, err := enc.Encode(map[string]string{"cap": "USD/2 50", "suffix": "not valid"})
				require.NoError(t, err)
				_, execErr := vm.Exec(context.Background(), vm.NewVm(program), &vars, e2eStore{})
				require.Equal(t, vm.InvalidAccountName{Name: "users:not valid"}, execErr)
			})
		})
	}
}
