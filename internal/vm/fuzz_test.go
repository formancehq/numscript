package vm

import (
	"math/big"
	"testing"
)

// FuzzExec feeds arbitrary bytes as an instruction stream and asserts that a
// program which passes Verify never panics under Exec. Verify itself must also
// never panic on arbitrary input. (Exec no longer verifies, so running an
// unverified malformed program may legitimately panic — hence the Verify gate.)
func FuzzExec(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{byte(Op_LoadInt), 0, 0, 0})
	f.Add([]byte{byte(Op_PullAccount), 0, 0, 0xFF})

	pool := Program{
		IntsPool:    []big.Int{*big.NewInt(0), *big.NewInt(7)},
		StringsPool: []string{"world", "dest"},
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		instrs := make([]Instruction, len(data)/4)
		for i := range instrs {
			off := i * 4
			instrs[i] = Instruction{data[off], data[off+1], data[off+2], data[off+3]}
		}
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panicked: %v", r)
			}
		}()

		// size as the assembler would, then let Verify be the gate: only a
		// verified program is required not to panic under Exec.
		prog := sizeProgram(Program{Instructions: instrs, IntsPool: pool.IntsPool, StringsPool: pool.StringsPool})
		if prog.Verify() != nil {
			return
		}
		_, _ = Exec(t.Context(), NewVm(prog), nil, mockStore{})
	})
}
