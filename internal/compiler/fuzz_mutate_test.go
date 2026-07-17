package compiler_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
)

// a script exercising most opcodes: inorder+max source (jump, min_int),
// allotment destination (mk_portion, sub_portion, mk_allotment), and no vars.
const mutateBaseSrc = `send [USD/2 10] (
	source = {
		max [USD/2 5] from @a
		@b
	}
	destination = {
		1/2 to @c
		remaining to @d
	}
)`

// Properties 1 & 3: take valid compiled bytecode, mutate it, and if the mutant
// still passes the verifier, executing it must never crash.
func FuzzMutatedBytecode(f *testing.F) {
	parsed := parser.Parse(mutateBaseSrc)
	if len(parsed.Errors) != 0 {
		f.Fatalf("parse: %v", parsed.Errors)
	}
	_, base, cErr := compiler.Compile(parsed.Value)
	if cErr != nil {
		f.Fatalf("compile: %v", cErr)
	}

	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(3),
		{Account: "b", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	f.Add([]byte{0, 0})
	f.Add([]byte{4, 255})
	f.Add([]byte{1, 9, 8, 2, 12, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		instrs := make([]vm.Instruction, len(base.Instructions))
		copy(instrs, base.Instructions)
		if len(instrs) == 0 {
			return
		}

		flat := make([]byte, len(instrs)*4)
		for k, ins := range instrs {
			flat[k*4], flat[k*4+1], flat[k*4+2], flat[k*4+3] = ins.Opcode, ins.A, ins.B, ins.C
		}
		for i := 0; i+1 < len(data); i += 2 {
			flat[int(data[i])%len(flat)] = data[i+1]
		}
		for k := range instrs {
			instrs[k] = vm.Instruction{Opcode: flat[k*4], A: flat[k*4+1], B: flat[k*4+2], C: flat[k*4+3]}
		}

		// keep base's declared counts; a mutation that points an instruction at a
		// register outside those bounds must be caught by Verify (coherence).
		prog := base
		prog.Instructions = instrs
		if prog.Verify() != nil {
			return // mutation broke the sanity checks: nothing more to prove
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("verified program panicked on Exec: %v", r)
			}
		}()
		_, _ = vm.Exec(t.Context(), vm.NewVm(prog), nil, store)
	})
}
