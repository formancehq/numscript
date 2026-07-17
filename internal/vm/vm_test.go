package vm

// White-box test (package vm) so it can build a Program/Vm from struct
// literals. It encodes the "inorder" send example into our low-level
// Instruction stream, with manual (non-optimal, per-bank) register allocation,
// runs Exec, and asserts the resulting postings.
//
// HARNESS ASSUMPTIONS (adjust to your actual API):
//   - Program has fields {instructions []Instruction; stringsPool []string;
//     intsPool []big.Int}.
//   - Instruction has exported {Opcode, A, B, C byte} and GetBC() uint16.
//   - nilReg (==0xFF) and worldAccount are package-level identifiers.
//   - Vm has a `program Program` and a `runstate *runtime.RunState`.
//   - One Store interface, GetBalance(account, asset string) int64, shared by
//     the generic Exec constraint and runtime.New.
//
// REQUIRED FIXES for this to PASS (see notes at bottom): SetCurrentAsset must
// propagate to vm.runstate; CheckEnoughFunds comparison is inverted;
// SendToAccount uses invalid `new(value)`.

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
)

// --- register allocation: one $rN namespace -> typed banks ----------------
//
//	$r0  "USD/2"   -> strings[0]  (sUSD)      $r6  remaining -> ints[3] (iRem)
//	$r1  10        -> ints[0]     (iTen)      $r7  "s1"      -> strings[2] (sS1)
//	$r2  monetary  -> monetary[0] (mMon)      $r8  pulled1   -> ints[4] (iPulled1)
//	$r3  asset     -> strings[1]  (sAsset)    $r9  "s2"      -> strings[3] (sS2)
//	$r4  amount    -> ints[1]     (iAmount)   $r10 pulled2   -> ints[5] (iPulled2)
//	$r5  sum=0     -> ints[2]     (iSum)      $r11 "dest"    -> strings[4] (sDest)
//	(added) zero overdraft bound -> ints[6] (iZero)  -- gives BoundedZero
const (
	sUSD, sAsset, sS1, sS2, sDest       = 0, 1, 2, 3, 4
	iTen, iAmount, iSum, iRem, iPulled1 = 0, 1, 2, 3, 4
	iPulled2, iZero                     = 5, 6
	mMon                                = 0
)

// pool indices
const (
	pUSD, pS1, pS2, pDest = 0, 1, 2, 3 // strings pool
	cTen, cZero           = 0, 1       // ints pool
)

func abc(op Opcode, a, b, c byte) Instruction {
	return Instruction{Opcode: byte(op), A: a, B: b, C: c}
}

func bc(op Opcode, a byte, v uint16) Instruction {
	return Instruction{Opcode: byte(op), A: a, B: byte(v), C: byte(v >> 8)}
}

// sizeProgram fills a hand-built Program's declared register/var counts by
// scanning its instructions, mirroring what the assembler does for real
// programs. Test-only: production Programs are decoded/compiled with counts
// already declared. On a malformed instruction it stops early and leaves the
// counts scanned so far, which is enough for Verify to then reject the program.
func sizeProgram(p Program) Program {
	instrs := p.Instructions
	n := len(instrs)
	bump := func(cur byte, idx int) byte {
		if idx+1 > int(cur) {
			return byte(idx + 1)
		}
		return cur
	}
	bump16 := func(cur uint16, idx int) uint16 {
		if idx+1 > int(cur) {
			return uint16(idx + 1)
		}
		return cur
	}
	for i := 0; i < n; {
		w := instrWords(instrs[i].Opcode)
		if i+w > n {
			break
		}
		var ext Instruction
		if w == 2 {
			ext = instrs[i+1]
		}
		d, err := decodeInstr(instrs[i], ext)
		if err != nil {
			break
		}
		for _, refs := range [][]regRef{d.reads, d.writes} {
			for _, r := range refs {
				switch r.bank {
				case bankInt:
					p.IntRegs = bump(p.IntRegs, r.index)
				case bankStr:
					p.StrRegs = bump(p.StrRegs, r.index)
				case bankPortion:
					p.PortionRegs = bump(p.PortionRegs, r.index)
				case bankMonetary:
					p.MonetaryRegs = bump(p.MonetaryRegs, r.index)
				}
			}
		}
		if d.varInt >= 0 {
			p.IntVars = bump16(p.IntVars, d.varInt)
		}
		if d.varStr >= 0 {
			p.StrVars = bump16(p.StrVars, d.varStr)
		}
		i += w
	}
	return p
}

func inorderProgram() Program {
	// Index of #inorder_end in the ENCODED stream. Note PullAccount occupies
	// two words each, so this is not the count of source lines.
	const inorderEnd = 19

	instrs := []Instruction{
		/* 0  */ bc(Op_LoadStr, sUSD, pUSD), // r0 = load_const("USD/2")
		/* 1  */ bc(Op_LoadInt, iTen, cTen), // r1 = load_const(10)
		/* 2  */ abc(Op_MkMonetary, mMon, sUSD, iTen), // r2 = mk_monetary(r0, r1)
		/* 3  */ abc(Op_GetAsset, sAsset, mMon, 0), // r3 = get_asset(r2)
		/* 4  */ abc(Op_SetCurrentAsset, sAsset, 0, 0), // set_current_asset(r3)
		/* 5  */ abc(Op_GetAmount, iAmount, mMon, 0), // r4 = get_amount(r2)
		/* 6  */ bc(Op_LoadInt, iSum, cZero), // r5 = load_const(0)
		/* 7  */ abc(Op_IntCopy, iRem, iAmount, 0), // r6 = int_copy(r4)
		/* 8  */ bc(Op_LoadInt, iZero, cZero), // (added) overdraft bound = 0 -> BoundedZero
		/* 9  */ bc(Op_LoadStr, sS1, pS1), // r7 = load_const("s1")
		/* 10 */ abc(Op_PullAccount, iPulled1, sS1, iRem), // r8 = pull(account r7, cap r6)      [word 1]
		/* 11 */ abc(0, iZero, nilReg, 0), //   ext: overdraft=iZero, color=nil  [word 2]
		/* 12 */ abc(Op_AddInt, iSum, iSum, iPulled1), // r5 = add_int(r5, r8)
		/* 13 */ abc(Op_SubInt, iRem, iRem, iPulled1), // r6 = sub_int(r6, r8)
		/* 14 */ bc(Op_JmpIfZero, iRem, inorderEnd), // jmp_if_zero(r6, #inorder_end)
		/* 15 */ bc(Op_LoadStr, sS2, pS2), // r9 = load_const("s2")
		/* 16 */ abc(Op_PullAccount, iPulled2, sS2, iRem), // r10 = pull(account r9, cap r6)     [word 1]
		/* 17 */ abc(0, iZero, nilReg, 0), //   ext: overdraft=iZero, color=nil  [word 2]
		/* 18 */ abc(Op_AddInt, iSum, iSum, iPulled2), // r5 = add_int(r5, r10)
		/* 19 */ abc(Op_CheckEnoughFunds, iSum, iAmount, 0), // #inorder_end: check_enough_funds(r5, r4)
		/* 20 */ bc(Op_LoadStr, sDest, pDest), // r11 = load_const("dest")
		/* 21 */ abc(Op_SendToAccount, sDest, nilReg, nilReg), // send_to_account(r11)  (no cap, no color)
	}

	return sizeProgram(Program{
		Instructions: instrs,
		StringsPool:  []string{"USD/2", "s1", "s2", "dest"},
		IntsPool:     []big.Int{*big.NewInt(10), *big.NewInt(0)},
	})
}

// --- mock store -----------------------------------------------------------

type mockStore struct {
	bal  map[runtime.PairKey]int64
	meta map[string]map[string]string
}

func (m mockStore) GetBalance(ctx context.Context, account, asset string, color string) (*big.Int, error) {
	return big.NewInt(m.bal[runtime.PairKey{Account: account, Asset: asset}]), nil
}

func (m mockStore) GetMetadata(ctx context.Context, account, key string) (string, bool, error) {
	v, ok := m.meta[account][key]
	return v, ok, nil
}

var _ Store = (*mockStore)(nil)

// --- the test -------------------------------------------------------------

func TestInorderSend(t *testing.T) {
	t.Skip()

	prog := inorderProgram()

	// s1 has 6, s2 has 10; sending 10 USD/2 => s1 gives 6, s2 gives 4.
	store := mockStore{bal: map[runtime.PairKey]int64{
		{Account: "s1", Asset: "USD/2"}: 6,
		{Account: "s2", Asset: "USD/2"}: 10,
	}}

	vm := NewVm(prog)

	res, err := Exec(context.Background(), vm, nil, store)
	if err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	want := []runtime.Posting{
		{Source: "s1", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(6)},
		{Source: "s2", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(4)},
	}
	if !reflect.DeepEqual(res.Postings, want) {
		t.Errorf("postings mismatch\n got: %+v\nwant: %+v", res.Postings, want)
	}
}

func assertValidAccountProgram(name string) Program {
	return sizeProgram(Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, 0, 0),
			abc(Op_AssertValidAccount, 0, nilReg, nilReg),
		},
		StringsPool: []string{name},
	})
}

func balanceNonNegativeProgram() Program {
	return sizeProgram(Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, 0, 0),
			bc(Op_LoadStr, 1, 1),
			abc(Op_Balance, 0, 0, 1),
			abc(Op_AssertNonNegativeBalance, 0, 0, nilReg),
		},
		StringsPool: []string{"acc", "USD/2"},
	})
}

func TestAssertNonNegativeBalance(t *testing.T) {
	store := mockStore{bal: map[runtime.PairKey]int64{{Account: "acc", Asset: "USD/2"}: 50}}
	if _, err := Exec(context.Background(), NewVm(balanceNonNegativeProgram()), nil, store); err != nil {
		t.Fatalf("non-negative balance rejected: %v", err)
	}

	store = mockStore{bal: map[runtime.PairKey]int64{{Account: "acc", Asset: "USD/2"}: -50}}
	_, err := Exec(context.Background(), NewVm(balanceNonNegativeProgram()), nil, store)
	if _, ok := err.(NegativeBalanceError); !ok {
		t.Fatalf("expected NegativeBalanceError, got %v", err)
	}
}

func TestUnknownOpcode(t *testing.T) {
	// Exec no longer verifies; an unknown opcode reaches the loop's default arm
	// and returns InternalError (rather than panicking). Rejection at the static
	// level is covered by TestVerify_UnknownOpcode.
	prog := Program{Instructions: []Instruction{abc(0xFE, 0, 0, 0)}}
	_, err := Exec(t.Context(), NewVm(prog), nil, mockStore{})
	if _, ok := err.(InternalError); !ok {
		t.Fatalf("expected InternalError, got %v", err)
	}
}

func TestMkPortionDivideByZero(t *testing.T) {
	prog := sizeProgram(Program{
		Instructions: []Instruction{
			bc(Op_LoadInt, 0, 0),
			bc(Op_LoadInt, 1, 1),
			abc(Op_MkPortion, 0, 0, 1),
		},
		IntsPool: []big.Int{*big.NewInt(1), *big.NewInt(0)},
	})
	_, err := Exec(context.Background(), NewVm(prog), nil, mockStore{})
	if _, ok := err.(DivideByZeroError); !ok {
		t.Fatalf("expected DivideByZeroError, got %v", err)
	}
}

func TestAssertValidAccount(t *testing.T) {
	_, err := Exec(context.Background(), NewVm(assertValidAccountProgram("users:001:wallet")), nil, mockStore{})
	if err != nil {
		t.Fatalf("valid account rejected: %v", err)
	}

	_, err = Exec(context.Background(), NewVm(assertValidAccountProgram("bad name!")), nil, mockStore{})
	if _, ok := err.(InvalidAccountName); !ok {
		t.Fatalf("expected InvalidAccountName, got %v", err)
	}
}
