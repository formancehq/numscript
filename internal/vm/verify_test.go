package vm

import (
	"math/big"
	"testing"
)

func mustReject(t *testing.T, p Program) {
	t.Helper()
	_, err := Exec(NewVm(p), nil, mockStore{})
	if _, ok := err.(MalformedProgramError); !ok {
		t.Fatalf("expected MalformedProgramError, got %v", err)
	}
}

func TestVerify_UnknownOpcode(t *testing.T) {
	mustReject(t, Program{Instructions: []Instruction{abc(0xFE, 0, 0, 0)}})
}

func TestVerify_TruncatedMultiWord(t *testing.T) {
	// a lone Op_PullAccount with no trailing ext word
	mustReject(t, Program{Instructions: []Instruction{abc(Op_PullAccount, 0, 0, nilReg)}})
}

func TestVerify_ConstIndexOutOfRange(t *testing.T) {
	// LoadInt referring to pool index 3 in an empty pool
	mustReject(t, Program{Instructions: []Instruction{bc(Op_LoadInt, 0, 3)}})
}

func TestVerify_BackwardJump(t *testing.T) {
	mustReject(t, Program{Instructions: []Instruction{
		bc(Op_LoadInt, 0, 0),
		bc(Op_JmpIfZero, 0, 0), // targets instruction 0 (backward)
	}, IntsPool: []big.Int{*big.NewInt(0)}})
}

func TestVerify_JumpOutOfRange(t *testing.T) {
	mustReject(t, Program{Instructions: []Instruction{
		bc(Op_LoadInt, 0, 0),
		bc(Op_JmpIfZero, 0, 99),
	}, IntsPool: []big.Int{*big.NewInt(0)}})
}

func TestVerify_MissingVars(t *testing.T) {
	// reads var int 0 but no vars are passed
	_, err := Exec(NewVm(Program{Instructions: []Instruction{bc(Op_LoadVarInt, 0, 0)}}), nil, mockStore{})
	if _, ok := err.(MalformedProgramError); !ok {
		t.Fatalf("expected MalformedProgramError, got %v", err)
	}
}

func TestVerify_SizesBanksToNeed(t *testing.T) {
	// a program using int reg 5 must get an int bank of at least 6
	p := Program{Instructions: []Instruction{bc(Op_LoadInt, 5, 0)}, IntsPool: []big.Int{*big.NewInt(1)}}
	vm := NewVm(p)
	if _, err := Exec(vm, nil, mockStore{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vm.intsRegs) != 6 {
		t.Fatalf("int bank size = %d, want 6", len(vm.intsRegs))
	}
}
