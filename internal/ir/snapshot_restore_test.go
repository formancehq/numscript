package ir

import (
	"testing"

	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

func TestSnapshotRestore_Typecheck(t *testing.T) {
	// Snapshot defines an int register; Restore reads it back
	require.NoError(t, Typecheck([]Instr{
		Snapshot{Dest: 0},
		Restore{Mark: 0},
	}))

	// Restore reading a non-int register is a type error
	require.Error(t, Typecheck([]Instr{
		LoadStr{Dest: 0, Value: "USD/2"},
		Restore{Mark: 0},
	}))
}

func TestSnapshotRestore_Dump(t *testing.T) {
	out := Dump([]Instr{
		Snapshot{Dest: 1},
		Restore{Mark: 1},
	})
	require.Contains(t, out, "$r1 = snapshot()")
	require.Contains(t, out, "restore($r1)")
}

func TestSnapshotRestore_Assemble(t *testing.T) {
	prog, err := Assemble([]Instr{
		Snapshot{Dest: 1},
		Restore{Mark: 1},
	})
	require.NoError(t, err)
	require.Equal(t, []vm.Instruction{
		{Opcode: byte(vm.Op_Snapshot), A: 0, B: 0xFF, C: 0xFF},
		{Opcode: byte(vm.Op_Restore), A: 0, B: 0xFF, C: 0xFF},
	}, prog.Instructions)
}
