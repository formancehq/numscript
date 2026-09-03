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
		BinaryOp{Op: OpMakePortion{}, Left: 1, Right: 1, Dest: 3},
	})
	require.NoError(t, err)

	require.Equal(t, byte(1), prog.MaxRegString, "one str reg")
	require.Equal(t, byte(1), prog.MaxRegInt, "one int reg")
	require.Equal(t, byte(1), prog.MaxRegPortion, "one portion reg")
	require.Equal(t, byte(0), prog.MaxRegBool, "no bool reg")
}

// The bool value is in the opcode, so the two constants differ only there,
// and bool registers are indexed in their own bank. Neither constant is
// read again after being set, so the allocator frees the first one's slot
// right after it and reuses it for the second — they end up sharing bank
// index 0.
func TestAssemble_ConstBool(t *testing.T) {
	prog, err := Assemble([]Instr{
		LoadInt{Dest: 0, Value: *big.NewInt(1)},
		ConstBool{Dest: 1, Value: true},
		ConstBool{Dest: 2, Value: false},
	})
	require.NoError(t, err)

	require.Equal(t, []vm.Instruction{
		vm.NewBC(vm.Op_LoadInt, 0, 0),
		{Opcode: byte(vm.Op_ConstTrue), A: 0, B: 0xFF, C: 0xFF},
		{Opcode: byte(vm.Op_ConstFalse), A: 0, B: 0xFF, C: 0xFF},
	}, prog.Instructions)

	require.Equal(t, byte(1), prog.MaxRegInt)
	require.Equal(t, byte(1), prog.MaxRegBool)
}

// 255 registers per bank, because 0xFF is the "operand unset" sentinel. The
// allocator reuses a slot once its register's last reference has passed, so
// the cap is only reachable when that many registers are genuinely live at
// once — not merely ever defined (see TestAssemble_SingleUseRegistersDontAccumulate).
func TestAssemble_RegisterBankOverflow(t *testing.T) {
	// liveInts defines n distinct int registers, then reads every one of
	// them back (via CheckEnoughFunds, which takes two int operands and
	// produces no destination register, so the read-back itself never adds
	// extra pressure), so all n stay simultaneously live right up until the
	// last definition runs — the peak liveness a real allocator has to
	// accommodate.
	liveInts := func(n int) []Instr {
		instrs := make([]Instr, 0, 2*n)
		for i := range n {
			instrs = append(instrs, LoadInt{Dest: Reg(i), Value: *big.NewInt(int64(i))})
		}
		for i := range n {
			instrs = append(instrs, CheckEnoughFunds{Got: Reg(i), Needed: Reg(i)})
		}
		return instrs
	}

	t.Run("255 simultaneously live registers fit", func(t *testing.T) {
		prog, err := Assemble(liveInts(255))
		require.NoError(t, err)
		require.Equal(t, byte(255), prog.MaxRegInt)
	})

	t.Run("256 simultaneously live registers do not", func(t *testing.T) {
		_, err := Assemble(liveInts(256))
		require.ErrorContains(t, err, "register bank overflow")
	})

	t.Run("banks are counted separately", func(t *testing.T) {
		instrs := liveInts(255)
		for i := range 255 {
			instrs = append(instrs, LoadStr{Dest: Reg(2000 + i), Value: "x"})
		}
		for i := range 255 {
			instrs = append(instrs, AssertSameAsset{Left: Reg(2000 + i), Right: Reg(2000 + i)})
		}
		_, err := Assemble(instrs)
		require.NoError(t, err)
	})
}

// Unlike liveInts above, each register here is used exactly once, at
// definition, and never read again — so however many are defined, only one
// physical slot is ever needed at a time.
func TestAssemble_SingleUseRegistersDontAccumulate(t *testing.T) {
	instrs := make([]Instr, 300)
	for i := range instrs {
		instrs[i] = LoadInt{Dest: Reg(i), Value: *big.NewInt(int64(i))}
	}
	prog, err := Assemble(instrs)
	require.NoError(t, err)
	require.Equal(t, byte(1), prog.MaxRegInt)
}

func TestAssemble_JmpDelta(t *testing.T) {
	t.Run("delta counts the instructions skipped", func(t *testing.T) {
		prog, err := Assemble([]Instr{
			ConstBool{Dest: 0, Value: true},         // 0
			JmpIfFalse{Cond: 0, Target: "end"},      // 1
			LoadInt{Dest: 1, Value: *big.NewInt(1)}, // 2
			LoadInt{Dest: 2, Value: *big.NewInt(2)}, // 3
			LabelMarker{Label: "end"},               // -> 4
		})
		require.NoError(t, err)

		require.Equal(t, vm.NewBC(vm.Op_JmpIfFalse, 0, 2), prog.Instructions[1])
	})

	// the two conditional jumps differ only in the opcode
	t.Run("jmp_if_true emits its own opcode", func(t *testing.T) {
		prog, err := Assemble([]Instr{
			ConstBool{Dest: 0, Value: true},
			JmpIfTrue{Cond: 0, Target: "end"},
			LoadInt{Dest: 1, Value: *big.NewInt(1)},
			LabelMarker{Label: "end"},
		})
		require.NoError(t, err)

		require.Equal(t, vm.NewBC(vm.Op_JmpIfTrue, 0, 1), prog.Instructions[1])
	})

	t.Run("jump to the immediately following instruction has delta 0", func(t *testing.T) {
		prog, err := Assemble([]Instr{
			ConstBool{Dest: 0, Value: true},
			JmpIfFalse{Cond: 0, Target: "end"},
			LabelMarker{Label: "end"},
			LoadInt{Dest: 1, Value: *big.NewInt(1)},
		})
		require.NoError(t, err)

		require.Equal(t, vm.NewBC(vm.Op_JmpIfFalse, 0, 0), prog.Instructions[1])
	})

	t.Run("backward jump is rejected", func(t *testing.T) {
		_, err := Assemble([]Instr{
			LabelMarker{Label: "start"},
			ConstBool{Dest: 0, Value: true},
			JmpIfFalse{Cond: 0, Target: "start"},
		})
		require.ErrorContains(t, err, "backward jump")
	})

	t.Run("jump to itself is rejected", func(t *testing.T) {
		_, err := Assemble([]Instr{
			ConstBool{Dest: 0, Value: true},
			LabelMarker{Label: "self"},
			JmpIfTrue{Cond: 0, Target: "self"},
		})
		require.ErrorContains(t, err, "backward jump")
	})

	t.Run("unconditional jmp patches its delta", func(t *testing.T) {
		prog, err := Assemble([]Instr{
			Jmp{Target: "end"},
			LoadInt{Dest: 0, Value: *big.NewInt(0)},
			LabelMarker{Label: "end"},
		})
		require.NoError(t, err)

		// one instruction (the load) sits between the jump and the label
		require.Equal(t, vm.NewBC(vm.Op_Jmp, 0, 1), prog.Instructions[0])
	})

	t.Run("backward unconditional jmp is rejected", func(t *testing.T) {
		_, err := Assemble([]Instr{
			LabelMarker{Label: "start"},
			Jmp{Target: "start"},
		})
		require.ErrorContains(t, err, "backward jump")
	})
}
