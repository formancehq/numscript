package parser

import (
	"fmt"
	"math/big"
	"slices"
	"strings"
)

// A hand-written, error-tolerant recursive descent parser.
//
// The parser never gives up: when an expected token is missing, it reports an
// error and keeps going without consuming anything; when an unexpected token
// is found, it skips tokens until it reaches a token that some enclosing
// construct knows how to handle (see the "stop" token sets).
// This way, the AST of an incomplete program (e.g. while it's being typed)
// keeps as much information as possible.
//
// Missing nodes are represented as nil.

type ParserError struct {
	Range
	Msg string
}

func (e ParserError) Error() string {
	return e.Msg
}

type ParseResult struct {
	Source string
	Value  Program
	Errors []ParserError
}

func Parse(input string) ParseResult {
	tokens, comments, lexErrors := lex(input)
	p := newParser(tokens, -1)
	return p.finish(input, comments, lexErrors)
}

func ParseErrorsToString(errors []ParserError, source string) string {
	buf := "Got errors while parsing:\n"
	for _, err := range errors {
		buf += err.Msg + "\n" + err.ShowOnSource(source) + "\n"
	}
	return buf
}

// Returns (a, b) representing a/(10^b)
func ParsePercentageRatio(source string) (*big.Int, int, error) {
	source = strings.TrimSuffix(source, "%")

	scale := 0
	if i := strings.Index(source, "."); i != -1 {
		scale = len(source) - i - 1
	}

	intPart := strings.ReplaceAll(source, ".", "")
	num, ok := new(big.Int).SetString(intPart, 10)
	if !ok {
		return nil, 0, fmt.Errorf("unexpected invalid string literal: %s", source)
	}

	return num, scale, nil
}

type tokenSet uint64

func setOf(kinds ...tokenKind) tokenSet {
	var s tokenSet
	for _, k := range kinds {
		s |= 1 << k
	}
	return s
}

func (s tokenSet) has(k tokenKind) bool {
	return s&(1<<k) != 0
}

var (
	// Tokens that only appear at the beginning of a statement (or EOF):
	// we never skip past them while recovering from errors.
	stmtStop = setOf(tkSend, tkSave, tkVars, tkEOF)

	exprStart = setOf(
		tkVariable, tkAsset, tkString, tkAt, tkNumber, tkPercentage,
		tkLBracket, tkMinus, tkLParen, tkIdentifier, tkOverdraft,
	)

	// Tokens after which a '-' is a prefix operator rather than an infix one
	prefixPosition = setOf(tkPlus, tkMinus, tkDiv, tkLParen, tkLBracket, tkComma)
)

var (
	hintAccount  = []string{"account"}
	hintMonetary = []string{"monetary"}
	hintPortion  = []string{"portion"}
	hintAsset    = []string{"asset"}
	hintNumber   = []string{"number"}
	hintString   = []string{"string"}
)

type parser struct {
	tokens []token
	pos    int
	errors []ParserError

	// index of the token where the last error was reported. Used to avoid cascading errors
	lastErrorTok int

	// end of the last consumed token
	prevEnd Position

	// when >= 0, infix operators at or after this token index are not consumed (used for backtracking)
	stopAt int

	// --- completion

	// index of the token under the cursor, or -1 when not computing completions
	cursorTok  int
	frozen     bool
	recorded   bool
	completion CompletionContext
	fnFrames   []FnCallFrame
}

func newParser(tokens []token, cursorTok int) *parser {
	return &parser{
		tokens:       tokens,
		lastErrorTok: -1,
		stopAt:       -1,
		cursorTok:    cursorTok,
	}
}

func (p *parser) finish(input string, comments []Comment, lexErrors []ParserError) ParseResult {
	program := p.parseProgram()
	program.Comments = comments

	errors := append(lexErrors, p.errors...)
	slices.SortStableFunc(errors, func(a, b ParserError) int {
		if a.Start.Line != b.Start.Line {
			return a.Start.Line - b.Start.Line
		}
		return a.Start.Character - b.Start.Character
	})

	return ParseResult{
		Source: input,
		Value:  program,
		Errors: errors,
	}
}

// --- Token stream utilities

func (p *parser) cur() token {
	return p.tokens[p.pos]
}

func (p *parser) peek(n int) token {
	i := min(p.pos+n, len(p.tokens)-1)
	return p.tokens[i]
}

func (p *parser) atCursor() bool {
	return p.pos == p.cursorTok && !p.frozen
}

// Checks the current token kind, and records it as a possible completion
func (p *parser) at(kind tokenKind) bool {
	if p.atCursor() {
		p.recordCompletion()
	}
	if p.atCursor() && kind.isKeyword() {
		for text, k := range keywords {
			if k == kind && !slices.Contains(p.completion.Keywords, text) {
				p.completion.Keywords = append(p.completion.Keywords, text)
			}
		}
	}
	return p.cur().Kind == kind
}

func (p *parser) atExprStart(hints []string) bool {
	p.recordSlot(&ExprSlot{Types: hints})
	return exprStart.has(p.cur().Kind)
}

func (p *parser) recordSlot(slot CompletionSlot) {
	if p.atCursor() {
		p.recordCompletion()
		p.completion.Slots = append(p.completion.Slots, slot)
	}
}

func (p *parser) recordCompletion() {
	if !p.recorded {
		p.recorded = true
		p.completion.FnCalls = slices.Clone(p.fnFrames)
	}
}

func (p *parser) bump() token {
	t := p.cur()
	if t.Kind == tkEOF {
		return t
	}
	if p.cursorTok >= 0 && p.pos >= p.cursorTok {
		if p.atCursor() {
			p.recordCompletion()
		}
		p.frozen = true
	}
	p.pos++
	p.prevEnd = t.End
	return t
}

func (p *parser) eat(kind tokenKind) bool {
	if p.at(kind) {
		p.bump()
		return true
	}
	return false
}

func (p *parser) expect(kind tokenKind) bool {
	if p.eat(kind) {
		return true
	}
	p.errExpected(kind.String())
	return false
}

// Like expect, but when the token is missing it skips tokens until either the
// expected token or a token in the stop set is found
func (p *parser) expectRecover(kind tokenKind, stop tokenSet) bool {
	if p.eat(kind) {
		return true
	}
	p.errExpected(kind.String())
	p.skipUntil(stop | setOf(kind))
	return p.eat(kind)
}

// Skip tokens until one in the given set is found. Balanced groups are skipped as a whole.
func (p *parser) skipUntil(stop tokenSet) {
	depth := 0
	for {
		kind := p.cur().Kind
		if kind == tkEOF || (depth == 0 && stop.has(kind)) || stmtStop.has(kind) {
			return
		}
		switch kind {
		case tkLParen, tkLBrace, tkLBracket:
			depth++
		case tkRParen, tkRBrace, tkRBracket:
			if depth > 0 {
				depth--
			}
		}
		p.bump()
	}
}

func (p *parser) sameLine() bool {
	return p.cur().Start.Line == p.prevEnd.Line
}

func (p *parser) rangeFrom(start Position) Range {
	end := p.prevEnd
	if end.Line < start.Line || (end.Line == start.Line && end.Character < start.Character) {
		end = start
	}
	return Range{Start: start, End: end}
}

// --- Errors

func (p *parser) report(rng Range, msg string) {
	if p.cursorTok >= 0 && p.pos >= p.cursorTok {
		p.frozen = true
	}
	if p.lastErrorTok == p.pos {
		return
	}
	p.lastErrorTok = p.pos

	if rng.Start == rng.End {
		rng.End.Character++
	}
	p.errors = append(p.errors, ParserError{Range: rng, Msg: msg})
}

func (p *parser) errExpected(what string) {
	cur := p.cur()
	msg := fmt.Sprintf("expected %s, found %s", what, cur.describe())

	// When the unexpected token is on a different line, the error is likely
	// about something missing at the end of the previous line (unless the token
	// is an identifier, which is likely a typo)
	rng := cur.Range
	if p.pos > 0 && !p.sameLine() && cur.Kind != tkIdentifier {
		rng = Range{Start: p.prevEnd, End: p.prevEnd}
	}
	p.report(rng, msg)
}

// Reports an error on the current token, which is about to be skipped
func (p *parser) errUnexpected(what string) {
	cur := p.cur()
	p.report(cur.Range, fmt.Sprintf("expected %s, found %s", what, cur.describe()))
}

// --- Backtracking

type checkpoint struct {
	pos          int
	nErrors      int
	lastErrorTok int
	prevEnd      Position
	frozen       bool
	recorded     bool
	completion   CompletionContext
}

func (p *parser) checkpoint() checkpoint {
	return checkpoint{
		pos:          p.pos,
		nErrors:      len(p.errors),
		lastErrorTok: p.lastErrorTok,
		prevEnd:      p.prevEnd,
		frozen:       p.frozen,
		recorded:     p.recorded,
		completion:   p.completion,
	}
}

func (p *parser) restore(c checkpoint) {
	p.pos = c.pos
	p.errors = p.errors[:c.nErrors]
	p.lastErrorTok = c.lastErrorTok
	p.prevEnd = c.prevEnd
	p.frozen = c.frozen
	p.recorded = c.recorded
	p.completion = c.completion
}

// --- Program

func (p *parser) parseProgram() Program {
	var program Program

	if p.at(tkHashBang) {
		program.Flags = p.parseFeatureDecl()
	}

	if p.at(tkVars) {
		program.Vars = p.parseVars()
	}

	for {
		p.recordSlot(&StatementSlot{})
		isSend := p.at(tkSend)
		isSave := p.at(tkSave)
		isEOF := p.at(tkEOF)

		switch {
		case isSend:
			program.Statements = append(program.Statements, p.parseSend())

		case isSave:
			program.Statements = append(program.Statements, p.parseSave())

		case p.atFnCall():
			program.Statements = append(program.Statements, p.parseFnCall())

		case isEOF:
			return program

		default:
			if p.cur().Kind == tkVars {
				p.report(p.cur().Range, "the vars block must be declared before any statement")
			} else {
				p.errUnexpected("a statement")
			}
			p.bump()
			p.skipUntilStatement()
		}
	}
}

func (p *parser) atFnCall() bool {
	kind := p.cur().Kind
	return (kind == tkIdentifier || kind == tkOverdraft) && p.peek(1).Kind == tkLParen
}

func (p *parser) skipUntilStatement() {
	for !stmtStop.has(p.cur().Kind) && !p.atFnCall() {
		p.bump()
	}
}

// '#!' '[' 'feature' '(' (STRING (',' STRING)*)? ')' ']'
func (p *parser) parseFeatureDecl() []StringLiteral {
	p.bump()

	var flags []StringLiteral
	if !p.expect(tkLBracket) || !p.expect(tkFeature) || !p.expect(tkLParen) {
		p.skipUntil(setOf(tkRBracket))
		p.eat(tkRBracket)
		return flags
	}

	if p.cur().Kind == tkString {
		for {
			if !p.at(tkString) {
				p.errExpected("a string")
				break
			}
			flags = append(flags, *parseString(p.bump()))
			if !p.eat(tkComma) {
				break
			}
		}
	}

	p.expect(tkRParen)
	p.expectRecover(tkRBracket, stmtStop)
	return flags
}

// 'vars' '{' varDeclaration* '}'
func (p *parser) parseVars() *VarDeclarations {
	start := p.bump().Start
	vars := &VarDeclarations{}

	if p.expect(tkLBrace) {
		for {
			p.recordSlot(&VarTypeSlot{})
			if p.at(tkIdentifier) && p.peek(1).Kind != tkLParen {
				vars.Declarations = append(vars.Declarations, p.parseVarDeclaration())
				continue
			}

			if p.eat(tkRBrace) {
				break
			}

			if stmtStop.has(p.cur().Kind) || p.atFnCall() {
				p.errExpected("`}`")
				break
			}

			p.errUnexpected("a variable declaration or `}`")
			p.bump()
		}
	}

	vars.Range = p.rangeFrom(start)
	return vars
}

// IDENTIFIER VARIABLE_NAME ('=' valueExpr)?
func (p *parser) parseVarDeclaration() VarDeclaration {
	typeTk := p.bump()
	decl := VarDeclaration{
		Type: &TypeDecl{Range: typeTk.Range, Name: typeTk.Text},
	}

	if p.at(tkVariable) {
		decl.Name = parseVariable(p.bump())
	} else {
		p.errExpected("a variable name")
	}

	if p.eat(tkEq) {
		var hints []string
		if typeTk.Text != "" {
			hints = []string{typeTk.Text}
		}

		origin := p.parseExpr(hints)
		if origin != nil {
			decl.Origin = &origin
		}
	}

	decl.Range = p.rangeFrom(typeTk.Start)
	return decl
}

// --- Statements

// 'send' sentValue '(' 'source' '=' source 'destination' '=' destination ')'
func (p *parser) parseSend() *SendStatement {
	start := p.bump().Start
	stmt := &SendStatement{
		SentValue: p.parseSentValue(),
	}

	p.expectRecover(tkLParen, setOf(tkSource, tkDestination, tkRParen))

	if p.expectKeywordClause(tkSource) {
		stmt.Source = p.parseSource(setOf(tkDestination, tkRParen), 0)
	}

	if !p.at(tkDestination) {
		p.errExpected(tkDestination.String())
		p.skipUntil(setOf(tkDestination, tkRParen) | p.misspelledKeywordSet())
	}

	if p.expectKeywordClause(tkDestination) {
		stmt.Destination = p.parseDestination(setOf(tkRParen), 0)
	}

	p.expectRecover(tkRParen, 0)

	stmt.Range = p.rangeFrom(start)
	return stmt
}

func (p *parser) misspelledKeywordSet() tokenSet {
	if p.cur().Kind == tkIdentifier && p.peek(1).Kind == tkEq {
		return setOf(tkIdentifier)
	}
	return 0
}

// Parses `<keyword> =`. A misspelled keyword (an identifier followed by '=') is
// reported but accepted, so that the rest of the clause can still be parsed.
func (p *parser) expectKeywordClause(keyword tokenKind) bool {
	if !p.at(keyword) {
		if p.cur().Kind == tkIdentifier && p.peek(1).Kind == tkEq {
			p.errExpected(keyword.String())
			p.bump()
		} else if p.cur().Kind == tkEq {
			p.errExpected(keyword.String())
		} else {
			p.errExpected(keyword.String())
			return false
		}
	} else {
		p.bump()
	}

	p.expect(tkEq)
	return true
}

// 'save' sentValue 'from' valueExpr
func (p *parser) parseSave() *SaveStatement {
	start := p.bump().Start
	stmt := &SaveStatement{
		SentValue: p.parseSentValue(),
	}

	if p.expect(tkFrom) || (p.sameLine() && exprStart.has(p.cur().Kind)) {
		stmt.Account = p.parseExpr(hintAccount)
	}

	stmt.Range = p.rangeFrom(start)
	return stmt
}

// valueExpr | '[' valueExpr '*' ']'
func (p *parser) parseSentValue() SentValue {
	isBracket := p.at(tkLBracket)
	isExpr := p.atExprStart(hintMonetary)

	if isBracket {
		start := p.cur().Start
		lit, sendAll := p.parseBracketLiteral(true)
		if sendAll != nil {
			return sendAll
		}
		expr := p.parseBinaryRest(start, lit, 0, hintMonetary)
		return &SentValueLiteral{Range: expr.GetRange(), Monetary: expr}
	}

	if isExpr {
		expr := p.parseExpr(hintMonetary)
		if expr == nil {
			return nil
		}
		return &SentValueLiteral{Range: expr.GetRange(), Monetary: expr}
	}

	p.errExpected("a monetary value")
	return nil
}

// --- Sources

func (p *parser) atSourceStart() bool {
	isBrace := p.at(tkLBrace)
	isOneof := p.at(tkOneof)
	isMax := p.at(tkMax)
	isExpr := p.atExprStart(hintAccount)
	return isBrace || isOneof || isMax || isExpr
}

func (p *parser) parseSource(stop tokenSet, follow tokenSet) Source {
	start := p.cur().Start

	isBrace := p.at(tkLBrace)
	isOneof := p.at(tkOneof)
	isMax := p.at(tkMax)
	isExpr := p.atExprStart(hintAccount)

	switch {
	case isBrace:
		return p.parseBraceSource(stop)

	case isOneof:
		p.bump()
		src := &SourceOneof{}
		if p.expect(tkLBrace) {
			if p.at(tkRBrace) {
				p.errExpected("a source")
			}
			src.Sources = p.parseInorderSources(nil, stop)
		}
		src.Range = p.rangeFrom(start)
		return src

	case isMax:
		p.bump()
		src := &SourceCapped{
			Cap: p.parseExpr(hintMonetary),
		}
		if p.expect(tkFrom) || p.sameLine() {
			if p.atSourceStart() {
				src.From = p.parseSource(stop, follow)
			} else {
				p.errExpected("a source")
			}
		}
		src.Range = p.rangeFrom(start)
		return src

	case isExpr:
		addr := p.parseExprBeforeAllotment(hintAccount, follow)
		return p.parseSourceAccountSuffix(start, addr, follow)

	default:
		p.errExpected("a source")
		return nil
	}
}

// Parses what follows the address of a source account:
//
//	('\' valueExpr)? ('allowing' 'unbounded' 'overdraft' | 'allowing' 'overdraft' 'up' 'to' valueExpr)?
//	| 'with' 'scaling' 'through' valueExpr
func (p *parser) parseSourceAccountSuffix(start Position, addr ValueExpr, follow tokenSet) Source {
	var color ValueExpr
	if p.eat(tkRestrict) {
		color = p.parseExprBeforeAllotment(hintString, follow)
	}

	if p.eat(tkAllowing) {
		if p.eat(tkUnbounded) {
			p.expect(tkOverdraft)
			return &SourceOverdraft{
				Range:   p.rangeFrom(start),
				Color:   color,
				Address: addr,
			}
		}

		if p.eat(tkOverdraft) {
			p.expect(tkUp)
			p.expect(tkTo)
			bound := p.parseExprBeforeAllotment(hintMonetary, follow)
			return &SourceOverdraft{
				Range:   p.rangeFrom(start),
				Color:   color,
				Address: addr,
				Bounded: &bound,
			}
		}

		p.errExpected("`unbounded` or `overdraft`")
	} else if p.at(tkWith) {
		withTk := p.bump()
		if color != nil {
			p.report(withTk.Range, "a color restriction cannot be used together with scaling")
		}
		p.expect(tkScaling)
		p.expect(tkThrough)
		through := p.parseExprBeforeAllotment(hintAccount, follow)
		return &SourceWithScaling{
			Range:   p.rangeFrom(start),
			Address: addr,
			Through: through,
		}
	}

	if addr == nil && color == nil {
		return nil
	}

	return &SourceAccount{
		Range:     p.rangeFrom(start),
		Color:     color,
		ValueExpr: addr,
	}
}

// Parses either an allotment or an inorder source, starting from the '{'
func (p *parser) parseBraceSource(stop tokenSet) Source {
	start := p.bump().Start

	isRemaining := p.at(tkRemaining)
	isExpr := p.atExprStart([]string{"portion", "account"})

	if isRemaining {
		return p.parseSourceAllotment(start, nil, Position{}, stop)
	}

	if isExpr {
		exprStart := p.cur().Start
		expr := p.parseExpr([]string{"portion", "account"})
		if p.at(tkFrom) {
			return p.parseSourceAllotment(start, expr, exprStart, stop)
		}

		var sources []Source
		first := p.parseSourceAccountSuffix(exprStart, expr, 0)
		if first != nil {
			sources = append(sources, first)
		}
		return &SourceInorder{
			Sources: p.parseInorderSources(sources, stop),
			Range:   p.rangeFrom(start),
		}
	}

	return &SourceInorder{
		Sources: p.parseInorderSources(nil, stop),
		Range:   p.rangeFrom(start),
	}
}

// source* '}'
func (p *parser) parseInorderSources(sources []Source, stop tokenSet) []Source {
	for {
		if p.atSourceStart() {
			src := p.parseSource(stop|setOf(tkRBrace), 0)
			if src != nil {
				sources = append(sources, src)
			}
			continue
		}

		if p.eat(tkRBrace) {
			return sources
		}

		if stop.has(p.cur().Kind) || stmtStop.has(p.cur().Kind) {
			p.errExpected("`}`")
			return sources
		}

		p.errUnexpected("a source or `}`")
		p.bump()
	}
}

// (allotment 'from' source)+ '}'
func (p *parser) parseSourceAllotment(start Position, first ValueExpr, firstStart Position, stop tokenSet) *SourceAllotment {
	src := &SourceAllotment{}
	inner := stop | setOf(tkRBrace)

	parseItem := func(itemStart Position, allotment AllotmentValue) {
		item := SourceAllotmentItem{Allotment: allotment}
		if p.expect(tkFrom) || p.sameLine() {
			if p.atSourceStart() {
				item.From = p.parseSource(inner, setOf(tkFrom))
			} else {
				p.errExpected("a source")
			}
		}
		item.Range = p.rangeFrom(itemStart)
		src.Items = append(src.Items, item)
	}

	if first != nil {
		parseItem(firstStart, &ValueExprAllotment{Value: first})
	}

	for {
		isRemaining := p.at(tkRemaining)
		isExpr := p.atExprStart(hintPortion)
		if isRemaining || isExpr {
			itemStart := p.cur().Start
			parseItem(itemStart, p.parseAllotment())
			continue
		}

		if p.eat(tkRBrace) {
			break
		}

		if stop.has(p.cur().Kind) || stmtStop.has(p.cur().Kind) {
			p.errExpected("`}`")
			break
		}

		p.errUnexpected("an allotment or `}`")
		p.bump()
	}

	src.Range = p.rangeFrom(start)
	return src
}

// Pre: at 'remaining' or at the beginning of an expression
func (p *parser) parseAllotment() AllotmentValue {
	if p.at(tkRemaining) {
		return &RemainingAllotment{Range: p.bump().Range}
	}

	expr := p.parseExpr(hintPortion)
	if expr == nil {
		return nil
	}
	return &ValueExprAllotment{Value: expr}
}

// --- Destinations

func (p *parser) parseDestination(stop tokenSet, follow tokenSet) Destination {
	start := p.cur().Start

	isBrace := p.at(tkLBrace)
	isOneof := p.at(tkOneof)
	isExpr := p.atExprStart(hintAccount)

	switch {
	case isBrace:
		p.bump()
		isMax := p.at(tkMax)
		isRemaining := p.at(tkRemaining)
		isExpr := p.atExprStart(hintPortion)

		if isMax || (!isRemaining && !isExpr) {
			clauses, remaining := p.parseInorderClauses(stop)
			return &DestinationInorder{
				Clauses:   clauses,
				Remaining: remaining,
				Range:     p.rangeFrom(start),
			}
		}

		return p.parseDestinationAllotment(start, stop)

	case isOneof:
		p.bump()
		dest := &DestinationOneof{}
		if p.expect(tkLBrace) {
			dest.Clauses, dest.Remaining = p.parseInorderClauses(stop)
		}
		dest.Range = p.rangeFrom(start)
		return dest

	case isExpr:
		expr := p.parseExprBeforeAllotment(hintAccount, follow)
		if expr == nil {
			return nil
		}
		return &DestinationAccount{ValueExpr: expr}

	default:
		p.errExpected("a destination")
		return nil
	}
}

// ('max' valueExpr keptOrDestination)* 'remaining' keptOrDestination '}'
func (p *parser) parseInorderClauses(stop tokenSet) ([]CappedKeptOrDestination, KeptOrDestination) {
	var clauses []CappedKeptOrDestination
	inner := stop | setOf(tkRBrace, tkMax, tkRemaining)

	for {
		if p.at(tkMax) {
			start := p.bump().Start
			clause := CappedKeptOrDestination{
				Cap: p.parseExpr(hintMonetary),
			}
			clause.To = p.parseKeptOrDestination(inner, 0)
			clause.Range = p.rangeFrom(start)
			clauses = append(clauses, clause)
			continue
		}

		if p.eat(tkRemaining) {
			remaining := p.parseKeptOrDestination(stop|setOf(tkRBrace), 0)
			p.expectRecover(tkRBrace, stop)
			return clauses, remaining
		}

		if p.at(tkRBrace) {
			p.errExpected("`max` or `remaining`")
			p.bump()
			return clauses, nil
		}

		if stop.has(p.cur().Kind) || stmtStop.has(p.cur().Kind) {
			p.errExpected("`remaining`")
			return clauses, nil
		}

		p.errUnexpected("`max` or `remaining`")
		p.bump()
	}
}

// (allotment keptOrDestination)+ '}'
func (p *parser) parseDestinationAllotment(start Position, stop tokenSet) *DestinationAllotment {
	dest := &DestinationAllotment{}
	inner := stop | setOf(tkRBrace)

	for {
		isRemaining := p.at(tkRemaining)
		isExpr := p.atExprStart(hintPortion)
		if isRemaining || isExpr {
			itemStart := p.cur().Start
			item := DestinationAllotmentItem{
				Allotment: p.parseAllotment(),
			}
			item.To = p.parseKeptOrDestination(inner, setOf(tkTo, tkKept))
			item.Range = p.rangeFrom(itemStart)
			dest.Items = append(dest.Items, item)
			continue
		}

		if p.eat(tkRBrace) {
			break
		}

		if stop.has(p.cur().Kind) || stmtStop.has(p.cur().Kind) {
			p.errExpected("`}`")
			break
		}

		p.errUnexpected("an allotment or `}`")
		p.bump()
	}

	dest.Range = p.rangeFrom(start)
	return dest
}

// 'to' destination | 'kept'
func (p *parser) parseKeptOrDestination(stop tokenSet, follow tokenSet) KeptOrDestination {
	isTo := p.at(tkTo)
	isKept := p.at(tkKept)

	switch {
	case isTo:
		p.bump()
		return &DestinationTo{
			Destination: p.parseDestination(stop, follow),
		}

	case isKept:
		return &DestinationKept{Range: p.bump().Range}

	default:
		p.errExpected("`to` or `kept`")
		return nil
	}
}

// --- Expressions

func infixPrecedence(kind tokenKind) int {
	switch kind {
	case tkPlus, tkMinus:
		return 1
	case tkDiv:
		return 2
	default:
		return 0
	}
}

// Parses an expression, reporting an error (and returning nil) when no expression is found
func (p *parser) parseExpr(hints []string) ValueExpr {
	return p.parseBinary(1, hints)
}

func (p *parser) parseBinary(minPrecedence int, hints []string) ValueExpr {
	if !p.atExprStart(hints) {
		p.errExpected("an expression")
		return nil
	}

	start := p.cur().Start
	left := p.parseUnary(hints)
	return p.parseBinaryRest(start, left, minPrecedence, hints)
}

// Precedence climbing: all the operators are left-associative
func (p *parser) parseBinaryRest(start Position, left ValueExpr, minPrecedence int, hints []string) ValueExpr {
	for {
		if p.stopAt >= 0 && p.pos >= p.stopAt {
			return left
		}

		opTk := p.cur()
		precedence := infixPrecedence(opTk.Kind)
		if precedence == 0 || precedence < minPrecedence {
			return left
		}
		p.bump()

		rightHints := hints
		if opTk.Kind == tkDiv {
			rightHints = hintNumber
		}
		right := p.parseBinary(precedence+1, rightHints)

		left = &BinaryInfix{
			Range:    p.rangeFrom(start),
			Operator: InfixOperator(opTk.Text),
			Left:     left,
			Right:    right,
		}
	}
}

// '-' valueExpr | primary
func (p *parser) parseUnary(hints []string) ValueExpr {
	if p.cur().Kind != tkMinus {
		return p.parsePrimary(hints)
	}

	opTk := p.bump()
	var expr ValueExpr
	if p.atExprStart(hints) {
		expr = p.parseUnary(hints)
	} else {
		p.errExpected("an expression")
	}

	return &Prefix{
		Range:    p.rangeFrom(opTk.Start),
		Operator: PrefixOperator(opTk.Text),
		Expr:     expr,
	}
}

// Pre: the current token is in the exprStart set
func (p *parser) parsePrimary(hints []string) ValueExpr {
	tk := p.cur()
	switch tk.Kind {
	case tkVariable:
		v := parseVariable(p.bump())
		if v == nil {
			return nil
		}
		return v

	case tkAsset:
		p.bump()
		return &AssetLiteral{Range: tk.Range, Asset: tk.Text}

	case tkString:
		return parseString(p.bump())

	case tkNumber:
		return parseNumberLiteral(p.bump())

	case tkPercentage:
		p.bump()
		return parsePercentageRatio(tk.Text, tk.Range)

	case tkAt:
		return p.parseAccountLiteral()

	case tkLBracket:
		lit, _ := p.parseBracketLiteral(false)
		return lit

	case tkLParen:
		p.bump()
		expr := p.parseExpr(hints)
		p.expect(tkRParen)
		return expr

	case tkIdentifier, tkOverdraft:
		if p.peek(1).Kind == tkLParen {
			return p.parseFnCall()
		}
		p.errUnexpected("an expression")
		p.bump()
		return nil

	default:
		p.errExpected("an expression")
		return nil
	}
}

// '@' accountPart (':' accountPart)*
func (p *parser) parseAccountLiteral() ValueExpr {
	atTk := p.bump()
	slot := &AccountPartSlot{LiteralStart: atTk.Start}

	var parts []AccountNamePart
	p.parseAccountPart(&parts, slot)
	for p.cur().Kind == tkColon {
		p.bump()
		parts = append(parts, AccountTextPart{Name: ":"})
		p.parseAccountPart(&parts, slot)
	}

	if len(parts) == 0 {
		return nil
	}

	return &AccountInterpLiteral{
		Range: p.rangeFrom(atTk.Start),
		Parts: parts,
	}
}

func (p *parser) parseAccountPart(parts *[]AccountNamePart, slot *AccountPartSlot) {
	p.recordSlot(slot)

	tk := p.cur()
	// Account parts must not be separated by whitespace from the preceding '@' or ':'
	adjacent := p.pos > 0 && p.tokens[p.pos-1].EndOffset == tk.StartOffset

	switch {
	case tk.Kind == tkAccountText:
		p.bump()
		*parts = append(*parts, AccountTextPart{Name: tk.Text})

	case tk.Kind == tkVariable && adjacent:
		v := parseVariable(p.bump())
		if v != nil {
			*parts = append(*parts, v)
		}

	default:
		p.errExpected("an account name")
	}
}

// Parses either a monetary literal ('[' valueExpr valueExpr ']')
// or, when allowSendAll is true, a sendAll literal ('[' valueExpr '*' ']')
func (p *parser) parseBracketLiteral(allowSendAll bool) (ValueExpr, *SentValueAll) {
	openTk := p.bump()
	start := openTk.Start

	asset := p.parseJuxtaposedExpr(hintAsset,
		func() bool {
			kind := p.cur().Kind
			return !exprStart.has(kind) && (!allowSendAll || kind != tkStar)
		},
		func() bool {
			return p.parseExpr(hintNumber) != nil && p.cur().Kind == tkRBracket
		},
	)

	if allowSendAll && p.eat(tkStar) {
		p.expectRecover(tkRBracket, setOf(tkLParen))
		return nil, &SentValueAll{
			Range: p.rangeFrom(start),
			Asset: asset,
		}
	}

	var amount ValueExpr
	if p.atExprStart(hintNumber) {
		amount = p.parseExpr(hintNumber)
	} else {
		p.errExpected("an amount")
	}

	p.expectRecover(tkRBracket, setOf(tkLParen))

	return &MonetaryLiteral{
		Range:  p.rangeFrom(start),
		Asset:  asset,
		Amount: amount,
	}, nil
}

// Parses an expression that may be directly followed by an allotment, e.g. the
// `@a` in `{ 1/2 to @a -1/2 to @b }`, where the greedy parse would read `@a - 1/2`.
// follow is the set of tokens that follows an allotment in that position.
func (p *parser) parseExprBeforeAllotment(hints []string, follow tokenSet) ValueExpr {
	if follow == 0 {
		return p.parseExpr(hints)
	}

	return p.parseJuxtaposedExpr(hints,
		func() bool { return follow.has(p.cur().Kind) },
		func() bool { return p.parseExpr(hintPortion) != nil && follow.has(p.cur().Kind) },
	)
}

// In some places of the grammar, two expressions are not separated by any token,
// (e.g. in the `[$asset -1]` monetary literal). When the greedy parse of the first
// expression consumes the beginning of the second one (needsSplit returns true after it),
// we look for a binary '-' that can be reinterpreted as the prefix operator of the
// second expression. The rightmost split for which isViable succeeds wins (that's the
// choice the ANTLR parser used to make).
//
// isViable is called after the first expression of the split, and it's allowed to consume tokens:
// the parser state is restored afterwards.
func (p *parser) parseJuxtaposedExpr(hints []string, needsSplit func() bool, isViable func() bool) ValueExpr {
	before := p.checkpoint()
	expr := p.parseExpr(hints)
	if !needsSplit() {
		return expr
	}

	greedyEnd := p.pos

	var candidates []int
	depth := 0
	for i := before.pos; i < greedyEnd; i++ {
		switch p.tokens[i].Kind {
		case tkLParen, tkLBracket:
			depth++
		case tkRParen, tkRBracket:
			depth--
		case tkMinus:
			if depth == 0 && i > before.pos && !prefixPosition.has(p.tokens[i-1].Kind) {
				candidates = append(candidates, i)
			}
		}
	}
	if len(candidates) == 0 {
		return expr
	}

	prevStopAt := p.stopAt
	defer func() { p.stopAt = prevStopAt }()

	for _, splitAt := range slices.Backward(candidates) {
		p.restore(before)
		p.lastErrorTok = -1

		p.stopAt = splitAt
		splitExpr := p.parseExpr(hints)
		p.stopAt = prevStopAt

		if p.pos != splitAt || len(p.errors) != before.nErrors {
			continue
		}

		afterSplit := p.checkpoint()
		viable := isViable() && len(p.errors) == before.nErrors
		p.restore(afterSplit)
		if viable {
			return splitExpr
		}
	}

	// No viable split: parse the expression greedily again
	p.restore(before)
	expr = p.parseExpr(hints)
	if p.pos != greedyEnd {
		panic("unreachable: the re-parse of an expression diverged")
	}
	return expr
}

// (IDENTIFIER | 'overdraft') '(' (valueExpr (',' valueExpr)*)? ')'
func (p *parser) parseFnCall() *FnCall {
	nameTk := p.bump()
	p.bump() // '('

	p.fnFrames = append(p.fnFrames, FnCallFrame{Name: nameTk.Text})
	defer func() { p.fnFrames = p.fnFrames[:len(p.fnFrames)-1] }()

	call := &FnCall{
		Caller: &FnCallIdentifier{Range: nameTk.Range, Name: nameTk.Text},
	}

	isExpr := p.atExprStart(nil)
	if isExpr || !p.at(tkRParen) {
		for {
			arg := p.parseExpr(nil)
			if arg != nil {
				call.Args = append(call.Args, arg)
			}

			if !p.eat(tkComma) {
				break
			}
			p.fnFrames[len(p.fnFrames)-1].ArgIndex++
		}
	}

	if !p.eat(tkRParen) {
		p.errExpected("`,` or `)`")
	}

	call.Range = p.rangeFrom(nameTk.Start)
	return call
}

// --- Literals

func parseVariable(tk token) *Variable {
	if len(tk.Text) <= 1 {
		// an invalid `$`: the error was already reported by the lexer
		return nil
	}
	return &Variable{Range: tk.Range, Name: tk.Text[1:]}
}

func parseString(tk token) *StringLiteral {
	inner := tk.Text[1:]
	// an unterminated string literal has no closing quote (the error is reported by the lexer)
	if len(tk.Text) >= 2 && strings.HasSuffix(tk.Text, `"`) {
		inner = inner[:len(inner)-1]
	}
	return &StringLiteral{Range: tk.Range, String: inner}
}

func parseNumberLiteral(tk token) *NumberLiteral {
	amt, ok := new(big.Int).SetString(strings.ReplaceAll(tk.Text, "_", ""), 10)
	if !ok {
		// This should be unreachable for tokens emitted by the lexer.
		// Degrade gracefully to a zero-value literal instead of panicking,
		// so that a malformed token cannot crash the process (e.g. the LSP server).
		amt = big.NewInt(0)
	}

	return &NumberLiteral{Range: tk.Range, Number: amt}
}

func parsePercentageRatio(source string, range_ Range) *PercentageLiteral {
	num, floatingDigits, err := ParsePercentageRatio(source)
	if err != nil {
		// This should be unreachable for tokens emitted by the lexer.
		// Degrade gracefully to a zero-value literal instead of panicking,
		// so that a malformed token cannot crash the process (e.g. the LSP server).
		return &PercentageLiteral{
			Range:  range_,
			Amount: big.NewInt(0),
		}
	}

	return &PercentageLiteral{
		Range:          range_,
		Amount:         num,
		FloatingDigits: floatingDigits,
	}
}
