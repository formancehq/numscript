package compiler_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/funds"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/vm"
)

// A script covering most of the instruction set: a oneof source (marks, and the
// jumps around them), a capped source (lt_int plus copies), an allotment
// destination (the portion ops and the leftover fixup), a balance() read, and a
// set_account_meta. Deliberately no vars, so the mutants run against nil.
const mutateBaseSrc = `send [USD/2 10] (
	source = {
		max [USD/2 5] from @a
		oneof {
			@b
			@c
		}
		@d
	}
	destination = {
		1/2 to @e
		remaining to @f
	}
)
set_account_meta(@e, "k", "v")
`

// FuzzMutatedBytecode is the converse of FuzzExec: instead of random bytes it
// starts from bytecode the compiler really emitted and corrupts it, so the
// mutants stay close enough to valid that they get past the cheap checks and
// reach the interesting ones.
//
// The property is the same either way — if the verifier accepts it, Exec must
// not crash on it.
func FuzzMutatedBytecode(f *testing.F) {
	parsed := parser.Parse(mutateBaseSrc)
	if len(parsed.Errors) != 0 {
		f.Fatalf("parse: %v", parsed.Errors)
	}
	featureFlags := map[string]struct{}{"experimental-oneof": {}}
	_, base, cErr := compiler.Compile(parsed.Value, featureFlags)
	if cErr != nil {
		f.Fatalf("compile: %v", cErr)
	}
	if err := vm.Verify(base); err != nil {
		f.Fatalf("the unmutated program must verify: %v", err)
	}

	store := e2eStore{balances: map[funds.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2"}: big.NewInt(3),
		{Account: "b", Asset: "USD/2"}: big.NewInt(4),
		{Account: "c", Asset: "USD/2"}: big.NewInt(4),
		{Account: "d", Asset: "USD/2"}: big.NewInt(100),
	}}

	// The base script reads no variables, but a mutation can turn any byte into
	// an Op_LoadVar*, so the mutants have to be verified against the vars they
	// will actually be given. Verify alone does not look at them.
	vars := &vm.Vars{
		StringsPool: []string{"v"},
		IntsPool:    []big.Int{*big.NewInt(1)},
	}

	f.Add([]byte{0, 0})
	f.Add([]byte{4, 255})
	f.Add([]byte{1, 9, 8, 2, 12, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(base.Instructions) == 0 {
			return
		}

		flat := make([]byte, len(base.Instructions)*4)
		for k, ins := range base.Instructions {
			flat[k*4], flat[k*4+1], flat[k*4+2], flat[k*4+3] = ins.Opcode, ins.A, ins.B, ins.C
		}
		// data is read as (offset, value) pairs. The offset is a uint16 rather
		// than the byte the original used, which could only ever reach the first
		// 256 bytes — this program is longer than that.
		for i := 0; i+2 < len(data); i += 3 {
			off := (int(data[i])<<8 | int(data[i+1])) % len(flat)
			flat[off] = data[i+2]
		}

		instrs := make([]vm.Instruction, len(base.Instructions))
		for k := range instrs {
			instrs[k] = vm.Instruction{Opcode: flat[k*4], A: flat[k*4+1], B: flat[k*4+2], C: flat[k*4+3]}
		}

		prog := base
		prog.Instructions = instrs

		if vm.VerifyWithVars(prog, vars) != nil {
			return // the mutation broke something static: nothing left to prove
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("verified program panicked in Exec: %v", r)
			}
		}()
		_, _ = vm.Exec(context.Background(), vm.NewVm(prog), vars, store)
	})
}
