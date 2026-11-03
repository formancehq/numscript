package prop

import (
	"fmt"

	"github.com/formancehq/numscript/internal/parser"
)

type tokKind int

const (
	tokEOF tokKind = iota
	tokInt
	tokName
	tokAccount
	tokAsset
	tokVar
	tokString
	tokLBracket
	tokLParen
	tokRParen
	tokComma
	tokColon
	tokPlus
	tokMinus
	tokStar
	tokSlash
	tokPercent
	tokEq
	tokNe
	tokLt
	tokLe
	tokGt
	tokGe
	tokImplies
	tokIff
	// tokInvalid carries the lexical error in text and ends the token stream.
	tokInvalid
)

type token struct {
	kind tokKind
	text string
	rng  parser.Range
}

type lexer struct {
	src  []rune
	pos  int
	line int
	col  int
}

func lex(src string) []token {
	l := &lexer{src: []rune(src)}
	var toks []token
	for {
		t := l.next()
		toks = append(toks, t)
		if t.kind == tokEOF || t.kind == tokInvalid {
			return toks
		}
	}
}

func (l *lexer) peek() rune { return l.peekAt(0) }

func (l *lexer) peekAt(n int) rune {
	if l.pos+n >= len(l.src) {
		return -1
	}
	return l.src[l.pos+n]
}

func (l *lexer) advance() {
	if l.src[l.pos] == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
	l.pos++
}

func (l *lexer) position() parser.Position {
	return parser.Position{Line: l.line, Character: l.col}
}

func (l *lexer) tok(kind tokKind, start parser.Position, startPos int) token {
	return token{
		kind: kind,
		text: string(l.src[startPos:l.pos]),
		rng:  parser.Range{Start: start, End: l.position()},
	}
}

func (l *lexer) invalid(start parser.Position, msg string) token {
	return token{kind: tokInvalid, text: msg, rng: parser.Range{Start: start, End: l.position()}}
}

func isDigit(r rune) bool    { return r >= '0' && r <= '9' }
func isLower(r rune) bool    { return r >= 'a' && r <= 'z' }
func isUpper(r rune) bool    { return r >= 'A' && r <= 'Z' }
func isNameChar(r rune) bool { return isLower(r) || isDigit(r) || r == '_' }
func isAccountChar(r rune) bool {
	return isLower(r) || isUpper(r) || isDigit(r) || r == '_' || r == '-'
}

func (l *lexer) skipSpaceAndComments() *token {
	for {
		switch {
		case l.peek() == ' ', l.peek() == '\t', l.peek() == '\r', l.peek() == '\n':
			l.advance()
		case l.peek() == '/' && l.peekAt(1) == '/':
			for l.peek() != -1 && l.peek() != '\n' {
				l.advance()
			}
		case l.peek() == '/' && l.peekAt(1) == '*':
			start := l.position()
			l.advance()
			l.advance()
			for l.peek() != '*' || l.peekAt(1) != '/' {
				if l.peek() == -1 {
					t := l.invalid(start, "unterminated block comment")
					return &t
				}
				l.advance()
			}
			l.advance()
			l.advance()
		default:
			return nil
		}
	}
}

func (l *lexer) next() token {
	if t := l.skipSpaceAndComments(); t != nil {
		return *t
	}
	start := l.position()
	startPos := l.pos
	c := l.peek()
	if c == -1 {
		return token{kind: tokEOF, rng: parser.Range{Start: start, End: start}}
	}
	switch {
	case c == '@':
		l.advance()
		if !isAccountChar(l.peek()) {
			return l.invalid(start, "expected an account name after '@'")
		}
		for {
			for isAccountChar(l.peek()) {
				l.advance()
			}
			if l.peek() != ':' {
				break
			}
			l.advance()
			if !isAccountChar(l.peek()) {
				return l.invalid(start, "expected an account segment after ':'")
			}
		}
		return l.tok(tokAccount, start, startPos)
	case c == '$':
		l.advance()
		for isNameChar(l.peek()) {
			l.advance()
		}
		return l.tok(tokVar, start, startPos)
	case c == '"':
		l.advance()
		for l.peek() != '"' {
			if l.peek() == -1 {
				return l.invalid(start, "unterminated string")
			}
			l.advance()
		}
		l.advance()
		return l.tok(tokString, start, startPos)
	case isDigit(c):
		for isDigit(l.peek()) {
			l.advance()
		}
		for l.peek() == '_' && isDigit(l.peekAt(1)) {
			l.advance()
			for isDigit(l.peek()) {
				l.advance()
			}
		}
		return l.tok(tokInt, start, startPos)
	case isUpper(c):
		for isUpper(l.peek()) || isDigit(l.peek()) {
			l.advance()
		}
		if l.peek() == '/' && isDigit(l.peekAt(1)) {
			l.advance()
			for isDigit(l.peek()) {
				l.advance()
			}
		}
		return l.tok(tokAsset, start, startPos)
	case isLower(c) || c == '_':
		for isNameChar(l.peek()) {
			l.advance()
		}
		return l.tok(tokName, start, startPos)
	}

	l.advance()
	single := func(kind tokKind) token { return l.tok(kind, start, startPos) }
	double := func(kind tokKind) token {
		l.advance()
		return l.tok(kind, start, startPos)
	}
	switch c {
	case '[':
		return single(tokLBracket)
	case '(':
		return single(tokLParen)
	case ')':
		return single(tokRParen)
	case ',':
		return single(tokComma)
	case ':':
		return single(tokColon)
	case '+':
		return single(tokPlus)
	case '-':
		return single(tokMinus)
	case '*':
		return single(tokStar)
	case '/':
		return single(tokSlash)
	case '%':
		return single(tokPercent)
	case '<':
		if l.peek() == '=' && l.peekAt(1) == '>' {
			l.advance()
			return double(tokIff)
		}
		if l.peek() == '=' {
			return double(tokLe)
		}
		return single(tokLt)
	case '>':
		if l.peek() == '=' {
			return double(tokGe)
		}
		return single(tokGt)
	case '=':
		if l.peek() == '>' {
			return double(tokImplies)
		}
		if l.peek() == '=' {
			return double(tokEq)
		}
	case '!':
		if l.peek() == '=' {
			return double(tokNe)
		}
	}
	return l.invalid(start, fmt.Sprintf("unexpected character '%c'", c))
}
