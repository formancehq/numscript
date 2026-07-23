package compiler

// Symbolic interpreter: walks the compiler IR ([]irInstr) and emits SMT-LIB2
// text that encodes the script's execution semantics as free-variable
// constraints (starting balances, vars and metadata are the unknowns). The
// emitted formula is meant to be fed to a solver (z3) together with a
// user-supplied query — see the `verify` package.
//
// This mirrors vm.Exec's opcode switch case-by-case, but instead of physical
// typed register banks it keeps a map[reg]term of SMT expressions. Register
// values are represented as SMT *expressions* (not mutable named constants),
// so in-place mutations (`$r5 += $r9`) are just a map reassignment and need no
// SSA versioning. The two places where a chain of updates would otherwise blow
// up the expression exponentially — the per-pull `available` amount and the
// per-send drained amount — are bound to fresh declared constants, which keeps
// the whole encoding linear in program size.
//
// Scope (v1): the "core scalar-send" opcode set exercised by the compiler
// snapshot fixtures. Anything outside it returns an *UnsupportedOpError rather
// than silently producing a wrong answer. See the deferred list in the switch
// default and in the package docs.

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/parser"
)

// worldAccountName mirrors vm.worldAccount: an infinite source, never
// consulted for a balance.
const worldAccountName = "world"

// AccountAsset keys the per-(account, asset) observable symbols.
type AccountAsset struct {
	Account string
	Asset   string
}

// SymbolTable maps the script's observable quantities to the SMT symbol names
// that stand for them in the emitted formula. The verify layer resolves DSL
// identifiers through this table.
//
// Sent/Received are already gated on failure: on an aborted run they are 0
// (a real abort discards every posting). end_balance and volumes are derived
// from these by the query layer as start + received - sent and received - sent.
type SymbolTable struct {
	// Fail is the name of the Bool symbol that is true iff the script aborts.
	Fail string
	// Start[k] is the free Int symbol for account k's starting balance
	// (asserted >= 0; absent for @world, which has no balance).
	Start map[AccountAsset]string
	// Sent[k] / Received[k] are the failure-gated Int symbols for how much
	// account k sent / received. Absent means "no activity" (i.e. 0).
	Sent     map[AccountAsset]string
	Received map[AccountAsset]string
	// Vars maps a numscript unknown (a var or a metadata read) to its free
	// SMT symbol. Keys are DSL identifiers, e.g. `var(0)` or `meta("acc","k")`.
	Vars map[string]string
	// order preserves the account/asset keys in first-touch order for
	// deterministic output and iteration.
	Order []AccountAsset
}

// Encoding is the result of symbolically interpreting a program.
type Encoding struct {
	// SMTLIB is the emitted preamble: all declare-const / assert statements
	// encoding the script. It has no (check-sat) — the query layer appends the
	// goal, (check-sat) and (get-model).
	SMTLIB  string
	Symbols SymbolTable
}

// UnsupportedOpError is returned when the IR uses a construct outside the v1
// core scalar-send scope. It names the offending instruction so callers can
// surface a precise "not yet supported" message.
type UnsupportedOpError struct {
	Op     string
	Reason string
}

func (e *UnsupportedOpError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("symbolic: unsupported op %s", e.Op)
	}
	return fmt.Sprintf("symbolic: unsupported op %s: %s", e.Op, e.Reason)
}

// SymbolicEncodeSource parses and compiles a numscript source string to IR,
// then symbolically encodes it. Convenience entry point for the verify layer
// and CLI.
func SymbolicEncodeSource(source string) (*Encoding, error) {
	program := parser.Parse(source)
	if len(program.Errors) != 0 {
		return nil, fmt.Errorf("parse errors: %s", parser.ParseErrorsToString(program.Errors, source))
	}
	compiled, cErr := compileProgramToIR(program.Value)
	if cErr != nil {
		return nil, fmt.Errorf("compile error: %v", cErr)
	}
	return simulate(compiled.instructions)
}

// termKind tags what a register currently holds.
type termKind uint8

const (
	tInt termKind = iota
	tStr
	tPortion
	tMonetary
)

// term is a tagged union standing for a register's current value.
type term struct {
	kind termKind

	i string // tInt: an SMT Int expression (symbol or composed)

	s string // tStr: the literal string value (v1: strings are literals)

	pNum, pDen *big.Int // tPortion: exact literal rational num/den

	mAsset string // tMonetary: asset literal
	mAmt   string // tMonetary: amount SMT Int expression
}

// interp holds the running symbolic-interpreter state.
type interp struct {
	regs map[reg]term

	// lines accumulates declare-const / assert statements in emission order.
	lines []string

	// currentAsset is the asset set by set_current_asset (a literal in core).
	currentAsset string

	// per-(account,asset) running SMT expressions
	curBalance map[AccountAsset]string // current balance expr (start minus debits)
	sentRaw    map[AccountAsset]string // ungated sent expr
	recvRaw    map[AccountAsset]string // ungated received expr

	// pool is the running total of pulled-but-unsent funds (the FIFO queue's
	// scalar total; color is out of scope so ordering is irrelevant).
	pool string

	// failExpr is the OR of every failure condition seen so far.
	failExpr string

	sym  SymbolTable
	seen map[AccountAsset]bool // account/asset already declared

	availCount int
	sendCount  int
}

func simulate(instrs []irInstr) (*Encoding, error) {
	it := &interp{
		regs:       map[reg]term{},
		curBalance: map[AccountAsset]string{},
		sentRaw:    map[AccountAsset]string{},
		recvRaw:    map[AccountAsset]string{},
		pool:       "0",
		failExpr:   "false",
		seen:       map[AccountAsset]bool{},
		sym: SymbolTable{
			Fail:     "fail",
			Start:    map[AccountAsset]string{},
			Sent:     map[AccountAsset]string{},
			Received: map[AccountAsset]string{},
			Vars:     map[string]string{},
		},
	}

	for _, instr := range instrs {
		if err := it.step(instr); err != nil {
			return nil, err
		}
	}

	it.finish()

	return &Encoding{
		SMTLIB:  strings.Join(it.lines, "\n") + "\n",
		Symbols: it.sym,
	}, nil
}

// --- SMT expression helpers -------------------------------------------------

func smtMax0(x string) string { return fmt.Sprintf("(ite (< %s 0) 0 %s)", x, x) }

func smtMin(a, b string) string { return fmt.Sprintf("(ite (<= %s %s) %s %s)", a, b, a, b) }

func smtAdd(a, b string) string { return fmt.Sprintf("(+ %s %s)", a, b) }

func smtSub(a, b string) string { return fmt.Sprintf("(- %s %s)", a, b) }

func (it *interp) emit(line string) { it.lines = append(it.lines, line) }

func (it *interp) declareInt(name, expr string) {
	it.emit(fmt.Sprintf("(declare-const %s Int)", name))
	it.emit(fmt.Sprintf("(assert (= %s %s))", name, expr))
}

func (it *interp) orFail(cond string) {
	if cond == "false" {
		return
	}
	if it.failExpr == "false" {
		it.failExpr = cond
		return
	}
	it.failExpr = fmt.Sprintf("(or %s %s)", it.failExpr, cond)
}

// sanitize turns an account/asset literal into a valid SMT symbol fragment.
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "x"
	}
	return b.String()
}

// touchAccount lazily declares start_<acct>_<asset> the first time an account
// is used. @world gets no balance symbol (it is an infinite source).
func (it *interp) touchAccount(account, asset string) {
	key := AccountAsset{account, asset}
	if it.seen[key] {
		return
	}
	it.seen[key] = true
	it.sym.Order = append(it.sym.Order, key)

	if account == worldAccountName {
		// no balance; still register the key so end_balance queries resolve
		it.curBalance[key] = "0"
		return
	}
	name := fmt.Sprintf("start_%s_%s", sanitize(account), sanitize(asset))
	it.emit(fmt.Sprintf("(declare-const %s Int)", name))
	it.emit(fmt.Sprintf("(assert (>= %s 0))", name))
	it.sym.Start[key] = name
	it.curBalance[key] = name
}

// --- per-instruction interpretation ----------------------------------------

func (it *interp) step(instr irInstr) error {
	switch in := instr.(type) {
	case loadInt:
		v := in.value // big.Int
		it.regs[in.dest] = term{kind: tInt, i: (&v).String()}
		return nil

	case loadStr:
		it.regs[in.dest] = term{kind: tStr, s: in.value}
		return nil

	case loadVar:
		return it.loadVar(in)

	case metaVar:
		return it.metaVar(in)

	case unaryOp:
		return it.unaryOp(in)

	case binaryOp:
		return it.binaryOp(in)

	case setCurrentAsset:
		t := it.regs[in.asset]
		if t.kind != tStr {
			return &UnsupportedOpError{Op: "set_current_asset", Reason: "symbolic asset"}
		}
		it.currentAsset = t.s
		return nil

	case pullAccount:
		return it.pullAccount(in)

	case sendToAccount:
		return it.sendToAccount(in)

	case checkEnoughFunds:
		got := it.regs[in.got].i
		needed := it.regs[in.needed].i
		it.orFail(fmt.Sprintf("(< %s %s)", got, needed))
		return nil

	case assertSameAsset:
		l := it.regs[in.left]
		r := it.regs[in.right]
		if l.kind != tStr || r.kind != tStr {
			return &UnsupportedOpError{Op: "assert_same_asset", Reason: "symbolic asset"}
		}
		if l.s != r.s {
			it.orFail("true") // guaranteed asset mismatch -> always aborts
		}
		return nil

	case assertValidAccount:
		// literal accounts coming from source are valid by construction;
		// nothing to assert. (Symbolic accounts are out of scope.)
		if it.regs[in.account].kind != tStr {
			return &UnsupportedOpError{Op: "assert_valid_account", Reason: "symbolic account"}
		}
		return nil

	case jmpIfZero, labelMarker:
		// Forward-only jumps. In compiled source blocks the only jump idiom is
		// "sequential source, jump out once the remaining need hits zero". A
		// later pull with cap == 0 already yields available == 0
		// (min(max(0,bal),0) == 0), so simulating straight through without
		// honoring the jump produces an identical result. We therefore treat
		// jumps and labels as no-ops in v1. (A general fork/merge is future
		// work for non-source-block jump patterns.)
		return nil

	default:
		return &UnsupportedOpError{Op: fmt.Sprintf("%T", instr)}
	}
}

func (it *interp) loadVar(in loadVar) error {
	if _, ok := in.typ.(varInt); !ok {
		return &UnsupportedOpError{Op: "load_var", Reason: fmt.Sprintf("non-int var type %s", in.typ)}
	}
	name := fmt.Sprintf("var_%d", in.index)
	id := fmt.Sprintf("var(%d)", in.index)
	if _, ok := it.sym.Vars[id]; !ok {
		it.emit(fmt.Sprintf("(declare-const %s Int)", name))
		it.sym.Vars[id] = name
	}
	it.regs[in.dest] = term{kind: tInt, i: name}
	return nil
}

func (it *interp) metaVar(in metaVar) error {
	if _, ok := in.typ.(metaInt); !ok {
		return &UnsupportedOpError{Op: "meta_var", Reason: fmt.Sprintf("non-int meta type %s", in.typ)}
	}
	acct := it.regs[in.account]
	key := it.regs[in.key]
	if acct.kind != tStr || key.kind != tStr {
		return &UnsupportedOpError{Op: "meta_var", Reason: "symbolic account/key"}
	}
	name := fmt.Sprintf("meta_%s_%s", sanitize(acct.s), sanitize(key.s))
	id := fmt.Sprintf("meta(%q,%q)", acct.s, key.s)
	if _, ok := it.sym.Vars[id]; !ok {
		it.emit(fmt.Sprintf("(declare-const %s Int)", name))
		it.sym.Vars[id] = name
	}
	it.regs[in.dest] = term{kind: tInt, i: name}
	return nil
}

func (it *interp) unaryOp(in unaryOp) error {
	arg := it.regs[in.arg]
	switch in.op.(type) {
	case opIntCopy:
		it.regs[in.dest] = arg
	case opPortionCopy:
		it.regs[in.dest] = arg
	case opNegInt:
		it.regs[in.dest] = term{kind: tInt, i: fmt.Sprintf("(- 0 %s)", arg.i)}
	case opGetAsset:
		if arg.kind != tMonetary {
			return &UnsupportedOpError{Op: "get_asset", Reason: "arg is not monetary"}
		}
		it.regs[in.dest] = term{kind: tStr, s: arg.mAsset}
	case opGetAmount:
		if arg.kind != tMonetary {
			return &UnsupportedOpError{Op: "get_amount", Reason: "arg is not monetary"}
		}
		it.regs[in.dest] = term{kind: tInt, i: arg.mAmt}
	default:
		return &UnsupportedOpError{Op: fmt.Sprintf("unary %s", in.op)}
	}
	return nil
}

func (it *interp) binaryOp(in binaryOp) error {
	l := it.regs[in.left]
	r := it.regs[in.right]
	switch in.op.(type) {
	case opAddInt:
		it.regs[in.dest] = term{kind: tInt, i: smtAdd(l.i, r.i)}
	case opSubInt:
		it.regs[in.dest] = term{kind: tInt, i: smtSub(l.i, r.i)}
	case opMinInt:
		it.regs[in.dest] = term{kind: tInt, i: smtMin(l.i, r.i)}
	case opMakeMonetary:
		if l.kind != tStr {
			return &UnsupportedOpError{Op: "mk_monetary", Reason: "symbolic asset"}
		}
		it.regs[in.dest] = term{kind: tMonetary, mAsset: l.s, mAmt: r.i}
	case opMakePortion:
		// literal-fold num/den to an exact rational (both trace to loadInt).
		num, nok := new(big.Int).SetString(l.i, 10)
		den, dok := new(big.Int).SetString(r.i, 10)
		if !nok || !dok {
			return &UnsupportedOpError{Op: "mk_portion", Reason: "symbolic portion (num/den not literal)"}
		}
		it.regs[in.dest] = term{kind: tPortion, pNum: num, pDen: den}
	default:
		return &UnsupportedOpError{Op: fmt.Sprintf("binary %s", in.op)}
	}
	return nil
}

func (it *interp) pullAccount(in pullAccount) error {
	if in.color != nil {
		return &UnsupportedOpError{Op: "pull_account", Reason: "colored source (queue model not implemented)"}
	}
	acctT := it.regs[in.account]
	if acctT.kind != tStr {
		return &UnsupportedOpError{Op: "pull_account", Reason: "symbolic account"}
	}
	account := acctT.s
	asset := it.currentAsset
	it.touchAccount(account, asset)
	key := AccountAsset{account, asset}

	if in.cap == nil {
		// PullUncapped path (no cap register). Deferred in v1.
		return &UnsupportedOpError{Op: "pull_account", Reason: "uncapped pull"}
	}
	capExpr := it.regs[*in.cap].i

	availName := fmt.Sprintf("avail_%d", it.availCount)
	it.availCount++

	var availExpr string
	if account == worldAccountName {
		// world: infinite source, available = max(0, cap), no balance consulted.
		availExpr = smtMax0(capExpr)
	} else {
		curBal := it.curBalance[key]
		od := "0"
		if in.overdraft != nil {
			od = it.regs[*in.overdraft].i
		}
		// eff = max(0, currentBal + max(0, overdraft)); avail = max(0, min(cap, eff))
		eff0 := fmt.Sprintf("(+ %s (ite (> %s 0) %s 0))", curBal, od, od)
		eff := smtMax0(eff0)
		availExpr = smtMax0(smtMin(capExpr, eff))
	}

	it.declareInt(availName, availExpr)

	// debit the source balance in place (skip for world, which has none)
	if account != worldAccountName {
		it.curBalance[key] = smtSub(it.curBalance[key], availName)
	}
	// record as sent, and add to the pool of unsent funds
	it.sentRaw[key] = addOrInit(it.sentRaw[key], availName)
	it.pool = smtAdd(it.pool, availName)

	it.regs[in.dest] = term{kind: tInt, i: availName}
	return nil
}

func (it *interp) sendToAccount(in sendToAccount) error {
	amtName := fmt.Sprintf("sent_%d", it.sendCount)
	it.sendCount++

	var amtExpr string
	if in.cap == nil {
		amtExpr = it.pool // drain everything
	} else {
		amtExpr = smtMin(it.regs[*in.cap].i, it.pool)
	}
	it.declareInt(amtName, amtExpr)
	it.pool = smtSub(it.pool, amtName)

	if in.account != nil {
		acctT := it.regs[*in.account]
		if acctT.kind != tStr {
			return &UnsupportedOpError{Op: "send_to_account", Reason: "symbolic account"}
		}
		account := acctT.s
		asset := it.currentAsset
		it.touchAccount(account, asset)
		key := AccountAsset{account, asset}
		it.recvRaw[key] = addOrInit(it.recvRaw[key], amtName)
	}
	// nil account => refund/keep path: funds leave the pool, credit no one.
	return nil
}

func addOrInit(existing, add string) string {
	if existing == "" {
		return add
	}
	return smtAdd(existing, add)
}

// finish binds the final observable symbols (fail, and the failure-gated
// sent/received per account) after all instructions have been processed.
func (it *interp) finish() {
	it.emit(fmt.Sprintf("(declare-const %s Bool)", it.sym.Fail))
	it.emit(fmt.Sprintf("(assert (= %s %s))", it.sym.Fail, it.failExpr))

	for _, key := range it.sym.Order {
		if raw, ok := it.sentRaw[key]; ok {
			name := fmt.Sprintf("sent_total_%s_%s", sanitize(key.Account), sanitize(key.Asset))
			gated := fmt.Sprintf("(ite %s 0 %s)", it.sym.Fail, raw)
			it.declareInt(name, gated)
			it.sym.Sent[key] = name
		}
		if raw, ok := it.recvRaw[key]; ok {
			name := fmt.Sprintf("recv_total_%s_%s", sanitize(key.Account), sanitize(key.Asset))
			gated := fmt.Sprintf("(ite %s 0 %s)", it.sym.Fail, raw)
			it.declareInt(name, gated)
			it.sym.Received[key] = name
		}
	}
}
