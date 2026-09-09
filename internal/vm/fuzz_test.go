package vm

import (
	"context"
	"math/big"
	"testing"
)

// FuzzExec reads arbitrary bytes as an instruction stream and asserts that
// anything the verifier accepts, the VM can run without panicking.
//
// This is the verifier's contract stated as a test. It says nothing about
// programs the verifier rejects: Exec is entitled to crash on those, which is
// why it is Exec's caller that has to decide whether a program is trusted.
//
// Note VerifyWithVars rather than Verify — a program that loads a variable is
// only safe against the vars it will actually be given, and Verify alone does
// not look at them.
func FuzzExec(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{byte(Op_LoadInt), 0, 0, 0})
	f.Add([]byte{byte(Op_PullAccount), 0, 0, nilReg})               // truncated: no ext word
	f.Add([]byte{byte(Op_LoadStr), 0, 0, 0, byte(Op_Jmp), 0, 1, 0}) // jump past the end
	f.Add([]byte{byte(Op_ConstTrue), 0, 0, 0, byte(Op_JmpIfTrue), 0, 0, 0})
	f.Add([]byte{byte(Op_MarkPush), 0, 0, 0, byte(Op_MarkEnd), 1, 0, 0})

	stringsPool := []string{"world", "dest", "USD/2"}
	intsPool := []big.Int{*big.NewInt(0), *big.NewInt(7)}
	vars := &Vars{
		StringsPool: []string{"var0"},
		IntsPool:    []big.Int{*big.NewInt(3)},
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		instrs := make([]Instruction, len(data)/4)
		for i := range instrs {
			off := i * 4
			instrs[i] = Instruction{data[off], data[off+1], data[off+2], data[off+3]}
		}

		prog := fullBanks(Program{
			Instructions: instrs,
			StringsPool:  stringsPool,
			IntsPool:     intsPool,
		})

		if VerifyWithVars(prog, vars) != nil {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("verified program panicked in Exec: %v", r)
			}
		}()
		_, _ = Exec(context.Background(), NewVm(prog), vars, mockStore{})
	})
}
