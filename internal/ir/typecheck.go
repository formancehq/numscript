package ir

import "fmt"

// regType is the type of a virtual register. It mirrors the VM register banks;
// every register has exactly one type for its whole life. A monetary is not one
// of them: it is a (regStr asset, regInt amount) pair.
type regType int

const (
	regInt regType = iota
	regStr
	regPortion
	regBool
)

func (t regType) String() string {
	switch t {
	case regInt:
		return "int"
	case regStr:
		return "string"
	case regPortion:
		return "portion"
	case regBool:
		return "bool"
	default:
		return "?"
	}
}

// bytecodeTypechecker validates a IR stream one instruction at
// a time. It remembers the type each register was written with; a later read of
// that register with a different type, or a read before any write, is a bug in
// the code that produced the instructions. The state is updated as each
// instruction checks out.
type bytecodeTypechecker struct {
	types map[Reg]regType
}

func newBytecodeTypechecker() *bytecodeTypechecker {
	return &bytecodeTypechecker{types: map[Reg]regType{}}
}

// use asserts r was already written with type want.
func (tc *bytecodeTypechecker) use(r Reg, want regType) error {
	got, ok := tc.types[r]
	if !ok {
		return fmt.Errorf("register %s read as %s before being written", r, want)
	}
	if got != want {
		return fmt.Errorf("register %s read as %s but holds %s", r, want, got)
	}
	return nil
}

func (tc *bytecodeTypechecker) useOpt(r *Reg, want regType) error {
	if r == nil {
		return nil
	}
	return tc.use(*r, want)
}

// def records that r now holds type t, rejecting a write that changes its type.
func (tc *bytecodeTypechecker) def(r Reg, t regType) error {
	if got, ok := tc.types[r]; ok && got != t {
		return fmt.Errorf("register %s written as %s but already holds %s", r, t, got)
	}
	tc.types[r] = t
	return nil
}

// check typechecks a single instruction, updating the state on success.
func (tc *bytecodeTypechecker) check(instr Instr) error {
	switch i := instr.(type) {
	case LoadInt:
		return tc.def(i.Dest, regInt)
	case LoadStr:
		return tc.def(i.Dest, regStr)
	case ConstBool:
		return tc.def(i.Dest, regBool)
	case LoadVar:
		t, err := varRegType(i.Typ)
		if err != nil {
			return err
		}
		return tc.def(i.Dest, t)

	case UnaryOp:
		dest, arg, err := unOpRegTypes(i.Op)
		if err != nil {
			return err
		}
		return firstErr(tc.use(i.Arg, arg), tc.def(i.Dest, dest))
	case BinaryOp:
		dest, left, right, err := binOpRegTypes(i.Op)
		if err != nil {
			return err
		}
		return firstErr(tc.use(i.Left, left), tc.use(i.Right, right), tc.def(i.Dest, dest))

	case PullAccount:
		return firstErr(
			tc.use(i.Account, regStr),
			tc.useOpt(i.Cap, regInt),
			tc.useOpt(i.Overdraft, regInt),
			tc.useOpt(i.Color, regStr),
			tc.def(i.Dest, regInt),
		)
	case SendToAccount:
		return firstErr(tc.useOpt(i.Account, regStr), tc.useOpt(i.Cap, regInt))
	case Save:
		return firstErr(tc.use(i.Account, regStr), tc.use(i.Asset, regStr), tc.useOpt(i.Amount, regInt))

	case MakeAllotment:
		if err := tc.use(i.Amount, regInt); err != nil {
			return err
		}
		for _, p := range i.Portions {
			if err := tc.use(p, regPortion); err != nil {
				return err
			}
		}
		for _, d := range i.Dest {
			if err := tc.def(d, regInt); err != nil {
				return err
			}
		}
		return nil

	case CheckEnoughFunds:
		return firstErr(tc.use(i.Got, regInt), tc.use(i.Needed, regInt))
	case AssertLeftover:
		return tc.use(i.Portion, regPortion)
	case SetCurrentAsset:
		return tc.use(i.Asset, regStr)
	case AssertSameAsset:
		return firstErr(tc.use(i.Left, regStr), tc.use(i.Right, regStr))
	case AssertValidAccount:
		return tc.use(i.Account, regStr)
	case AssertValidColor:
		return tc.use(i.Color, regStr)
	case AssertNonNegativeBalance:
		return firstErr(tc.use(i.Balance, regInt), tc.use(i.Account, regStr))

	case SetTxMeta:
		return firstErr(tc.use(i.Key, regStr), tc.use(i.Value, regStr))
	case SetAccountMeta:
		return firstErr(tc.use(i.Account, regStr), tc.use(i.Key, regStr), tc.use(i.Value, regStr))
	case MetaVar:
		t, err := metaRegType(i.Typ)
		if err != nil {
			return err
		}
		return firstErr(tc.use(i.Account, regStr), tc.use(i.Key, regStr), tc.def(i.Dest, t))
	case MetaMonetary:
		return firstErr(
			tc.use(i.Account, regStr),
			tc.use(i.Key, regStr),
			tc.def(i.DestAsset, regStr),
			tc.def(i.DestAmount, regInt),
		)
	case FetchBalance:
		return firstErr(tc.use(i.Account, regStr), tc.use(i.Asset, regStr), tc.def(i.Dest, regInt))

	case JmpIfFalse:
		return tc.use(i.Cond, regBool)
	case JmpIfTrue:
		return tc.use(i.Cond, regBool)
	case Jmp:
		return nil
	case LabelMarker:
		return nil

	case Snapshot:
		return tc.def(i.Dest, regInt)
	case Restore:
		return tc.use(i.Mark, regInt)

	default:
		return fmt.Errorf("bytecode typechecker: unhandled instruction %T", instr)
	}
}

func Typecheck(instrs []Instr) error {
	tc := newBytecodeTypechecker()
	for pos, instr := range instrs {
		if err := tc.check(instr); err != nil {
			return fmt.Errorf("at instruction %d (%s): %w", pos, instr, err)
		}
	}
	return nil
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func varRegType(t VarType) (regType, error) {
	switch t.(type) {
	case VarInt:
		return regInt, nil
	case VarStr:
		return regStr, nil
	default:
		return 0, fmt.Errorf("bytecode typechecker: unknown var type %T", t)
	}
}

func metaRegType(t MetaType) (regType, error) {
	switch t.(type) {
	case MetaStr:
		return regStr, nil
	case MetaInt:
		return regInt, nil
	case MetaPortion:
		return regPortion, nil
	default:
		return 0, fmt.Errorf("bytecode typechecker: unknown meta type %T", t)
	}
}

func unOpRegTypes(op UnKind) (dest, arg regType, err error) {
	switch op.(type) {
	case OpIntCopy:
		return regInt, regInt, nil
	case OpPortionCopy:
		return regPortion, regPortion, nil
	case OpStrCopy:
		return regStr, regStr, nil
	case OpBoolCopy:
		return regBool, regBool, nil
	case OpNegInt:
		return regInt, regInt, nil
	case OpIntToString:
		return regStr, regInt, nil
	case OpIsZero:
		return regBool, regInt, nil
	case OpNot:
		return regBool, regBool, nil
	case OpPortionToString:
		return regStr, regPortion, nil
	default:
		return 0, 0, fmt.Errorf("bytecode typechecker: unknown unary op %T", op)
	}
}

func binOpRegTypes(op BinKind) (dest, left, right regType, err error) {
	switch op.(type) {
	case OpAddInt, OpSubInt:
		return regInt, regInt, regInt, nil
	case OpAddString:
		return regStr, regStr, regStr, nil
	case OpStrEq:
		return regBool, regStr, regStr, nil
	case OpLtInt, OpEqInt:
		return regBool, regInt, regInt, nil
	case OpLtPortion, OpEqPortion:
		return regBool, regPortion, regPortion, nil
	case OpSubPortion:
		return regPortion, regPortion, regPortion, nil
	case OpMakePortion:
		return regPortion, regInt, regInt, nil
	case OpMonetaryToString:
		return regStr, regStr, regInt, nil
	default:
		return 0, 0, 0, fmt.Errorf("bytecode typechecker: unknown binary op %T", op)
	}
}
