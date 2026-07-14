package vm

import (
	"math/big"
	"testing"
)

// FuzzExec feeds arbitrary bytes as an instruction stream and asserts the VM
// never panics: the verifier must reject a malformed program, or execution must
// return a result/error. It never crashes the host.
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
		prog := Program{Instructions: instrs, IntsPool: pool.IntsPool, StringsPool: pool.StringsPool}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Exec panicked: %v", r)
			}
		}()
		_, _ = Exec(NewVm(prog), nil, mockStore{})
	})
}
