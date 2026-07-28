package compiler

import (
	"testing"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

func TestSnapshotRestore_Typecheck(t *testing.T) {
	// Snapshot defines an int register; Restore reads it back
	require.NoError(t, ir.Typecheck([]ir.Instr{
		ir.Snapshot{Dest: 0},
		ir.Restore{Mark: 0},
	}))

	// ir.Restore reading a non-int register is a type error
	require.Error(t, ir.Typecheck([]ir.Instr{
		ir.LoadStr{Dest: 0, Value: "USD/2"},
		ir.Restore{Mark: 0},
	}))
}

func TestSnapshotRestore_Dump(t *testing.T) {
	out := ir.Dump([]ir.Instr{
		ir.Snapshot{Dest: 1},
		ir.Restore{Mark: 1},
	})
	require.Contains(t, out, "$r1 = snapshot()")
	require.Contains(t, out, "restore($r1)")
}

func TestSnapshotRestore_Assemble(t *testing.T) {
	prog, err := ir.Assemble([]ir.Instr{
		ir.Snapshot{Dest: 1},
		ir.Restore{Mark: 1},
	})
	require.NoError(t, err)
	require.Equal(t, []vm.Instruction{
		{Opcode: byte(vm.Op_Snapshot), A: 0, B: 0xFF, C: 0xFF},
		{Opcode: byte(vm.Op_Restore), A: 0, B: 0xFF, C: 0xFF},
	}, prog.Instructions)
}
