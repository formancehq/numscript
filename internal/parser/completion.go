package parser

import "unicode/utf8"

// A syntactic position where the parser was expecting something when it reached the cursor
type CompletionSlot interface{ completionSlot() }

// An expression is expected. Types contains the types the expression is expected to have,
// when they are known from the syntactic context (e.g. "account" for a source).
type ExprSlot struct{ Types []string }

// The type of a variable declaration is expected
type VarTypeSlot struct{}

// A statement is expected
type StatementSlot struct{}

// A part of an account literal (e.g. `@users:$id`) is expected
type AccountPartSlot struct{ LiteralStart Position }

func (*ExprSlot) completionSlot()        {}
func (*VarTypeSlot) completionSlot()     {}
func (*StatementSlot) completionSlot()   {}
func (*AccountPartSlot) completionSlot() {}

// A function call enclosing the cursor
type FnCallFrame struct {
	Name string
	// The index of the argument the cursor is on
	ArgIndex int
}

// What the parser expected at the cursor position
type CompletionContext struct {
	// The keywords that are valid at the cursor
	Keywords []string

	Slots []CompletionSlot

	// The function calls enclosing the cursor (innermost last)
	FnCalls []FnCallFrame

	// The range of the word under the cursor, that a completion would replace.
	// It's an empty range at the cursor when no word is being typed.
	Range Range

	// The part of the word under the cursor that precedes the cursor
	Prefix string
}

// Tokens that can be extended by typing more characters: when the cursor is right after
// one of them, that token is the one being completed
func (k tokenKind) isWordLike() bool {
	switch k {
	case tkIdentifier, tkVariable, tkAccountText, tkAsset, tkNumber:
		return true
	default:
		return k.isKeyword()
	}
}

// Parses the input and returns what the parser expected at the cursor position.
//
// Returns a nil context when the cursor is inside a comment or a string literal.
func ParseForCompletion(input string, cursor Position) (ParseResult, *CompletionContext) {
	tokens, comments, lexErrors := lex(input)

	for _, comment := range comments {
		if posLt(comment.Start, cursor) && !posLt(comment.End, cursor) {
			return Parse(input), nil
		}
	}

	cursorTok := -1
	completionRange := Range{Start: cursor, End: cursor}
	prefix := ""

	for index, tk := range tokens {
		if tk.Kind == tkEOF || !posLt(tk.Start, cursor) {
			cursorTok = index
			break
		}

		containsCursor := posLt(cursor, tk.End)
		touchesCursor := tk.End == cursor && tk.Kind.isWordLike()
		if !containsCursor && !touchesCursor {
			continue
		}

		if tk.Kind == tkString {
			return Parse(input), nil
		}

		cursorTok = index
		completionRange = tk.Range
		// tokens never span multiple lines
		prefix = runePrefix(tk.Text, cursor.Character-tk.Start.Character)
		break
	}

	p := newParser(tokens, cursorTok)
	result := p.finish(input, comments, lexErrors)

	ctx := p.completion
	ctx.Range = completionRange
	ctx.Prefix = prefix
	return result, &ctx
}

func posLt(a, b Position) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Character < b.Character
}

func runePrefix(s string, n int) string {
	offset := 0
	for range n {
		if offset >= len(s) {
			break
		}
		_, size := utf8.DecodeRuneInString(s[offset:])
		offset += size
	}
	return s[:offset]
}
