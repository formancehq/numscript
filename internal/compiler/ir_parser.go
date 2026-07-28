package compiler

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/formancehq/numscript/internal/irparser"
	"github.com/formancehq/numscript/internal/parser"
)

// TransformError is an error during AST → irInstr transformation.
type TransformError struct {
	Range parser.Range
	Msg   string
}

func (e TransformError) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Range.Start.Line+1, e.Range.Start.Character+1, e.Msg)
}

// transformer carries the state shared by the whole transformation.
type transformer struct {
	// labelPos maps each label defined in the program to its position, so a jump
	// can be checked to both resolve and go forward.
	labelPos map[string]int
	// stmtPos is the position of the statement being transformed.
	stmtPos int
	// regByName binds each register name to the logical register it got on its
	// first appearance.
	regByName map[string]reg
	nextReg   reg
}

// freshReg allocates a register bound to no name.
func (t *transformer) freshReg() reg {
	r := t.nextReg
	t.nextReg++
	return r
}

// resolveReg returns the logical register a name refers to, allocating one on
// the name's first appearance.
func (t *transformer) resolveReg(rr irparser.RegRef) reg {
	if r, ok := t.regByName[rr.Name]; ok {
		return r
	}
	r := t.freshReg()
	t.regByName[rr.Name] = r
	return r
}

// Transform converts a parsed IR AST into a slice of irInstr.
// It returns all instructions and any errors encountered.
func Transform(prog irparser.Program) ([]irInstr, []TransformError) {
	var instrs []irInstr
	var errs []TransformError

	t := &transformer{labelPos: map[string]int{}, regByName: map[string]reg{}}
	// First pass: collect the labels and where they sit.
	for pos, stmt := range prog.Stmts {
		if ls, ok := stmt.(*irparser.LabelStmt); ok {
			if _, seen := t.labelPos[ls.Name]; seen {
				errs = append(errs, TransformError{Range: ls.Range, Msg: fmt.Sprintf("duplicate label #%s", ls.Name)})
			}
			t.labelPos[ls.Name] = pos
		}
	}

	for pos, stmt := range prog.Stmts {
		switch s := stmt.(type) {
		case *irparser.LabelStmt:
			instrs = append(instrs, labelMarker{label: label(s.Name)})
		case *irparser.InstrStmt:
			t.stmtPos = pos
			instr, err := t.transformInstr(s)
			if err != nil {
				errs = append(errs, *err)
				continue
			}
			instrs = append(instrs, instr)
		}
	}

	if len(errs) > 0 {
		return instrs, errs
	}
	return instrs, nil
}

func (t *transformer) transformInstr(s *irparser.InstrStmt) (irInstr, *TransformError) {
	switch {
	case s.Const != nil:
		return t.transformConst(s)
	case s.Call != nil:
		return t.transformCall(s)
	case s.Infix != nil:
		return t.transformInfix(s, s.Infix, "infix")
	case s.CompoundAssign != nil:
		return t.transformInfix(s, s.CompoundAssign, "compound assign")
	default:
		return nil, &TransformError{Range: s.Range, Msg: "empty instruction"}
	}
}

// ---- const assignment ----

func (t *transformer) transformConst(s *irparser.InstrStmt) (irInstr, *TransformError) {
	if s.Dest == nil || s.Dest.Kind != irparser.DestReg {
		return nil, &TransformError{Range: s.Range, Msg: "const assignment requires a single register dest"}
	}
	dest := t.resolveReg(s.Dest.Regs[0])

	switch s.Const.Kind {
	case irparser.ConstString:
		return loadStr{dest: dest, value: *s.Const.StrVal}, nil
	case irparser.ConstInt:
		n, ok := new(big.Int).SetString(*s.Const.IntVal, 10)
		if !ok {
			return nil, &TransformError{Range: s.Const.Range, Msg: fmt.Sprintf("invalid integer: %q", *s.Const.IntVal)}
		}
		return loadInt{dest: dest, value: *n}, nil
	default:
		return nil, &TransformError{Range: s.Range, Msg: "unknown const kind"}
	}
}

// ---- infix / compound assign ----

// transformInfix handles both `$d = $l + $r` and `$d += $r`: the parser gives
// the compound form the same shape, with Left repeated as the dest.
func (t *transformer) transformInfix(s *irparser.InstrStmt, infix *irparser.Infix, what string) (irInstr, *TransformError) {
	if s.Dest == nil || s.Dest.Kind != irparser.DestReg {
		return nil, &TransformError{Range: s.Range, Msg: what + " requires a single register dest"}
	}
	dest := t.resolveReg(s.Dest.Regs[0])
	left := t.resolveReg(infix.Left)
	right := t.resolveReg(infix.Right)

	var op binKind
	switch infix.Op {
	case "+":
		op = opAddInt{}
	case "-":
		op = opSubInt{}
	default:
		return nil, &TransformError{Range: infix.Range, Msg: fmt.Sprintf("unknown infix operator: %q", infix.Op)}
	}
	return binaryOp{op: op, dest: dest, left: left, right: right}, nil
}

// ---- call instructions ----

// argParser reads the args of one instruction call. Every accessor reports what
// is wrong with the arg it wanted and returns a zero value, so a caller can read
// its args straight into an irInstr literal and let transformCall check ap.errs
// once at the end. Positional accessors consume in call order: composite literal
// operands are evaluated left to right, so `f{a: ap.reg(), b: ap.reg()}` reads
// the args in that order.
type argParser struct {
	t         *transformer
	args      []irparser.Arg
	pos       int
	seenLabel map[string]bool
	errs      *[]TransformError
	// callRange is where to point an error about an arg that isn't there.
	callRange parser.Range
}

func (t *transformer) newArgParser(call *irparser.InstrCall, errs *[]TransformError) *argParser {
	return &argParser{
		t:         t,
		args:      call.Args,
		seenLabel: map[string]bool{},
		errs:      errs,
		callRange: call.Range,
	}
}

// next consumes the next positional arg, checking it holds the expected kind. It
// reports missing and mistyped args itself and returns nil in both cases, so a
// caller only reads what it needs and the error shows up in ap.errs.
func (ap *argParser) next(want irparser.ValueKind) *irparser.Value {
	if ap.pos >= len(ap.args) {
		ap.addErr(ap.callRange, "missing %s argument", valueKindStr(want))
		return nil
	}
	a := ap.args[ap.pos]
	ap.pos++
	if a.Value.Kind != want {
		ap.addErr(a.Range, "expected %s, got %s", valueKindStr(want), valueKindStr(a.Value.Kind))
		return nil
	}
	return &a.Value
}

// reg consumes the next positional arg as a register.
func (ap *argParser) reg() reg {
	v := ap.next(irparser.ValReg)
	if v == nil {
		return 0
	}
	return ap.t.resolveReg(*v.Reg)
}

// optLabeledReg consumes an optional labeled arg with the given label as a register.
func (ap *argParser) optLabeledReg(label string) *reg {
	a, ok := ap.labeledArg(label)
	if !ok {
		return nil
	}
	if a.Value.Kind != irparser.ValReg {
		ap.addErr(a.Range, "labeled arg %q: expected register, got %s", label, valueKindStr(a.Value.Kind))
		return nil
	}
	r := ap.t.resolveReg(*a.Value.Reg)
	return &r
}

// reqLabeledReg is optLabeledReg for a label the instruction can't do without.
func (ap *argParser) reqLabeledReg(label string) reg {
	r := ap.optLabeledReg(label)
	if r == nil {
		ap.addErr(ap.callRange, "missing labeled argument %q", label)
		return 0
	}
	return *r
}

// labeledArg finds a labeled arg by name. Returns nil if not found.
func (ap *argParser) labeledArg(label string) (*irparser.Arg, bool) {
	if ap.seenLabel[label] {
		return nil, false
	}
	for i := ap.pos; i < len(ap.args); i++ {
		if ap.args[i].Label == label {
			ap.seenLabel[label] = true
			return &ap.args[i], true
		}
	}
	// also check already-consumed positional area
	for i := 0; i < ap.pos; i++ {
		if ap.args[i].Label == label {
			ap.seenLabel[label] = true
			return &ap.args[i], true
		}
	}
	return nil, false
}

// labelRef consumes the next positional arg as a label reference. It returns ""
// when the arg is missing or isn't one.
func (ap *argParser) labelRef() label {
	v := ap.next(irparser.ValLabel)
	if v == nil {
		return ""
	}
	return label(*v.Label)
}

// intLit consumes the next positional arg as an integer literal.
func (ap *argParser) intLit() uint16 {
	v := ap.next(irparser.ValInt)
	if v == nil {
		return 0
	}
	n, err := strconv.ParseUint(*v.Int, 10, 16)
	if err != nil {
		ap.addErr(v.Range, "integer literal out of range (0-65535): %s", *v.Int)
		return 0
	}
	return uint16(n)
}

// regList consumes the next positional arg as a register list.
func (ap *argParser) regList() []reg {
	v := ap.next(irparser.ValRegList)
	if v == nil {
		return nil
	}
	regs := make([]reg, len(*v.Regs))
	for i, r := range *v.Regs {
		regs[i] = ap.t.resolveReg(r)
	}
	return regs
}

func (ap *argParser) addErr(rng parser.Range, format string, args ...any) {
	*ap.errs = append(*ap.errs, TransformError{Range: rng, Msg: fmt.Sprintf(format, args...)})
}

func valueKindStr(k irparser.ValueKind) string {
	switch k {
	case irparser.ValReg:
		return "register"
	case irparser.ValLabel:
		return "label"
	case irparser.ValInt:
		return "integer literal"
	case irparser.ValRegList:
		return "register list"
	default:
		return "unknown"
	}
}

func (t *transformer) transformCall(s *irparser.InstrStmt) (irInstr, *TransformError) {
	var errs []TransformError
	ap := t.newArgParser(s.Call, &errs)

	// A labeled arg is looked up by name, so a repeated one would silently lose
	// every occurrence but the first.
	seen := map[string]bool{}
	for _, a := range s.Call.Args {
		if a.Label == "" {
			continue
		}
		if seen[a.Label] {
			ap.addErr(a.Range, "duplicate labeled argument %q", a.Label)
		}
		seen[a.Label] = true
	}

	// Resolve dest
	var dest reg
	var dests []reg
	if s.Dest != nil {
		switch s.Dest.Kind {
		case irparser.DestReg:
			dest = t.resolveReg(s.Dest.Regs[0])
		case irparser.DestDiscard:
			// `_` only exists in the text: desugar it to a fresh register, which
			// no other statement can name and nothing reads back.
			dest = t.freshReg()
		case irparser.DestList:
			dests = make([]reg, len(s.Dest.Regs))
			for i, r := range s.Dest.Regs {
				dests[i] = t.resolveReg(r)
			}
		}
	}

	name, typeParam := s.Call.Name, s.Call.TypeParam

	// load_var and meta are the only instructions parameterized by a type.
	if typeParam != "" && name != "load_var" && name != "meta" {
		return nil, &TransformError{Range: s.Call.Range, Msg: fmt.Sprintf("%s doesn't take a type parameter", name)}
	}

	var instr irInstr

	switch name {
	case "load_var":
		var typ varType
		switch typeParam {
		case "int":
			typ = varInt{}
		case "str":
			typ = varStr{}
		default:
			return nil, &TransformError{Range: s.Call.Range, Msg: fmt.Sprintf("load_var: expected type parameter int or str, got %q", typeParam)}
		}
		instr = loadVar{dest: dest, typ: typ, index: ap.intLit()}

	case "meta":
		var typ metaType
		switch typeParam {
		case "str":
			typ = metaStr{}
		case "int":
			typ = metaInt{}
		case "portion":
			typ = metaPortion{}
		case "monetary":
			typ = metaMonetary{}
		default:
			return nil, &TransformError{Range: s.Call.Range, Msg: fmt.Sprintf("meta: expected type parameter str, int, portion or monetary, got %q", typeParam)}
		}
		instr = metaVar{dest: dest, typ: typ, account: ap.reg(), key: ap.reg()}

	case "balance":
		instr = fetchBalance{dest: dest, account: ap.reg(), asset: ap.reg()}

	case "mk_monetary":
		instr = ap.binaryOp(dest, opMakeMonetary{})
	case "mk_portion":
		instr = ap.binaryOp(dest, opMakePortion{})
	case "add_int":
		instr = ap.binaryOp(dest, opAddInt{})
	case "sub_int":
		instr = ap.binaryOp(dest, opSubInt{})
	case "add_string":
		instr = ap.binaryOp(dest, opAddString{})
	case "sub_portion":
		instr = ap.binaryOp(dest, opSubPortion{})
	case "min_int":
		instr = ap.binaryOp(dest, opMinInt{})

	case "int_copy":
		instr = unaryOp{dest: dest, op: opIntCopy{}, arg: ap.reg()}
	case "portion_copy":
		instr = unaryOp{dest: dest, op: opPortionCopy{}, arg: ap.reg()}
	case "get_asset":
		instr = unaryOp{dest: dest, op: opGetAsset{}, arg: ap.reg()}
	case "get_amount":
		instr = unaryOp{dest: dest, op: opGetAmount{}, arg: ap.reg()}
	case "neg_int":
		instr = unaryOp{dest: dest, op: opNegInt{}, arg: ap.reg()}
	case "int_to_string":
		instr = unaryOp{dest: dest, op: opIntToString{}, arg: ap.reg()}
	case "portion_to_string":
		instr = unaryOp{dest: dest, op: opPortionToString{}, arg: ap.reg()}
	case "monetary_to_string":
		instr = unaryOp{dest: dest, op: opMonetaryToString{}, arg: ap.reg()}

	case "pull_account":
		instr = pullAccount{
			dest:      dest,
			account:   ap.reqLabeledReg("account"),
			cap:       ap.optLabeledReg("cap"),
			overdraft: ap.optLabeledReg("overdraft"),
			color:     ap.optLabeledReg("color"),
		}
	case "send_to_account":
		instr = sendToAccount{account: ap.optLabeledReg("account"), cap: ap.optLabeledReg("cap")}
	case "save":
		instr = save{
			account: ap.reqLabeledReg("account"),
			asset:   ap.reqLabeledReg("asset"),
			amount:  ap.optLabeledReg("amount"),
		}

	case "mk_allot":
		if len(dests) == 0 {
			ap.addErr(s.Range, "mk_allot requires a dest list")
			break
		}
		instr = makeAllotment{dest: dests, amount: ap.reg(), portions: ap.regList()}

	case "check_enough_funds":
		instr = checkEnoughFunds{got: ap.reg(), needed: ap.reg()}
	case "assert_leftover":
		instr = assertLeftover{portion: ap.reg(), exact: false}
	case "assert_leftover_exact":
		instr = assertLeftover{portion: ap.reg(), exact: true}
	case "set_current_asset":
		instr = setCurrentAsset{asset: ap.reg()}
	case "assert_same_asset":
		instr = assertSameAsset{left: ap.reg(), right: ap.reg()}
	case "assert_valid_account":
		instr = assertValidAccount{account: ap.reg()}
	case "assert_non_negative_balance":
		instr = assertNonNegativeBalance{balance: ap.reg(), account: ap.reg()}

	case "snapshot":
		instr = snapshot{dest: dest}
	case "restore":
		instr = restore{mark: ap.reg()}

	case "set_tx_meta":
		instr = setTxMeta{key: ap.reg(), value: ap.reg()}
	case "set_account_meta":
		instr = setAccountMeta{account: ap.reg(), key: ap.reg(), value: ap.reg()}

	case "jmp_if_zero":
		cond, target := ap.reg(), ap.labelRef()
		labelPos, defined := t.labelPos[string(target)]
		switch {
		case target == "":
			// labelRef already said what was wrong
		case !defined:
			ap.addErr(s.Range, "jmp_if_zero: label %s is not defined in the program", target)
		case labelPos < t.stmtPos:
			// The VM only allows jumping forward: that's what makes every program
			// terminate, so a backward jump is rejected here rather than assembled.
			ap.addErr(s.Range, "jmp_if_zero: label %s is behind the jump (jumps must go forward)", target)
		default:
			instr = jmpIfZero{cond: cond, target: target}
		}

	default:
		return nil, &TransformError{Range: s.Call.Range, Msg: fmt.Sprintf("unknown instruction: %s", name)}
	}

	// Check for unconsumed args (skip labeled args that were already seen)
	for ap.pos < len(ap.args) {
		a := ap.args[ap.pos]
		if a.Label != "" && ap.seenLabel[a.Label] {
			ap.pos++
			continue
		}
		ap.addErr(a.Range, "unexpected extra argument")
		ap.pos++
	}
	// Also check labeled args that weren't consumed
	for i := range ap.args {
		if ap.args[i].Label != "" && !ap.seenLabel[ap.args[i].Label] {
			ap.addErr(ap.args[i].Range, "unknown labeled argument %q", ap.args[i].Label)
		}
	}

	if len(errs) > 0 {
		// Return first error along with the instruction
		return instr, &errs[0]
	}
	return instr, nil
}

// binaryOp reads the two register args every binary instruction takes.
func (ap *argParser) binaryOp(dest reg, op binKind) binaryOp {
	return binaryOp{dest: dest, op: op, left: ap.reg(), right: ap.reg()}
}
