package compiler

import (
	"testing"

	"github.com/formancehq/numscript/internal/vm"
)

func TestAssemble_AddInt(t *testing.T) {
	// Three distinct virtual int registers map to the first three int-bank
	// indices in first-use order.
	prog, err := Assemble([]vInstr{
		binaryOp{op: opAddInt{}, dest: 10, left: 20, right: 30},
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
	prog, err := Assemble([]vInstr{
		// reg 7 -> 0, reg 8 -> 1 ; dest==left==7
		binaryOp{op: opAddInt{}, dest: 7, left: 7, right: 8},
		// reg 9 -> 2 ; reuses 7->0 and 8->1
		binaryOp{op: opAddInt{}, dest: 9, left: 7, right: 8},
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
