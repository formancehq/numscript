package compiler

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytecodeTypecheck_Valid(t *testing.T) {
	// $0 = 1; $1 = 2; $2 = $0 + $1   (all int)
	instrs := []irInstr{
		loadInt{dest: 0, value: *big.NewInt(1)},
		loadInt{dest: 1, value: *big.NewInt(2)},
		binaryOp{op: opAddInt{}, dest: 2, left: 0, right: 1},
	}
	require.NoError(t, typecheckInstructions(instrs))
}

func TestBytecodeTypecheck_UseBeforeWrite(t *testing.T) {
	// reads $0 as int before it is ever written
	instrs := []irInstr{
		unaryOp{op: opNegInt{}, dest: 1, arg: 0},
	}
	require.Error(t, typecheckInstructions(instrs))
}

func TestBytecodeTypecheck_WrongType(t *testing.T) {
	// $0 is a string, then used where an int is expected
	instrs := []irInstr{
		loadStr{dest: 0, value: "USD/2"},
		unaryOp{op: opNegInt{}, dest: 1, arg: 0},
	}
	require.Error(t, typecheckInstructions(instrs))
}

func TestBytecodeTypecheck_RedefinedWithDifferentType(t *testing.T) {
	// $0 written as int, then overwritten as string
	instrs := []irInstr{
		loadInt{dest: 0, value: *big.NewInt(1)},
		loadStr{dest: 0, value: "x"},
	}
	require.Error(t, typecheckInstructions(instrs))
}
