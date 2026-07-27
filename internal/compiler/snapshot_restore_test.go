package compiler

import (
	"testing"

	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

func TestSnapshotRestore_Typecheck(t *testing.T) {
	// snapshot defines an int reg; restore reads it back
	require.NoError(t, typecheckInstructions([]irInstr{
		snapshot{dest: 0},
		restore{mark: 0},
	}))

	// restore reading a non-int register is a type error
	require.Error(t, typecheckInstructions([]irInstr{
		loadStr{dest: 0, value: "USD/2"},
		restore{mark: 0},
	}))
}

func TestSnapshotRestore_Dump(t *testing.T) {
	out := dump([]irInstr{
		snapshot{dest: 1},
		restore{mark: 1},
	})
	require.Contains(t, out, "$r1 = snapshot()")
	require.Contains(t, out, "restore($r1)")
}

func TestSnapshotRestore_Assemble(t *testing.T) {
	prog, err := assembleProgram([]irInstr{
		snapshot{dest: 1},
		restore{mark: 1},
	})
	require.NoError(t, err)
	require.Equal(t, []vm.Instruction{
		{Opcode: byte(vm.Op_Snapshot), A: 0, B: 0xFF, C: 0xFF},
		{Opcode: byte(vm.Op_Restore), A: 0, B: 0xFF, C: 0xFF},
	}, prog.Instructions)
}
