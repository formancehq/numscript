package ir

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

func TestAssemble_AddInt(t *testing.T) {
	// Three distinct virtual int registers map to the first three int-bank
	// indices in first-use order.
	prog, err := Assemble([]Instr{
		BinaryOp{Op: OpAddInt{}, Dest: 10, Left: 20, Right: 30},
	})
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	instrs := prog.Instructions
	if len(instrs) != 1 {
		t.Fatalf("got %d instructions, want 1", len(instrs))
	}
	want := vm.Instruction{Opcode: byte(vm.Op_AddInt), A: 0, B: 1, C: 2}
	if instrs[0] != want {
		t.Errorf("got %+v, want %+v", instrs[0], want)
	}
}

func TestAssemble_AddInt_ReusesRegisterIndices(t *testing.T) {
	// A virtual register reused across operands/instructions keeps the same
	// bank index; new ones get fresh indices in first-use order.
	prog, err := Assemble([]Instr{
		// Reg 7 -> 0, Reg 8 -> 1 ; dest==left==7
		BinaryOp{Op: OpAddInt{}, Dest: 7, Left: 7, Right: 8},
		// Reg 9 -> 2 ; reuses 7->0 and 8->1
		BinaryOp{Op: OpAddInt{}, Dest: 9, Left: 7, Right: 8},
	})
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	got := prog.Instructions
	want := []vm.Instruction{
		{Opcode: byte(vm.Op_AddInt), A: 0, B: 0, C: 1},
		{Opcode: byte(vm.Op_AddInt), A: 2, B: 0, C: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d instructions, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("instr[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestAssemble_Empty(t *testing.T) {
	prog, err := Assemble(nil)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if len(prog.Instructions) != 0 {
		t.Errorf("expected no instructions, got %d", len(prog.Instructions))
	}
}

func TestAssemble_MaxRegPerBank(t *testing.T) {
	prog, err := Assemble([]Instr{
		LoadStr{Dest: 0, Value: "USD/2"},
		LoadInt{Dest: 1, Value: *big.NewInt(10)},
		BinaryOp{Op: OpMakeMonetary{}, Left: 0, Right: 1, Dest: 2},
		BinaryOp{Op: OpMakePortion{}, Left: 1, Right: 1, Dest: 3},
	})
	require.NoError(t, err)

	require.Equal(t, byte(1), prog.MaxRegString, "one str reg")
	require.Equal(t, byte(1), prog.MaxRegInt, "one int reg")
	require.Equal(t, byte(1), prog.MaxRegMonetary, "one monetary reg")
	require.Equal(t, byte(1), prog.MaxRegPortion, "one portion reg")
}

// 255 registers per bank, because 0xFF is the "operand unset" sentinel.
func TestAssemble_RegisterBankOverflow(t *testing.T) {
	loadInts := func(n int) []Instr {
		instrs := make([]Instr, n)
		for i := range instrs {
			instrs[i] = LoadInt{Dest: Reg(i), Value: *big.NewInt(int64(i))}
		}
		return instrs
	}

	t.Run("255 registers fit", func(t *testing.T) {
		prog, err := Assemble(loadInts(255))
		require.NoError(t, err)
		require.Equal(t, byte(255), prog.MaxRegInt)
	})

	t.Run("256 do not", func(t *testing.T) {
		_, err := Assemble(loadInts(256))
		require.ErrorContains(t, err, "register bank overflow")
	})

	t.Run("banks are counted separately", func(t *testing.T) {
		instrs := loadInts(255)
		for i := range 255 {
			instrs = append(instrs, LoadStr{Dest: Reg(1000 + i), Value: "x"})
		}
		_, err := Assemble(instrs)
		require.NoError(t, err)
	})
}

func TestAssemble_JmpIfZeroDelta(t *testing.T) {
	t.Run("delta counts the instructions skipped", func(t *testing.T) {
		prog, err := Assemble([]Instr{
			LoadInt{Dest: 0, Value: *big.NewInt(0)}, // 0
			JmpIfZero{Cond: 0, Target: "end"},       // 1
			LoadInt{Dest: 1, Value: *big.NewInt(1)}, // 2
			LoadInt{Dest: 2, Value: *big.NewInt(2)}, // 3
			LabelMarker{Label: "end"},               // -> 4
		})
		require.NoError(t, err)

		require.Equal(t, vm.NewBC(vm.Op_JmpIfZero, 0, 2), prog.Instructions[1])
	})

	t.Run("jump to the immediately following instruction has delta 0", func(t *testing.T) {
		prog, err := Assemble([]Instr{
			LoadInt{Dest: 0, Value: *big.NewInt(0)},
			JmpIfZero{Cond: 0, Target: "end"},
			LabelMarker{Label: "end"},
			LoadInt{Dest: 1, Value: *big.NewInt(1)},
		})
		require.NoError(t, err)

		require.Equal(t, vm.NewBC(vm.Op_JmpIfZero, 0, 0), prog.Instructions[1])
	})

	t.Run("backward jump is rejected", func(t *testing.T) {
		_, err := Assemble([]Instr{
			LabelMarker{Label: "start"},
			LoadInt{Dest: 0, Value: *big.NewInt(0)},
			JmpIfZero{Cond: 0, Target: "start"},
		})
		require.ErrorContains(t, err, "backward jump")
	})

	t.Run("jump to itself is rejected", func(t *testing.T) {
		_, err := Assemble([]Instr{
			LoadInt{Dest: 0, Value: *big.NewInt(0)},
			LabelMarker{Label: "self"},
			JmpIfZero{Cond: 0, Target: "self"},
		})
		require.ErrorContains(t, err, "backward jump")
	})
}

// mk_allot reserves its blocks contiguously, so one that doesn't fit must error
// rather than wrap around the bank.
func TestAssemble_ContiguousOverflow(t *testing.T) {
	// one portion register, then an allotment whose dest block can't fit
	portion := []Instr{
		LoadInt{Dest: 0, Value: *big.NewInt(1)},
		BinaryOp{Op: OpMakePortion{}, Left: 0, Right: 0, Dest: 1},
	}

	dests := make([]Reg, 255)
	portions := make([]Reg, 255)
	for i := range dests {
		dests[i] = Reg(100 + i)
		portions[i] = 1
	}

	_, err := Assemble(append(portion, MakeAllotment{Dest: dests, Amount: 0, Portions: portions}))
	require.ErrorContains(t, err, "register bank overflow")
}
