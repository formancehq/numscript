package prop

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/parser"
)

const (
	precLowest = iota
	precIff
	precImplies
	precOr
	precAnd
	precNot
	precCmp
	precSum
	precTerm
	precUnary
	precAtom
)

type assoc int

const (
	assocLeft assoc = iota
	assocRight
	assocNone
)

type opInfo struct {
	op    BinaryOp
	prec  int
	assoc assoc
}

var binaryOps = map[BinaryOp]opInfo{
	Iff:     {Iff, precIff, assocNone},
	Implies: {Implies, precImplies, assocRight},
	Or:      {Or, precOr, assocLeft},
	And:     {And, precAnd, assocLeft},
	Eq:      {Eq, precCmp, assocNone},
	Ne:      {Ne, precCmp, assocNone},
	Lt:      {Lt, precCmp, assocNone},
	Le:      {Le, precCmp, assocNone},
	Gt:      {Gt, precCmp, assocNone},
	Ge:      {Ge, precCmp, assocNone},
	Add:     {Add, precSum, assocLeft},
	Sub:     {Sub, precSum, assocLeft},
	Mul:     {Mul, precTerm, assocLeft},
	Div:     {Div, precTerm, assocLeft},
	Mod:     {Mod, precTerm, assocLeft},
}

func binaryOp(t token) (opInfo, bool) {
	var op BinaryOp
	switch t.kind {
	case tokIff:
		op = Iff
	case tokImplies:
		op = Implies
	case tokEq:
		op = Eq
	case tokNe:
		op = Ne
	case tokLt:
		op = Lt
	case tokLe:
		op = Le
	case tokGt:
		op = Gt
	case tokGe:
		op = Ge
	case tokPlus:
		op = Add
	case tokMinus:
		op = Sub
	case tokStar:
		op = Mul
	case tokSlash:
		op = Div
	case tokPercent:
		op = Mod
	case tokName:
		switch t.text {
		case "and":
			op = And
		case "or":
			op = Or
		default:
			return opInfo{}, false
		}
	default:
		return opInfo{}, false
	}
	return binaryOps[op], true
}

var reservedNames = map[string]string{
	"forall":   "quantifiers are not supported yet",
	"exists":   "quantifiers are not supported yet",
	"in":       "quantifiers are not supported yet",
	"accounts": "quantifiers are not supported yet",
	"assets":   "quantifiers are not supported yet",
	"sum":      "sums are not supported yet",
	"meta":     "metadata is not supported yet",
}

var keywords = map[string]bool{
	"before": true, "after": true, "sent": true, "received": true, "moved": true, "delta": true,
	"success": true, "fails": true, "and": true, "or": true, "not": true, "true": true, "false": true,
	"prop": true, "min": true, "max": true,
}

type bailout struct{}

type parserState struct {
	toks []token
	i    int
	errs []Error
}

func newParser(src string) *parserState {
	return &parserState{toks: lex(src)}
}

// run executes f, stopping at the first failure.
func (p *parserState) run(f func()) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(bailout); !ok {
				panic(r)
			}
		}
	}()
	f()
}

func (p *parserState) fail(rng parser.Range, msg string) {
	p.errs = append(p.errs, Error{Range: rng, Msg: msg})
	panic(bailout{})
}

// cur is the current token; reaching a lexical error reports it.
func (p *parserState) cur() token {
	t := p.toks[p.i]
	if t.kind == tokInvalid {
		p.fail(t.rng, t.text)
	}
	return t
}

func (p *parserState) peek() token {
	if p.i+1 < len(p.toks) {
		return p.toks[p.i+1]
	}
	return p.toks[len(p.toks)-1]
}

func (p *parserState) advance() token {
	t := p.toks[p.i]
	if p.i < len(p.toks)-1 {
		p.i++
	}
	return t
}

func found(t token) string {
	if t.kind == tokEOF {
		return "end of input"
	}
	return "'" + t.text + "'"
}

func (p *parserState) expect(kind tokKind, what string) token {
	t := p.cur()
	if t.kind != kind {
		p.fail(t.rng, fmt.Sprintf("expected %s but found %s", what, found(t)))
	}
	return p.advance()
}

func (p *parserState) isName(text string) bool {
	t := p.cur()
	return t.kind == tokName && t.text == text
}

func span(a, b parser.Range) parser.Range {
	return parser.Range{Start: a.Start, End: b.End}
}

func ParseFile(src string) ([]Property, []Error) {
	p := newParser(src)
	var props []Property
	seen := map[string]bool{}
	p.run(func() {
		for p.cur().kind != tokEOF {
			kw := p.cur()
			if !p.isName("prop") {
				p.fail(kw.rng, "expected 'prop' but found "+found(kw))
			}
			p.advance()
			name := p.expect(tokName, "a property name")
			p.expect(tokColon, "':'")
			f := p.parseBinary(precLowest)
			if seen[name.text] {
				p.errs = append(p.errs, Error{Range: name.rng, Msg: fmt.Sprintf("property %q is defined twice", name.text)})
			}
			seen[name.text] = true
			props = append(props, Property{Name: name.text, Formula: f, Range: span(kw.rng, f.GetRange())})
		}
	})
	if len(p.errs) == 0 && len(props) == 0 {
		p.errs = append(p.errs, Error{Msg: "no properties found"})
	}
	if len(p.errs) > 0 {
		return nil, p.errs
	}
	return props, nil
}

func ParseInline(src string) (Property, []Error) {
	p := newParser(src)
	var prop Property
	p.run(func() {
		var start parser.Range
		named := false
		if t := p.cur(); t.kind == tokName && p.peek().kind == tokColon {
			p.advance()
			p.advance()
			prop.Name = t.text
			start = t.rng
			named = true
		}
		f := p.parseBinary(precLowest)
		if t := p.cur(); t.kind != tokEOF {
			p.fail(t.rng, "expected end of input but found "+found(t))
		}
		if !named {
			start = f.GetRange()
		}
		prop.Formula = f
		prop.Range = span(start, f.GetRange())
	})
	if len(p.errs) > 0 {
		return Property{}, p.errs
	}
	return prop, nil
}

func (p *parserState) parseBinary(minPrec int) Expr {
	l := p.parsePrefix()
	for {
		info, ok := binaryOp(p.cur())
		if !ok || info.prec < minPrec {
			return l
		}
		p.advance()
		next := info.prec + 1
		if info.assoc == assocRight {
			next = info.prec
		}
		r := p.parseBinary(next)
		l = &Binary{Range: span(l.GetRange(), r.GetRange()), Op: info.op, L: l, R: r}
		if info.assoc == assocNone {
			if again, ok := binaryOp(p.cur()); ok && again.prec == info.prec {
				if info.prec == precIff {
					p.fail(p.cur().rng, "'<=>' does not chain: write (a <=> b) <=> c")
				}
				p.fail(p.cur().rng, "comparisons do not chain: write (a < b) and (b < c)")
			}
		}
	}
}

func (p *parserState) parsePrefix() Expr {
	t := p.cur()
	switch {
	case t.kind == tokMinus:
		p.advance()
		x := p.parsePrefix()
		return &Unary{Range: span(t.rng, x.GetRange()), Op: Neg, X: x}
	case t.kind == tokName && t.text == "not":
		p.advance()
		x := p.parseBinary(precCmp)
		return &Unary{Range: span(t.rng, x.GetRange()), Op: Not, X: x}
	}
	return p.parsePrimary()
}

func (p *parserState) parsePrimary() Expr {
	t := p.cur()
	switch t.kind {
	case tokInt:
		p.advance()
		v, ok := new(big.Int).SetString(strings.ReplaceAll(t.text, "_", ""), 10)
		if !ok {
			p.fail(t.rng, "invalid integer "+t.text)
		}
		return &IntLit{Range: t.rng, Value: v}
	case tokLParen:
		p.advance()
		e := p.parseBinary(precLowest)
		p.expect(tokRParen, "')'")
		return e
	case tokName:
		return p.parseName(t)
	case tokVar:
		p.fail(t.rng, "script variables are not supported in properties yet")
	case tokString:
		p.fail(t.rng, "strings are not supported in properties")
	case tokLBracket:
		p.fail(t.rng, "monetary literals are not allowed in properties; amounts are plain integers in the asset's minor unit: write 100")
	case tokEOF:
		p.fail(t.rng, "unexpected end of input")
	}
	p.fail(t.rng, "expected an expression but found "+found(t))
	return nil
}

func (p *parserState) parseName(t token) Expr {
	switch t.text {
	case "true", "false":
		p.advance()
		return &BoolLit{Range: t.rng, Value: t.text == "true"}
	case "success":
		p.advance()
		return &Success{Range: t.rng}
	case "fails":
		p.advance()
		return &Fails{Range: t.rng}
	case "before":
		return p.parseAtomCall(Before)
	case "after":
		return p.parseAtomCall(After)
	case "sent":
		return p.parseAtomCall(Sent)
	case "received":
		return p.parseAtomCall(Received)
	case "moved":
		return p.parseAtomCall(Moved)
	case "delta":
		p.advance()
		p.expect(tokLParen, "'('")
		acc := p.expect(tokAccount, "an account")
		p.expect(tokComma, "','")
		asset := p.expect(tokAsset, "an asset")
		end := p.expect(tokRParen, "')'")
		return &Delta{Range: span(t.rng, end.rng), Account: acc.text[1:], Asset: asset.text}
	case "min", "max":
		return p.parseMinMax(t)
	}
	if msg, ok := reservedNames[t.text]; ok {
		p.fail(t.rng, msg)
	}
	if keywords[t.text] {
		p.fail(t.rng, "expected an expression but found "+found(t))
	}
	p.fail(t.rng, fmt.Sprintf("unknown name '%s'", t.text))
	return nil
}

func (p *parserState) parseAtomCall(kind AtomKind) Expr {
	name := p.advance()
	p.expect(tokLParen, "'('")
	atom := Atom{Kind: kind}
	atom.Account = p.expect(tokAccount, "an account").text[1:]
	p.expect(tokComma, "','")
	if kind == Moved {
		atom.Account2 = p.expect(tokAccount, "an account").text[1:]
		p.expect(tokComma, "','")
	}
	atom.Asset = p.expect(tokAsset, "an asset").text
	end := p.expect(tokRParen, "')'")
	return &AtomExpr{Range: span(name.rng, end.rng), Atom: atom}
}

func (p *parserState) parseMinMax(name token) Expr {
	p.advance()
	p.expect(tokLParen, "'('")
	var args []Expr
	for {
		args = append(args, p.parseBinary(precSum))
		t := p.cur()
		if t.kind == tokComma {
			p.advance()
			continue
		}
		if t.kind == tokRParen {
			break
		}
		p.fail(t.rng, "expected ',' or ')' but found "+found(t))
	}
	end := p.advance()
	rng := span(name.rng, end.rng)
	if len(args) < 2 {
		p.fail(rng, name.text+" takes at least two arguments")
	}
	return &MinMax{Range: rng, IsMax: name.text == "max", Args: args}
}
