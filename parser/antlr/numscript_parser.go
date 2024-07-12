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
		"", "", "", "", "", "'vars'", "'max'", "'source'", "'destination'",
		"'send'", "'from'", "'to'", "'('", "')'", "'['", "']'", "'{'", "'}'",
		"'='",
	}
	staticData.SymbolicNames = []string{
		"", "WS", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "VARS", "MAX",
		"SOURCE", "DESTINATION", "SEND", "FROM", "TO", "LPARENS", "RPARENS",
		"LBRACKET", "RBRACKET", "LBRACE", "RBRACE", "EQ", "RATIO_PORTION_LITERAL",
		"PERCENTAGE_PORTION_LITERAL", "TYPE_IDENT", "NUMBER", "VARIABLE_NAME",
		"ACCOUNT", "ASSET",
	}
	staticData.RuleNames = []string{
		"portion", "varDeclaration", "varsDeclaration", "program", "monetaryLit",
		"cap", "source", "allotmentClauseSrc", "destination", "allotmentClauseDest",
		"sendExpr", "statement",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 25, 128, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 1, 0, 1, 0, 3, 0, 27, 8, 0, 1, 1, 1, 1, 1, 1, 1, 2, 1,
		2, 1, 2, 5, 2, 35, 8, 2, 10, 2, 12, 2, 38, 9, 2, 1, 2, 1, 2, 1, 3, 3, 3,
		43, 8, 3, 1, 3, 5, 3, 46, 8, 3, 10, 3, 12, 3, 49, 9, 3, 1, 4, 1, 4, 1,
		4, 1, 4, 1, 4, 1, 5, 1, 5, 3, 5, 58, 8, 5, 1, 6, 1, 6, 1, 6, 1, 6, 4, 6,
		64, 8, 6, 11, 6, 12, 6, 65, 1, 6, 1, 6, 1, 6, 1, 6, 5, 6, 72, 8, 6, 10,
		6, 12, 6, 75, 9, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 3, 6, 83, 8, 6,
		1, 7, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 8, 4, 8, 93, 8, 8, 11, 8,
		12, 8, 94, 1, 8, 1, 8, 1, 8, 1, 8, 5, 8, 101, 8, 8, 10, 8, 12, 8, 104,
		9, 8, 1, 8, 3, 8, 107, 8, 8, 1, 9, 1, 9, 1, 9, 1, 9, 1, 10, 1, 10, 3, 10,
		115, 8, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1,
		11, 1, 11, 1, 11, 1, 11, 0, 0, 12, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20,
		22, 0, 0, 132, 0, 26, 1, 0, 0, 0, 2, 28, 1, 0, 0, 0, 4, 31, 1, 0, 0, 0,
		6, 42, 1, 0, 0, 0, 8, 50, 1, 0, 0, 0, 10, 57, 1, 0, 0, 0, 12, 82, 1, 0,
		0, 0, 14, 84, 1, 0, 0, 0, 16, 106, 1, 0, 0, 0, 18, 108, 1, 0, 0, 0, 20,
		114, 1, 0, 0, 0, 22, 116, 1, 0, 0, 0, 24, 27, 5, 19, 0, 0, 25, 27, 5, 20,
		0, 0, 26, 24, 1, 0, 0, 0, 26, 25, 1, 0, 0, 0, 27, 1, 1, 0, 0, 0, 28, 29,
		5, 21, 0, 0, 29, 30, 5, 23, 0, 0, 30, 3, 1, 0, 0, 0, 31, 32, 5, 5, 0, 0,
		32, 36, 5, 16, 0, 0, 33, 35, 3, 2, 1, 0, 34, 33, 1, 0, 0, 0, 35, 38, 1,
		0, 0, 0, 36, 34, 1, 0, 0, 0, 36, 37, 1, 0, 0, 0, 37, 39, 1, 0, 0, 0, 38,
		36, 1, 0, 0, 0, 39, 40, 5, 17, 0, 0, 40, 5, 1, 0, 0, 0, 41, 43, 3, 4, 2,
		0, 42, 41, 1, 0, 0, 0, 42, 43, 1, 0, 0, 0, 43, 47, 1, 0, 0, 0, 44, 46,
		3, 22, 11, 0, 45, 44, 1, 0, 0, 0, 46, 49, 1, 0, 0, 0, 47, 45, 1, 0, 0,
		0, 47, 48, 1, 0, 0, 0, 48, 7, 1, 0, 0, 0, 49, 47, 1, 0, 0, 0, 50, 51, 5,
		14, 0, 0, 51, 52, 5, 25, 0, 0, 52, 53, 5, 22, 0, 0, 53, 54, 5, 15, 0, 0,
		54, 9, 1, 0, 0, 0, 55, 58, 3, 8, 4, 0, 56, 58, 5, 23, 0, 0, 57, 55, 1,
		0, 0, 0, 57, 56, 1, 0, 0, 0, 58, 11, 1, 0, 0, 0, 59, 83, 5, 24, 0, 0, 60,
		83, 5, 23, 0, 0, 61, 63, 5, 16, 0, 0, 62, 64, 3, 14, 7, 0, 63, 62, 1, 0,
		0, 0, 64, 65, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 65, 66, 1, 0, 0, 0, 66, 67,
		1, 0, 0, 0, 67, 68, 5, 17, 0, 0, 68, 83, 1, 0, 0, 0, 69, 73, 5, 16, 0,
		0, 70, 72, 3, 12, 6, 0, 71, 70, 1, 0, 0, 0, 72, 75, 1, 0, 0, 0, 73, 71,
		1, 0, 0, 0, 73, 74, 1, 0, 0, 0, 74, 76, 1, 0, 0, 0, 75, 73, 1, 0, 0, 0,
		76, 83, 5, 17, 0, 0, 77, 78, 5, 6, 0, 0, 78, 79, 3, 10, 5, 0, 79, 80, 5,
		10, 0, 0, 80, 81, 3, 12, 6, 0, 81, 83, 1, 0, 0, 0, 82, 59, 1, 0, 0, 0,
		82, 60, 1, 0, 0, 0, 82, 61, 1, 0, 0, 0, 82, 69, 1, 0, 0, 0, 82, 77, 1,
		0, 0, 0, 83, 13, 1, 0, 0, 0, 84, 85, 3, 0, 0, 0, 85, 86, 5, 10, 0, 0, 86,
		87, 3, 12, 6, 0, 87, 15, 1, 0, 0, 0, 88, 107, 5, 24, 0, 0, 89, 107, 5,
		23, 0, 0, 90, 92, 5, 16, 0, 0, 91, 93, 3, 18, 9, 0, 92, 91, 1, 0, 0, 0,
		93, 94, 1, 0, 0, 0, 94, 92, 1, 0, 0, 0, 94, 95, 1, 0, 0, 0, 95, 96, 1,
		0, 0, 0, 96, 97, 5, 17, 0, 0, 97, 107, 1, 0, 0, 0, 98, 102, 5, 16, 0, 0,
		99, 101, 3, 16, 8, 0, 100, 99, 1, 0, 0, 0, 101, 104, 1, 0, 0, 0, 102, 100,
		1, 0, 0, 0, 102, 103, 1, 0, 0, 0, 103, 105, 1, 0, 0, 0, 104, 102, 1, 0,
		0, 0, 105, 107, 5, 17, 0, 0, 106, 88, 1, 0, 0, 0, 106, 89, 1, 0, 0, 0,
		106, 90, 1, 0, 0, 0, 106, 98, 1, 0, 0, 0, 107, 17, 1, 0, 0, 0, 108, 109,
		3, 0, 0, 0, 109, 110, 5, 11, 0, 0, 110, 111, 3, 16, 8, 0, 111, 19, 1, 0,
		0, 0, 112, 115, 3, 8, 4, 0, 113, 115, 5, 23, 0, 0, 114, 112, 1, 0, 0, 0,
		114, 113, 1, 0, 0, 0, 115, 21, 1, 0, 0, 0, 116, 117, 5, 9, 0, 0, 117, 118,
		3, 20, 10, 0, 118, 119, 5, 12, 0, 0, 119, 120, 5, 7, 0, 0, 120, 121, 5,
		18, 0, 0, 121, 122, 3, 12, 6, 0, 122, 123, 5, 8, 0, 0, 123, 124, 5, 18,
		0, 0, 124, 125, 3, 16, 8, 0, 125, 126, 5, 13, 0, 0, 126, 23, 1, 0, 0, 0,
		12, 26, 36, 42, 47, 57, 65, 73, 82, 94, 102, 106, 114,
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
	NumscriptParserEOF                        = antlr.TokenEOF
	NumscriptParserWS                         = 1
	NumscriptParserNEWLINE                    = 2
	NumscriptParserMULTILINE_COMMENT          = 3
	NumscriptParserLINE_COMMENT               = 4
	NumscriptParserVARS                       = 5
	NumscriptParserMAX                        = 6
	NumscriptParserSOURCE                     = 7
	NumscriptParserDESTINATION                = 8
	NumscriptParserSEND                       = 9
	NumscriptParserFROM                       = 10
	NumscriptParserTO                         = 11
	NumscriptParserLPARENS                    = 12
	NumscriptParserRPARENS                    = 13
	NumscriptParserLBRACKET                   = 14
	NumscriptParserRBRACKET                   = 15
	NumscriptParserLBRACE                     = 16
	NumscriptParserRBRACE                     = 17
	NumscriptParserEQ                         = 18
	NumscriptParserRATIO_PORTION_LITERAL      = 19
	NumscriptParserPERCENTAGE_PORTION_LITERAL = 20
	NumscriptParserTYPE_IDENT                 = 21
	NumscriptParserNUMBER                     = 22
	NumscriptParserVARIABLE_NAME              = 23
	NumscriptParserACCOUNT                    = 24
	NumscriptParserASSET                      = 25
)

// NumscriptParser rules.
const (
	NumscriptParserRULE_portion             = 0
	NumscriptParserRULE_varDeclaration      = 1
	NumscriptParserRULE_varsDeclaration     = 2
	NumscriptParserRULE_program             = 3
	NumscriptParserRULE_monetaryLit         = 4
	NumscriptParserRULE_cap                 = 5
	NumscriptParserRULE_source              = 6
	NumscriptParserRULE_allotmentClauseSrc  = 7
	NumscriptParserRULE_destination         = 8
	NumscriptParserRULE_allotmentClauseDest = 9
	NumscriptParserRULE_sendExpr            = 10
	NumscriptParserRULE_statement           = 11
)

// IPortionContext is an interface to support dynamic dispatch.
type IPortionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsPortionContext differentiates from other interfaces.
	IsPortionContext()
}

type PortionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPortionContext() *PortionContext {
	var p = new(PortionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_portion
	return p
}

func InitEmptyPortionContext(p *PortionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_portion
}

func (*PortionContext) IsPortionContext() {}

func NewPortionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PortionContext {
	var p = new(PortionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_portion

	return p
}

func (s *PortionContext) GetParser() antlr.Parser { return s.parser }

func (s *PortionContext) CopyAll(ctx *PortionContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *PortionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PortionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type PercentageContext struct {
	PortionContext
}

func NewPercentageContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PercentageContext {
	var p = new(PercentageContext)

	InitEmptyPortionContext(&p.PortionContext)
	p.parser = parser
	p.CopyAll(ctx.(*PortionContext))

	return p
}

func (s *PercentageContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PercentageContext) PERCENTAGE_PORTION_LITERAL() antlr.TerminalNode {
	return s.GetToken(NumscriptParserPERCENTAGE_PORTION_LITERAL, 0)
}

func (s *PercentageContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterPercentage(s)
	}
}

func (s *PercentageContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitPercentage(s)
	}
}

func (s *PercentageContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitPercentage(s)

	default:
		return t.VisitChildren(s)
	}
}

type RatioContext struct {
	PortionContext
}

func NewRatioContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *RatioContext {
	var p = new(RatioContext)

	InitEmptyPortionContext(&p.PortionContext)
	p.parser = parser
	p.CopyAll(ctx.(*PortionContext))

	return p
}

func (s *RatioContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RatioContext) RATIO_PORTION_LITERAL() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRATIO_PORTION_LITERAL, 0)
}

func (s *RatioContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterRatio(s)
	}
}

func (s *RatioContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitRatio(s)
	}
}

func (s *RatioContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitRatio(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) Portion() (localctx IPortionContext) {
	localctx = NewPortionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, NumscriptParserRULE_portion)
	p.SetState(26)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserRATIO_PORTION_LITERAL:
		localctx = NewRatioContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(24)
			p.Match(NumscriptParserRATIO_PORTION_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserPERCENTAGE_PORTION_LITERAL:
		localctx = NewPercentageContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(25)
			p.Match(NumscriptParserPERCENTAGE_PORTION_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
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

// IVarDeclarationContext is an interface to support dynamic dispatch.
type IVarDeclarationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetType_ returns the type_ token.
	GetType_() antlr.Token

	// GetName returns the name token.
	GetName() antlr.Token

	// SetType_ sets the type_ token.
	SetType_(antlr.Token)

	// SetName sets the name token.
	SetName(antlr.Token)

	// Getter signatures
	TYPE_IDENT() antlr.TerminalNode
	VARIABLE_NAME() antlr.TerminalNode

	// IsVarDeclarationContext differentiates from other interfaces.
	IsVarDeclarationContext()
}

type VarDeclarationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	type_  antlr.Token
	name   antlr.Token
}

func NewEmptyVarDeclarationContext() *VarDeclarationContext {
	var p = new(VarDeclarationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_varDeclaration
	return p
}

func InitEmptyVarDeclarationContext(p *VarDeclarationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_varDeclaration
}

func (*VarDeclarationContext) IsVarDeclarationContext() {}

func NewVarDeclarationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *VarDeclarationContext {
	var p = new(VarDeclarationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_varDeclaration

	return p
}

func (s *VarDeclarationContext) GetParser() antlr.Parser { return s.parser }

func (s *VarDeclarationContext) GetType_() antlr.Token { return s.type_ }

func (s *VarDeclarationContext) GetName() antlr.Token { return s.name }

func (s *VarDeclarationContext) SetType_(v antlr.Token) { s.type_ = v }

func (s *VarDeclarationContext) SetName(v antlr.Token) { s.name = v }

func (s *VarDeclarationContext) TYPE_IDENT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserTYPE_IDENT, 0)
}

func (s *VarDeclarationContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARIABLE_NAME, 0)
}

func (s *VarDeclarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VarDeclarationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *VarDeclarationContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterVarDeclaration(s)
	}
}

func (s *VarDeclarationContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitVarDeclaration(s)
	}
}

func (s *VarDeclarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitVarDeclaration(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) VarDeclaration() (localctx IVarDeclarationContext) {
	localctx = NewVarDeclarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, NumscriptParserRULE_varDeclaration)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(28)

		var _m = p.Match(NumscriptParserTYPE_IDENT)

		localctx.(*VarDeclarationContext).type_ = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(29)

		var _m = p.Match(NumscriptParserVARIABLE_NAME)

		localctx.(*VarDeclarationContext).name = _m
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

// IVarsDeclarationContext is an interface to support dynamic dispatch.
type IVarsDeclarationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	VARS() antlr.TerminalNode
	LBRACE() antlr.TerminalNode
	RBRACE() antlr.TerminalNode
	AllVarDeclaration() []IVarDeclarationContext
	VarDeclaration(i int) IVarDeclarationContext

	// IsVarsDeclarationContext differentiates from other interfaces.
	IsVarsDeclarationContext()
}

type VarsDeclarationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyVarsDeclarationContext() *VarsDeclarationContext {
	var p = new(VarsDeclarationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_varsDeclaration
	return p
}

func InitEmptyVarsDeclarationContext(p *VarsDeclarationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_varsDeclaration
}

func (*VarsDeclarationContext) IsVarsDeclarationContext() {}

func NewVarsDeclarationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *VarsDeclarationContext {
	var p = new(VarsDeclarationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_varsDeclaration

	return p
}

func (s *VarsDeclarationContext) GetParser() antlr.Parser { return s.parser }

func (s *VarsDeclarationContext) VARS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARS, 0)
}

func (s *VarsDeclarationContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACE, 0)
}

func (s *VarsDeclarationContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACE, 0)
}

func (s *VarsDeclarationContext) AllVarDeclaration() []IVarDeclarationContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IVarDeclarationContext); ok {
			len++
		}
	}

	tst := make([]IVarDeclarationContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IVarDeclarationContext); ok {
			tst[i] = t.(IVarDeclarationContext)
			i++
		}
	}

	return tst
}

func (s *VarsDeclarationContext) VarDeclaration(i int) IVarDeclarationContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVarDeclarationContext); ok {
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

	return t.(IVarDeclarationContext)
}

func (s *VarsDeclarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VarsDeclarationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *VarsDeclarationContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterVarsDeclaration(s)
	}
}

func (s *VarsDeclarationContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitVarsDeclaration(s)
	}
}

func (s *VarsDeclarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitVarsDeclaration(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) VarsDeclaration() (localctx IVarsDeclarationContext) {
	localctx = NewVarsDeclarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, NumscriptParserRULE_varsDeclaration)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(31)
		p.Match(NumscriptParserVARS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(32)
		p.Match(NumscriptParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(36)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NumscriptParserTYPE_IDENT {
		{
			p.SetState(33)
			p.VarDeclaration()
		}

		p.SetState(38)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(39)
		p.Match(NumscriptParserRBRACE)
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

// IProgramContext is an interface to support dynamic dispatch.
type IProgramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	VarsDeclaration() IVarsDeclarationContext
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

func (s *ProgramContext) VarsDeclaration() IVarsDeclarationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVarsDeclarationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVarsDeclarationContext)
}

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
	p.EnterRule(localctx, 6, NumscriptParserRULE_program)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(42)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NumscriptParserVARS {
		{
			p.SetState(41)
			p.VarsDeclaration()
		}

	}
	p.SetState(47)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NumscriptParserSEND {
		{
			p.SetState(44)
			p.Statement()
		}

		p.SetState(49)
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
	p.EnterRule(localctx, 8, NumscriptParserRULE_monetaryLit)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(50)
		p.Match(NumscriptParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(51)

		var _m = p.Match(NumscriptParserASSET)

		localctx.(*MonetaryLitContext).asset = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(52)

		var _m = p.Match(NumscriptParserNUMBER)

		localctx.(*MonetaryLitContext).amt = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(53)
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

// ICapContext is an interface to support dynamic dispatch.
type ICapContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsCapContext differentiates from other interfaces.
	IsCapContext()
}

type CapContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCapContext() *CapContext {
	var p = new(CapContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_cap
	return p
}

func InitEmptyCapContext(p *CapContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_cap
}

func (*CapContext) IsCapContext() {}

func NewCapContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CapContext {
	var p = new(CapContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_cap

	return p
}

func (s *CapContext) GetParser() antlr.Parser { return s.parser }

func (s *CapContext) CopyAll(ctx *CapContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *CapContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CapContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type LitCapContext struct {
	CapContext
}

func NewLitCapContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LitCapContext {
	var p = new(LitCapContext)

	InitEmptyCapContext(&p.CapContext)
	p.parser = parser
	p.CopyAll(ctx.(*CapContext))

	return p
}

func (s *LitCapContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LitCapContext) MonetaryLit() IMonetaryLitContext {
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

func (s *LitCapContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterLitCap(s)
	}
}

func (s *LitCapContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitLitCap(s)
	}
}

func (s *LitCapContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitLitCap(s)

	default:
		return t.VisitChildren(s)
	}
}

type VarCapContext struct {
	CapContext
}

func NewVarCapContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VarCapContext {
	var p = new(VarCapContext)

	InitEmptyCapContext(&p.CapContext)
	p.parser = parser
	p.CopyAll(ctx.(*CapContext))

	return p
}

func (s *VarCapContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VarCapContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARIABLE_NAME, 0)
}

func (s *VarCapContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterVarCap(s)
	}
}

func (s *VarCapContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitVarCap(s)
	}
}

func (s *VarCapContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitVarCap(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) Cap_() (localctx ICapContext) {
	localctx = NewCapContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, NumscriptParserRULE_cap)
	p.SetState(57)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserLBRACKET:
		localctx = NewLitCapContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(55)
			p.MonetaryLit()
		}

	case NumscriptParserVARIABLE_NAME:
		localctx = NewVarCapContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(56)
			p.Match(NumscriptParserVARIABLE_NAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
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

func (s *SourceContext) CopyAll(ctx *SourceContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *SourceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SourceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type SrcVariableContext struct {
	SourceContext
}

func NewSrcVariableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcVariableContext {
	var p = new(SrcVariableContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcVariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcVariableContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARIABLE_NAME, 0)
}

func (s *SrcVariableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSrcVariable(s)
	}
}

func (s *SrcVariableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSrcVariable(s)
	}
}

func (s *SrcVariableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitSrcVariable(s)

	default:
		return t.VisitChildren(s)
	}
}

type SrcSeqContext struct {
	SourceContext
}

func NewSrcSeqContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcSeqContext {
	var p = new(SrcSeqContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcSeqContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcSeqContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACE, 0)
}

func (s *SrcSeqContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACE, 0)
}

func (s *SrcSeqContext) AllSource() []ISourceContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISourceContext); ok {
			len++
		}
	}

	tst := make([]ISourceContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISourceContext); ok {
			tst[i] = t.(ISourceContext)
			i++
		}
	}

	return tst
}

func (s *SrcSeqContext) Source(i int) ISourceContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISourceContext); ok {
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

	return t.(ISourceContext)
}

func (s *SrcSeqContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSrcSeq(s)
	}
}

func (s *SrcSeqContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSrcSeq(s)
	}
}

func (s *SrcSeqContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitSrcSeq(s)

	default:
		return t.VisitChildren(s)
	}
}

type SrcAllotmentContext struct {
	SourceContext
}

func NewSrcAllotmentContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcAllotmentContext {
	var p = new(SrcAllotmentContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcAllotmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcAllotmentContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACE, 0)
}

func (s *SrcAllotmentContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACE, 0)
}

func (s *SrcAllotmentContext) AllAllotmentClauseSrc() []IAllotmentClauseSrcContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAllotmentClauseSrcContext); ok {
			len++
		}
	}

	tst := make([]IAllotmentClauseSrcContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAllotmentClauseSrcContext); ok {
			tst[i] = t.(IAllotmentClauseSrcContext)
			i++
		}
	}

	return tst
}

func (s *SrcAllotmentContext) AllotmentClauseSrc(i int) IAllotmentClauseSrcContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAllotmentClauseSrcContext); ok {
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

	return t.(IAllotmentClauseSrcContext)
}

func (s *SrcAllotmentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSrcAllotment(s)
	}
}

func (s *SrcAllotmentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSrcAllotment(s)
	}
}

func (s *SrcAllotmentContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitSrcAllotment(s)

	default:
		return t.VisitChildren(s)
	}
}

type SrcCappedContext struct {
	SourceContext
}

func NewSrcCappedContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcCappedContext {
	var p = new(SrcCappedContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcCappedContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcCappedContext) MAX() antlr.TerminalNode {
	return s.GetToken(NumscriptParserMAX, 0)
}

func (s *SrcCappedContext) Cap_() ICapContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICapContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICapContext)
}

func (s *SrcCappedContext) FROM() antlr.TerminalNode {
	return s.GetToken(NumscriptParserFROM, 0)
}

func (s *SrcCappedContext) Source() ISourceContext {
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

func (s *SrcCappedContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSrcCapped(s)
	}
}

func (s *SrcCappedContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSrcCapped(s)
	}
}

func (s *SrcCappedContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitSrcCapped(s)

	default:
		return t.VisitChildren(s)
	}
}

type SrcAccountContext struct {
	SourceContext
}

func NewSrcAccountContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcAccountContext {
	var p = new(SrcAccountContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcAccountContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcAccountContext) ACCOUNT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserACCOUNT, 0)
}

func (s *SrcAccountContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSrcAccount(s)
	}
}

func (s *SrcAccountContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSrcAccount(s)
	}
}

func (s *SrcAccountContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitSrcAccount(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) Source() (localctx ISourceContext) {
	localctx = NewSourceContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, NumscriptParserRULE_source)
	var _la int

	p.SetState(82)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSrcAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(59)
			p.Match(NumscriptParserACCOUNT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewSrcVariableContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(60)
			p.Match(NumscriptParserVARIABLE_NAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewSrcAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(61)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(63)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == NumscriptParserRATIO_PORTION_LITERAL || _la == NumscriptParserPERCENTAGE_PORTION_LITERAL {
			{
				p.SetState(62)
				p.AllotmentClauseSrc()
			}

			p.SetState(65)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(67)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		localctx = NewSrcSeqContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(69)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(73)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&25231424) != 0 {
			{
				p.SetState(70)
				p.Source()
			}

			p.SetState(75)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(76)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		localctx = NewSrcCappedContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(77)
			p.Match(NumscriptParserMAX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(78)
			p.Cap_()
		}
		{
			p.SetState(79)
			p.Match(NumscriptParserFROM)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(80)
			p.Source()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
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

// IAllotmentClauseSrcContext is an interface to support dynamic dispatch.
type IAllotmentClauseSrcContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Portion() IPortionContext
	FROM() antlr.TerminalNode
	Source() ISourceContext

	// IsAllotmentClauseSrcContext differentiates from other interfaces.
	IsAllotmentClauseSrcContext()
}

type AllotmentClauseSrcContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAllotmentClauseSrcContext() *AllotmentClauseSrcContext {
	var p = new(AllotmentClauseSrcContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_allotmentClauseSrc
	return p
}

func InitEmptyAllotmentClauseSrcContext(p *AllotmentClauseSrcContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_allotmentClauseSrc
}

func (*AllotmentClauseSrcContext) IsAllotmentClauseSrcContext() {}

func NewAllotmentClauseSrcContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AllotmentClauseSrcContext {
	var p = new(AllotmentClauseSrcContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_allotmentClauseSrc

	return p
}

func (s *AllotmentClauseSrcContext) GetParser() antlr.Parser { return s.parser }

func (s *AllotmentClauseSrcContext) Portion() IPortionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPortionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPortionContext)
}

func (s *AllotmentClauseSrcContext) FROM() antlr.TerminalNode {
	return s.GetToken(NumscriptParserFROM, 0)
}

func (s *AllotmentClauseSrcContext) Source() ISourceContext {
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

func (s *AllotmentClauseSrcContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AllotmentClauseSrcContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AllotmentClauseSrcContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterAllotmentClauseSrc(s)
	}
}

func (s *AllotmentClauseSrcContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitAllotmentClauseSrc(s)
	}
}

func (s *AllotmentClauseSrcContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitAllotmentClauseSrc(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) AllotmentClauseSrc() (localctx IAllotmentClauseSrcContext) {
	localctx = NewAllotmentClauseSrcContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, NumscriptParserRULE_allotmentClauseSrc)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(84)
		p.Portion()
	}
	{
		p.SetState(85)
		p.Match(NumscriptParserFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(86)
		p.Source()
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

// IDestinationContext is an interface to support dynamic dispatch.
type IDestinationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsDestinationContext differentiates from other interfaces.
	IsDestinationContext()
}

type DestinationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDestinationContext() *DestinationContext {
	var p = new(DestinationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_destination
	return p
}

func InitEmptyDestinationContext(p *DestinationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_destination
}

func (*DestinationContext) IsDestinationContext() {}

func NewDestinationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DestinationContext {
	var p = new(DestinationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_destination

	return p
}

func (s *DestinationContext) GetParser() antlr.Parser { return s.parser }

func (s *DestinationContext) CopyAll(ctx *DestinationContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *DestinationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestinationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type DestVariableContext struct {
	DestinationContext
}

func NewDestVariableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestVariableContext {
	var p = new(DestVariableContext)

	InitEmptyDestinationContext(&p.DestinationContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestinationContext))

	return p
}

func (s *DestVariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestVariableContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARIABLE_NAME, 0)
}

func (s *DestVariableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestVariable(s)
	}
}

func (s *DestVariableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestVariable(s)
	}
}

func (s *DestVariableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitDestVariable(s)

	default:
		return t.VisitChildren(s)
	}
}

type DestAccountContext struct {
	DestinationContext
}

func NewDestAccountContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestAccountContext {
	var p = new(DestAccountContext)

	InitEmptyDestinationContext(&p.DestinationContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestinationContext))

	return p
}

func (s *DestAccountContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestAccountContext) ACCOUNT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserACCOUNT, 0)
}

func (s *DestAccountContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestAccount(s)
	}
}

func (s *DestAccountContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestAccount(s)
	}
}

func (s *DestAccountContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitDestAccount(s)

	default:
		return t.VisitChildren(s)
	}
}

type DestAllotmentContext struct {
	DestinationContext
}

func NewDestAllotmentContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestAllotmentContext {
	var p = new(DestAllotmentContext)

	InitEmptyDestinationContext(&p.DestinationContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestinationContext))

	return p
}

func (s *DestAllotmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestAllotmentContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACE, 0)
}

func (s *DestAllotmentContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACE, 0)
}

func (s *DestAllotmentContext) AllAllotmentClauseDest() []IAllotmentClauseDestContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAllotmentClauseDestContext); ok {
			len++
		}
	}

	tst := make([]IAllotmentClauseDestContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAllotmentClauseDestContext); ok {
			tst[i] = t.(IAllotmentClauseDestContext)
			i++
		}
	}

	return tst
}

func (s *DestAllotmentContext) AllotmentClauseDest(i int) IAllotmentClauseDestContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAllotmentClauseDestContext); ok {
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

	return t.(IAllotmentClauseDestContext)
}

func (s *DestAllotmentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestAllotment(s)
	}
}

func (s *DestAllotmentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestAllotment(s)
	}
}

func (s *DestAllotmentContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitDestAllotment(s)

	default:
		return t.VisitChildren(s)
	}
}

type DestSeqContext struct {
	DestinationContext
}

func NewDestSeqContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestSeqContext {
	var p = new(DestSeqContext)

	InitEmptyDestinationContext(&p.DestinationContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestinationContext))

	return p
}

func (s *DestSeqContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestSeqContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACE, 0)
}

func (s *DestSeqContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACE, 0)
}

func (s *DestSeqContext) AllDestination() []IDestinationContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IDestinationContext); ok {
			len++
		}
	}

	tst := make([]IDestinationContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IDestinationContext); ok {
			tst[i] = t.(IDestinationContext)
			i++
		}
	}

	return tst
}

func (s *DestSeqContext) Destination(i int) IDestinationContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDestinationContext); ok {
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

	return t.(IDestinationContext)
}

func (s *DestSeqContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestSeq(s)
	}
}

func (s *DestSeqContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestSeq(s)
	}
}

func (s *DestSeqContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitDestSeq(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) Destination() (localctx IDestinationContext) {
	localctx = NewDestinationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, NumscriptParserRULE_destination)
	var _la int

	p.SetState(106)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext()) {
	case 1:
		localctx = NewDestAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(88)
			p.Match(NumscriptParserACCOUNT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewDestVariableContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(89)
			p.Match(NumscriptParserVARIABLE_NAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewDestAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(90)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(92)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == NumscriptParserRATIO_PORTION_LITERAL || _la == NumscriptParserPERCENTAGE_PORTION_LITERAL {
			{
				p.SetState(91)
				p.AllotmentClauseDest()
			}

			p.SetState(94)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(96)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		localctx = NewDestSeqContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(98)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(102)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&25231360) != 0 {
			{
				p.SetState(99)
				p.Destination()
			}

			p.SetState(104)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(105)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
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

// IAllotmentClauseDestContext is an interface to support dynamic dispatch.
type IAllotmentClauseDestContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Portion() IPortionContext
	TO() antlr.TerminalNode
	Destination() IDestinationContext

	// IsAllotmentClauseDestContext differentiates from other interfaces.
	IsAllotmentClauseDestContext()
}

type AllotmentClauseDestContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAllotmentClauseDestContext() *AllotmentClauseDestContext {
	var p = new(AllotmentClauseDestContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_allotmentClauseDest
	return p
}

func InitEmptyAllotmentClauseDestContext(p *AllotmentClauseDestContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_allotmentClauseDest
}

func (*AllotmentClauseDestContext) IsAllotmentClauseDestContext() {}

func NewAllotmentClauseDestContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AllotmentClauseDestContext {
	var p = new(AllotmentClauseDestContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_allotmentClauseDest

	return p
}

func (s *AllotmentClauseDestContext) GetParser() antlr.Parser { return s.parser }

func (s *AllotmentClauseDestContext) Portion() IPortionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPortionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPortionContext)
}

func (s *AllotmentClauseDestContext) TO() antlr.TerminalNode {
	return s.GetToken(NumscriptParserTO, 0)
}

func (s *AllotmentClauseDestContext) Destination() IDestinationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDestinationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDestinationContext)
}

func (s *AllotmentClauseDestContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AllotmentClauseDestContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AllotmentClauseDestContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterAllotmentClauseDest(s)
	}
}

func (s *AllotmentClauseDestContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitAllotmentClauseDest(s)
	}
}

func (s *AllotmentClauseDestContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitAllotmentClauseDest(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) AllotmentClauseDest() (localctx IAllotmentClauseDestContext) {
	localctx = NewAllotmentClauseDestContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, NumscriptParserRULE_allotmentClauseDest)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(108)
		p.Portion()
	}
	{
		p.SetState(109)
		p.Match(NumscriptParserTO)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(110)
		p.Destination()
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

// ISendExprContext is an interface to support dynamic dispatch.
type ISendExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsSendExprContext differentiates from other interfaces.
	IsSendExprContext()
}

type SendExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySendExprContext() *SendExprContext {
	var p = new(SendExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_sendExpr
	return p
}

func InitEmptySendExprContext(p *SendExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_sendExpr
}

func (*SendExprContext) IsSendExprContext() {}

func NewSendExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SendExprContext {
	var p = new(SendExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_sendExpr

	return p
}

func (s *SendExprContext) GetParser() antlr.Parser { return s.parser }

func (s *SendExprContext) CopyAll(ctx *SendExprContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *SendExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SendExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type SendVariableContext struct {
	SendExprContext
}

func NewSendVariableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SendVariableContext {
	var p = new(SendVariableContext)

	InitEmptySendExprContext(&p.SendExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*SendExprContext))

	return p
}

func (s *SendVariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SendVariableContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARIABLE_NAME, 0)
}

func (s *SendVariableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSendVariable(s)
	}
}

func (s *SendVariableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSendVariable(s)
	}
}

func (s *SendVariableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitSendVariable(s)

	default:
		return t.VisitChildren(s)
	}
}

type SendMonContext struct {
	SendExprContext
}

func NewSendMonContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SendMonContext {
	var p = new(SendMonContext)

	InitEmptySendExprContext(&p.SendExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*SendExprContext))

	return p
}

func (s *SendMonContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SendMonContext) MonetaryLit() IMonetaryLitContext {
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

func (s *SendMonContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSendMon(s)
	}
}

func (s *SendMonContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSendMon(s)
	}
}

func (s *SendMonContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NumscriptVisitor:
		return t.VisitSendMon(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NumscriptParser) SendExpr() (localctx ISendExprContext) {
	localctx = NewSendExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, NumscriptParserRULE_sendExpr)
	p.SetState(114)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserLBRACKET:
		localctx = NewSendMonContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(112)
			p.MonetaryLit()
		}

	case NumscriptParserVARIABLE_NAME:
		localctx = NewSendVariableContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(113)
			p.Match(NumscriptParserVARIABLE_NAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
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
	SendExpr() ISendExprContext
	LPARENS() antlr.TerminalNode
	SOURCE() antlr.TerminalNode
	AllEQ() []antlr.TerminalNode
	EQ(i int) antlr.TerminalNode
	Source() ISourceContext
	DESTINATION() antlr.TerminalNode
	Destination() IDestinationContext
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

func (s *StatementContext) SendExpr() ISendExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISendExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISendExprContext)
}

func (s *StatementContext) LPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLPARENS, 0)
}

func (s *StatementContext) SOURCE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserSOURCE, 0)
}

func (s *StatementContext) AllEQ() []antlr.TerminalNode {
	return s.GetTokens(NumscriptParserEQ)
}

func (s *StatementContext) EQ(i int) antlr.TerminalNode {
	return s.GetToken(NumscriptParserEQ, i)
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

func (s *StatementContext) DESTINATION() antlr.TerminalNode {
	return s.GetToken(NumscriptParserDESTINATION, 0)
}

func (s *StatementContext) Destination() IDestinationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDestinationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDestinationContext)
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
	p.EnterRule(localctx, 22, NumscriptParserRULE_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(116)
		p.Match(NumscriptParserSEND)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(117)
		p.SendExpr()
	}
	{
		p.SetState(118)
		p.Match(NumscriptParserLPARENS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(119)
		p.Match(NumscriptParserSOURCE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(120)
		p.Match(NumscriptParserEQ)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(121)
		p.Source()
	}
	{
		p.SetState(122)
		p.Match(NumscriptParserDESTINATION)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(123)
		p.Match(NumscriptParserEQ)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(124)
		p.Destination()
	}
	{
		p.SetState(125)
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
