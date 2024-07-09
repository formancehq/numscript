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
		"", "", "", "", "", "'source'", "'send'", "'('", "')'", "'['", "']'",
		"'='",
	}
	staticData.SymbolicNames = []string{
		"", "WS", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "SOURCE",
		"SEND", "LPARENS", "RPARENS", "LBRACKET", "RBRACKET", "EQ", "NUMBER",
		"VARIABLE_NAME", "ACCOUNT", "ASSET",
	}
	staticData.RuleNames = []string{
		"program", "monetaryLit", "source", "statement",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 15, 30, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 1, 0, 5,
		0, 10, 8, 0, 10, 0, 12, 0, 13, 9, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2,
		1, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 0, 0, 4, 0,
		2, 4, 6, 0, 0, 26, 0, 11, 1, 0, 0, 0, 2, 14, 1, 0, 0, 0, 4, 19, 1, 0, 0,
		0, 6, 21, 1, 0, 0, 0, 8, 10, 3, 6, 3, 0, 9, 8, 1, 0, 0, 0, 10, 13, 1, 0,
		0, 0, 11, 9, 1, 0, 0, 0, 11, 12, 1, 0, 0, 0, 12, 1, 1, 0, 0, 0, 13, 11,
		1, 0, 0, 0, 14, 15, 5, 9, 0, 0, 15, 16, 5, 15, 0, 0, 16, 17, 5, 12, 0,
		0, 17, 18, 5, 10, 0, 0, 18, 3, 1, 0, 0, 0, 19, 20, 5, 14, 0, 0, 20, 5,
		1, 0, 0, 0, 21, 22, 5, 6, 0, 0, 22, 23, 3, 2, 1, 0, 23, 24, 5, 7, 0, 0,
		24, 25, 5, 5, 0, 0, 25, 26, 5, 11, 0, 0, 26, 27, 3, 4, 2, 0, 27, 28, 5,
		8, 0, 0, 28, 7, 1, 0, 0, 0, 1, 11,
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
	NumscriptParserWS                = 1
	NumscriptParserNEWLINE           = 2
	NumscriptParserMULTILINE_COMMENT = 3
	NumscriptParserLINE_COMMENT      = 4
	NumscriptParserSOURCE            = 5
	NumscriptParserSEND              = 6
	NumscriptParserLPARENS           = 7
	NumscriptParserRPARENS           = 8
	NumscriptParserLBRACKET          = 9
	NumscriptParserRBRACKET          = 10
	NumscriptParserEQ                = 11
	NumscriptParserNUMBER            = 12
	NumscriptParserVARIABLE_NAME     = 13
	NumscriptParserACCOUNT           = 14
	NumscriptParserASSET             = 15
)

// NumscriptParser rules.
const (
	NumscriptParserRULE_program     = 0
	NumscriptParserRULE_monetaryLit = 1
	NumscriptParserRULE_source      = 2
	NumscriptParserRULE_statement   = 3
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
	p.SetState(11)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NumscriptParserSEND {
		{
			p.SetState(8)
			p.Statement()
		}

		p.SetState(13)
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
		p.SetState(14)
		p.Match(NumscriptParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(15)

		var _m = p.Match(NumscriptParserASSET)

		localctx.(*MonetaryLitContext).asset = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(16)

		var _m = p.Match(NumscriptParserNUMBER)

		localctx.(*MonetaryLitContext).amt = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(17)
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

// ISourceContext is an interface to support dynamic dispatch.
type ISourceContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ACCOUNT() antlr.TerminalNode

	// IsSourceContext differentiates from other interfaces.
	IsSourceContext()
}

type SourceContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySourceContext() *SourceContext {
	var p = new(SourceContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_source
	return p
}

func InitEmptySourceContext(p *SourceContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_source
}

func (*SourceContext) IsSourceContext() {}

func NewSourceContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SourceContext {
	var p = new(SourceContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_source

	return p
}

func (s *SourceContext) GetParser() antlr.Parser { return s.parser }

func (s *SourceContext) ACCOUNT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserACCOUNT, 0)
}

func (s *SourceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SourceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SourceContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSource(s)
	}
}

func (s *SourceContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSource(s)
	}
}

func (s *SourceContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitSource(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) Source() (localctx ISourceContext) {
	localctx = NewSourceContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, NumscriptParserRULE_source)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(19)
		p.Match(NumscriptParserACCOUNT)
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
	SOURCE() antlr.TerminalNode
	EQ() antlr.TerminalNode
	Source() ISourceContext
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

func (s *StatementContext) SOURCE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserSOURCE, 0)
}

func (s *StatementContext) EQ() antlr.TerminalNode {
	return s.GetToken(NumscriptParserEQ, 0)
}

func (s *StatementContext) Source() ISourceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISourceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISourceContext)
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
	p.EnterRule(localctx, 6, NumscriptParserRULE_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(21)
		p.Match(NumscriptParserSEND)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(22)
		p.MonetaryLit()
	}
	{
		p.SetState(23)
		p.Match(NumscriptParserLPARENS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(24)
		p.Match(NumscriptParserSOURCE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(25)
		p.Match(NumscriptParserEQ)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(26)
		p.Source()
	}
	{
		p.SetState(27)
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
