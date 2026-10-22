package smt

import (
	"fmt"
	"math/big"
	"strings"
)

// sexpr is a node of solver output: an atom or a list.
type sexpr struct {
	atom   string
	list   []sexpr
	isList bool
}

func (s sexpr) String() string {
	if !s.isList {
		return s.atom
	}
	parts := make([]string, len(s.list))
	for i, c := range s.list {
		parts[i] = c.String()
	}
	return "(" + strings.Join(parts, " ") + ")"
}

// parseSexprs reads every top-level s-expression of text. Quoted symbols
// |..| and string literals ".." are kept as single atoms, with the delimiters
// stripped for symbols.
func parseSexprs(text string) ([]sexpr, error) {
	p := &sexprParser{src: text}
	var out []sexpr
	for {
		p.skipSpace()
		if p.pos >= len(p.src) {
			return out, nil
		}
		node, err := p.next()
		if err != nil {
			return nil, err
		}
		out = append(out, node)
	}
}

type sexprParser struct {
	src string
	pos int
}

func (p *sexprParser) skipSpace() {
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch c {
		case ' ', '\n', '\t', '\r':
			p.pos++
		case ';':
			for p.pos < len(p.src) && p.src[p.pos] != '\n' {
				p.pos++
			}
		default:
			return
		}
	}
}

func (p *sexprParser) next() (sexpr, error) {
	p.skipSpace()
	if p.pos >= len(p.src) {
		return sexpr{}, fmt.Errorf("smt: unexpected end of solver output")
	}
	switch c := p.src[p.pos]; c {
	case '(':
		p.pos++
		node := sexpr{isList: true}
		for {
			p.skipSpace()
			if p.pos >= len(p.src) {
				return sexpr{}, fmt.Errorf("smt: unclosed list in solver output")
			}
			if p.src[p.pos] == ')' {
				p.pos++
				return node, nil
			}
			child, err := p.next()
			if err != nil {
				return sexpr{}, err
			}
			node.list = append(node.list, child)
		}
	case ')':
		return sexpr{}, fmt.Errorf("smt: unexpected ')' in solver output")
	case '|':
		end := strings.IndexByte(p.src[p.pos+1:], '|')
		if end < 0 {
			return sexpr{}, fmt.Errorf("smt: unclosed |symbol| in solver output")
		}
		atom := p.src[p.pos+1 : p.pos+1+end]
		p.pos += end + 2
		return sexpr{atom: atom}, nil
	case '"':
		var sb strings.Builder
		p.pos++
		for {
			if p.pos >= len(p.src) {
				return sexpr{}, fmt.Errorf("smt: unclosed string in solver output")
			}
			if p.src[p.pos] == '"' {
				if p.pos+1 < len(p.src) && p.src[p.pos+1] == '"' {
					sb.WriteByte('"')
					p.pos += 2
					continue
				}
				p.pos++
				return sexpr{atom: sb.String()}, nil
			}
			sb.WriteByte(p.src[p.pos])
			p.pos++
		}
	default:
		start := p.pos
		for p.pos < len(p.src) && !strings.ContainsRune(" \n\t\r()|\"", rune(p.src[p.pos])) {
			p.pos++
		}
		return sexpr{atom: p.src[start:p.pos]}, nil
	}
}

// parseValue decodes a value as printed by get-value: a numeral, (- numeral),
// true or false.
func parseValue(s sexpr) (Value, error) {
	if !s.isList {
		switch s.atom {
		case "true":
			return BoolValue(true), nil
		case "false":
			return BoolValue(false), nil
		}
		n, ok := new(big.Int).SetString(s.atom, 10)
		if !ok {
			return Value{}, fmt.Errorf("smt: cannot parse value %q", s.atom)
		}
		return IntValue(n), nil
	}
	if len(s.list) == 2 && !s.list[0].isList && s.list[0].atom == "-" {
		inner, err := parseValue(s.list[1])
		if err != nil {
			return Value{}, err
		}
		if inner.Sort != SortInt {
			return Value{}, fmt.Errorf("smt: cannot negate %s", s)
		}
		return IntValue(new(big.Int).Neg(inner.Int)), nil
	}
	return Value{}, fmt.Errorf("smt: cannot parse value %s", s)
}
