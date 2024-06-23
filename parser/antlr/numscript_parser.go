// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type NumscriptParser struct {
	*antlr.BaseParser
}

var NumscriptParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func numscriptParserInit() {
	staticData := &NumscriptParserStaticData
	staticData.LiteralNames = []string{
		"", "", "", "", "", "'send'", "'('", "')'", "'['", "']'", "'='",
	}
	staticData.SymbolicNames = []string{
		"", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "WS", "SEND", "LPARENS",
		"RPARENS", "LBRACKET", "RBRACKET", "EQ", "NUMBER", "VARIABLE_NAME",
		"ACCOUNT", "ASSET",
	}
	staticData.RuleNames = []string{
		"program", "monetaryLit", "statement",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 14, 23, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 1, 0, 5, 0, 8, 8, 0,
		10, 0, 12, 0, 11, 9, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 2,
		1, 2, 1, 2, 1, 2, 0, 0, 3, 0, 2, 4, 0, 0, 20, 0, 9, 1, 0, 0, 0, 2, 12,
		1, 0, 0, 0, 4, 17, 1, 0, 0, 0, 6, 8, 3, 4, 2, 0, 7, 6, 1, 0, 0, 0, 8, 11,
		1, 0, 0, 0, 9, 7, 1, 0, 0, 0, 9, 10, 1, 0, 0, 0, 10, 1, 1, 0, 0, 0, 11,
		9, 1, 0, 0, 0, 12, 13, 5, 8, 0, 0, 13, 14, 5, 14, 0, 0, 14, 15, 5, 11,
		0, 0, 15, 16, 5, 9, 0, 0, 16, 3, 1, 0, 0, 0, 17, 18, 5, 5, 0, 0, 18, 19,
		3, 2, 1, 0, 19, 20, 5, 6, 0, 0, 20, 21, 5, 7, 0, 0, 21, 5, 1, 0, 0, 0,
		1, 9,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// NumscriptParserInit initializes any static state used to implement NumscriptParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewNumscriptParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func NumscriptParserInit() {
	staticData := &NumscriptParserStaticData
	staticData.once.Do(numscriptParserInit)
}

// NewNumscriptParser produces a new parser instance for the optional input antlr.TokenStream.
func NewNumscriptParser(input antlr.TokenStream) *NumscriptParser {
	NumscriptParserInit()
	this := new(NumscriptParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &NumscriptParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Numscript.g4"

	return this
}

// NumscriptParser tokens.
const (
	NumscriptParserEOF               = antlr.TokenEOF
	NumscriptParserNEWLINE           = 1
	NumscriptParserMULTILINE_COMMENT = 2
	NumscriptParserLINE_COMMENT      = 3
	NumscriptParserWS                = 4
	NumscriptParserSEND              = 5
	NumscriptParserLPARENS           = 6
	NumscriptParserRPARENS           = 7
	NumscriptParserLBRACKET          = 8
	NumscriptParserRBRACKET          = 9
	NumscriptParserEQ                = 10
	NumscriptParserNUMBER            = 11
	NumscriptParserVARIABLE_NAME     = 12
	NumscriptParserACCOUNT           = 13
	NumscriptParserASSET             = 14
)

// NumscriptParser rules.
const (
	NumscriptParserRULE_program     = 0
	NumscriptParserRULE_monetaryLit = 1
	NumscriptParserRULE_statement   = 2
)

// IProgramContext is an interface to support dynamic dispatch.
type IProgramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllStatement() []IStatementContext
	Statement(i int) IStatementContext

	// IsProgramContext differentiates from other interfaces.
	IsProgramContext()
}

type ProgramContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyProgramContext() *ProgramContext {
	var p = new(ProgramContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_program
	return p
}

func InitEmptyProgramContext(p *ProgramContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_program
}

func (*ProgramContext) IsProgramContext() {}

func NewProgramContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ProgramContext {
	var p = new(ProgramContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_program

	return p
}

func (s *ProgramContext) GetParser() antlr.Parser { return s.parser }

func (s *ProgramContext) AllStatement() []IStatementContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStatementContext); ok {
			len++
		}
	}

	tst := make([]IStatementContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStatementContext); ok {
			tst[i] = t.(IStatementContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Statement(i int) IStatementContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatementContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStatementContext)
}

func (s *ProgramContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ProgramContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ProgramContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterProgram(s)
	}
}

func (s *ProgramContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitProgram(s)
	}
}

func (s *ProgramContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitProgram(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) Program() (localctx IProgramContext) {
	localctx = NewProgramContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, NumscriptParserRULE_program)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(9)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NumscriptParserSEND {
		{
			p.SetState(6)
			p.Statement()
		}

		p.SetState(11)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMonetaryLitContext is an interface to support dynamic dispatch.
type IMonetaryLitContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetAsset returns the asset token.
	GetAsset() antlr.Token

	// GetAmt returns the amt token.
	GetAmt() antlr.Token

	// SetAsset sets the asset token.
	SetAsset(antlr.Token)

	// SetAmt sets the amt token.
	SetAmt(antlr.Token)

	// Getter signatures
	LBRACKET() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode
	ASSET() antlr.TerminalNode
	NUMBER() antlr.TerminalNode

	// IsMonetaryLitContext differentiates from other interfaces.
	IsMonetaryLitContext()
}

type MonetaryLitContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	asset  antlr.Token
	amt    antlr.Token
}

func NewEmptyMonetaryLitContext() *MonetaryLitContext {
	var p = new(MonetaryLitContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_monetaryLit
	return p
}

func InitEmptyMonetaryLitContext(p *MonetaryLitContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_monetaryLit
}

func (*MonetaryLitContext) IsMonetaryLitContext() {}

func NewMonetaryLitContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MonetaryLitContext {
	var p = new(MonetaryLitContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_monetaryLit

	return p
}

func (s *MonetaryLitContext) GetParser() antlr.Parser { return s.parser }

func (s *MonetaryLitContext) GetAsset() antlr.Token { return s.asset }

func (s *MonetaryLitContext) GetAmt() antlr.Token { return s.amt }

func (s *MonetaryLitContext) SetAsset(v antlr.Token) { s.asset = v }

func (s *MonetaryLitContext) SetAmt(v antlr.Token) { s.amt = v }

func (s *MonetaryLitContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACKET, 0)
}

func (s *MonetaryLitContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACKET, 0)
}

func (s *MonetaryLitContext) ASSET() antlr.TerminalNode {
	return s.GetToken(NumscriptParserASSET, 0)
}

func (s *MonetaryLitContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(NumscriptParserNUMBER, 0)
}

func (s *MonetaryLitContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MonetaryLitContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MonetaryLitContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterMonetaryLit(s)
	}
}

func (s *MonetaryLitContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitMonetaryLit(s)
	}
}

func (s *MonetaryLitContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitMonetaryLit(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) MonetaryLit() (localctx IMonetaryLitContext) {
	localctx = NewMonetaryLitContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, NumscriptParserRULE_monetaryLit)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(12)
		p.Match(NumscriptParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(13)

		var _m = p.Match(NumscriptParserASSET)

		localctx.(*MonetaryLitContext).asset = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(14)

		var _m = p.Match(NumscriptParserNUMBER)

		localctx.(*MonetaryLitContext).amt = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(15)
		p.Match(NumscriptParserRBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IStatementContext is an interface to support dynamic dispatch.
type IStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SEND() antlr.TerminalNode
	MonetaryLit() IMonetaryLitContext
	LPARENS() antlr.TerminalNode
	RPARENS() antlr.TerminalNode

	// IsStatementContext differentiates from other interfaces.
	IsStatementContext()
}

type StatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStatementContext() *StatementContext {
	var p = new(StatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_statement
	return p
}

func InitEmptyStatementContext(p *StatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_statement
}

func (*StatementContext) IsStatementContext() {}

func NewStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StatementContext {
	var p = new(StatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_statement

	return p
}

func (s *StatementContext) GetParser() antlr.Parser { return s.parser }

func (s *StatementContext) SEND() antlr.TerminalNode {
	return s.GetToken(NumscriptParserSEND, 0)
}

func (s *StatementContext) MonetaryLit() IMonetaryLitContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMonetaryLitContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMonetaryLitContext)
}

func (s *StatementContext) LPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLPARENS, 0)
}

func (s *StatementContext) RPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRPARENS, 0)
}

func (s *StatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StatementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterStatement(s)
	}
}

func (s *StatementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitStatement(s)
	}
}

func (s *StatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) Statement() (localctx IStatementContext) {
	localctx = NewStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, NumscriptParserRULE_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(17)
		p.Match(NumscriptParserSEND)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(18)
		p.MonetaryLit()
	}
	{
		p.SetState(19)
		p.Match(NumscriptParserLPARENS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(20)
		p.Match(NumscriptParserRPARENS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
