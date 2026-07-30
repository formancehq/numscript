package ir

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/formancehq/numscript/internal/ir/internal/syntax"
	"github.com/formancehq/numscript/internal/parser"
)

// Error is something wrong with an IR text: either the grammar rejected it, or
// it doesn't describe a well-formed instruction stream.
type Error struct {
	Range parser.Range
	Msg   string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Range.Start.Line+1, e.Range.Start.Character+1, e.Msg)
}

// Parse reads an IR text into the instruction stream it describes.
//
// It checks the grammar and everything the grammar can't express: that
// instructions exist and take the arguments they were given, that labels resolve
// and are unique, and that jumps go forward. It does not typecheck the registers
// — that's Typecheck.
func Parse(src string) ([]Instr, []Error) {
	parsed := syntax.Parse(src)
	if len(parsed.Errors) > 0 {
		errs := make([]Error, len(parsed.Errors))
		for i, e := range parsed.Errors {
			errs[i] = Error{Range: e.Range, Msg: e.Msg}
		}
		return nil, errs
	}
	return transform(parsed.Value)
}

// transformer carries the state shared by the whole transformation.
type transformer struct {
	// labelPos maps each label defined in the program to its position, so a jump
	// can be checked to both resolve and go forward.
	labelPos map[string]int
	// stmtPos is the position of the statement being transformed.
	stmtPos int
	// regByName binds each register name to the logical register it got on its
	// first appearance, and nameByReg maps it back for error messages.
	regByName map[string]Reg
	nameByReg map[Reg]string
	nextReg   Reg
	// written records the registers an instruction has assigned to so far, so a
	// read of one that was never written can be reported.
	written map[Reg]bool
}

// regName spells a register the way the text did.
func (t *transformer) regName(r Reg) string {
	if name, ok := t.nameByReg[r]; ok {
		return name
	}
	return r.String()
}

// freshReg allocates a register bound to no name.
func (t *transformer) freshReg() Reg {
	r := t.nextReg
	t.nextReg++
	return r
}

// resolveReg returns the logical register a name refers to, allocating one on
// the name's first appearance.
func (t *transformer) resolveReg(rr syntax.RegRef) Reg {
	if r, ok := t.regByName[rr.Name]; ok {
		return r
	}
	r := t.freshReg()
	t.regByName[rr.Name] = r
	t.nameByReg[r] = rr.Name
	return r
}

// transform converts a parsed IR AST into a slice of Instr.
// It returns all instructions and any errors encountered.
func transform(prog syntax.Program) ([]Instr, []Error) {
	var instrs []Instr
	var errs []Error

	t := &transformer{
		labelPos:  map[string]int{},
		regByName: map[string]Reg{},
		nameByReg: map[Reg]string{},
		written:   map[Reg]bool{},
	}
	// First pass: collect the labels and where they sit.
	for pos, stmt := range prog.Stmts {
		if ls, ok := stmt.(*syntax.LabelStmt); ok {
			if _, seen := t.labelPos[ls.Name]; seen {
				errs = append(errs, Error{Range: ls.Range, Msg: fmt.Sprintf("duplicate label #%s", ls.Name)})
			}
			t.labelPos[ls.Name] = pos
		}
	}

	for pos, stmt := range prog.Stmts {
		switch s := stmt.(type) {
		case *syntax.LabelStmt:
			instrs = append(instrs, LabelMarker{Label: Label(s.Name)})
		case *syntax.InstrStmt:
			t.stmtPos = pos
			instr, err := t.transformInstr(s)
			if err != nil {
				errs = append(errs, *err)
				continue
			}

			// Jumps only go forward, so text order is execution order: a read with
			// no earlier write can't be reached by any path.
			for _, r := range instr.sources() {
				if !t.written[r] {
					errs = append(errs, Error{
						Range: s.Range,
						Msg:   fmt.Sprintf("register %s is read but never written", t.regName(r)),
					})
				}
			}
			for _, r := range instr.dests() {
				t.written[r] = true
			}

			instrs = append(instrs, instr)
		}
	}

	if len(errs) > 0 {
		return instrs, errs
	}
	return instrs, nil
}

func (t *transformer) transformInstr(s *syntax.InstrStmt) (Instr, *Error) {
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
		return nil, &Error{Range: s.Range, Msg: "empty instruction"}
	}
}

// ---- const assignment ----

func (t *transformer) transformConst(s *syntax.InstrStmt) (Instr, *Error) {
	if s.Dest == nil || s.Dest.Kind != syntax.DestReg {
		return nil, &Error{Range: s.Range, Msg: "const assignment requires a single register dest"}
	}
	dest := t.resolveReg(s.Dest.Regs[0])

	switch s.Const.Kind {
	case syntax.ConstString:
		return LoadStr{Dest: dest, Value: *s.Const.StrVal}, nil
	case syntax.ConstInt:
		n, ok := new(big.Int).SetString(*s.Const.IntVal, 10)
		if !ok {
			return nil, &Error{Range: s.Const.Range, Msg: fmt.Sprintf("invalid integer: %q", *s.Const.IntVal)}
		}
		return LoadInt{Dest: dest, Value: *n}, nil
	case syntax.ConstBool:
		return ConstBool{Dest: dest, Value: *s.Const.BoolVal}, nil
	default:
		return nil, &Error{Range: s.Range, Msg: "unknown const kind"}
	}
}

// ---- infix / compound assign ----

// transformInfix handles both `$d = $l + $r` and `$d += $r`: the parser gives
// the compound form the same shape, with left repeated as the dest.
func (t *transformer) transformInfix(s *syntax.InstrStmt, infix *syntax.Infix, what string) (Instr, *Error) {
	if s.Dest == nil || s.Dest.Kind != syntax.DestReg {
		return nil, &Error{Range: s.Range, Msg: what + " requires a single register dest"}
	}
	dest := t.resolveReg(s.Dest.Regs[0])
	left := t.resolveReg(infix.Left)
	right := t.resolveReg(infix.Right)

	var op BinKind
	switch infix.Op {
	case "+":
		op = OpAddInt{}
	case "-":
		op = OpSubInt{}
	default:
		return nil, &Error{Range: infix.Range, Msg: fmt.Sprintf("unknown infix operator: %q", infix.Op)}
	}
	return BinaryOp{Op: op, Dest: dest, Left: left, Right: right}, nil
}

// ---- call instructions ----

// argParser reads the args of one instruction call. Accessors report a bad arg
// themselves and return a zero value, so callers read args straight into an Instr
// literal and transformCall checks ap.errs once at the end. Composite literal
// operands evaluate left to right, so `f{a: ap.reg(), b: ap.reg()}` reads in order.
type argParser struct {
	t         *transformer
	args      []syntax.Arg
	pos       int
	seenLabel map[string]bool
	errs      *[]Error
	// callRange is where to point an error about an arg that isn't there.
	callRange parser.Range
}

func (t *transformer) newArgParser(call *syntax.InstrCall, errs *[]Error) *argParser {
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
func (ap *argParser) next(want syntax.ValueKind) *syntax.Value {
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
func (ap *argParser) reg() Reg {
	v := ap.next(syntax.ValReg)
	if v == nil {
		return 0
	}
	return ap.t.resolveReg(*v.Reg)
}

// optLabeledReg consumes an optional labeled arg with the given label as a register.
func (ap *argParser) optLabeledReg(name string) *Reg {
	a, ok := ap.labeledArg(name)
	if !ok {
		return nil
	}
	if a.Value.Kind != syntax.ValReg {
		ap.addErr(a.Range, "labeled arg %q: expected register, got %s", name, valueKindStr(a.Value.Kind))
		return nil
	}
	r := ap.t.resolveReg(*a.Value.Reg)
	return &r
}

// reqLabeledReg is optLabeledReg for a label the instruction can't do without.
func (ap *argParser) reqLabeledReg(name string) Reg {
	r := ap.optLabeledReg(name)
	if r == nil {
		ap.addErr(ap.callRange, "missing labeled argument %q", name)
		return 0
	}
	return *r
}

// labeledArg finds a labeled arg by name. Returns nil if not found.
func (ap *argParser) labeledArg(name string) (*syntax.Arg, bool) {
	if ap.seenLabel[name] {
		return nil, false
	}
	for i := ap.pos; i < len(ap.args); i++ {
		if ap.args[i].Label == name {
			ap.seenLabel[name] = true
			return &ap.args[i], true
		}
	}
	// also check already-consumed positional area
	for i := 0; i < ap.pos; i++ {
		if ap.args[i].Label == name {
			ap.seenLabel[name] = true
			return &ap.args[i], true
		}
	}
	return nil, false
}

// labelRef consumes the next positional arg as a label reference. It returns ""
// when the arg is missing or isn't one.
func (ap *argParser) labelRef() Label {
	v := ap.next(syntax.ValLabel)
	if v == nil {
		return ""
	}
	return Label(*v.Label)
}

// intLit consumes the next positional arg as an integer literal.
func (ap *argParser) intLit() uint16 {
	v := ap.next(syntax.ValInt)
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
func (ap *argParser) regList() []Reg {
	v := ap.next(syntax.ValRegList)
	if v == nil {
		return nil
	}
	regs := make([]Reg, len(*v.Regs))
	for i, r := range *v.Regs {
		regs[i] = ap.t.resolveReg(r)
	}
	return regs
}

func (ap *argParser) addErr(rng parser.Range, format string, args ...any) {
	*ap.errs = append(*ap.errs, Error{Range: rng, Msg: fmt.Sprintf(format, args...)})
}

func valueKindStr(k syntax.ValueKind) string {
	switch k {
	case syntax.ValReg:
		return "register"
	case syntax.ValLabel:
		return "label"
	case syntax.ValInt:
		return "integer literal"
	case syntax.ValRegList:
		return "register list"
	default:
		return "unknown"
	}
}

func (t *transformer) transformCall(s *syntax.InstrStmt) (Instr, *Error) {
	var errs []Error
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
	var dest Reg
	var dests []Reg
	if s.Dest != nil {
		switch s.Dest.Kind {
		case syntax.DestReg:
			dest = t.resolveReg(s.Dest.Regs[0])
		case syntax.DestDiscard:
			// `_` only exists in the text: desugar it to a fresh register, which
			// no other statement can name and nothing reads back.
			dest = t.freshReg()
		case syntax.DestList:
			dests = make([]Reg, len(s.Dest.Regs))
			for i, r := range s.Dest.Regs {
				dests[i] = t.resolveReg(r)
			}
		}
	}

	name, typeParam := s.Call.Name, s.Call.TypeParam

	// load_var and meta are the only instructions parameterized by a type.
	if typeParam != "" && name != "load_var" && name != "meta" {
		return nil, &Error{Range: s.Call.Range, Msg: fmt.Sprintf("%s doesn't take a type parameter", name)}
	}

	var instr Instr

	switch name {
	case "load_var":
		var typ VarType
		switch typeParam {
		case "int":
			typ = VarInt{}
		case "str":
			typ = VarStr{}
		default:
			return nil, &Error{Range: s.Call.Range, Msg: fmt.Sprintf("load_var: expected type parameter int or str, got %q", typeParam)}
		}
		instr = LoadVar{Dest: dest, Typ: typ, Index: ap.intLit()}

	case "meta":
		var typ MetaType
		switch typeParam {
		case "str":
			typ = MetaStr{}
		case "int":
			typ = MetaInt{}
		case "portion":
			typ = MetaPortion{}
		default:
			return nil, &Error{Range: s.Call.Range, Msg: fmt.Sprintf("meta: expected type parameter str, int or portion, got %q", typeParam)}
		}
		instr = MetaVar{Dest: dest, Typ: typ, Account: ap.reg(), Key: ap.reg()}

	case "balance":
		instr = FetchBalance{Dest: dest, Account: ap.reg(), Asset: ap.reg()}

	case "monetary_to_string":
		instr = ap.BinaryOp(dest, OpMonetaryToString{})
	case "mk_portion":
		instr = ap.BinaryOp(dest, OpMakePortion{})
	case "add_int":
		instr = ap.BinaryOp(dest, OpAddInt{})
	case "sub_int":
		instr = ap.BinaryOp(dest, OpSubInt{})
	case "add_string":
		instr = ap.BinaryOp(dest, OpAddString{})
	case "str_eq":
		instr = ap.BinaryOp(dest, OpStrEq{})
	case "sub_portion":
		instr = ap.BinaryOp(dest, OpSubPortion{})
	case "min_int":
		instr = ap.BinaryOp(dest, OpMinInt{})

	case "int_copy":
		instr = UnaryOp{Dest: dest, Op: OpIntCopy{}, Arg: ap.reg()}
	case "portion_copy":
		instr = UnaryOp{Dest: dest, Op: OpPortionCopy{}, Arg: ap.reg()}
	case "neg_int":
		instr = UnaryOp{Dest: dest, Op: OpNegInt{}, Arg: ap.reg()}
	case "int_to_string":
		instr = UnaryOp{Dest: dest, Op: OpIntToString{}, Arg: ap.reg()}
	case "is_zero":
		instr = UnaryOp{Dest: dest, Op: OpIsZero{}, Arg: ap.reg()}
	case "portion_to_string":
		instr = UnaryOp{Dest: dest, Op: OpPortionToString{}, Arg: ap.reg()}

	case "pull_account":
		instr = PullAccount{
			Dest:      dest,
			Account:   ap.reqLabeledReg("account"),
			Cap:       ap.optLabeledReg("cap"),
			Overdraft: ap.optLabeledReg("overdraft"),
			Color:     ap.optLabeledReg("color"),
		}
	case "send_to_account":
		instr = SendToAccount{Account: ap.optLabeledReg("account"), Cap: ap.optLabeledReg("cap")}
	case "save":
		instr = Save{
			Account: ap.reqLabeledReg("account"),
			Asset:   ap.reqLabeledReg("asset"),
			Amount:  ap.optLabeledReg("amount"),
		}

	case "mk_allot":
		if len(dests) == 0 {
			ap.addErr(s.Range, "mk_allot requires a dest list")
			break
		}
		instr = MakeAllotment{Dest: dests, Amount: ap.reg(), Portions: ap.regList()}

	case "meta_monetary":
		if len(dests) != 2 {
			ap.addErr(s.Range, "meta_monetary requires a dest list of 2 registers (asset, amount)")
			break
		}
		instr = MetaMonetary{
			DestAsset:  dests[0],
			DestAmount: dests[1],
			Account:    ap.reg(),
			Key:        ap.reg(),
		}

	case "check_enough_funds":
		instr = CheckEnoughFunds{Got: ap.reg(), Needed: ap.reg()}
	case "assert_leftover":
		instr = AssertLeftover{Portion: ap.reg(), Exact: false}
	case "assert_leftover_exact":
		instr = AssertLeftover{Portion: ap.reg(), Exact: true}
	case "set_current_asset":
		instr = SetCurrentAsset{Asset: ap.reg()}
	case "assert_same_asset":
		instr = AssertSameAsset{Left: ap.reg(), Right: ap.reg()}
	case "assert_valid_account":
		instr = AssertValidAccount{Account: ap.reg()}
	case "assert_valid_color":
		instr = AssertValidColor{Color: ap.reg()}
	case "assert_non_negative_balance":
		instr = AssertNonNegativeBalance{Balance: ap.reg(), Account: ap.reg()}

	case "snapshot":
		instr = Snapshot{Dest: dest}
	case "restore":
		instr = Restore{Mark: ap.reg()}

	case "set_tx_meta":
		instr = SetTxMeta{Key: ap.reg(), Value: ap.reg()}
	case "set_account_meta":
		instr = SetAccountMeta{Account: ap.reg(), Key: ap.reg(), Value: ap.reg()}

	case "jmp_if_false":
		cond, target := ap.reg(), ap.labelRef()
		if t.checkJmpTarget(ap, s, name, target) {
			instr = JmpIfFalse{Cond: cond, Target: target}
		}

	case "jmp_if_true":
		cond, target := ap.reg(), ap.labelRef()
		if t.checkJmpTarget(ap, s, name, target) {
			instr = JmpIfTrue{Cond: cond, Target: target}
		}

	case "jmp":
		target := ap.labelRef()
		if t.checkJmpTarget(ap, s, name, target) {
			instr = Jmp{Target: target}
		}

	default:
		return nil, &Error{Range: s.Call.Range, Msg: fmt.Sprintf("unknown instruction: %s", name)}
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

// checkJmpTarget reports whether target is a label the jump named name may
// reach: one that is defined, and not behind the jump. The VM only allows
// jumping forward — that's what makes every program terminate — so a backward
// jump is rejected here rather than assembled.
func (t *transformer) checkJmpTarget(ap *argParser, s *syntax.InstrStmt, name string, target Label) bool {
	labelPos, defined := t.labelPos[string(target)]
	switch {
	case target == "":
		// labelRef already said what was wrong
		return false
	case !defined:
		ap.addErr(s.Range, "%s: label %s is not defined in the program", name, target)
		return false
	case labelPos < t.stmtPos:
		ap.addErr(s.Range, "%s: label %s is behind the jump (jumps must go forward)", name, target)
		return false
	default:
		return true
	}
}

// BinaryOp reads the two register args every binary instruction takes.
func (ap *argParser) BinaryOp(dest Reg, op BinKind) BinaryOp {
	return BinaryOp{Dest: dest, Op: op, Left: ap.reg(), Right: ap.reg()}
}
