// Code generated from IR.g4 by ANTLR 4.13.2. DO NOT EDIT.

package antlrParser // IR
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

type IRParser struct {
	*antlr.BaseParser
}

var IRParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func irParserInit() {
	staticData := &IRParserStaticData
	staticData.LiteralNames = []string{
		"", "':'", "", "", "", "", "", "", "", "", "'('", "')'", "'['", "']'",
		"','", "'='", "'+'", "'-'", "'+='", "'-='", "'<'", "'>'", "'_'",
	}
	staticData.SymbolicNames = []string{
		"", "", "WS", "NEWLINE", "TYPE_KEYWORD", "REG", "LABEL", "INT", "STRING",
		"IDENTIFIER", "LPAREN", "RPAREN", "LBRACKET", "RBRACKET", "COMMA", "EQ",
		"PLUS", "MINUS", "PLUS_EQ", "MINUS_EQ", "LT", "GT", "UNDERSCORE",
	}
	staticData.RuleNames = []string{
		"program", "line", "labelMarker", "instruction", "dest", "regList",
		"instrCall", "instrName", "typeName", "args", "arg", "value", "const_",
		"reg",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 22, 125, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 1, 0, 5, 0, 30, 8, 0, 10,
		0, 12, 0, 33, 9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 3, 1, 39, 8, 1, 1, 2, 1, 2,
		1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3,
		1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 62, 8, 3, 1, 4, 1, 4, 1,
		4, 1, 4, 1, 4, 1, 4, 3, 4, 70, 8, 4, 1, 5, 1, 5, 1, 5, 5, 5, 75, 8, 5,
		10, 5, 12, 5, 78, 9, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7,
		1, 7, 1, 7, 3, 7, 90, 8, 7, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 5, 9, 97, 8,
		9, 10, 9, 12, 9, 100, 9, 9, 3, 9, 102, 8, 9, 1, 10, 1, 10, 1, 10, 1, 10,
		3, 10, 108, 8, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 3,
		11, 117, 8, 11, 1, 12, 1, 12, 3, 12, 121, 8, 12, 1, 13, 1, 13, 1, 13, 0,
		0, 14, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 0, 3, 1, 0, 16,
		17, 1, 0, 18, 19, 2, 0, 4, 4, 9, 9, 127, 0, 31, 1, 0, 0, 0, 2, 38, 1, 0,
		0, 0, 4, 40, 1, 0, 0, 0, 6, 61, 1, 0, 0, 0, 8, 69, 1, 0, 0, 0, 10, 71,
		1, 0, 0, 0, 12, 79, 1, 0, 0, 0, 14, 84, 1, 0, 0, 0, 16, 91, 1, 0, 0, 0,
		18, 101, 1, 0, 0, 0, 20, 107, 1, 0, 0, 0, 22, 116, 1, 0, 0, 0, 24, 120,
		1, 0, 0, 0, 26, 122, 1, 0, 0, 0, 28, 30, 3, 2, 1, 0, 29, 28, 1, 0, 0, 0,
		30, 33, 1, 0, 0, 0, 31, 29, 1, 0, 0, 0, 31, 32, 1, 0, 0, 0, 32, 34, 1,
		0, 0, 0, 33, 31, 1, 0, 0, 0, 34, 35, 5, 0, 0, 1, 35, 1, 1, 0, 0, 0, 36,
		39, 3, 4, 2, 0, 37, 39, 3, 6, 3, 0, 38, 36, 1, 0, 0, 0, 38, 37, 1, 0, 0,
		0, 39, 3, 1, 0, 0, 0, 40, 41, 5, 6, 0, 0, 41, 5, 1, 0, 0, 0, 42, 43, 3,
		8, 4, 0, 43, 44, 5, 15, 0, 0, 44, 45, 3, 12, 6, 0, 45, 62, 1, 0, 0, 0,
		46, 62, 3, 12, 6, 0, 47, 48, 3, 8, 4, 0, 48, 49, 5, 15, 0, 0, 49, 50, 3,
		24, 12, 0, 50, 62, 1, 0, 0, 0, 51, 52, 3, 8, 4, 0, 52, 53, 5, 15, 0, 0,
		53, 54, 3, 26, 13, 0, 54, 55, 7, 0, 0, 0, 55, 56, 3, 26, 13, 0, 56, 62,
		1, 0, 0, 0, 57, 58, 3, 26, 13, 0, 58, 59, 7, 1, 0, 0, 59, 60, 3, 26, 13,
		0, 60, 62, 1, 0, 0, 0, 61, 42, 1, 0, 0, 0, 61, 46, 1, 0, 0, 0, 61, 47,
		1, 0, 0, 0, 61, 51, 1, 0, 0, 0, 61, 57, 1, 0, 0, 0, 62, 7, 1, 0, 0, 0,
		63, 70, 3, 26, 13, 0, 64, 70, 5, 22, 0, 0, 65, 66, 5, 12, 0, 0, 66, 67,
		3, 10, 5, 0, 67, 68, 5, 13, 0, 0, 68, 70, 1, 0, 0, 0, 69, 63, 1, 0, 0,
		0, 69, 64, 1, 0, 0, 0, 69, 65, 1, 0, 0, 0, 70, 9, 1, 0, 0, 0, 71, 76, 3,
		26, 13, 0, 72, 73, 5, 14, 0, 0, 73, 75, 3, 26, 13, 0, 74, 72, 1, 0, 0,
		0, 75, 78, 1, 0, 0, 0, 76, 74, 1, 0, 0, 0, 76, 77, 1, 0, 0, 0, 77, 11,
		1, 0, 0, 0, 78, 76, 1, 0, 0, 0, 79, 80, 3, 14, 7, 0, 80, 81, 5, 10, 0,
		0, 81, 82, 3, 18, 9, 0, 82, 83, 5, 11, 0, 0, 83, 13, 1, 0, 0, 0, 84, 89,
		5, 9, 0, 0, 85, 86, 5, 20, 0, 0, 86, 87, 3, 16, 8, 0, 87, 88, 5, 21, 0,
		0, 88, 90, 1, 0, 0, 0, 89, 85, 1, 0, 0, 0, 89, 90, 1, 0, 0, 0, 90, 15,
		1, 0, 0, 0, 91, 92, 7, 2, 0, 0, 92, 17, 1, 0, 0, 0, 93, 98, 3, 20, 10,
		0, 94, 95, 5, 14, 0, 0, 95, 97, 3, 20, 10, 0, 96, 94, 1, 0, 0, 0, 97, 100,
		1, 0, 0, 0, 98, 96, 1, 0, 0, 0, 98, 99, 1, 0, 0, 0, 99, 102, 1, 0, 0, 0,
		100, 98, 1, 0, 0, 0, 101, 93, 1, 0, 0, 0, 101, 102, 1, 0, 0, 0, 102, 19,
		1, 0, 0, 0, 103, 108, 3, 22, 11, 0, 104, 105, 5, 9, 0, 0, 105, 106, 5,
		1, 0, 0, 106, 108, 3, 22, 11, 0, 107, 103, 1, 0, 0, 0, 107, 104, 1, 0,
		0, 0, 108, 21, 1, 0, 0, 0, 109, 117, 3, 26, 13, 0, 110, 117, 5, 6, 0, 0,
		111, 117, 5, 7, 0, 0, 112, 113, 5, 12, 0, 0, 113, 114, 3, 10, 5, 0, 114,
		115, 5, 13, 0, 0, 115, 117, 1, 0, 0, 0, 116, 109, 1, 0, 0, 0, 116, 110,
		1, 0, 0, 0, 116, 111, 1, 0, 0, 0, 116, 112, 1, 0, 0, 0, 117, 23, 1, 0,
		0, 0, 118, 121, 5, 8, 0, 0, 119, 121, 5, 7, 0, 0, 120, 118, 1, 0, 0, 0,
		120, 119, 1, 0, 0, 0, 121, 25, 1, 0, 0, 0, 122, 123, 5, 5, 0, 0, 123, 27,
		1, 0, 0, 0, 11, 31, 38, 61, 69, 76, 89, 98, 101, 107, 116, 120,
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

// IRParserInit initializes any static state used to implement IRParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewIRParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func IRParserInit() {
	staticData := &IRParserStaticData
	staticData.once.Do(irParserInit)
}

// NewIRParser produces a new parser instance for the optional input antlr.TokenStream.
func NewIRParser(input antlr.TokenStream) *IRParser {
	IRParserInit()
	this := new(IRParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &IRParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "IR.g4"

	return this
}

// IRParser tokens.
const (
	IRParserEOF          = antlr.TokenEOF
	IRParserT__0         = 1
	IRParserWS           = 2
	IRParserNEWLINE      = 3
	IRParserTYPE_KEYWORD = 4
	IRParserREG          = 5
	IRParserLABEL        = 6
	IRParserINT          = 7
	IRParserSTRING       = 8
	IRParserIDENTIFIER   = 9
	IRParserLPAREN       = 10
	IRParserRPAREN       = 11
	IRParserLBRACKET     = 12
	IRParserRBRACKET     = 13
	IRParserCOMMA        = 14
	IRParserEQ           = 15
	IRParserPLUS         = 16
	IRParserMINUS        = 17
	IRParserPLUS_EQ      = 18
	IRParserMINUS_EQ     = 19
	IRParserLT           = 20
	IRParserGT           = 21
	IRParserUNDERSCORE   = 22
)

// IRParser rules.
const (
	IRParserRULE_program     = 0
	IRParserRULE_line        = 1
	IRParserRULE_labelMarker = 2
	IRParserRULE_instruction = 3
	IRParserRULE_dest        = 4
	IRParserRULE_regList     = 5
	IRParserRULE_instrCall   = 6
	IRParserRULE_instrName   = 7
	IRParserRULE_typeName    = 8
	IRParserRULE_args        = 9
	IRParserRULE_arg         = 10
	IRParserRULE_value       = 11
	IRParserRULE_const_      = 12
	IRParserRULE_reg         = 13
)

// IProgramContext is an interface to support dynamic dispatch.
type IProgramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllLine() []ILineContext
	Line(i int) ILineContext

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
	p.RuleIndex = IRParserRULE_program
	return p
}

func InitEmptyProgramContext(p *ProgramContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_program
}

func (*ProgramContext) IsProgramContext() {}

func NewProgramContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ProgramContext {
	var p = new(ProgramContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_program

	return p
}

func (s *ProgramContext) GetParser() antlr.Parser { return s.parser }

func (s *ProgramContext) EOF() antlr.TerminalNode {
	return s.GetToken(IRParserEOF, 0)
}

func (s *ProgramContext) AllLine() []ILineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILineContext); ok {
			len++
		}
	}

	tst := make([]ILineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILineContext); ok {
			tst[i] = t.(ILineContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Line(i int) ILineContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILineContext); ok {
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

	return t.(ILineContext)
}

func (s *ProgramContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ProgramContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ProgramContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterProgram(s)
	}
}

func (s *ProgramContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitProgram(s)
	}
}

func (p *IRParser) Program() (localctx IProgramContext) {
	localctx = NewProgramContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, IRParserRULE_program)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(31)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&4199008) != 0 {
		{
			p.SetState(28)
			p.Line()
		}

		p.SetState(33)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(34)
		p.Match(IRParserEOF)
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

// ILineContext is an interface to support dynamic dispatch.
type ILineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LabelMarker() ILabelMarkerContext
	Instruction() IInstructionContext

	// IsLineContext differentiates from other interfaces.
	IsLineContext()
}

type LineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLineContext() *LineContext {
	var p = new(LineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_line
	return p
}

func InitEmptyLineContext(p *LineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_line
}

func (*LineContext) IsLineContext() {}

func NewLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LineContext {
	var p = new(LineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_line

	return p
}

func (s *LineContext) GetParser() antlr.Parser { return s.parser }

func (s *LineContext) LabelMarker() ILabelMarkerContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILabelMarkerContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILabelMarkerContext)
}

func (s *LineContext) Instruction() IInstructionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInstructionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInstructionContext)
}

func (s *LineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterLine(s)
	}
}

func (s *LineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitLine(s)
	}
}

func (p *IRParser) Line() (localctx ILineContext) {
	localctx = NewLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, IRParserRULE_line)
	p.SetState(38)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case IRParserLABEL:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(36)
			p.LabelMarker()
		}

	case IRParserREG, IRParserIDENTIFIER, IRParserLBRACKET, IRParserUNDERSCORE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(37)
			p.Instruction()
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

// ILabelMarkerContext is an interface to support dynamic dispatch.
type ILabelMarkerContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LABEL() antlr.TerminalNode

	// IsLabelMarkerContext differentiates from other interfaces.
	IsLabelMarkerContext()
}

type LabelMarkerContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLabelMarkerContext() *LabelMarkerContext {
	var p = new(LabelMarkerContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_labelMarker
	return p
}

func InitEmptyLabelMarkerContext(p *LabelMarkerContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_labelMarker
}

func (*LabelMarkerContext) IsLabelMarkerContext() {}

func NewLabelMarkerContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LabelMarkerContext {
	var p = new(LabelMarkerContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_labelMarker

	return p
}

func (s *LabelMarkerContext) GetParser() antlr.Parser { return s.parser }

func (s *LabelMarkerContext) LABEL() antlr.TerminalNode {
	return s.GetToken(IRParserLABEL, 0)
}

func (s *LabelMarkerContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LabelMarkerContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LabelMarkerContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterLabelMarker(s)
	}
}

func (s *LabelMarkerContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitLabelMarker(s)
	}
}

func (p *IRParser) LabelMarker() (localctx ILabelMarkerContext) {
	localctx = NewLabelMarkerContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, IRParserRULE_labelMarker)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(40)
		p.Match(IRParserLABEL)
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

// IInstructionContext is an interface to support dynamic dispatch.
type IInstructionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsInstructionContext differentiates from other interfaces.
	IsInstructionContext()
}

type InstructionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInstructionContext() *InstructionContext {
	var p = new(InstructionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_instruction
	return p
}

func InitEmptyInstructionContext(p *InstructionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_instruction
}

func (*InstructionContext) IsInstructionContext() {}

func NewInstructionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InstructionContext {
	var p = new(InstructionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_instruction

	return p
}

func (s *InstructionContext) GetParser() antlr.Parser { return s.parser }

func (s *InstructionContext) CopyAll(ctx *InstructionContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *InstructionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InstructionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type InstrNoDestContext struct {
	InstructionContext
}

func NewInstrNoDestContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InstrNoDestContext {
	var p = new(InstrNoDestContext)

	InitEmptyInstructionContext(&p.InstructionContext)
	p.parser = parser
	p.CopyAll(ctx.(*InstructionContext))

	return p
}

func (s *InstrNoDestContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InstrNoDestContext) InstrCall() IInstrCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInstrCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInstrCallContext)
}

func (s *InstrNoDestContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterInstrNoDest(s)
	}
}

func (s *InstrNoDestContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitInstrNoDest(s)
	}
}

type CompoundAssignInstrContext struct {
	InstructionContext
	left  IRegContext
	op    antlr.Token
	right IRegContext
}

func NewCompoundAssignInstrContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CompoundAssignInstrContext {
	var p = new(CompoundAssignInstrContext)

	InitEmptyInstructionContext(&p.InstructionContext)
	p.parser = parser
	p.CopyAll(ctx.(*InstructionContext))

	return p
}

func (s *CompoundAssignInstrContext) GetOp() antlr.Token { return s.op }

func (s *CompoundAssignInstrContext) SetOp(v antlr.Token) { s.op = v }

func (s *CompoundAssignInstrContext) GetLeft() IRegContext { return s.left }

func (s *CompoundAssignInstrContext) GetRight() IRegContext { return s.right }

func (s *CompoundAssignInstrContext) SetLeft(v IRegContext) { s.left = v }

func (s *CompoundAssignInstrContext) SetRight(v IRegContext) { s.right = v }

func (s *CompoundAssignInstrContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CompoundAssignInstrContext) AllReg() []IRegContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRegContext); ok {
			len++
		}
	}

	tst := make([]IRegContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRegContext); ok {
			tst[i] = t.(IRegContext)
			i++
		}
	}

	return tst
}

func (s *CompoundAssignInstrContext) Reg(i int) IRegContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRegContext); ok {
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

	return t.(IRegContext)
}

func (s *CompoundAssignInstrContext) PLUS_EQ() antlr.TerminalNode {
	return s.GetToken(IRParserPLUS_EQ, 0)
}

func (s *CompoundAssignInstrContext) MINUS_EQ() antlr.TerminalNode {
	return s.GetToken(IRParserMINUS_EQ, 0)
}

func (s *CompoundAssignInstrContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterCompoundAssignInstr(s)
	}
}

func (s *CompoundAssignInstrContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitCompoundAssignInstr(s)
	}
}

type ConstAssignContext struct {
	InstructionContext
}

func NewConstAssignContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConstAssignContext {
	var p = new(ConstAssignContext)

	InitEmptyInstructionContext(&p.InstructionContext)
	p.parser = parser
	p.CopyAll(ctx.(*InstructionContext))

	return p
}

func (s *ConstAssignContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConstAssignContext) Dest() IDestContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDestContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDestContext)
}

func (s *ConstAssignContext) EQ() antlr.TerminalNode {
	return s.GetToken(IRParserEQ, 0)
}

func (s *ConstAssignContext) Const_() IConst_Context {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConst_Context); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConst_Context)
}

func (s *ConstAssignContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterConstAssign(s)
	}
}

func (s *ConstAssignContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitConstAssign(s)
	}
}

type InstrWithDestContext struct {
	InstructionContext
}

func NewInstrWithDestContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InstrWithDestContext {
	var p = new(InstrWithDestContext)

	InitEmptyInstructionContext(&p.InstructionContext)
	p.parser = parser
	p.CopyAll(ctx.(*InstructionContext))

	return p
}

func (s *InstrWithDestContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InstrWithDestContext) Dest() IDestContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDestContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDestContext)
}

func (s *InstrWithDestContext) EQ() antlr.TerminalNode {
	return s.GetToken(IRParserEQ, 0)
}

func (s *InstrWithDestContext) InstrCall() IInstrCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInstrCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInstrCallContext)
}

func (s *InstrWithDestContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterInstrWithDest(s)
	}
}

func (s *InstrWithDestContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitInstrWithDest(s)
	}
}

type InfixInstrContext struct {
	InstructionContext
	left  IRegContext
	op    antlr.Token
	right IRegContext
}

func NewInfixInstrContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InfixInstrContext {
	var p = new(InfixInstrContext)

	InitEmptyInstructionContext(&p.InstructionContext)
	p.parser = parser
	p.CopyAll(ctx.(*InstructionContext))

	return p
}

func (s *InfixInstrContext) GetOp() antlr.Token { return s.op }

func (s *InfixInstrContext) SetOp(v antlr.Token) { s.op = v }

func (s *InfixInstrContext) GetLeft() IRegContext { return s.left }

func (s *InfixInstrContext) GetRight() IRegContext { return s.right }

func (s *InfixInstrContext) SetLeft(v IRegContext) { s.left = v }

func (s *InfixInstrContext) SetRight(v IRegContext) { s.right = v }

func (s *InfixInstrContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InfixInstrContext) Dest() IDestContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDestContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDestContext)
}

func (s *InfixInstrContext) EQ() antlr.TerminalNode {
	return s.GetToken(IRParserEQ, 0)
}

func (s *InfixInstrContext) AllReg() []IRegContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRegContext); ok {
			len++
		}
	}

	tst := make([]IRegContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRegContext); ok {
			tst[i] = t.(IRegContext)
			i++
		}
	}

	return tst
}

func (s *InfixInstrContext) Reg(i int) IRegContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRegContext); ok {
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

	return t.(IRegContext)
}

func (s *InfixInstrContext) PLUS() antlr.TerminalNode {
	return s.GetToken(IRParserPLUS, 0)
}

func (s *InfixInstrContext) MINUS() antlr.TerminalNode {
	return s.GetToken(IRParserMINUS, 0)
}

func (s *InfixInstrContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterInfixInstr(s)
	}
}

func (s *InfixInstrContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitInfixInstr(s)
	}
}

func (p *IRParser) Instruction() (localctx IInstructionContext) {
	localctx = NewInstructionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, IRParserRULE_instruction)
	var _la int

	p.SetState(61)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext()) {
	case 1:
		localctx = NewInstrWithDestContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(42)
			p.Dest()
		}
		{
			p.SetState(43)
			p.Match(IRParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(44)
			p.InstrCall()
		}

	case 2:
		localctx = NewInstrNoDestContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(46)
			p.InstrCall()
		}

	case 3:
		localctx = NewConstAssignContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(47)
			p.Dest()
		}
		{
			p.SetState(48)
			p.Match(IRParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(49)
			p.Const_()
		}

	case 4:
		localctx = NewInfixInstrContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(51)
			p.Dest()
		}
		{
			p.SetState(52)
			p.Match(IRParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(53)

			var _x = p.Reg()

			localctx.(*InfixInstrContext).left = _x
		}
		{
			p.SetState(54)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*InfixInstrContext).op = _lt

			_la = p.GetTokenStream().LA(1)

			if !(_la == IRParserPLUS || _la == IRParserMINUS) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*InfixInstrContext).op = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(55)

			var _x = p.Reg()

			localctx.(*InfixInstrContext).right = _x
		}

	case 5:
		localctx = NewCompoundAssignInstrContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(57)

			var _x = p.Reg()

			localctx.(*CompoundAssignInstrContext).left = _x
		}
		{
			p.SetState(58)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*CompoundAssignInstrContext).op = _lt

			_la = p.GetTokenStream().LA(1)

			if !(_la == IRParserPLUS_EQ || _la == IRParserMINUS_EQ) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*CompoundAssignInstrContext).op = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(59)

			var _x = p.Reg()

			localctx.(*CompoundAssignInstrContext).right = _x
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

// IDestContext is an interface to support dynamic dispatch.
type IDestContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsDestContext differentiates from other interfaces.
	IsDestContext()
}

type DestContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDestContext() *DestContext {
	var p = new(DestContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_dest
	return p
}

func InitEmptyDestContext(p *DestContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_dest
}

func (*DestContext) IsDestContext() {}

func NewDestContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DestContext {
	var p = new(DestContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_dest

	return p
}

func (s *DestContext) GetParser() antlr.Parser { return s.parser }

func (s *DestContext) CopyAll(ctx *DestContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *DestContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type DestRegContext struct {
	DestContext
}

func NewDestRegContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestRegContext {
	var p = new(DestRegContext)

	InitEmptyDestContext(&p.DestContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestContext))

	return p
}

func (s *DestRegContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestRegContext) Reg() IRegContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRegContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRegContext)
}

func (s *DestRegContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterDestReg(s)
	}
}

func (s *DestRegContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitDestReg(s)
	}
}

type DestDiscardContext struct {
	DestContext
}

func NewDestDiscardContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestDiscardContext {
	var p = new(DestDiscardContext)

	InitEmptyDestContext(&p.DestContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestContext))

	return p
}

func (s *DestDiscardContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestDiscardContext) UNDERSCORE() antlr.TerminalNode {
	return s.GetToken(IRParserUNDERSCORE, 0)
}

func (s *DestDiscardContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterDestDiscard(s)
	}
}

func (s *DestDiscardContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitDestDiscard(s)
	}
}

type DestListContext struct {
	DestContext
}

func NewDestListContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestListContext {
	var p = new(DestListContext)

	InitEmptyDestContext(&p.DestContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestContext))

	return p
}

func (s *DestListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestListContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(IRParserLBRACKET, 0)
}

func (s *DestListContext) RegList() IRegListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRegListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRegListContext)
}

func (s *DestListContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(IRParserRBRACKET, 0)
}

func (s *DestListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterDestList(s)
	}
}

func (s *DestListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitDestList(s)
	}
}

func (p *IRParser) Dest() (localctx IDestContext) {
	localctx = NewDestContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, IRParserRULE_dest)
	p.SetState(69)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case IRParserREG:
		localctx = NewDestRegContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(63)
			p.Reg()
		}

	case IRParserUNDERSCORE:
		localctx = NewDestDiscardContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(64)
			p.Match(IRParserUNDERSCORE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case IRParserLBRACKET:
		localctx = NewDestListContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(65)
			p.Match(IRParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(66)
			p.RegList()
		}
		{
			p.SetState(67)
			p.Match(IRParserRBRACKET)
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

// IRegListContext is an interface to support dynamic dispatch.
type IRegListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllReg() []IRegContext
	Reg(i int) IRegContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsRegListContext differentiates from other interfaces.
	IsRegListContext()
}

type RegListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRegListContext() *RegListContext {
	var p = new(RegListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_regList
	return p
}

func InitEmptyRegListContext(p *RegListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_regList
}

func (*RegListContext) IsRegListContext() {}

func NewRegListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RegListContext {
	var p = new(RegListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_regList

	return p
}

func (s *RegListContext) GetParser() antlr.Parser { return s.parser }

func (s *RegListContext) AllReg() []IRegContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRegContext); ok {
			len++
		}
	}

	tst := make([]IRegContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRegContext); ok {
			tst[i] = t.(IRegContext)
			i++
		}
	}

	return tst
}

func (s *RegListContext) Reg(i int) IRegContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRegContext); ok {
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

	return t.(IRegContext)
}

func (s *RegListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(IRParserCOMMA)
}

func (s *RegListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(IRParserCOMMA, i)
}

func (s *RegListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RegListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RegListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterRegList(s)
	}
}

func (s *RegListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitRegList(s)
	}
}

func (p *IRParser) RegList() (localctx IRegListContext) {
	localctx = NewRegListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, IRParserRULE_regList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(71)
		p.Reg()
	}
	p.SetState(76)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == IRParserCOMMA {
		{
			p.SetState(72)
			p.Match(IRParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(73)
			p.Reg()
		}

		p.SetState(78)
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

// IInstrCallContext is an interface to support dynamic dispatch.
type IInstrCallContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	InstrName() IInstrNameContext
	LPAREN() antlr.TerminalNode
	Args() IArgsContext
	RPAREN() antlr.TerminalNode

	// IsInstrCallContext differentiates from other interfaces.
	IsInstrCallContext()
}

type InstrCallContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInstrCallContext() *InstrCallContext {
	var p = new(InstrCallContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_instrCall
	return p
}

func InitEmptyInstrCallContext(p *InstrCallContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_instrCall
}

func (*InstrCallContext) IsInstrCallContext() {}

func NewInstrCallContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InstrCallContext {
	var p = new(InstrCallContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_instrCall

	return p
}

func (s *InstrCallContext) GetParser() antlr.Parser { return s.parser }

func (s *InstrCallContext) InstrName() IInstrNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInstrNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInstrNameContext)
}

func (s *InstrCallContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(IRParserLPAREN, 0)
}

func (s *InstrCallContext) Args() IArgsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArgsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArgsContext)
}

func (s *InstrCallContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(IRParserRPAREN, 0)
}

func (s *InstrCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InstrCallContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InstrCallContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterInstrCall(s)
	}
}

func (s *InstrCallContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitInstrCall(s)
	}
}

func (p *IRParser) InstrCall() (localctx IInstrCallContext) {
	localctx = NewInstrCallContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, IRParserRULE_instrCall)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(79)
		p.InstrName()
	}
	{
		p.SetState(80)
		p.Match(IRParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(81)
		p.Args()
	}
	{
		p.SetState(82)
		p.Match(IRParserRPAREN)
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

// IInstrNameContext is an interface to support dynamic dispatch.
type IInstrNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	LT() antlr.TerminalNode
	TypeName() ITypeNameContext
	GT() antlr.TerminalNode

	// IsInstrNameContext differentiates from other interfaces.
	IsInstrNameContext()
}

type InstrNameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInstrNameContext() *InstrNameContext {
	var p = new(InstrNameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_instrName
	return p
}

func InitEmptyInstrNameContext(p *InstrNameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_instrName
}

func (*InstrNameContext) IsInstrNameContext() {}

func NewInstrNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InstrNameContext {
	var p = new(InstrNameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_instrName

	return p
}

func (s *InstrNameContext) GetParser() antlr.Parser { return s.parser }

func (s *InstrNameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(IRParserIDENTIFIER, 0)
}

func (s *InstrNameContext) LT() antlr.TerminalNode {
	return s.GetToken(IRParserLT, 0)
}

func (s *InstrNameContext) TypeName() ITypeNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeNameContext)
}

func (s *InstrNameContext) GT() antlr.TerminalNode {
	return s.GetToken(IRParserGT, 0)
}

func (s *InstrNameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InstrNameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InstrNameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterInstrName(s)
	}
}

func (s *InstrNameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitInstrName(s)
	}
}

func (p *IRParser) InstrName() (localctx IInstrNameContext) {
	localctx = NewInstrNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, IRParserRULE_instrName)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(84)
		p.Match(IRParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(89)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == IRParserLT {
		{
			p.SetState(85)
			p.Match(IRParserLT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(86)
			p.TypeName()
		}
		{
			p.SetState(87)
			p.Match(IRParserGT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
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

// ITypeNameContext is an interface to support dynamic dispatch.
type ITypeNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TYPE_KEYWORD() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode

	// IsTypeNameContext differentiates from other interfaces.
	IsTypeNameContext()
}

type TypeNameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTypeNameContext() *TypeNameContext {
	var p = new(TypeNameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_typeName
	return p
}

func InitEmptyTypeNameContext(p *TypeNameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_typeName
}

func (*TypeNameContext) IsTypeNameContext() {}

func NewTypeNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TypeNameContext {
	var p = new(TypeNameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_typeName

	return p
}

func (s *TypeNameContext) GetParser() antlr.Parser { return s.parser }

func (s *TypeNameContext) TYPE_KEYWORD() antlr.TerminalNode {
	return s.GetToken(IRParserTYPE_KEYWORD, 0)
}

func (s *TypeNameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(IRParserIDENTIFIER, 0)
}

func (s *TypeNameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TypeNameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TypeNameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterTypeName(s)
	}
}

func (s *TypeNameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitTypeName(s)
	}
}

func (p *IRParser) TypeName() (localctx ITypeNameContext) {
	localctx = NewTypeNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, IRParserRULE_typeName)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(91)
		_la = p.GetTokenStream().LA(1)

		if !(_la == IRParserTYPE_KEYWORD || _la == IRParserIDENTIFIER) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
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

// IArgsContext is an interface to support dynamic dispatch.
type IArgsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllArg() []IArgContext
	Arg(i int) IArgContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsArgsContext differentiates from other interfaces.
	IsArgsContext()
}

type ArgsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArgsContext() *ArgsContext {
	var p = new(ArgsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_args
	return p
}

func InitEmptyArgsContext(p *ArgsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_args
}

func (*ArgsContext) IsArgsContext() {}

func NewArgsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArgsContext {
	var p = new(ArgsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_args

	return p
}

func (s *ArgsContext) GetParser() antlr.Parser { return s.parser }

func (s *ArgsContext) AllArg() []IArgContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IArgContext); ok {
			len++
		}
	}

	tst := make([]IArgContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IArgContext); ok {
			tst[i] = t.(IArgContext)
			i++
		}
	}

	return tst
}

func (s *ArgsContext) Arg(i int) IArgContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArgContext); ok {
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

	return t.(IArgContext)
}

func (s *ArgsContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(IRParserCOMMA)
}

func (s *ArgsContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(IRParserCOMMA, i)
}

func (s *ArgsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArgsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArgsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterArgs(s)
	}
}

func (s *ArgsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitArgs(s)
	}
}

func (p *IRParser) Args() (localctx IArgsContext) {
	localctx = NewArgsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, IRParserRULE_args)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(101)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&4832) != 0 {
		{
			p.SetState(93)
			p.Arg()
		}
		p.SetState(98)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == IRParserCOMMA {
			{
				p.SetState(94)
				p.Match(IRParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(95)
				p.Arg()
			}

			p.SetState(100)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
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

// IArgContext is an interface to support dynamic dispatch.
type IArgContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsArgContext differentiates from other interfaces.
	IsArgContext()
}

type ArgContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArgContext() *ArgContext {
	var p = new(ArgContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_arg
	return p
}

func InitEmptyArgContext(p *ArgContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_arg
}

func (*ArgContext) IsArgContext() {}

func NewArgContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArgContext {
	var p = new(ArgContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_arg

	return p
}

func (s *ArgContext) GetParser() antlr.Parser { return s.parser }

func (s *ArgContext) CopyAll(ctx *ArgContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ArgContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArgContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type PositionalArgContext struct {
	ArgContext
}

func NewPositionalArgContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PositionalArgContext {
	var p = new(PositionalArgContext)

	InitEmptyArgContext(&p.ArgContext)
	p.parser = parser
	p.CopyAll(ctx.(*ArgContext))

	return p
}

func (s *PositionalArgContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PositionalArgContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *PositionalArgContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterPositionalArg(s)
	}
}

func (s *PositionalArgContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitPositionalArg(s)
	}
}

type LabeledArgContext struct {
	ArgContext
}

func NewLabeledArgContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LabeledArgContext {
	var p = new(LabeledArgContext)

	InitEmptyArgContext(&p.ArgContext)
	p.parser = parser
	p.CopyAll(ctx.(*ArgContext))

	return p
}

func (s *LabeledArgContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LabeledArgContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(IRParserIDENTIFIER, 0)
}

func (s *LabeledArgContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *LabeledArgContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterLabeledArg(s)
	}
}

func (s *LabeledArgContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitLabeledArg(s)
	}
}

func (p *IRParser) Arg() (localctx IArgContext) {
	localctx = NewArgContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, IRParserRULE_arg)
	p.SetState(107)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case IRParserREG, IRParserLABEL, IRParserINT, IRParserLBRACKET:
		localctx = NewPositionalArgContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(103)
			p.Value()
		}

	case IRParserIDENTIFIER:
		localctx = NewLabeledArgContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(104)
			p.Match(IRParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(105)
			p.Match(IRParserT__0)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(106)
			p.Value()
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

// IValueContext is an interface to support dynamic dispatch.
type IValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsValueContext differentiates from other interfaces.
	IsValueContext()
}

type ValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueContext() *ValueContext {
	var p = new(ValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_value
	return p
}

func InitEmptyValueContext(p *ValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_value
}

func (*ValueContext) IsValueContext() {}

func NewValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueContext {
	var p = new(ValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_value

	return p
}

func (s *ValueContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueContext) CopyAll(ctx *ValueContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ValRegContext struct {
	ValueContext
}

func NewValRegContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ValRegContext {
	var p = new(ValRegContext)

	InitEmptyValueContext(&p.ValueContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueContext))

	return p
}

func (s *ValRegContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValRegContext) Reg() IRegContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRegContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRegContext)
}

func (s *ValRegContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterValReg(s)
	}
}

func (s *ValRegContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitValReg(s)
	}
}

type ValIntContext struct {
	ValueContext
}

func NewValIntContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ValIntContext {
	var p = new(ValIntContext)

	InitEmptyValueContext(&p.ValueContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueContext))

	return p
}

func (s *ValIntContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValIntContext) INT() antlr.TerminalNode {
	return s.GetToken(IRParserINT, 0)
}

func (s *ValIntContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterValInt(s)
	}
}

func (s *ValIntContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitValInt(s)
	}
}

type ValLabelContext struct {
	ValueContext
}

func NewValLabelContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ValLabelContext {
	var p = new(ValLabelContext)

	InitEmptyValueContext(&p.ValueContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueContext))

	return p
}

func (s *ValLabelContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValLabelContext) LABEL() antlr.TerminalNode {
	return s.GetToken(IRParserLABEL, 0)
}

func (s *ValLabelContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterValLabel(s)
	}
}

func (s *ValLabelContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitValLabel(s)
	}
}

type ValRegListContext struct {
	ValueContext
}

func NewValRegListContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ValRegListContext {
	var p = new(ValRegListContext)

	InitEmptyValueContext(&p.ValueContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueContext))

	return p
}

func (s *ValRegListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValRegListContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(IRParserLBRACKET, 0)
}

func (s *ValRegListContext) RegList() IRegListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRegListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRegListContext)
}

func (s *ValRegListContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(IRParserRBRACKET, 0)
}

func (s *ValRegListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterValRegList(s)
	}
}

func (s *ValRegListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitValRegList(s)
	}
}

func (p *IRParser) Value() (localctx IValueContext) {
	localctx = NewValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, IRParserRULE_value)
	p.SetState(116)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case IRParserREG:
		localctx = NewValRegContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(109)
			p.Reg()
		}

	case IRParserLABEL:
		localctx = NewValLabelContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(110)
			p.Match(IRParserLABEL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case IRParserINT:
		localctx = NewValIntContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(111)
			p.Match(IRParserINT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case IRParserLBRACKET:
		localctx = NewValRegListContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(112)
			p.Match(IRParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(113)
			p.RegList()
		}
		{
			p.SetState(114)
			p.Match(IRParserRBRACKET)
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

// IConst_Context is an interface to support dynamic dispatch.
type IConst_Context interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsConst_Context differentiates from other interfaces.
	IsConst_Context()
}

type Const_Context struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConst_Context() *Const_Context {
	var p = new(Const_Context)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_const_
	return p
}

func InitEmptyConst_Context(p *Const_Context) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_const_
}

func (*Const_Context) IsConst_Context() {}

func NewConst_Context(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Const_Context {
	var p = new(Const_Context)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_const_

	return p
}

func (s *Const_Context) GetParser() antlr.Parser { return s.parser }

func (s *Const_Context) CopyAll(ctx *Const_Context) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *Const_Context) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Const_Context) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ConstStringContext struct {
	Const_Context
}

func NewConstStringContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConstStringContext {
	var p = new(ConstStringContext)

	InitEmptyConst_Context(&p.Const_Context)
	p.parser = parser
	p.CopyAll(ctx.(*Const_Context))

	return p
}

func (s *ConstStringContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConstStringContext) STRING() antlr.TerminalNode {
	return s.GetToken(IRParserSTRING, 0)
}

func (s *ConstStringContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterConstString(s)
	}
}

func (s *ConstStringContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitConstString(s)
	}
}

type ConstIntContext struct {
	Const_Context
}

func NewConstIntContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConstIntContext {
	var p = new(ConstIntContext)

	InitEmptyConst_Context(&p.Const_Context)
	p.parser = parser
	p.CopyAll(ctx.(*Const_Context))

	return p
}

func (s *ConstIntContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConstIntContext) INT() antlr.TerminalNode {
	return s.GetToken(IRParserINT, 0)
}

func (s *ConstIntContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterConstInt(s)
	}
}

func (s *ConstIntContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitConstInt(s)
	}
}

func (p *IRParser) Const_() (localctx IConst_Context) {
	localctx = NewConst_Context(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, IRParserRULE_const_)
	p.SetState(120)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case IRParserSTRING:
		localctx = NewConstStringContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(118)
			p.Match(IRParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case IRParserINT:
		localctx = NewConstIntContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(119)
			p.Match(IRParserINT)
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

// IRegContext is an interface to support dynamic dispatch.
type IRegContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	REG() antlr.TerminalNode

	// IsRegContext differentiates from other interfaces.
	IsRegContext()
}

type RegContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRegContext() *RegContext {
	var p = new(RegContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_reg
	return p
}

func InitEmptyRegContext(p *RegContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = IRParserRULE_reg
}

func (*RegContext) IsRegContext() {}

func NewRegContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RegContext {
	var p = new(RegContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = IRParserRULE_reg

	return p
}

func (s *RegContext) GetParser() antlr.Parser { return s.parser }

func (s *RegContext) REG() antlr.TerminalNode {
	return s.GetToken(IRParserREG, 0)
}

func (s *RegContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RegContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RegContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.EnterReg(s)
	}
}

func (s *RegContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(IRListener); ok {
		listenerT.ExitReg(s)
	}
}

func (p *IRParser) Reg() (localctx IRegContext) {
	localctx = NewRegContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, IRParserRULE_reg)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(122)
		p.Match(IRParserREG)
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
