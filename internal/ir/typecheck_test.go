package ir

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytecodeTypecheck_Valid(t *testing.T) {
	// $0 = 1; $1 = 2; $2 = $0 + $1   (all int)
	instrs := []Instr{
		LoadInt{Dest: 0, Value: *big.NewInt(1)},
		LoadInt{Dest: 1, Value: *big.NewInt(2)},
		BinaryOp{Op: OpAddInt{}, Dest: 2, Left: 0, Right: 1},
	}
	require.NoError(t, Typecheck(instrs))
}

func TestBytecodeTypecheck_UseBeforeWrite(t *testing.T) {
	// reads $0 as int before it is ever written
	instrs := []Instr{
		UnaryOp{Op: OpNegInt{}, Dest: 1, Arg: 0},
	}
	require.ErrorContains(t, Typecheck(instrs), "read as int before being written")
}

func TestBytecodeTypecheck_WrongType(t *testing.T) {
	// $0 is a string, then used where an int is expected
	instrs := []Instr{
		LoadStr{Dest: 0, Value: "USD/2"},
		UnaryOp{Op: OpNegInt{}, Dest: 1, Arg: 0},
	}
	require.ErrorContains(t, Typecheck(instrs), "read as int but holds string")
}

func TestBytecodeTypecheck_RedefinedWithDifferentType(t *testing.T) {
	// $0 written as int, then overwritten as string
	instrs := []Instr{
		LoadInt{Dest: 0, Value: *big.NewInt(1)},
		LoadStr{Dest: 0, Value: "x"},
	}
	require.ErrorContains(t, Typecheck(instrs), "written as string but already holds int")
}

func TestBytecodeTypecheck_ErrorLocatesTheInstruction(t *testing.T) {
	instrs := []Instr{
		LoadStr{Dest: 0, Value: "src"},
		LoadStr{Dest: 1, Value: "dest"},
		CheckEnoughFunds{Got: 0, Needed: 1},
	}
	err := Typecheck(instrs)
	require.ErrorContains(t, err, "at instruction 2")
	require.ErrorContains(t, err, "check_enough_funds($r0, $r1)")
}

// Every register operand must be rejected when it names a register of the wrong
// bank. Each case is the prelude plus one instruction with exactly one bad operand.
func TestBytecodeTypecheck_OperandTypes(t *testing.T) {
	intReg, strReg, portionReg := Reg(0), Reg(1), Reg(2)
	prelude := []Instr{
		LoadInt{Dest: intReg, Value: *big.NewInt(1)},
		LoadStr{Dest: strReg, Value: "USD/2"},
		BinaryOp{Op: OpMakePortion{}, Dest: portionReg, Left: intReg, Right: intReg},
	}

	testCases := []struct {
		name  string
		instr Instr
	}{
		{"pull_account account", PullAccount{Dest: 9, Account: intReg}},
		{"pull_account cap", PullAccount{Dest: 9, Account: strReg, Cap: &strReg}},
		{"pull_account overdraft", PullAccount{Dest: 9, Account: strReg, Overdraft: &strReg}},
		{"pull_account color", PullAccount{Dest: 9, Account: strReg, Color: &intReg}},
		{"send_to_account account", SendToAccount{Account: &intReg}},
		{"send_to_account cap", SendToAccount{Account: &strReg, Cap: &strReg}},
		{"save account", Save{Account: intReg, Asset: strReg}},
		{"save asset", Save{Account: strReg, Asset: intReg}},
		{"save amount", Save{Account: strReg, Asset: strReg, Amount: &strReg}},
		{"mk_allot amount", MakeAllotment{Dest: []Reg{9}, Amount: strReg, Portions: []Reg{portionReg}}},
		{"mk_allot portion", MakeAllotment{Dest: []Reg{9}, Amount: intReg, Portions: []Reg{intReg}}},
		{"mk_allot dest", MakeAllotment{Dest: []Reg{strReg}, Amount: intReg, Portions: []Reg{portionReg}}},
		{"check_enough_funds got", CheckEnoughFunds{Got: strReg, Needed: intReg}},
		{"check_enough_funds needed", CheckEnoughFunds{Got: intReg, Needed: strReg}},
		{"assert_leftover", AssertLeftover{Portion: intReg}},
		{"set_current_asset", SetCurrentAsset{Asset: intReg}},
		{"assert_same_asset left", AssertSameAsset{Left: intReg, Right: strReg}},
		{"assert_same_asset right", AssertSameAsset{Left: strReg, Right: intReg}},
		{"assert_valid_account", AssertValidAccount{Account: intReg}},
		{"assert_valid_color", AssertValidColor{Color: intReg}},
		{"assert_non_negative_balance balance", AssertNonNegativeBalance{Balance: strReg, Account: strReg}},
		{"assert_non_negative_balance account", AssertNonNegativeBalance{Balance: intReg, Account: intReg}},
		{"set_tx_meta key", SetTxMeta{Key: intReg, Value: strReg}},
		{"set_tx_meta value", SetTxMeta{Key: strReg, Value: intReg}},
		{"set_account_meta account", SetAccountMeta{Account: intReg, Key: strReg, Value: strReg}},
		{"set_account_meta key", SetAccountMeta{Account: strReg, Key: intReg, Value: strReg}},
		{"set_account_meta value", SetAccountMeta{Account: strReg, Key: strReg, Value: intReg}},
		{"meta account", MetaVar{Dest: 9, Account: intReg, Key: strReg, Typ: MetaStr{}}},
		{"meta key", MetaVar{Dest: 9, Account: strReg, Key: intReg, Typ: MetaStr{}}},
		{"meta_monetary account", MetaMonetary{DestAsset: 9, DestAmount: 10, Account: intReg, Key: strReg}},
		{"meta_monetary key", MetaMonetary{DestAsset: 9, DestAmount: 10, Account: strReg, Key: intReg}},
		{"meta_monetary dest asset", MetaMonetary{DestAsset: intReg, DestAmount: 10, Account: strReg, Key: strReg}},
		{"meta_monetary dest amount", MetaMonetary{DestAsset: 9, DestAmount: strReg, Account: strReg, Key: strReg}},
		{"balance account", FetchBalance{Dest: 9, Account: intReg, Asset: strReg}},
		{"balance asset", FetchBalance{Dest: 9, Account: strReg, Asset: intReg}},
		{"balance dest", FetchBalance{Dest: strReg, Account: strReg, Asset: strReg}},
		// a quantity is not a condition: that's the guarantee the bool bank buys
		{"jmp_if_false cond", JmpIfFalse{Cond: intReg, Target: "end"}},
		{"jmp_if_true cond", JmpIfTrue{Cond: strReg, Target: "end"}},
		{"is_zero arg", UnaryOp{Op: OpIsZero{}, Dest: 9, Arg: strReg}},
		{"str_eq left", BinaryOp{Op: OpStrEq{}, Dest: 9, Left: intReg, Right: strReg}},
		// each comparison takes its own bank and yields a bool; not takes a bool
		{"lt_int left", BinaryOp{Op: OpLtInt{}, Dest: 9, Left: strReg, Right: intReg}},
		{"lt_int right", BinaryOp{Op: OpLtInt{}, Dest: 9, Left: intReg, Right: portionReg}},
		{"eq_int left", BinaryOp{Op: OpEqInt{}, Dest: 9, Left: strReg, Right: intReg}},
		{"lt_portion left", BinaryOp{Op: OpLtPortion{}, Dest: 9, Left: intReg, Right: portionReg}},
		{"eq_portion right", BinaryOp{Op: OpEqPortion{}, Dest: 9, Left: portionReg, Right: intReg}},
		{"not arg", UnaryOp{Op: OpNot{}, Dest: 9, Arg: intReg}},
		{"restore mark", Restore{Mark: strReg}},
		{"unary arg", UnaryOp{Op: OpPortionToString{}, Dest: 9, Arg: intReg}},
		{"binary left", BinaryOp{Op: OpAddString{}, Dest: 9, Left: intReg, Right: strReg}},
		{"binary right", BinaryOp{Op: OpAddString{}, Dest: 9, Left: strReg, Right: intReg}},
		{"monetary_to_string asset", BinaryOp{Op: OpMonetaryToString{}, Dest: 9, Left: intReg, Right: intReg}},
		{"monetary_to_string amount", BinaryOp{Op: OpMonetaryToString{}, Dest: 9, Left: strReg, Right: strReg}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Error(t, Typecheck(append(append([]Instr{}, prelude...), tc.instr)))
		})
	}
}

// The dest bank of these instructions comes from a type tag, not from an operand.
// Reading the dest back as an int is what tells the two apart.
func TestBytecodeTypecheck_TaggedDests(t *testing.T) {
	str := Reg(0)
	prelude := []Instr{LoadStr{Dest: str, Value: "k"}}

	testCases := []struct {
		name      string
		instr     Instr
		destIsInt bool
	}{
		{"load_var<int>", LoadVar{Dest: 9, Typ: VarInt{}}, true},
		{"load_var<str>", LoadVar{Dest: 9, Typ: VarStr{}}, false},
		{"meta<str>", MetaVar{Dest: 9, Account: str, Key: str, Typ: MetaStr{}}, false},
		{"meta<int>", MetaVar{Dest: 9, Account: str, Key: str, Typ: MetaInt{}}, true},
		{"meta<portion>", MetaVar{Dest: 9, Account: str, Key: str, Typ: MetaPortion{}}, false},
		{"snapshot", Snapshot{Dest: 9}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// restore only accepts an int register
			instrs := append(append([]Instr{}, prelude...), tc.instr, Restore{Mark: 9})
			if tc.destIsInt {
				require.NoError(t, Typecheck(instrs))
			} else {
				require.Error(t, Typecheck(instrs))
			}
		})
	}
}

// A bool register is its own bank: nothing that takes an int accepts one, and it
// can't be rewritten as another type.
func TestBytecodeTypecheck_Bool(t *testing.T) {
	t.Run("const_true and const_false define a bool", func(t *testing.T) {
		require.NoError(t, Typecheck([]Instr{
			ConstBool{Dest: 0, Value: true},
			ConstBool{Dest: 1, Value: false},
		}))
	})

	t.Run("a bool is not an int", func(t *testing.T) {
		err := Typecheck([]Instr{
			ConstBool{Dest: 0, Value: true},
			Restore{Mark: 0},
		})
		require.ErrorContains(t, err, "read as int but holds bool")
	})

	t.Run("an int is not a bool", func(t *testing.T) {
		err := Typecheck([]Instr{
			LoadInt{Dest: 0, Value: *big.NewInt(1)},
			ConstBool{Dest: 0, Value: true},
		})
		require.ErrorContains(t, err, "written as bool but already holds int")
	})

	t.Run("rewriting a bool with the same type is allowed", func(t *testing.T) {
		require.NoError(t, Typecheck([]Instr{
			ConstBool{Dest: 0, Value: true},
			ConstBool{Dest: 0, Value: false},
		}))
	})

	// every comparison feeds a jump directly, and `not` composes with all of them
	// — which is what makes the derived operators expressible without opcodes
	t.Run("comparisons yield branchable bools", func(t *testing.T) {
		// operand register per bank, so each op is fed its own type
		prelude := []Instr{
			LoadInt{Dest: 0, Value: *big.NewInt(1)},
			BinaryOp{Op: OpMakePortion{}, Dest: 1, Left: 0, Right: 0},
			LoadStr{Dest: 2, Value: "x"},
		}
		ops := map[BinKind]Reg{
			OpLtInt{}:     0,
			OpEqInt{}:     0,
			OpLtPortion{}: 1,
			OpEqPortion{}: 1,
			OpStrEq{}:     2,
		}
		for op, operand := range ops {
			t.Run(op.String(), func(t *testing.T) {
				require.NoError(t, Typecheck(append(append([]Instr{}, prelude...),
					BinaryOp{Op: op, Dest: 9, Left: operand, Right: operand},
					UnaryOp{Op: OpNot{}, Dest: 10, Arg: 9},
					JmpIfTrue{Cond: 9, Target: "end"},
					JmpIfFalse{Cond: 10, Target: "end"},
					LabelMarker{Label: "end"},
				)))
			})
		}
	})

	// is_zero is in the comparison group too, and is the one unary member
	t.Run("is_zero yields a branchable bool", func(t *testing.T) {
		require.NoError(t, Typecheck([]Instr{
			LoadInt{Dest: 0, Value: *big.NewInt(1)},
			UnaryOp{Op: OpIsZero{}, Dest: 1, Arg: 0},
			JmpIfTrue{Cond: 1, Target: "end"},
			LabelMarker{Label: "end"},
		}))
	})

	t.Run("a comparison result is not an int", func(t *testing.T) {
		err := Typecheck([]Instr{
			LoadInt{Dest: 0, Value: *big.NewInt(1)},
			BinaryOp{Op: OpEqInt{}, Dest: 1, Left: 0, Right: 0},
			Restore{Mark: 1},
		})
		require.ErrorContains(t, err, "read as int but holds bool")
	})
}

func TestBytecodeTypecheck_LabelMarker(t *testing.T) {
	require.NoError(t, Typecheck([]Instr{LabelMarker{Label: "end"}}))
	require.Empty(t, LabelMarker{Label: "end"}.dests())
	require.Empty(t, LabelMarker{Label: "end"}.sources())
}

// --- An unknown instruction or type tag is a bug in whatever built the stream:
// reported as an error, never panicked.

type unknownInstr struct{}

func (unknownInstr) dests() []Reg              { return nil }
func (unknownInstr) sources() []Reg            { return nil }
func (unknownInstr) assemble(*assembler) error { return nil }
func (unknownInstr) String() string            { return "unknown_instr" }

type unknownUnOp struct{}

func (unknownUnOp) String() string  { return "unknown_un_op" }
func (unknownUnOp) sig() unaryOpSig { return unaryOpSig{} }

type unknownBinOp struct{}

func (unknownBinOp) String() string   { return "unknown_bin_op" }
func (unknownBinOp) sig() binaryOpSig { return binaryOpSig{} }

type unknownVarType struct{}

func (unknownVarType) String() string                             { return "unknown_var_type" }
func (unknownVarType) assembleLoad(*assembler, Reg, uint16) error { return nil }

type unknownMetaType struct{}

func (unknownMetaType) String() string                               { return "unknown_meta_type" }
func (unknownMetaType) assembleMeta(*assembler, Reg, Reg, Reg) error { return nil }

func TestBytecodeTypecheck_UnknownTags(t *testing.T) {
	str := Reg(0)
	prelude := []Instr{LoadStr{Dest: str, Value: "k"}}

	testCases := []struct {
		name  string
		instr Instr
		msg   string
	}{
		{"instruction", unknownInstr{}, "unhandled instruction"},
		{"unary op", UnaryOp{Op: unknownUnOp{}, Dest: 9, Arg: str}, "unknown unary op"},
		{"binary op", BinaryOp{Op: unknownBinOp{}, Dest: 9, Left: str, Right: str}, "unknown binary op"},
		{"var type", LoadVar{Dest: 9, Typ: unknownVarType{}}, "unknown var type"},
		{"meta type", MetaVar{Dest: 9, Account: str, Key: str, Typ: unknownMetaType{}}, "unknown meta type"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorContains(t, Typecheck(append(append([]Instr{}, prelude...), tc.instr)), tc.msg)
		})
	}
}

func TestRegTypeString(t *testing.T) {
	require.Equal(t, "int", regInt.String())
	require.Equal(t, "string", regStr.String())
	require.Equal(t, "portion", regPortion.String())
	require.Equal(t, "bool", regBool.String())
	require.Equal(t, "?", regType(99).String())
	require.Equal(t, "?", regType(42).String())
}
