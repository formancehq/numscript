package verify

// A small user-facing formula DSL and its compiler to SMT-LIB2.
//
// Grammar (lowest to highest precedence):
//
//	iff      := implies ( "<=>" implies )*
//	implies  := or ( "=>" or )*            (right-associative)
//	or       := and ( ("||" | "or") and )*
//	and      := cmp ( ("&&" | "and") cmp )*
//	cmp      := unary ( ("=="|"!="|"<"|"<="|">"|">=") unary )?
//	unary    := "!" unary | primary
//	primary  := "(" iff ")"
//	          | "true" | "false" | "fail"
//	          | INT
//	          | STRING
//	          | ident "(" args ")"     (predicate / monetary / accessor call)
//	          | ident                  (a var symbol, e.g. from a numscript var)
//
// Predicates start_balance/end_balance/sent/received/volumes take
// (accountString, assetString) and yield an Int. `fail` is a Bool. `mon`,
// `get_amount`, `get_asset` mirror the numscript monetary ops.

import (
	"fmt"
	"strings"

	"github.com/formancehq/numscript/internal/compiler"
)

// --- lexer ------------------------------------------------------------------

type tokKind int

const (
	tkEOF tokKind = iota
	tkIdent
	tkInt
	tkString
	tkLParen
	tkRParen
	tkComma
	tkOp // any operator/keyword-operator, value in tok.s
)

type token struct {
	kind tokKind
	s    string
}

func lex(src string) ([]token, error) {
	var toks []token
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '(':
			toks = append(toks, token{tkLParen, "("})
			i++
		case c == ')':
			toks = append(toks, token{tkRParen, ")"})
			i++
		case c == ',':
			toks = append(toks, token{tkComma, ","})
			i++
		case c == '"':
			j := i + 1
			for j < len(src) && src[j] != '"' {
				j++
			}
			if j >= len(src) {
				return nil, fmt.Errorf("unterminated string literal")
			}
			toks = append(toks, token{tkString, src[i+1 : j]})
			i = j + 1
		case c >= '0' && c <= '9':
			j := i
			for j < len(src) && src[j] >= '0' && src[j] <= '9' {
				j++
			}
			toks = append(toks, token{tkInt, src[i:j]})
			i = j
		case isIdentStart(c):
			j := i
			for j < len(src) && isIdentPart(src[j]) {
				j++
			}
			word := src[i:j]
			switch word {
			case "and", "or", "xor":
				toks = append(toks, token{tkOp, word})
			default:
				toks = append(toks, token{tkIdent, word})
			}
			i = j
		default:
			// multi/single char operators
			op, n := lexOp(src[i:])
			if n == 0 {
				return nil, fmt.Errorf("unexpected character %q", string(c))
			}
			toks = append(toks, token{tkOp, op})
			i += n
		}
	}
	toks = append(toks, token{tkEOF, ""})
	return toks, nil
}

func lexOp(s string) (string, int) {
	// longest-match multichar operators first
	for _, op := range []string{"<=>", "=>", "==", "!=", "<=", ">=", "&&", "||"} {
		if strings.HasPrefix(s, op) {
			return op, len(op)
		}
	}
	switch s[0] {
	case '!', '<', '>', '+', '-', '*':
		return string(s[0]), 1
	}
	return "", 0
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

// --- AST --------------------------------------------------------------------

type node interface{ isNode() }

type binNode struct {
	op    string
	left  node
	right node
}
type notNode struct{ arg node }
type callNode struct {
	name string
	args []node
}
type intNode struct{ v string }
type strNode struct{ v string }
type identNode struct{ name string }
type litBoolNode struct{ v bool }
type failNode struct{}

func (binNode) isNode()     {}
func (notNode) isNode()     {}
func (callNode) isNode()    {}
func (intNode) isNode()     {}
func (strNode) isNode()     {}
func (identNode) isNode()   {}
func (litBoolNode) isNode() {}
func (failNode) isNode()    {}

// --- parser (recursive descent) ---------------------------------------------

type parser struct {
	toks []token
	pos  int
}

func parseQuery(src string) (node, error) {
	toks, err := lex(src)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	n, err := p.parseIff()
	if err != nil {
		return nil, err
	}
	if p.peek().kind != tkEOF {
		return nil, fmt.Errorf("unexpected trailing input near %q", p.peek().s)
	}
	return n, nil
}

func (p *parser) peek() token { return p.toks[p.pos] }
func (p *parser) next() token { t := p.toks[p.pos]; p.pos++; return t }

func (p *parser) acceptOp(ops ...string) (string, bool) {
	t := p.peek()
	if t.kind != tkOp {
		return "", false
	}
	for _, op := range ops {
		if t.s == op {
			p.pos++
			return op, true
		}
	}
	return "", false
}

func (p *parser) parseIff() (node, error) {
	left, err := p.parseImplies()
	if err != nil {
		return nil, err
	}
	for {
		if _, ok := p.acceptOp("<=>"); !ok {
			return left, nil
		}
		right, err := p.parseImplies()
		if err != nil {
			return nil, err
		}
		left = binNode{"<=>", left, right}
	}
}

func (p *parser) parseImplies() (node, error) {
	left, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if _, ok := p.acceptOp("=>"); ok {
		right, err := p.parseImplies() // right-associative
		if err != nil {
			return nil, err
		}
		return binNode{"=>", left, right}, nil
	}
	return left, nil
}

func (p *parser) parseOr() (node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for {
		if _, ok := p.acceptOp("||", "or"); !ok {
			return left, nil
		}
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = binNode{"or", left, right}
	}
}

func (p *parser) parseAnd() (node, error) {
	left, err := p.parseCmp()
	if err != nil {
		return nil, err
	}
	for {
		if _, ok := p.acceptOp("&&", "and", "xor"); !ok {
			return left, nil
		}
		// note: xor handled here for precedence simplicity
		opTok := p.toks[p.pos-1].s
		right, err := p.parseCmp()
		if err != nil {
			return nil, err
		}
		if opTok == "xor" {
			left = binNode{"xor", left, right}
		} else {
			left = binNode{"and", left, right}
		}
	}
}

func (p *parser) parseCmp() (node, error) {
	left, err := p.parseAdd()
	if err != nil {
		return nil, err
	}
	if op, ok := p.acceptOp("==", "!=", "<", "<=", ">", ">="); ok {
		right, err := p.parseAdd()
		if err != nil {
			return nil, err
		}
		return binNode{op, left, right}, nil
	}
	return left, nil
}

func (p *parser) parseAdd() (node, error) {
	left, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for {
		op, ok := p.acceptOp("+", "-")
		if !ok {
			return left, nil
		}
		right, err := p.parseMul()
		if err != nil {
			return nil, err
		}
		left = binNode{op, left, right}
	}
}

func (p *parser) parseMul() (node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		if _, ok := p.acceptOp("*"); !ok {
			return left, nil
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = binNode{"*", left, right}
	}
}

func (p *parser) parseUnary() (node, error) {
	if _, ok := p.acceptOp("!"); ok {
		arg, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return notNode{arg}, nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (node, error) {
	t := p.next()
	switch t.kind {
	case tkLParen:
		n, err := p.parseIff()
		if err != nil {
			return nil, err
		}
		if p.next().kind != tkRParen {
			return nil, fmt.Errorf("expected ')'")
		}
		return n, nil
	case tkInt:
		return intNode{t.s}, nil
	case tkString:
		return strNode{t.s}, nil
	case tkIdent:
		switch t.s {
		case "true":
			return litBoolNode{true}, nil
		case "false":
			return litBoolNode{false}, nil
		case "fail":
			return failNode{}, nil
		}
		if p.peek().kind == tkLParen {
			return p.parseCall(t.s)
		}
		return identNode{t.s}, nil
	default:
		return nil, fmt.Errorf("unexpected token %q", t.s)
	}
}

func (p *parser) parseCall(name string) (node, error) {
	p.next() // consume '('
	var args []node
	if p.peek().kind != tkRParen {
		for {
			a, err := p.parseIff()
			if err != nil {
				return nil, err
			}
			args = append(args, a)
			if p.peek().kind == tkComma {
				p.next()
				continue
			}
			break
		}
	}
	if p.next().kind != tkRParen {
		return nil, fmt.Errorf("expected ')' in call to %s", name)
	}
	return callNode{name, args}, nil
}

// --- resolver: AST -> typed SMT expression ----------------------------------

type exprKind int

const (
	kindBool exprKind = iota
	kindInt
	kindStr
	kindMonetary
)

func (k exprKind) String() string {
	switch k {
	case kindBool:
		return "bool"
	case kindInt:
		return "int"
	case kindStr:
		return "string"
	case kindMonetary:
		return "monetary"
	default:
		return "?"
	}
}

// sExpr is a resolved DSL expression: an SMT fragment plus its type. For
// monetary values we also keep the (literal) asset so mismatched comparisons
// are a DSL error rather than a silent int comparison.
type sExpr struct {
	kind  exprKind
	smt   string // for bool/int: the SMT expression
	str   string // for kindStr: the literal string value
	asset string // for kindMonetary: the literal asset
}

func resolve(n node, sym compiler.SymbolTable) (sExpr, error) {
	switch v := n.(type) {
	case litBoolNode:
		if v.v {
			return sExpr{kind: kindBool, smt: "true"}, nil
		}
		return sExpr{kind: kindBool, smt: "false"}, nil
	case failNode:
		return sExpr{kind: kindBool, smt: sym.Fail}, nil
	case intNode:
		return sExpr{kind: kindInt, smt: v.v}, nil
	case strNode:
		return sExpr{kind: kindStr, str: v.v}, nil
	case identNode:
		if s, ok := sym.Vars[v.name]; ok {
			return sExpr{kind: kindInt, smt: s}, nil
		}
		return sExpr{}, fmt.Errorf("unknown identifier %q", v.name)
	case notNode:
		a, err := resolve(v.arg, sym)
		if err != nil {
			return sExpr{}, err
		}
		if a.kind != kindBool {
			return sExpr{}, fmt.Errorf("'!' expects bool, got %s", a.kind)
		}
		return sExpr{kind: kindBool, smt: fmt.Sprintf("(not %s)", a.smt)}, nil
	case callNode:
		return resolveCall(v, sym)
	case binNode:
		return resolveBin(v, sym)
	default:
		return sExpr{}, fmt.Errorf("unhandled node %T", n)
	}
}

func resolveBin(v binNode, sym compiler.SymbolTable) (sExpr, error) {
	l, err := resolve(v.left, sym)
	if err != nil {
		return sExpr{}, err
	}
	r, err := resolve(v.right, sym)
	if err != nil {
		return sExpr{}, err
	}

	switch v.op {
	case "and", "or":
		if l.kind != kindBool || r.kind != kindBool {
			return sExpr{}, fmt.Errorf("'%s' expects bool operands", v.op)
		}
		return sExpr{kind: kindBool, smt: fmt.Sprintf("(%s %s %s)", v.op, l.smt, r.smt)}, nil
	case "xor":
		if l.kind != kindBool || r.kind != kindBool {
			return sExpr{}, fmt.Errorf("'xor' expects bool operands")
		}
		return sExpr{kind: kindBool, smt: fmt.Sprintf("(xor %s %s)", l.smt, r.smt)}, nil
	case "=>":
		if l.kind != kindBool || r.kind != kindBool {
			return sExpr{}, fmt.Errorf("'=>' expects bool operands")
		}
		return sExpr{kind: kindBool, smt: fmt.Sprintf("(=> %s %s)", l.smt, r.smt)}, nil
	case "<=>":
		if l.kind != kindBool || r.kind != kindBool {
			return sExpr{}, fmt.Errorf("'<=>' expects bool operands")
		}
		return sExpr{kind: kindBool, smt: fmt.Sprintf("(= %s %s)", l.smt, r.smt)}, nil
	case "==", "!=":
		return resolveEq(v.op, l, r)
	case "<", "<=", ">", ">=":
		if l.kind != kindInt || r.kind != kindInt {
			return sExpr{}, fmt.Errorf("'%s' expects int operands, got %s and %s", v.op, l.kind, r.kind)
		}
		return sExpr{kind: kindBool, smt: fmt.Sprintf("(%s %s %s)", v.op, l.smt, r.smt)}, nil
	case "+", "-", "*":
		if l.kind != kindInt || r.kind != kindInt {
			return sExpr{}, fmt.Errorf("'%s' expects int operands, got %s and %s", v.op, l.kind, r.kind)
		}
		return sExpr{kind: kindInt, smt: fmt.Sprintf("(%s %s %s)", v.op, l.smt, r.smt)}, nil
	default:
		return sExpr{}, fmt.Errorf("unknown operator %q", v.op)
	}
}

func resolveEq(op string, l, r sExpr) (sExpr, error) {
	var body string
	switch {
	case l.kind == kindInt && r.kind == kindInt:
		body = fmt.Sprintf("(= %s %s)", l.smt, r.smt)
	case l.kind == kindMonetary && r.kind == kindMonetary:
		if l.asset != r.asset {
			return sExpr{}, fmt.Errorf("cannot compare monetary values of different assets (%q vs %q)", l.asset, r.asset)
		}
		body = fmt.Sprintf("(= %s %s)", l.smt, r.smt)
	case l.kind == kindBool && r.kind == kindBool:
		body = fmt.Sprintf("(= %s %s)", l.smt, r.smt)
	default:
		return sExpr{}, fmt.Errorf("cannot compare %s and %s", l.kind, r.kind)
	}
	if op == "!=" {
		body = fmt.Sprintf("(not %s)", body)
	}
	return sExpr{kind: kindBool, smt: body}, nil
}

func resolveCall(v callNode, sym compiler.SymbolTable) (sExpr, error) {
	switch v.name {
	case "start_balance", "end_balance", "sent", "received", "volumes":
		acct, asset, err := twoStringArgs(v, sym)
		if err != nil {
			return sExpr{}, err
		}
		return sExpr{kind: kindInt, smt: balanceExpr(v.name, acct, asset, sym)}, nil
	case "mon":
		if len(v.args) != 2 {
			return sExpr{}, fmt.Errorf("mon expects (asset, amount)")
		}
		asset, err := resolve(v.args[0], sym)
		if err != nil {
			return sExpr{}, err
		}
		amt, err := resolve(v.args[1], sym)
		if err != nil {
			return sExpr{}, err
		}
		if asset.kind != kindStr || amt.kind != kindInt {
			return sExpr{}, fmt.Errorf("mon expects (string asset, int amount)")
		}
		return sExpr{kind: kindMonetary, smt: amt.smt, asset: asset.str}, nil
	case "get_amount":
		if len(v.args) != 1 {
			return sExpr{}, fmt.Errorf("get_amount expects (monetary)")
		}
		m, err := resolve(v.args[0], sym)
		if err != nil {
			return sExpr{}, err
		}
		if m.kind != kindMonetary {
			return sExpr{}, fmt.Errorf("get_amount expects a monetary value")
		}
		return sExpr{kind: kindInt, smt: m.smt}, nil
	case "get_asset":
		if len(v.args) != 1 {
			return sExpr{}, fmt.Errorf("get_asset expects (monetary)")
		}
		m, err := resolve(v.args[0], sym)
		if err != nil {
			return sExpr{}, err
		}
		if m.kind != kindMonetary {
			return sExpr{}, fmt.Errorf("get_asset expects a monetary value")
		}
		return sExpr{kind: kindStr, str: m.asset}, nil
	default:
		return sExpr{}, fmt.Errorf("unknown predicate %q", v.name)
	}
}

func twoStringArgs(v callNode, sym compiler.SymbolTable) (string, string, error) {
	if len(v.args) != 2 {
		return "", "", fmt.Errorf("%s expects (account, asset)", v.name)
	}
	a, err := resolve(v.args[0], sym)
	if err != nil {
		return "", "", err
	}
	b, err := resolve(v.args[1], sym)
	if err != nil {
		return "", "", err
	}
	if a.kind != kindStr || b.kind != kindStr {
		return "", "", fmt.Errorf("%s expects string (account, asset)", v.name)
	}
	return a.str, b.str, nil
}

// balanceExpr builds the SMT integer expression for a balance-family predicate.
// Missing symbols default to "0" (an untouched account / no activity).
func balanceExpr(name, acct, asset string, sym compiler.SymbolTable) string {
	key := compiler.AccountAsset{Account: acct, Asset: asset}
	start := orZero(sym.Start[key])
	sent := orZero(sym.Sent[key])
	recv := orZero(sym.Received[key])
	switch name {
	case "start_balance":
		return start
	case "sent":
		return sent
	case "received":
		return recv
	case "volumes":
		return fmt.Sprintf("(- %s %s)", recv, sent)
	case "end_balance":
		// start + received - sent (received/sent are already failure-gated)
		return fmt.Sprintf("(- (+ %s %s) %s)", start, recv, sent)
	}
	return "0"
}

func orZero(s string) string {
	if s == "" {
		return "0"
	}
	return s
}
