package compiler

import "fmt"

// regType is the type of a virtual register. It mirrors the four VM register
// banks; every register has exactly one type for its whole life.
type regType int

const (
	regInt regType = iota
	regStr
	regPortion
	regMonetary
)

func (t regType) String() string {
	switch t {
	case regInt:
		return "int"
	case regStr:
		return "string"
	case regPortion:
		return "portion"
	case regMonetary:
		return "monetary"
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
	types map[reg]regType
}

func newBytecodeTypechecker() *bytecodeTypechecker {
	return &bytecodeTypechecker{types: map[reg]regType{}}
}

// use asserts r was already written with type want.
func (tc *bytecodeTypechecker) use(r reg, want regType) error {
	got, ok := tc.types[r]
	if !ok {
		return fmt.Errorf("register %s read as %s before being written", r, want)
	}
	if got != want {
		return fmt.Errorf("register %s read as %s but holds %s", r, want, got)
	}
	return nil
}

func (tc *bytecodeTypechecker) useOpt(r *reg, want regType) error {
	if r == nil {
		return nil
	}
	return tc.use(*r, want)
}

// def records that r now holds type t, rejecting a write that changes its type.
func (tc *bytecodeTypechecker) def(r reg, t regType) error {
	if got, ok := tc.types[r]; ok && got != t {
		return fmt.Errorf("register %s written as %s but already holds %s", r, t, got)
	}
	tc.types[r] = t
	return nil
}

// check typechecks a single instruction, updating the state on success.
func (tc *bytecodeTypechecker) check(instr irInstr) error {
	switch i := instr.(type) {
	case loadInt:
		return tc.def(i.dest, regInt)
	case loadStr:
		return tc.def(i.dest, regStr)
	case loadVar:
		t, err := varRegType(i.typ)
		if err != nil {
			return err
		}
		return tc.def(i.dest, t)

	case unaryOp:
		dest, arg, err := unOpRegTypes(i.op)
		if err != nil {
			return err
		}
		return firstErr(tc.use(i.arg, arg), tc.def(i.dest, dest))
	case binaryOp:
		dest, left, right, err := binOpRegTypes(i.op)
		if err != nil {
			return err
		}
		return firstErr(tc.use(i.left, left), tc.use(i.right, right), tc.def(i.dest, dest))

	case pullAccount:
		return firstErr(
			tc.use(i.account, regStr),
			tc.useOpt(i.cap, regInt),
			tc.useOpt(i.overdraft, regInt),
			tc.useOpt(i.color, regStr),
			tc.def(i.dest, regInt),
		)
	case sendToAccount:
		return firstErr(tc.useOpt(i.account, regStr), tc.useOpt(i.cap, regInt))
	case save:
		return firstErr(tc.use(i.account, regStr), tc.use(i.asset, regStr), tc.useOpt(i.amount, regInt))

	case makeAllotment:
		if err := tc.use(i.amount, regInt); err != nil {
			return err
		}
		for _, p := range i.portions {
			if err := tc.use(p, regPortion); err != nil {
				return err
			}
		}
		for _, d := range i.dest {
			if err := tc.def(d, regInt); err != nil {
				return err
			}
		}
		return nil

	case checkEnoughFunds:
		return firstErr(tc.use(i.got, regInt), tc.use(i.needed, regInt))
	case assertLeftover:
		return tc.use(i.portion, regPortion)
	case setCurrentAsset:
		return tc.use(i.asset, regStr)
	case assertSameAsset:
		return firstErr(tc.use(i.left, regStr), tc.use(i.right, regStr))
	case assertValidAccount:
		return tc.use(i.account, regStr)
	case assertNonNegativeBalance:
		return firstErr(tc.use(i.balance, regMonetary), tc.use(i.account, regStr))

	case setTxMeta:
		return firstErr(tc.use(i.key, regStr), tc.use(i.value, regStr))
	case setAccountMeta:
		return firstErr(tc.use(i.account, regStr), tc.use(i.key, regStr), tc.use(i.value, regStr))
	case metaVar:
		t, err := metaRegType(i.typ)
		if err != nil {
			return err
		}
		return firstErr(tc.use(i.account, regStr), tc.use(i.key, regStr), tc.def(i.dest, t))
	case fetchBalance:
		return firstErr(tc.use(i.account, regStr), tc.use(i.asset, regStr), tc.def(i.dest, regMonetary))

	case jmpIfZero:
		return tc.use(i.cond, regInt)
	case labelMarker:
		return nil

	case snapshot:
		return tc.def(i.dest, regInt)
	case restore:
		return tc.use(i.mark, regInt)

	default:
		return fmt.Errorf("bytecode typechecker: unhandled instruction %T", instr)
	}
}

func typecheckInstructions(instrs []irInstr) error {
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

func varRegType(t varType) (regType, error) {
	switch t.(type) {
	case varInt:
		return regInt, nil
	case varStr:
		return regStr, nil
	default:
		return 0, fmt.Errorf("bytecode typechecker: unknown var type %T", t)
	}
}

func metaRegType(t metaType) (regType, error) {
	switch t.(type) {
	case metaStr:
		return regStr, nil
	case metaInt:
		return regInt, nil
	case metaPortion:
		return regPortion, nil
	case metaMonetary:
		return regMonetary, nil
	default:
		return 0, fmt.Errorf("bytecode typechecker: unknown meta type %T", t)
	}
}

func unOpRegTypes(op unKind) (dest, arg regType, err error) {
	switch op.(type) {
	case opIntCopy:
		return regInt, regInt, nil
	case opPortionCopy:
		return regPortion, regPortion, nil
	case opGetAsset:
		return regStr, regMonetary, nil
	case opGetAmount:
		return regInt, regMonetary, nil
	case opNegInt:
		return regInt, regInt, nil
	case opIntToString:
		return regStr, regInt, nil
	case opPortionToString:
		return regStr, regPortion, nil
	case opMonetaryToString:
		return regStr, regMonetary, nil
	default:
		return 0, 0, fmt.Errorf("bytecode typechecker: unknown unary op %T", op)
	}
}

func binOpRegTypes(op binKind) (dest, left, right regType, err error) {
	switch op.(type) {
	case opMinInt, opAddInt, opSubInt:
		return regInt, regInt, regInt, nil
	case opAddString:
		return regStr, regStr, regStr, nil
	case opSubPortion:
		return regPortion, regPortion, regPortion, nil
	case opMakePortion:
		return regPortion, regInt, regInt, nil
	case opMakeMonetary:
		return regMonetary, regStr, regInt, nil
	default:
		return 0, 0, 0, fmt.Errorf("bytecode typechecker: unknown binary op %T", op)
	}
}
