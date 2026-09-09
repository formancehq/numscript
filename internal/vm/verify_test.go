package vm

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// fullBanks declares every bank at its maximum, so a test that isn't about the
// MaxReg check doesn't have to count registers.
func fullBanks(p Program) Program {
	p.MaxRegString, p.MaxRegInt, p.MaxRegPortion, p.MaxRegBool = 255, 255, 255, 255
	return p
}

func mustReject(t *testing.T, p Program) {
	t.Helper()
	require.Error(t, Verify(fullBanks(p)))
}

func mustAccept(t *testing.T, p Program) {
	t.Helper()
	require.NoError(t, Verify(fullBanks(p)))
}

func TestVerifyUnknownOpcode(t *testing.T) {
	mustReject(t, Program{Instructions: []Instruction{abc(0xFE, 0, 0, 0)}})
}

func TestVerifyTruncatedMultiWord(t *testing.T) {
	// each of these needs an ext word the stream doesn't have
	for _, op := range []Opcode{
		Op_PullAccount,
		Op_Save,
		Op_SetAccountMeta,
		Op_MetaStr,
		Op_MetaInt,
		Op_MetaPortion,
		Op_MetaMonetary,
		Op_Balance,
	} {
		mustReject(t, Program{
			Instructions: []Instruction{bc(Op_LoadStr, 0, 0), abc(op, 0, 0, 0)},
			StringsPool:  []string{"acc"},
		})
	}
}

func TestVerifyConstIndexOutOfRange(t *testing.T) {
	mustReject(t, Program{Instructions: []Instruction{bc(Op_LoadInt, 0, 3)}})
	mustReject(t, Program{Instructions: []Instruction{bc(Op_LoadStr, 0, 3)}})
}

func TestVerifyJumpPastEndIsFine(t *testing.T) {
	// pc runs off the end and the loop stops; nothing to guard
	mustAccept(t, Program{Instructions: []Instruction{
		abc(Op_ConstTrue, 0, nilReg, nilReg),
		bc(Op_JmpIfTrue, 0, 99),
	}})
}

func TestVerifyJumpIntoExtWord(t *testing.T) {
	// instruction 2 is Op_Balance's ext word, not an instruction boundary
	mustReject(t, Program{
		Instructions: []Instruction{
			abc(Op_ConstTrue, 0, nilReg, nilReg),
			bc(Op_JmpIfTrue, 0, 1),
			abc(Op_Balance, 0, 0, 0),
			abc(0, nilReg, nilReg, nilReg), // ext
		},
		StringsPool: []string{"acc"},
	})
}

func TestVerifyReadNotAssignedOnAllPaths(t *testing.T) {
	// int reg 1 is written only on the fall-through path, then read at the join
	mustReject(t, Program{
		Instructions: []Instruction{
			abc(Op_ConstTrue, 0, nilReg, nilReg), // 0: b0 = true
			bc(Op_JmpIfTrue, 0, 1),               // 1: skip instruction 2
			bc(Op_LoadInt, 1, 0),                 // 2: i1 = 0   (skipped when jumping)
			abc(Op_NegInt, 2, 1, nilReg),         // 3: i2 = -i1 (i1 maybe unassigned)
		},
		IntsPool: []big.Int{*big.NewInt(0)},
	})
}

func TestVerifyAssignedOnBothBranchesIsFine(t *testing.T) {
	// the same register written on either side of a diamond is assigned at the join
	mustAccept(t, Program{
		Instructions: []Instruction{
			abc(Op_ConstTrue, 0, nilReg, nilReg), // 0: b0 = true
			bc(Op_JmpIfTrue, 0, 2),               // 1: -> 4
			bc(Op_LoadInt, 1, 0),                 // 2: i1 = 0
			bc(Op_Jmp, 0, 1),                     // 3: -> 5
			bc(Op_LoadInt, 1, 0),                 // 4: i1 = 0
			abc(Op_NegInt, 2, 1, nilReg),         // 5: i2 = -i1
		},
		IntsPool: []big.Int{*big.NewInt(0)},
	})
}

// A type confusion at the bytecode level shows up as a read of a register that
// was never written: the bank is part of a register's identity, so string 0 and
// int 0 are different registers.
func TestVerifyTypeConfusionIsAnUnassignedRead(t *testing.T) {
	mustReject(t, Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, 0, 0),         // strings[0] = "x"
			abc(Op_NegInt, 1, 0, nilReg), // reads ints[0], never written
		},
		StringsPool: []string{"x"},
	})
}

func TestVerifyCurrentAssetNotSet(t *testing.T) {
	mustReject(t, Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, 0, 0),
			abc(Op_SendToAccount, 0, nilReg, nilReg),
		},
		StringsPool: []string{"dest"},
	})
}

func TestVerifyRegisterBeyondDeclaredMax(t *testing.T) {
	p := Program{
		Instructions: []Instruction{bc(Op_LoadInt, 5, 0)},
		IntsPool:     []big.Int{*big.NewInt(1)},
	}

	p.MaxRegInt = 5 // registers 0..4, so 5 is out
	require.Error(t, Verify(p))

	p.MaxRegInt = 6
	require.NoError(t, Verify(p))
}

func TestVerifyNilRegInNonOptionalOperand(t *testing.T) {
	// Op_SetCurrentAsset always dereferences A, so nilReg there is malformed
	mustReject(t, Program{Instructions: []Instruction{
		abc(Op_SetCurrentAsset, nilReg, nilReg, nilReg),
	}})
}

func TestVerifyNilRegInOptionalOperandIsFine(t *testing.T) {
	// Op_Save's amount is optional: nilReg means "save all"
	mustAccept(t, Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, 0, 0),
			bc(Op_LoadStr, 1, 1),
			abc(Op_Save, 0, 1, nilReg),
			abc(0, nilReg, nilReg, nilReg), // ext: no scope
		},
		StringsPool: []string{"acc", "USD/2"},
	})
}

func TestVerifyFlagOperand(t *testing.T) {
	// Exec tests `instr.A == 1`, so a 2 would silently commit a region meant to
	// rewind
	mustReject(t, Program{Instructions: []Instruction{
		abc(Op_MarkPush, nilReg, nilReg, nilReg),
		abc(Op_MarkEnd, 2, nilReg, nilReg),
	}})
	mustAccept(t, Program{Instructions: []Instruction{
		abc(Op_MarkPush, nilReg, nilReg, nilReg),
		abc(Op_MarkEnd, 1, nilReg, nilReg),
	}})
}

func TestVerifyWithVars(t *testing.T) {
	p := fullBanks(Program{Instructions: []Instruction{
		bc(Op_LoadVarInt, 0, 1), // reads int var 1, so 2 are needed
	}})

	require.NoError(t, Verify(p), "Verify alone says nothing about vars")

	require.Error(t, VerifyWithVars(p, nil))
	require.Error(t, VerifyWithVars(p, &Vars{IntsPool: []big.Int{*big.NewInt(0)}}))
	require.NoError(t, VerifyWithVars(p, &Vars{IntsPool: []big.Int{*big.NewInt(0), *big.NewInt(1)}}))
}

func TestVerifyWithVarsAllowsNilWhenNoneAreRead(t *testing.T) {
	require.NoError(t, VerifyWithVars(fullBanks(Program{
		Instructions: []Instruction{bc(Op_LoadInt, 0, 0)},
		IntsPool:     []big.Int{*big.NewInt(1)},
	}), nil))
}

func TestVerifyEmptyProgram(t *testing.T) {
	mustAccept(t, Program{})
}
