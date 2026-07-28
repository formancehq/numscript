package ir

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytecodeTypecheck_Valid(t *testing.T) {
	// $0 = 1; $1 = 2; $2 = $0 + $1   (all int)
	instrs := []Instr{
		LoadInt{Dest: 0, Value: *big.NewInt(1)},
		LoadInt{Dest: 1, Value: *big.NewInt(2)},
		BinaryOp{Op: OpAddInt{}, Dest: 2, Left: 0, Right: 1},
	}
	require.NoError(t, Typecheck(instrs))
}

func TestBytecodeTypecheck_UseBeforeWrite(t *testing.T) {
	// reads $0 as int before it is ever written
	instrs := []Instr{
		UnaryOp{Op: OpNegInt{}, Dest: 1, Arg: 0},
	}
	require.Error(t, Typecheck(instrs))
}

func TestBytecodeTypecheck_WrongType(t *testing.T) {
	// $0 is a string, then used where an int is expected
	instrs := []Instr{
		LoadStr{Dest: 0, Value: "USD/2"},
		UnaryOp{Op: OpNegInt{}, Dest: 1, Arg: 0},
	}
	require.Error(t, Typecheck(instrs))
}

func TestBytecodeTypecheck_RedefinedWithDifferentType(t *testing.T) {
	// $0 written as int, then overwritten as string
	instrs := []Instr{
		LoadInt{Dest: 0, Value: *big.NewInt(1)},
		LoadStr{Dest: 0, Value: "x"},
	}
	require.Error(t, Typecheck(instrs))
}
