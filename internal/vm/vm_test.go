package vm

// White-box tests (package vm) that build a Program from struct literals, so
// they can reach encodings the compiler doesn't emit. Behavioural VM tests are
// written in the IR textual format instead — see ir_test.go.

import (
	"context"
	"math/big"
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

func abc(op Opcode, a, b, c byte) Instruction {
	return Instruction{Opcode: byte(op), A: a, B: b, C: c}
}

func bc(op Opcode, a byte, v uint16) Instruction {
	return Instruction{Opcode: byte(op), A: a, B: byte(v), C: byte(v >> 8)}
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

func assertValidAccountProgram(name string) Program {
	return Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, 0, 0),
			abc(Op_AssertValidAccount, 0, nilReg, nilReg),
		},
		StringsPool: []string{name},
	}
}

func balanceNonNegativeProgram() Program {
	return Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, 0, 0),
			bc(Op_LoadStr, 1, 1),
			abc(Op_Balance, 0, 0, 1),
			abc(Op_AssertNonNegativeBalance, 0, 0, nilReg),
		},
		StringsPool: []string{"acc", "USD/2"},
	}
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
	prog := Program{Instructions: []Instruction{abc(0xFE, 0, 0, 0)}}
	_, err := Exec(context.Background(), NewVm(prog), nil, mockStore{})
	if _, ok := err.(InternalError); !ok {
		t.Fatalf("expected InternalError, got %v", err)
	}
}

func TestMkPortionDivideByZero(t *testing.T) {
	prog := Program{
		Instructions: []Instruction{
			bc(Op_LoadInt, 0, 0),
			bc(Op_LoadInt, 1, 1),
			abc(Op_MkPortion, 0, 0, 1),
		},
		IntsPool: []big.Int{*big.NewInt(1), *big.NewInt(0)},
	}
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
