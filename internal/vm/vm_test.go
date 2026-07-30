package vm

// White-box tests (package vm) that build a Program from struct literals, so
// they can reach encodings the compiler doesn't emit. Behavioural VM tests are
// written in the IR textual format instead — see ir_test.go.

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
	"github.com/stretchr/testify/require"
)

// --- register allocation: one $rN namespace -> typed banks ----------------
//
//	$r0  "USD/2"   -> strings[0]  (sUSD)      $r6  remaining -> ints[3] (iRem)
//	$r1  10        -> ints[0]     (iTen)      $r7  "s1"      -> strings[2] (sS1)
//	$r3  asset     -> strings[1]  (sAsset)    $r8  pulled1   -> ints[4] (iPulled1)
//	$r4  amount    -> ints[1]     (iAmount)   $r9  "s2"      -> strings[3] (sS2)
//	$r5  sum=0     -> ints[2]     (iSum)      $r10 pulled2   -> ints[5] (iPulled2)
//	                                          $r11 "dest"    -> strings[4] (sDest)
//	(added) zero overdraft bound -> ints[6] (iZero)  -- gives BoundedZero
const (
	sUSD, sAsset, sS1, sS2, sDest       = 0, 1, 2, 3, 4
	iTen, iAmount, iSum, iRem, iPulled1 = 0, 1, 2, 3, 4
	iPulled2, iZero                     = 5, 6
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

// Nothing reads a bool yet, so the bank itself is the only observable effect.
func TestConstBool(t *testing.T) {
	prog := Program{
		Instructions: []Instruction{
			abc(Op_ConstTrue, 0, nilReg, nilReg),
			abc(Op_ConstFalse, 1, nilReg, nilReg),
			// a register written twice keeps the last value
			abc(Op_ConstTrue, 2, nilReg, nilReg),
			abc(Op_ConstFalse, 2, nilReg, nilReg),
		},
	}

	vm := NewVm(prog)
	_, err := Exec(context.Background(), vm, nil, mockStore{})
	require.Nil(t, err)

	require.True(t, vm.boolsRegs[0])
	require.False(t, vm.boolsRegs[1])
	require.False(t, vm.boolsRegs[2])
	require.False(t, vm.boolsRegs[3], "untouched registers stay false")
}

// is_zero is the only projection from a quantity to a condition, so it has to
// agree with what Op_JmpIfZero used to test: sign, not magnitude.
func TestIsZero(t *testing.T) {
	testCases := []struct {
		name  string
		value int64
		want  bool
	}{
		{"zero", 0, true},
		{"positive", 7, false},
		{"negative", -7, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prog := Program{
				Instructions: []Instruction{
					bc(Op_LoadInt, 0, 0),
					abc(Op_IsZero, 0, 0, nilReg),
				},
				IntsPool: []big.Int{*big.NewInt(tc.value)},
			}

			vm := NewVm(prog)
			_, err := Exec(context.Background(), vm, nil, mockStore{})
			require.Nil(t, err)
			require.Equal(t, tc.want, vm.boolsRegs[0])
		})
	}
}

// Portion addition and subtraction, over unequal denominators so the result has
// to be a real rational sum rather than a numerator-wise one.
func TestPortionArithmetic(t *testing.T) {
	testCases := []struct {
		name             string
		op               Opcode
		numL, denL       int64
		numR, denR       int64
		wantNum, wantDen int64
	}{
		{"add with equal denominators", Op_AddPortion, 1, 4, 1, 4, 1, 2},
		{"add with unequal denominators", Op_AddPortion, 1, 6, 1, 3, 1, 2},
		{"add to a whole", Op_AddPortion, 1, 3, 2, 3, 1, 1},
		{"add past a whole", Op_AddPortion, 3, 4, 1, 2, 5, 4},
		{"sub with unequal denominators", Op_SubPortion, 1, 2, 1, 6, 1, 3},
		{"sub to zero", Op_SubPortion, 1, 3, 1, 3, 0, 1},
		{"sub below zero", Op_SubPortion, 1, 4, 1, 2, -1, 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prog := Program{
				Instructions: []Instruction{
					bc(Op_LoadInt, 0, 0), bc(Op_LoadInt, 1, 1),
					bc(Op_LoadInt, 2, 2), bc(Op_LoadInt, 3, 3),
					abc(Op_MkPortion, 0, 0, 1),
					abc(Op_MkPortion, 1, 2, 3),
					abc(tc.op, 2, 0, 1),
				},
				IntsPool: []big.Int{
					*big.NewInt(tc.numL), *big.NewInt(tc.denL),
					*big.NewInt(tc.numR), *big.NewInt(tc.denR),
				},
			}

			vm := NewVm(prog)
			_, err := Exec(context.Background(), vm, nil, mockStore{})
			require.Nil(t, err)
			want := big.NewRat(tc.wantNum, tc.wantDen)
			require.Zero(t, vm.portionsRegs[2].Cmp(want),
				"got %s, want %s", vm.portionsRegs[2].RatString(), want.RatString())
		})
	}
}

// One copy per bank. Each case writes a distinct value into reg 1, copies reg 1
// into reg 0, and checks reg 0 took it — so a copy wired to the wrong bank, or a
// no-op, fails.
func TestBankCopies(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		prog := Program{
			Instructions: []Instruction{
				bc(Op_LoadInt, 1, 0),
				abc(Op_IntCopy, 0, 1, nilReg),
			},
			IntsPool: []big.Int{*big.NewInt(-42)},
		}
		vm := NewVm(prog)
		_, err := Exec(context.Background(), vm, nil, mockStore{})
		require.Nil(t, err)
		require.Zero(t, vm.intsRegs[0].Cmp(big.NewInt(-42)))
	})

	t.Run("portion", func(t *testing.T) {
		prog := Program{
			Instructions: []Instruction{
				bc(Op_LoadInt, 0, 0),
				bc(Op_LoadInt, 1, 1),
				abc(Op_MkPortion, 1, 0, 1),
				abc(Op_PortionCopy, 0, 1, nilReg),
			},
			IntsPool: []big.Int{*big.NewInt(1), *big.NewInt(3)},
		}
		vm := NewVm(prog)
		_, err := Exec(context.Background(), vm, nil, mockStore{})
		require.Nil(t, err)
		require.Zero(t, vm.portionsRegs[0].Cmp(big.NewRat(1, 3)))
	})

	t.Run("str", func(t *testing.T) {
		prog := Program{
			Instructions: []Instruction{
				bc(Op_LoadStr, 1, 0),
				abc(Op_StrCopy, 0, 1, nilReg),
			},
			StringsPool: []string{"USD/2"},
		}
		vm := NewVm(prog)
		_, err := Exec(context.Background(), vm, nil, mockStore{})
		require.Nil(t, err)
		require.Equal(t, "USD/2", vm.stringsRegs[0])
	})

	t.Run("bool", func(t *testing.T) {
		prog := Program{
			Instructions: []Instruction{
				abc(Op_ConstTrue, 1, nilReg, nilReg),
				abc(Op_BoolCopy, 0, 1, nilReg),
				// and the false direction, over a register that already held true
				abc(Op_ConstTrue, 2, nilReg, nilReg),
				abc(Op_ConstFalse, 3, nilReg, nilReg),
				abc(Op_BoolCopy, 2, 3, nilReg),
			},
		}
		vm := NewVm(prog)
		_, err := Exec(context.Background(), vm, nil, mockStore{})
		require.Nil(t, err)
		require.True(t, vm.boolsRegs[0])
		require.False(t, vm.boolsRegs[2], "copying false over true")
	})
}

// A copy is a copy, not an alias: overwriting the source must not disturb the
// destination. Only the int and portion banks can get this wrong, since those two
// hold big values that are Set() into place rather than assigned.
func TestCopiesAreNotAliases(t *testing.T) {
	prog := Program{
		Instructions: []Instruction{
			bc(Op_LoadInt, 1, 0),          // $1 = 7
			abc(Op_IntCopy, 0, 1, nilReg), // $0 = copy $1
			bc(Op_LoadInt, 1, 1),          // $1 = 9
		},
		IntsPool: []big.Int{*big.NewInt(7), *big.NewInt(9)},
	}
	vm := NewVm(prog)
	_, err := Exec(context.Background(), vm, nil, mockStore{})
	require.Nil(t, err)
	require.Zero(t, vm.intsRegs[0].Cmp(big.NewInt(7)), "the copy tracked its source")
	require.Zero(t, vm.intsRegs[1].Cmp(big.NewInt(9)))
}

// The two int comparisons, over both signs. Each case also asserts the negation,
// so `!=` — which has no opcode — is covered wherever `==` is.
func TestIntComparisons(t *testing.T) {
	testCases := []struct {
		name        string
		op          Opcode
		left, right int64
		want        bool
	}{
		{"lt when less", Op_LtInt, 3, 7, true},
		{"lt when equal", Op_LtInt, 7, 7, false},
		{"lt when greater", Op_LtInt, 7, 3, false},
		{"lt across zero", Op_LtInt, -7, 3, true},
		{"lt on negatives", Op_LtInt, -7, -3, true},

		{"eq when equal", Op_EqInt, 7, 7, true},
		{"eq when different", Op_EqInt, 7, 3, false},
		{"eq on negatives", Op_EqInt, -7, -7, true},
		{"eq distinguishes sign", Op_EqInt, -7, 7, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prog := Program{
				Instructions: []Instruction{
					bc(Op_LoadInt, 0, 0),
					bc(Op_LoadInt, 1, 1),
					abc(tc.op, 0, 0, 1),
					abc(Op_Not, 1, 0, nilReg),
				},
				IntsPool: []big.Int{*big.NewInt(tc.left), *big.NewInt(tc.right)},
			}

			vm := NewVm(prog)
			_, err := Exec(context.Background(), vm, nil, mockStore{})
			require.Nil(t, err)
			require.Equal(t, tc.want, vm.boolsRegs[0])
			require.Equal(t, !tc.want, vm.boolsRegs[1], "not")
		})
	}
}

// The four derived operators have no opcodes: the front end normalises them onto
// Lt / Eq / Not. This checks each lowering against the operator it stands for,
// which is the property that makes leaving them out safe.
func TestDerivedComparisonLowerings(t *testing.T) {
	// each lowering as it would be emitted, over a grid that covers <, == and >
	values := []int64{-7, -1, 0, 1, 7}

	testCases := []struct {
		name string
		emit []Instruction // leaves the answer in bool reg 0
		want func(l, r int64) bool
	}{
		{
			name: "a > b  ->  Lt(b, a)",
			emit: []Instruction{abc(Op_LtInt, 0, 1, 0)},
			want: func(l, r int64) bool { return l > r },
		},
		{
			name: "a <= b  ->  Not(Lt(b, a))",
			emit: []Instruction{abc(Op_LtInt, 1, 1, 0), abc(Op_Not, 0, 1, nilReg)},
			want: func(l, r int64) bool { return l <= r },
		},
		{
			name: "a >= b  ->  Not(Lt(a, b))",
			emit: []Instruction{abc(Op_LtInt, 1, 0, 1), abc(Op_Not, 0, 1, nilReg)},
			want: func(l, r int64) bool { return l >= r },
		},
		{
			name: "a != b  ->  Not(Eq(a, b))",
			emit: []Instruction{abc(Op_EqInt, 1, 0, 1), abc(Op_Not, 0, 1, nilReg)},
			want: func(l, r int64) bool { return l != r },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, l := range values {
				for _, r := range values {
					instrs := []Instruction{bc(Op_LoadInt, 0, 0), bc(Op_LoadInt, 1, 1)}
					prog := Program{
						Instructions: append(instrs, tc.emit...),
						IntsPool:     []big.Int{*big.NewInt(l), *big.NewInt(r)},
					}

					vm := NewVm(prog)
					_, err := Exec(context.Background(), vm, nil, mockStore{})
					require.Nil(t, err)
					require.Equal(t, tc.want(l, r), vm.boolsRegs[0], "l=%d r=%d", l, r)
				}
			}
		})
	}
}

// Portion comparison is value comparison: big.Rat normalises on construction, so
// equal rationals with different spellings must compare equal.
func TestPortionComparisons(t *testing.T) {
	// builds two portions from (numL/denL, numR/denR) and compares them
	run := func(t *testing.T, op Opcode, numL, denL, numR, denR int64) bool {
		t.Helper()
		prog := Program{
			Instructions: []Instruction{
				bc(Op_LoadInt, 0, 0), bc(Op_LoadInt, 1, 1),
				bc(Op_LoadInt, 2, 2), bc(Op_LoadInt, 3, 3),
				abc(Op_MkPortion, 0, 0, 1), // portion 0 = numL/denL
				abc(Op_MkPortion, 1, 2, 3), // portion 1 = numR/denR
				abc(op, 0, 0, 1),
			},
			IntsPool: []big.Int{
				*big.NewInt(numL), *big.NewInt(denL),
				*big.NewInt(numR), *big.NewInt(denR),
			},
		}
		vm := NewVm(prog)
		_, err := Exec(context.Background(), vm, nil, mockStore{})
		require.Nil(t, err)
		return vm.boolsRegs[0]
	}

	t.Run("equality is by value, not by numerator/denominator", func(t *testing.T) {
		require.True(t, run(t, Op_EqPortion, 1, 2, 2, 4), "1/2 == 2/4")
		require.True(t, run(t, Op_EqPortion, 3, 9, 1, 3), "3/9 == 1/3")
		require.False(t, run(t, Op_EqPortion, 1, 2, 1, 3), "1/2 != 1/3")
	})

	t.Run("ordering", func(t *testing.T) {
		require.True(t, run(t, Op_LtPortion, 1, 3, 1, 2), "1/3 < 1/2")
		require.False(t, run(t, Op_LtPortion, 1, 2, 1, 3), "1/2 not < 1/3")
		require.False(t, run(t, Op_LtPortion, 1, 2, 2, 4), "equal values are not <")
	})
}

// The two conditional jumps are duals: each takes the edge the other doesn't.
func TestConditionalJumps(t *testing.T) {
	// jump over a CONST_TRUE writing bool reg 1, so reg 1 reports whether the
	// jump was taken
	prog := func(jmp Opcode, cond Opcode) Program {
		return Program{
			Instructions: []Instruction{
				abc(cond, 0, nilReg, nilReg),
				bc(jmp, 0, 1),
				abc(Op_ConstTrue, 1, nilReg, nilReg),
			},
		}
	}

	testCases := []struct {
		name  string
		jmp   Opcode
		cond  Opcode
		taken bool
	}{
		{"jmp_if_false on false", Op_JmpIfFalse, Op_ConstFalse, true},
		{"jmp_if_false on true", Op_JmpIfFalse, Op_ConstTrue, false},
		{"jmp_if_true on true", Op_JmpIfTrue, Op_ConstTrue, true},
		{"jmp_if_true on false", Op_JmpIfTrue, Op_ConstFalse, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vm := NewVm(prog(tc.jmp, tc.cond))
			_, err := Exec(context.Background(), vm, nil, mockStore{})
			require.Nil(t, err)
			require.Equal(t, tc.taken, !vm.boolsRegs[1], "jump taken")
		})
	}
}

func TestExecutionErrorMessages(t *testing.T) {
	testCases := []struct {
		name string
		err  ExecutionError
		msg  string
	}{
		{"MissingFundsError", MissingFundsError{Asset: "USD/2", Needed: big.NewInt(10), Got: big.NewInt(4)},
			"missing funds for asset USD/2: needed 10, got 4"},
		{"AssetMismatchError", AssetMismatchError{Expected: "USD/2", Got: "EUR/2"},
			"asset mismatch: expected USD/2, got EUR/2"},
		{"InvalidUncappedSource", InvalidUncappedSource{Account: "src"},
			"unbounded source is not allowed here: @src"},
		{"InvalidAllotmentSum", InvalidAllotmentSum{ActualSum: *big.NewRat(3, 2)},
			"invalid allotment: portions must sum to 1, got 3/2"},
		{"MetadataNotFoundError", MetadataNotFoundError{Account: "acc", Key: "k"},
			`metadata not found: acc["k"]`},
		{"BadMetaValueError", BadMetaValueError{Account: "acc", Key: "k", Raw: "oops"},
			`invalid metadata value for acc["k"]: "oops"`},
		{"InvalidAccountName", InvalidAccountName{Name: "not an account"},
			`invalid account name: "not an account"`},
		{"InvalidColor", InvalidColor{Color: "red"},
			`invalid color name: "red"`},
		{"NegativeBalanceError", NegativeBalanceError{Account: "src", Amount: *big.NewInt(-1)},
			"cannot fetch negative balance from account @src"},
		{"DivideByZeroError", DivideByZeroError{Numerator: *big.NewInt(7)},
			"cannot divide by zero (in 7/0)"},
		{"InternalError", InternalError{Err: errors.New("boom")}, "internal error: boom"},
		{"StoreError", StoreError{Wrapped: errors.New("store is down")}, "store error: store is down"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.msg, tc.err.Error())
		})
	}
}

// The two wrapping errors must stay unwrappable, so hosts can inspect the cause.
func TestExecutionErrorsUnwrap(t *testing.T) {
	cause := errors.New("cause")
	require.ErrorIs(t, InternalError{Err: cause}, cause)
	require.ErrorIs(t, StoreError{Wrapped: cause}, cause)
}
