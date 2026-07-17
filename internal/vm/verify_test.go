package vm

import (
	"math/big"
	"testing"
)

// mustReject asserts Verify rejects p. Programs are sized first (as the
// assembler would) so a rejection reflects the specific incoherence under test
// rather than an incidentally-too-small declared bank.
func mustReject(t *testing.T, p Program) {
	t.Helper()
	if err := sizeProgram(p).Verify(); err == nil {
		t.Fatalf("expected program to be rejected by Verify, got nil")
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
	// reads var int 0 but no vars are passed. This is an Exec-time guard (vars are
	// caller input, not part of the bytecode), not a Verify check.
	p := sizeProgram(Program{Instructions: []Instruction{bc(Op_LoadVarInt, 0, 0)}})
	_, err := Exec(t.Context(), NewVm(p), nil, mockStore{})
	if _, ok := err.(MalformedProgramError); !ok {
		t.Fatalf("expected MalformedProgramError, got %v", err)
	}
}

func TestVerify_ReadNotAssignedOnAllPaths(t *testing.T) {
	// r1 is written only on the fall-through path; the jump skips to instr 3
	// which reads it, so it is not assigned on every path.
	mustReject(t, Program{Instructions: []Instruction{
		bc(Op_LoadInt, 0, 0),         // 0: r0 = 0
		bc(Op_JmpIfZero, 0, 3),       // 1: if r0==0 skip to 3
		bc(Op_LoadInt, 1, 0),         // 2: r1 = 0 (skipped when jumping)
		abc(Op_NegInt, 2, 1, nilReg), // 3: r2 = -r1  (r1 maybe unassigned)
	}, IntsPool: []big.Int{*big.NewInt(0)}})
}

func TestVerify_CurrentAssetNotSet(t *testing.T) {
	// a send before any set_current_asset
	mustReject(t, Program{Instructions: []Instruction{
		bc(Op_LoadStr, 0, 0), // r0 = "dest"
		abc(Op_SendToAccount, 0, nilReg, nilReg),
	}, StringsPool: []string{"dest"}})
}

func TestNewVmSizesBanksFromDeclaredCounts(t *testing.T) {
	// NewVm allocates each bank to the program's declared count (no scanning).
	vm := NewVm(Program{IntRegs: 6, StrRegs: 2, PortionRegs: 1, MonetaryRegs: 3})
	if got := len(vm.intsRegs); got != 6 {
		t.Fatalf("int bank size = %d, want 6", got)
	}
	if got := len(vm.stringsRegs); got != 2 {
		t.Fatalf("string bank size = %d, want 2", got)
	}
	if got := len(vm.portionsRegs); got != 1 {
		t.Fatalf("portion bank size = %d, want 1", got)
	}
	if got := len(vm.monetariesRegs); got != 3 {
		t.Fatalf("monetary bank size = %d, want 3", got)
	}
}
