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
		return t.transformInfix(s)
	case s.CompoundAssign != nil:
		return t.transformCompoundAssign(s)
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

func (t *transformer) transformInfix(s *irparser.InstrStmt) (irInstr, *TransformError) {
	if s.Dest == nil || s.Dest.Kind != irparser.DestReg {
		return nil, &TransformError{Range: s.Range, Msg: "infix requires a single register dest"}
	}
	dest := t.resolveReg(s.Dest.Regs[0])
	left := t.resolveReg(s.Infix.Left)
	right := t.resolveReg(s.Infix.Right)

	op, err := infixOp(s.Infix.Op)
	if err != nil {
		return nil, &TransformError{Range: s.Infix.Range, Msg: err.Error()}
	}
	return binaryOp{op: op, dest: dest, left: left, right: right}, nil
}

func (t *transformer) transformCompoundAssign(s *irparser.InstrStmt) (irInstr, *TransformError) {
	if s.Dest == nil || s.Dest.Kind != irparser.DestReg {
		return nil, &TransformError{Range: s.Range, Msg: "compound assign requires a single register dest"}
	}
	dest := t.resolveReg(s.Dest.Regs[0])
	left := t.resolveReg(s.CompoundAssign.Left)
	right := t.resolveReg(s.CompoundAssign.Right)

	op, err := infixOp(s.CompoundAssign.Op)
	if err != nil {
		return nil, &TransformError{Range: s.CompoundAssign.Range, Msg: err.Error()}
	}
	return binaryOp{op: op, dest: dest, left: left, right: right}, nil
}

func infixOp(op string) (binKind, error) {
	switch op {
	case "+":
		return opAddInt{}, nil
	case "-":
		return opSubInt{}, nil
	default:
		return nil, fmt.Errorf("unknown infix operator: %q", op)
	}
}

// ---- call instructions ----

// argParser helps validate and extract args for a specific instruction.
type argParser struct {
	t         *transformer
	args      []irparser.Arg
	pos       int
	seenLabel map[string]bool
	errs      *[]TransformError
}

func (t *transformer) newArgParser(args []irparser.Arg, errs *[]TransformError) *argParser {
	return &argParser{
		t:         t,
		args:      args,
		seenLabel: map[string]bool{},
		errs:      errs,
	}
}

// nextReg consumes the next positional arg as a register.
func (ap *argParser) nextReg() (reg, bool) {
	if ap.pos >= len(ap.args) {
		return 0, false
	}
	a := ap.args[ap.pos]
	ap.pos++
	if a.Value.Kind != irparser.ValReg {
		ap.addErr(a.Range, "expected register, got %s", valueKindStr(a.Value.Kind))
		return 0, false
	}
	return ap.t.resolveReg(*a.Value.Reg), true
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

// nextLabel consumes the next positional arg as a label reference.
func (ap *argParser) nextLabel() (label, bool) {
	if ap.pos >= len(ap.args) {
		return "", false
	}
	a := ap.args[ap.pos]
	ap.pos++
	if a.Value.Kind != irparser.ValLabel {
		ap.addErr(a.Range, "expected label reference, got %s", valueKindStr(a.Value.Kind))
		return "", false
	}
	return label(*a.Value.Label), true
}

// nextIntLit consumes the next positional arg as an integer literal.
func (ap *argParser) nextIntLit() (uint16, bool) {
	if ap.pos >= len(ap.args) {
		return 0, false
	}
	a := ap.args[ap.pos]
	ap.pos++
	if a.Value.Kind != irparser.ValInt {
		ap.addErr(a.Range, "expected integer literal, got %s", valueKindStr(a.Value.Kind))
		return 0, false
	}
	n, err := strconv.ParseUint(*a.Value.Int, 10, 16)
	if err != nil {
		ap.addErr(a.Range, "integer literal out of range (0-65535): %s", *a.Value.Int)
		return 0, false
	}
	return uint16(n), true
}

// nextRegList consumes the next positional arg as a register list.
func (ap *argParser) nextRegList() ([]reg, bool) {
	if ap.pos >= len(ap.args) {
		return nil, false
	}
	a := ap.args[ap.pos]
	ap.pos++
	if a.Value.Kind != irparser.ValRegList {
		ap.addErr(a.Range, "expected register list, got %s", valueKindStr(a.Value.Kind))
		return nil, false
	}
	regs := make([]reg, len(*a.Value.Regs))
	for i, r := range *a.Value.Regs {
		regs[i] = ap.t.resolveReg(r)
	}
	return regs, true
}

func (ap *argParser) addErr(rng parser.Range, format string, args ...interface{}) {
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
	case irparser.ValBool:
		return "boolean"
	case irparser.ValRegList:
		return "register list"
	default:
		return "unknown"
	}
}

func (t *transformer) transformCall(s *irparser.InstrStmt) (irInstr, *TransformError) {
	var errs []TransformError
	ap := t.newArgParser(s.Call.Args, &errs)

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
	hasDest := s.Dest != nil
	if hasDest {
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

	key := instrKey{s.Call.Name, s.Call.TypeParam}

	var instr irInstr
	var err *TransformError

	switch key {
	case instrKey{"load_var", "int"}:
		instr, err = parseLoadVar(ap, dest, varInt{}, s.Range)
	case instrKey{"load_var", "str"}:
		instr, err = parseLoadVar(ap, dest, varStr{}, s.Range)
	case instrKey{"meta", "str"}:
		instr, err = parseMetaVar(ap, dest, metaStr{}, s.Range)
	case instrKey{"meta", "int"}:
		instr, err = parseMetaVar(ap, dest, metaInt{}, s.Range)
	case instrKey{"meta", "portion"}:
		instr, err = parseMetaVar(ap, dest, metaPortion{}, s.Range)
	case instrKey{"meta", "monetary"}:
		instr, err = parseMetaVar(ap, dest, metaMonetary{}, s.Range)
	case instrKey{"balance", ""}:
		instr, err = parseFetchBalance(ap, dest, s.Range)
	case instrKey{"mk_monetary", ""}:
		instr, err = parseBinaryCall(ap, dest, opMakeMonetary{}, s.Range)
	case instrKey{"mk_portion", ""}:
		instr, err = parseBinaryCall(ap, dest, opMakePortion{}, s.Range)
	case instrKey{"add_int", ""}:
		instr, err = parseBinaryCall(ap, dest, opAddInt{}, s.Range)
	case instrKey{"sub_int", ""}:
		instr, err = parseBinaryCall(ap, dest, opSubInt{}, s.Range)
	case instrKey{"add_string", ""}:
		instr, err = parseBinaryCall(ap, dest, opAddString{}, s.Range)
	case instrKey{"sub_portion", ""}:
		instr, err = parseBinaryCall(ap, dest, opSubPortion{}, s.Range)
	case instrKey{"min_int", ""}:
		instr, err = parseBinaryCall(ap, dest, opMinInt{}, s.Range)
	case instrKey{"int_copy", ""}:
		instr, err = parseUnaryCall(ap, dest, opIntCopy{}, s.Range)
	case instrKey{"portion_copy", ""}:
		instr, err = parseUnaryCall(ap, dest, opPortionCopy{}, s.Range)
	case instrKey{"get_asset", ""}:
		instr, err = parseUnaryCall(ap, dest, opGetAsset{}, s.Range)
	case instrKey{"get_amount", ""}:
		instr, err = parseUnaryCall(ap, dest, opGetAmount{}, s.Range)
	case instrKey{"neg_int", ""}:
		instr, err = parseUnaryCall(ap, dest, opNegInt{}, s.Range)
	case instrKey{"int_to_string", ""}:
		instr, err = parseUnaryCall(ap, dest, opIntToString{}, s.Range)
	case instrKey{"portion_to_string", ""}:
		instr, err = parseUnaryCall(ap, dest, opPortionToString{}, s.Range)
	case instrKey{"monetary_to_string", ""}:
		instr, err = parseUnaryCall(ap, dest, opMonetaryToString{}, s.Range)
	case instrKey{"pull_account", ""}:
		instr, err = parsePullAccount(ap, dest, s.Range)
	case instrKey{"send_to_account", ""}:
		instr, err = parseSendToAccount(ap, s.Range)
	case instrKey{"mk_allot", ""}:
		instr, err = parseMakeAllotment(ap, dests, s.Range)
	case instrKey{"check_enough_funds", ""}:
		instr, err = parseCheckEnoughFunds(ap, s.Range)
	case instrKey{"save", ""}:
		instr, err = parseSave(ap, s.Range)
	case instrKey{"assert_leftover", ""}:
		instr, err = parseAssertLeftover(ap, false, s.Range)
	case instrKey{"assert_leftover_exact", ""}:
		instr, err = parseAssertLeftover(ap, true, s.Range)
	case instrKey{"set_current_asset", ""}:
		instr, err = parseSetCurrentAsset(ap, s.Range)
	case instrKey{"assert_same_asset", ""}:
		instr, err = parseAssertSameAsset(ap, s.Range)
	case instrKey{"assert_valid_account", ""}:
		instr, err = parseAssertValidAccount(ap, s.Range)
	case instrKey{"assert_non_negative_balance", ""}:
		instr, err = parseAssertNonNegativeBalance(ap, s.Range)
	case instrKey{"set_tx_meta", ""}:
		instr, err = parseSetTxMeta(ap, s.Range)
	case instrKey{"set_account_meta", ""}:
		instr, err = parseSetAccountMeta(ap, s.Range)
	case instrKey{"jmp_if_zero", ""}:
		instr, err = t.parseJmpIfZero(ap, s.Range)
	default:
		return nil, &TransformError{Range: s.Call.Range, Msg: fmt.Sprintf("unknown instruction: %s", key)}
	}

	if err != nil {
		return nil, err
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

type instrKey struct {
	name, typeParam string
}

func (k instrKey) String() string {
	if k.typeParam == "" {
		return k.name
	}
	return fmt.Sprintf("%s<%s>", k.name, k.typeParam)
}

// ---- individual instruction parsers ----

func parseLoadVar(ap *argParser, dest reg, typ varType, rng parser.Range) (irInstr, *TransformError) {
	index, ok := ap.nextIntLit()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return loadVar{dest: dest, typ: typ, index: index}, nil
}

func parseMetaVar(ap *argParser, dest reg, typ metaType, rng parser.Range) (irInstr, *TransformError) {
	account, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	key, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return metaVar{dest: dest, typ: typ, account: account, key: key}, nil
}

func parseFetchBalance(ap *argParser, dest reg, rng parser.Range) (irInstr, *TransformError) {
	account, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	asset, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return fetchBalance{dest: dest, account: account, asset: asset}, nil
}

func parseBinaryCall(ap *argParser, dest reg, op binKind, rng parser.Range) (irInstr, *TransformError) {
	left, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	right, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return binaryOp{op: op, dest: dest, left: left, right: right}, nil
}

func parseUnaryCall(ap *argParser, dest reg, op unKind, rng parser.Range) (irInstr, *TransformError) {
	arg, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return unaryOp{op: op, dest: dest, arg: arg}, nil
}

func parsePullAccount(ap *argParser, dest reg, rng parser.Range) (irInstr, *TransformError) {
	account := ap.optLabeledReg("account")
	if account == nil {
		return nil, &TransformError{Range: rng, Msg: "pull_account requires labeled arg 'account'"}
	}
	cap := ap.optLabeledReg("cap")
	overdraft := ap.optLabeledReg("overdraft")
	color := ap.optLabeledReg("color")
	return pullAccount{dest: dest, account: *account, cap: cap, overdraft: overdraft, color: color}, nil
}

func parseSendToAccount(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	account := ap.optLabeledReg("account")
	cap := ap.optLabeledReg("cap")
	return sendToAccount{account: account, cap: cap}, nil
}

func parseMakeAllotment(ap *argParser, dests []reg, rng parser.Range) (irInstr, *TransformError) {
	if len(dests) == 0 {
		return nil, &TransformError{Range: rng, Msg: "mk_allot requires a dest list"}
	}
	amount, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	portions, ok := ap.nextRegList()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return makeAllotment{dest: dests, amount: amount, portions: portions}, nil
}

func parseCheckEnoughFunds(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	got, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	needed, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return checkEnoughFunds{got: got, needed: needed}, nil
}

func parseSave(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	account := ap.optLabeledReg("account")
	asset := ap.optLabeledReg("asset")
	amount := ap.optLabeledReg("amount")
	if account == nil || asset == nil {
		return nil, &TransformError{Range: rng, Msg: "save requires labeled args 'account' and 'asset'"}
	}
	return save{account: *account, asset: *asset, amount: amount}, nil
}

func parseAssertLeftover(ap *argParser, exact bool, rng parser.Range) (irInstr, *TransformError) {
	portion, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return assertLeftover{portion: portion, exact: exact}, nil
}

func parseSetCurrentAsset(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	asset, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return setCurrentAsset{asset: asset}, nil
}

func parseAssertSameAsset(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	left, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	right, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return assertSameAsset{left: left, right: right}, nil
}

func parseAssertValidAccount(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	account, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return assertValidAccount{account: account}, nil
}

func parseAssertNonNegativeBalance(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	balance, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	account, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return assertNonNegativeBalance{balance: balance, account: account}, nil
}

func parseSetTxMeta(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	key, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	value, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return setTxMeta{key: key, value: value}, nil
}

func parseSetAccountMeta(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	account, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	key, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	value, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	return setAccountMeta{account: account, key: key, value: value}, nil
}

func (t *transformer) parseJmpIfZero(ap *argParser, rng parser.Range) (irInstr, *TransformError) {
	cond, ok := ap.nextReg()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	target, ok := ap.nextLabel()
	if !ok {
		return nil, firstTransformErr(ap.errs)
	}
	labelPos, ok := t.labelPos[string(target)]
	if !ok {
		return nil, &TransformError{Range: rng, Msg: fmt.Sprintf("jmp_if_zero: label %s is not defined in the program", target)}
	}
	// The VM only allows jumping forward: that's what makes every program
	// terminate, so a backward jump is rejected here rather than assembled.
	if labelPos < t.stmtPos {
		return nil, &TransformError{Range: rng, Msg: fmt.Sprintf("jmp_if_zero: label %s is behind the jump (jumps must go forward)", target)}
	}
	return jmpIfZero{cond: cond, target: target}, nil
}

func firstTransformErr(errs *[]TransformError) *TransformError {
	if errs == nil || len(*errs) == 0 {
		return nil
	}
	e := (*errs)[0]
	return &e
}
