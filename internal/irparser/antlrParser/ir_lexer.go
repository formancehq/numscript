// Code generated from IR.g4 by ANTLR 4.13.2. DO NOT EDIT.

package antlrParser

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type IRLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var IRLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func irlexerLexerInit() {
	staticData := &IRLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "':'", "", "", "", "", "", "", "", "", "", "'('", "')'", "'['",
		"']'", "','", "'='", "'+'", "'-'", "'+='", "'-='", "'<'", "'>'", "'_'",
	}
	staticData.SymbolicNames = []string{
		"", "", "WS", "NEWLINE", "TYPE_KEYWORD", "BOOL", "REG", "LABEL", "INT",
		"STRING", "IDENTIFIER", "LPAREN", "RPAREN", "LBRACKET", "RBRACKET",
		"COMMA", "EQ", "PLUS", "MINUS", "PLUS_EQ", "MINUS_EQ", "LT", "GT", "UNDERSCORE",
	}
	staticData.RuleNames = []string{
		"T__0", "WS", "NEWLINE", "TYPE_KEYWORD", "BOOL", "REG", "LABEL", "INT",
		"STRING", "IDENTIFIER", "LPAREN", "RPAREN", "LBRACKET", "RBRACKET",
		"COMMA", "EQ", "PLUS", "MINUS", "PLUS_EQ", "MINUS_EQ", "LT", "GT", "UNDERSCORE",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 23, 164, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7,
		20, 2, 21, 7, 21, 2, 22, 7, 22, 1, 0, 1, 0, 1, 1, 4, 1, 51, 8, 1, 11, 1,
		12, 1, 52, 1, 1, 1, 1, 1, 2, 4, 2, 58, 8, 2, 11, 2, 12, 2, 59, 1, 2, 1,
		2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1,
		3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 85, 8, 3,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 3, 4, 96, 8, 4, 1,
		5, 1, 5, 1, 5, 5, 5, 101, 8, 5, 10, 5, 12, 5, 104, 9, 5, 1, 6, 1, 6, 1,
		6, 5, 6, 109, 8, 6, 10, 6, 12, 6, 112, 9, 6, 1, 7, 4, 7, 115, 8, 7, 11,
		7, 12, 7, 116, 1, 8, 1, 8, 1, 8, 1, 8, 5, 8, 123, 8, 8, 10, 8, 12, 8, 126,
		9, 8, 1, 8, 1, 8, 1, 9, 1, 9, 5, 9, 132, 8, 9, 10, 9, 12, 9, 135, 9, 9,
		1, 10, 1, 10, 1, 11, 1, 11, 1, 12, 1, 12, 1, 13, 1, 13, 1, 14, 1, 14, 1,
		15, 1, 15, 1, 16, 1, 16, 1, 17, 1, 17, 1, 18, 1, 18, 1, 18, 1, 19, 1, 19,
		1, 19, 1, 20, 1, 20, 1, 21, 1, 21, 1, 22, 1, 22, 0, 0, 23, 1, 1, 3, 2,
		5, 3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19, 10, 21, 11, 23, 12, 25,
		13, 27, 14, 29, 15, 31, 16, 33, 17, 35, 18, 37, 19, 39, 20, 41, 21, 43,
		22, 45, 23, 1, 0, 8, 2, 0, 9, 9, 32, 32, 2, 0, 10, 10, 13, 13, 3, 0, 65,
		90, 95, 95, 97, 122, 4, 0, 48, 57, 65, 90, 95, 95, 97, 122, 1, 0, 48, 57,
		3, 0, 10, 10, 13, 13, 34, 34, 1, 0, 97, 122, 3, 0, 48, 57, 95, 95, 97,
		122, 175, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1,
		0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0, 15,
		1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0, 0, 21, 1, 0, 0, 0, 0,
		23, 1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1, 0, 0, 0, 0, 29, 1, 0, 0, 0,
		0, 31, 1, 0, 0, 0, 0, 33, 1, 0, 0, 0, 0, 35, 1, 0, 0, 0, 0, 37, 1, 0, 0,
		0, 0, 39, 1, 0, 0, 0, 0, 41, 1, 0, 0, 0, 0, 43, 1, 0, 0, 0, 0, 45, 1, 0,
		0, 0, 1, 47, 1, 0, 0, 0, 3, 50, 1, 0, 0, 0, 5, 57, 1, 0, 0, 0, 7, 84, 1,
		0, 0, 0, 9, 95, 1, 0, 0, 0, 11, 97, 1, 0, 0, 0, 13, 105, 1, 0, 0, 0, 15,
		114, 1, 0, 0, 0, 17, 118, 1, 0, 0, 0, 19, 129, 1, 0, 0, 0, 21, 136, 1,
		0, 0, 0, 23, 138, 1, 0, 0, 0, 25, 140, 1, 0, 0, 0, 27, 142, 1, 0, 0, 0,
		29, 144, 1, 0, 0, 0, 31, 146, 1, 0, 0, 0, 33, 148, 1, 0, 0, 0, 35, 150,
		1, 0, 0, 0, 37, 152, 1, 0, 0, 0, 39, 155, 1, 0, 0, 0, 41, 158, 1, 0, 0,
		0, 43, 160, 1, 0, 0, 0, 45, 162, 1, 0, 0, 0, 47, 48, 5, 58, 0, 0, 48, 2,
		1, 0, 0, 0, 49, 51, 7, 0, 0, 0, 50, 49, 1, 0, 0, 0, 51, 52, 1, 0, 0, 0,
		52, 50, 1, 0, 0, 0, 52, 53, 1, 0, 0, 0, 53, 54, 1, 0, 0, 0, 54, 55, 6,
		1, 0, 0, 55, 4, 1, 0, 0, 0, 56, 58, 7, 1, 0, 0, 57, 56, 1, 0, 0, 0, 58,
		59, 1, 0, 0, 0, 59, 57, 1, 0, 0, 0, 59, 60, 1, 0, 0, 0, 60, 61, 1, 0, 0,
		0, 61, 62, 6, 2, 0, 0, 62, 6, 1, 0, 0, 0, 63, 64, 5, 105, 0, 0, 64, 65,
		5, 110, 0, 0, 65, 85, 5, 116, 0, 0, 66, 67, 5, 115, 0, 0, 67, 68, 5, 116,
		0, 0, 68, 85, 5, 114, 0, 0, 69, 70, 5, 112, 0, 0, 70, 71, 5, 111, 0, 0,
		71, 72, 5, 114, 0, 0, 72, 73, 5, 116, 0, 0, 73, 74, 5, 105, 0, 0, 74, 75,
		5, 111, 0, 0, 75, 85, 5, 110, 0, 0, 76, 77, 5, 109, 0, 0, 77, 78, 5, 111,
		0, 0, 78, 79, 5, 110, 0, 0, 79, 80, 5, 101, 0, 0, 80, 81, 5, 116, 0, 0,
		81, 82, 5, 97, 0, 0, 82, 83, 5, 114, 0, 0, 83, 85, 5, 121, 0, 0, 84, 63,
		1, 0, 0, 0, 84, 66, 1, 0, 0, 0, 84, 69, 1, 0, 0, 0, 84, 76, 1, 0, 0, 0,
		85, 8, 1, 0, 0, 0, 86, 87, 5, 116, 0, 0, 87, 88, 5, 114, 0, 0, 88, 89,
		5, 117, 0, 0, 89, 96, 5, 101, 0, 0, 90, 91, 5, 102, 0, 0, 91, 92, 5, 97,
		0, 0, 92, 93, 5, 108, 0, 0, 93, 94, 5, 115, 0, 0, 94, 96, 5, 101, 0, 0,
		95, 86, 1, 0, 0, 0, 95, 90, 1, 0, 0, 0, 96, 10, 1, 0, 0, 0, 97, 98, 5,
		36, 0, 0, 98, 102, 7, 2, 0, 0, 99, 101, 7, 3, 0, 0, 100, 99, 1, 0, 0, 0,
		101, 104, 1, 0, 0, 0, 102, 100, 1, 0, 0, 0, 102, 103, 1, 0, 0, 0, 103,
		12, 1, 0, 0, 0, 104, 102, 1, 0, 0, 0, 105, 106, 5, 35, 0, 0, 106, 110,
		7, 2, 0, 0, 107, 109, 7, 3, 0, 0, 108, 107, 1, 0, 0, 0, 109, 112, 1, 0,
		0, 0, 110, 108, 1, 0, 0, 0, 110, 111, 1, 0, 0, 0, 111, 14, 1, 0, 0, 0,
		112, 110, 1, 0, 0, 0, 113, 115, 7, 4, 0, 0, 114, 113, 1, 0, 0, 0, 115,
		116, 1, 0, 0, 0, 116, 114, 1, 0, 0, 0, 116, 117, 1, 0, 0, 0, 117, 16, 1,
		0, 0, 0, 118, 124, 5, 34, 0, 0, 119, 120, 5, 92, 0, 0, 120, 123, 5, 34,
		0, 0, 121, 123, 8, 5, 0, 0, 122, 119, 1, 0, 0, 0, 122, 121, 1, 0, 0, 0,
		123, 126, 1, 0, 0, 0, 124, 122, 1, 0, 0, 0, 124, 125, 1, 0, 0, 0, 125,
		127, 1, 0, 0, 0, 126, 124, 1, 0, 0, 0, 127, 128, 5, 34, 0, 0, 128, 18,
		1, 0, 0, 0, 129, 133, 7, 6, 0, 0, 130, 132, 7, 7, 0, 0, 131, 130, 1, 0,
		0, 0, 132, 135, 1, 0, 0, 0, 133, 131, 1, 0, 0, 0, 133, 134, 1, 0, 0, 0,
		134, 20, 1, 0, 0, 0, 135, 133, 1, 0, 0, 0, 136, 137, 5, 40, 0, 0, 137,
		22, 1, 0, 0, 0, 138, 139, 5, 41, 0, 0, 139, 24, 1, 0, 0, 0, 140, 141, 5,
		91, 0, 0, 141, 26, 1, 0, 0, 0, 142, 143, 5, 93, 0, 0, 143, 28, 1, 0, 0,
		0, 144, 145, 5, 44, 0, 0, 145, 30, 1, 0, 0, 0, 146, 147, 5, 61, 0, 0, 147,
		32, 1, 0, 0, 0, 148, 149, 5, 43, 0, 0, 149, 34, 1, 0, 0, 0, 150, 151, 5,
		45, 0, 0, 151, 36, 1, 0, 0, 0, 152, 153, 5, 43, 0, 0, 153, 154, 5, 61,
		0, 0, 154, 38, 1, 0, 0, 0, 155, 156, 5, 45, 0, 0, 156, 157, 5, 61, 0, 0,
		157, 40, 1, 0, 0, 0, 158, 159, 5, 60, 0, 0, 159, 42, 1, 0, 0, 0, 160, 161,
		5, 62, 0, 0, 161, 44, 1, 0, 0, 0, 162, 163, 5, 95, 0, 0, 163, 46, 1, 0,
		0, 0, 11, 0, 52, 59, 84, 95, 102, 110, 116, 122, 124, 133, 1, 6, 0, 0,
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

// IRLexerInit initializes any static state used to implement IRLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewIRLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func IRLexerInit() {
	staticData := &IRLexerLexerStaticData
	staticData.once.Do(irlexerLexerInit)
}

// NewIRLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewIRLexer(input antlr.CharStream) *IRLexer {
	IRLexerInit()
	l := new(IRLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &IRLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "IR.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// IRLexer tokens.
const (
	IRLexerT__0         = 1
	IRLexerWS           = 2
	IRLexerNEWLINE      = 3
	IRLexerTYPE_KEYWORD = 4
	IRLexerBOOL         = 5
	IRLexerREG          = 6
	IRLexerLABEL        = 7
	IRLexerINT          = 8
	IRLexerSTRING       = 9
	IRLexerIDENTIFIER   = 10
	IRLexerLPAREN       = 11
	IRLexerRPAREN       = 12
	IRLexerLBRACKET     = 13
	IRLexerRBRACKET     = 14
	IRLexerCOMMA        = 15
	IRLexerEQ           = 16
	IRLexerPLUS         = 17
	IRLexerMINUS        = 18
	IRLexerPLUS_EQ      = 19
	IRLexerMINUS_EQ     = 20
	IRLexerLT           = 21
	IRLexerGT           = 22
	IRLexerUNDERSCORE   = 23
)
