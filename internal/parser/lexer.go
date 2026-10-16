package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type tokenKind uint8

const (
	tkEOF tokenKind = iota

	// Keywords
	tkVars
	tkMax
	tkSource
	tkDestination
	tkSend
	tkFrom
	tkUp
	tkTo
	tkRemaining
	tkAllowing
	tkUnbounded
	tkOverdraft
	tkOneof
	tkKept
	tkSave
	tkWith
	tkScaling
	tkThrough
	tkFeature

	// Punctuation
	tkLParen
	tkRParen
	tkLBracket
	tkRBracket
	tkLBrace
	tkRBrace
	tkComma
	tkEq
	tkStar
	tkPlus
	tkMinus
	tkDiv
	tkRestrict
	tkHashBang
	tkColon
	tkAt

	// Literals
	tkPercentage
	tkString
	tkIdentifier
	tkNumber
	tkAsset
	tkVariable
	tkAccountText
)

var keywords = map[string]tokenKind{
	"vars":        tkVars,
	"max":         tkMax,
	"source":      tkSource,
	"destination": tkDestination,
	"send":        tkSend,
	"from":        tkFrom,
	"up":          tkUp,
	"to":          tkTo,
	"remaining":   tkRemaining,
	"allowing":    tkAllowing,
	"unbounded":   tkUnbounded,
	"overdraft":   tkOverdraft,
	"oneof":       tkOneof,
	"kept":        tkKept,
	"save":        tkSave,
	"with":        tkWith,
	"scaling":     tkScaling,
	"through":     tkThrough,
	"feature":     tkFeature,
}

var punctuation = map[tokenKind]string{
	tkLParen:   "(",
	tkRParen:   ")",
	tkLBracket: "[",
	tkRBracket: "]",
	tkLBrace:   "{",
	tkRBrace:   "}",
	tkComma:    ",",
	tkEq:       "=",
	tkStar:     "*",
	tkPlus:     "+",
	tkMinus:    "-",
	tkDiv:      "/",
	tkRestrict: "\\",
	tkHashBang: "#!",
	tkColon:    ":",
	tkAt:       "@",
}

func (k tokenKind) isKeyword() bool {
	return k >= tkVars && k <= tkFeature
}

// A short human readable description, used in error messages
func (k tokenKind) String() string {
	if k.isKeyword() {
		for text, kind := range keywords {
			if kind == k {
				return "`" + text + "`"
			}
		}
	}
	if text, ok := punctuation[k]; ok {
		return "`" + text + "`"
	}

	switch k {
	case tkEOF:
		return "end of file"
	case tkPercentage:
		return "percentage"
	case tkString:
		return "string"
	case tkIdentifier:
		return "identifier"
	case tkNumber:
		return "number"
	case tkAsset:
		return "asset"
	case tkVariable:
		return "variable"
	case tkAccountText:
		return "account name"
	default:
		return "token"
	}
}

type token struct {
	Kind tokenKind
	Text string
	Range
	// byte offsets in the source
	StartOffset, EndOffset int
}

func (t token) describe() string {
	switch t.Kind {
	case tkEOF:
		return "end of file"
	case tkString:
		return "string " + t.Text
	default:
		return "`" + t.Text + "`"
	}
}

type lexer struct {
	src    string
	offset int
	pos    Position

	// true right after '@' or ':', where account name parts are lexed
	accountMode bool

	tokens   []token
	comments []Comment
	errors   []ParserError
}

func lex(src string) ([]token, []Comment, []ParserError) {
	l := lexer{src: src}
	l.run()
	return l.tokens, l.comments, l.errors
}

func (l *lexer) peekAt(n int) byte {
	if l.offset+n >= len(l.src) {
		return 0
	}
	return l.src[l.offset+n]
}

func (l *lexer) advance() {
	r, size := utf8.DecodeRuneInString(l.src[l.offset:])
	l.offset += size
	if r == '\n' {
		l.pos.Line++
		l.pos.Character = 0
	} else {
		l.pos.Character++
	}
}

func (l *lexer) advanceWhile(pred func(byte) bool) {
	for l.offset < len(l.src) && pred(l.src[l.offset]) {
		l.advance()
	}
}

func (l *lexer) emit(kind tokenKind, startOffset int, startPos Position) {
	l.tokens = append(l.tokens, token{
		Kind:        kind,
		Text:        l.src[startOffset:l.offset],
		Range:       Range{Start: startPos, End: l.pos},
		StartOffset: startOffset,
		EndOffset:   l.offset,
	})
}

func (l *lexer) errorFrom(startPos Position, msg string) {
	l.errors = append(l.errors, ParserError{
		Range: Range{Start: startPos, End: l.pos},
		Msg:   msg,
	})
}

func isLower(c byte) bool        { return c >= 'a' && c <= 'z' }
func isUpper(c byte) bool        { return c >= 'A' && c <= 'Z' }
func isDigit(c byte) bool        { return c >= '0' && c <= '9' }
func isIdentTail(c byte) bool    { return isLower(c) || c == '_' }
func isVarNameStart(c byte) bool { return isLower(c) || c == '_' }
func isVarNameTail(c byte) bool  { return isLower(c) || isDigit(c) || c == '_' }
func isAssetTail(c byte) bool    { return isUpper(c) || isDigit(c) }
func isAccountText(c byte) bool {
	return isLower(c) || isUpper(c) || isDigit(c) || c == '_' || c == '-'
}
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

// Characters that cannot start any token. Consecutive invalid chars are reported as a single error.
func (l *lexer) isInvalidStart() bool {
	c := l.src[l.offset]
	if isWhitespace(c) || isLower(c) || isUpper(c) || isDigit(c) {
		return false
	}
	switch c {
	case '(', ')', '[', ']', '{', '}', ',', '=', '*', '+', '-', '/', '\\', ':', '@', '"', '$':
		return false
	case '#':
		return l.peekAt(1) != '!'
	default:
		return true
	}
}

func (l *lexer) run() {
	for {
		if l.accountMode {
			l.accountMode = false
			l.lexAccountPart()
		}

		l.advanceWhile(isWhitespace)

		startOffset, startPos := l.offset, l.pos
		if l.offset >= len(l.src) {
			l.emit(tkEOF, startOffset, startPos)
			return
		}

		c := l.src[l.offset]
		switch {
		case c == '/' && l.peekAt(1) == '/':
			l.advanceWhile(func(c byte) bool { return c != '\n' && c != '\r' })
			l.comments = append(l.comments, Comment{
				Content: l.src[startOffset+2 : l.offset],
				Range:   Range{Start: startPos, End: l.pos},
			})

		case c == '/' && l.peekAt(1) == '*':
			l.lexMultilineComment(startPos)

		case isLower(c):
			l.advance()
			l.advanceWhile(isIdentTail)
			kind, ok := keywords[l.src[startOffset:l.offset]]
			if !ok {
				kind = tkIdentifier
			}
			l.emit(kind, startOffset, startPos)

		case isUpper(c):
			l.advance()
			l.advanceWhile(isAssetTail)
			if l.peekAt(0) == '/' && isDigit(l.peekAt(1)) {
				l.advance()
				l.advanceWhile(isDigit)
			}
			l.emit(tkAsset, startOffset, startPos)

		case isDigit(c):
			l.lexNumber(startOffset, startPos)

		case c == '"':
			l.lexString(startOffset, startPos)

		case c == '$':
			l.lexVariable(startOffset, startPos)

		case c == '@':
			l.advance()
			l.emit(tkAt, startOffset, startPos)
			l.accountMode = true

		case c == ':':
			l.advance()
			l.emit(tkColon, startOffset, startPos)
			l.accountMode = true

		case c == '#' && l.peekAt(1) == '!':
			l.advance()
			l.advance()
			l.emit(tkHashBang, startOffset, startPos)

		default:
			kind, ok := singleCharTokens[c]
			if ok {
				l.advance()
				l.emit(kind, startOffset, startPos)
				continue
			}

			for l.offset < len(l.src) && l.isInvalidStart() {
				l.advance()
			}
			l.errorFrom(startPos, fmt.Sprintf("invalid character: `%s`", l.src[startOffset:l.offset]))
		}
	}
}

var singleCharTokens = map[byte]tokenKind{
	'(':  tkLParen,
	')':  tkRParen,
	'[':  tkLBracket,
	']':  tkRBracket,
	'{':  tkLBrace,
	'}':  tkRBrace,
	',':  tkComma,
	'=':  tkEq,
	'*':  tkStar,
	'+':  tkPlus,
	'-':  tkMinus,
	'/':  tkDiv,
	'\\': tkRestrict,
}

// Lexes the part of an account literal right after '@' or ':'.
// No whitespace is allowed there: when no part is found, no token is emitted
// and the parser reports the missing part.
func (l *lexer) lexAccountPart() {
	if l.offset >= len(l.src) {
		return
	}

	startOffset, startPos := l.offset, l.pos
	c := l.src[l.offset]
	switch {
	case c == '$' && isVarNameStart(l.peekAt(1)):
		l.lexVariable(startOffset, startPos)
	case isAccountText(c):
		l.advanceWhile(isAccountText)
		l.emit(tkAccountText, startOffset, startPos)
	}
}

func (l *lexer) lexVariable(startOffset int, startPos Position) {
	l.advance() // '$'
	if !isVarNameStart(l.peekAt(0)) {
		// Emit an (invalid) variable anyway, so that the parser knows a variable is being typed
		l.emit(tkVariable, startOffset, startPos)
		l.errorFrom(startPos, "invalid variable name: expected a lowercase letter or `_` after `$`")
		return
	}
	l.advanceWhile(isVarNameTail)
	l.emit(tkVariable, startOffset, startPos)
}

func (l *lexer) lexNumber(startOffset int, startPos Position) {
	l.advanceWhile(isDigit)

	// percentage: [0-9]+ ('.' [0-9]+)? '%'
	if l.peekAt(0) == '%' {
		l.advance()
		l.emit(tkPercentage, startOffset, startPos)
		return
	}
	if l.peekAt(0) == '.' && isDigit(l.peekAt(1)) {
		i := 1
		for isDigit(l.peekAt(i)) {
			i++
		}
		if l.peekAt(i) == '%' {
			for range i + 1 {
				l.advance()
			}
			l.emit(tkPercentage, startOffset, startPos)
			return
		}
	}

	// number: [0-9]+ ('_' [0-9]+)*
	for l.peekAt(0) == '_' && isDigit(l.peekAt(1)) {
		l.advance()
		l.advanceWhile(isDigit)
	}
	l.emit(tkNumber, startOffset, startPos)
}

// STRING: '"' ('\\"' | ~[\r\n"])* '"'
//
// We emulate the longest-match semantics: an escaped quote may also be the closing one.
func (l *lexer) lexString(startOffset int, startPos Position) {
	end := -1 // offset right after the closing quote
	i := l.offset + 1
	for i < len(l.src) {
		c := l.src[i]
		if c == '\n' || c == '\r' {
			break
		}
		if c == '"' {
			end = i + 1
			if l.src[i-1] != '\\' || i-1 == l.offset {
				break
			}
		}
		i++
	}

	if end == -1 {
		// Unterminated string: take the rest of the line, so that the following lines are not affected
		for l.offset < i {
			l.advance()
		}
		l.emit(tkString, startOffset, startPos)
		l.errorFrom(startPos, "unterminated string literal")
		return
	}

	for l.offset < end {
		l.advance()
	}
	l.emit(tkString, startOffset, startPos)
}

// MULTILINE_COMMENT: '/*' (MULTILINE_COMMENT | .)*? '*/'
//
// A nested '/*' opens a nested comment when that leads to a terminated comment,
// otherwise it's treated as plain text (this mimics the ANTLR lexer).
func (l *lexer) lexMultilineComment(startPos Position) {
	memo := map[int]int{}

	// returns the offset right after the '*/' closing the comment whose content starts at i, or -1
	var matchFrom func(i int) int
	matchFrom = func(start int) int {
		if end, ok := memo[start]; ok {
			return end
		}
		i := start

		end := -1
		for i < len(l.src) {
			if strings.HasPrefix(l.src[i:], "*/") {
				end = i + 2
				break
			}
			if strings.HasPrefix(l.src[i:], "/*") {
				if nestedEnd := matchFrom(i + 2); nestedEnd != -1 {
					if outerEnd := matchFrom(nestedEnd); outerEnd != -1 {
						end = outerEnd
						break
					}
				}
			}
			i++
		}

		memo[start] = end
		return end
	}

	end := matchFrom(l.offset + 2)
	if end == -1 {
		end = len(l.src)
		defer l.errorFrom(startPos, "unterminated comment")
	}
	for l.offset < end {
		l.advance()
	}
}
