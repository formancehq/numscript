// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser

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

type NumscriptLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var NumscriptLexerLexerStaticData struct {
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

func numscriptlexerLexerInit() {
	staticData := &NumscriptLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "", "", "", "", "'send'", "'('", "')'", "'['", "']'", "'='",
	}
	staticData.SymbolicNames = []string{
		"", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "WS", "SEND", "LPARENS",
		"RPARENS", "LBRACKET", "RBRACKET", "EQ", "NUMBER", "VARIABLE_NAME",
		"ACCOUNT", "ASSET",
	}
	staticData.RuleNames = []string{
		"NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "WS", "SEND", "LPARENS",
		"RPARENS", "LBRACKET", "RBRACKET", "EQ", "NUMBER", "VARIABLE_NAME",
		"ACCOUNT", "ASSET",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 14, 123, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 1, 0, 4, 0, 31, 8,
		0, 11, 0, 12, 0, 32, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 5, 1, 40, 8, 1, 10,
		1, 12, 1, 43, 9, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 2, 1,
		2, 5, 2, 54, 8, 2, 10, 2, 12, 2, 57, 9, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 3,
		4, 3, 64, 8, 3, 11, 3, 12, 3, 65, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 4, 1,
		4, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7, 1, 7, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10, 4,
		10, 86, 8, 10, 11, 10, 12, 10, 87, 1, 11, 1, 11, 4, 11, 92, 8, 11, 11,
		11, 12, 11, 93, 1, 11, 5, 11, 97, 8, 11, 10, 11, 12, 11, 100, 9, 11, 1,
		12, 1, 12, 4, 12, 104, 8, 12, 11, 12, 12, 12, 105, 1, 12, 1, 12, 4, 12,
		110, 8, 12, 11, 12, 12, 12, 111, 5, 12, 114, 8, 12, 10, 12, 12, 12, 117,
		9, 12, 1, 13, 4, 13, 120, 8, 13, 11, 13, 12, 13, 121, 2, 41, 55, 0, 14,
		1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19, 10, 21, 11,
		23, 12, 25, 13, 27, 14, 1, 0, 7, 2, 0, 10, 10, 13, 13, 3, 0, 9, 10, 13,
		13, 32, 32, 1, 0, 48, 57, 2, 0, 95, 95, 97, 122, 3, 0, 48, 57, 95, 95,
		97, 122, 5, 0, 45, 45, 48, 57, 65, 90, 95, 95, 97, 122, 2, 0, 47, 57, 65,
		90, 134, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1,
		0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0, 15,
		1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0, 0, 21, 1, 0, 0, 0, 0,
		23, 1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1, 0, 0, 0, 1, 30, 1, 0, 0, 0,
		3, 34, 1, 0, 0, 0, 5, 49, 1, 0, 0, 0, 7, 63, 1, 0, 0, 0, 9, 69, 1, 0, 0,
		0, 11, 74, 1, 0, 0, 0, 13, 76, 1, 0, 0, 0, 15, 78, 1, 0, 0, 0, 17, 80,
		1, 0, 0, 0, 19, 82, 1, 0, 0, 0, 21, 85, 1, 0, 0, 0, 23, 89, 1, 0, 0, 0,
		25, 101, 1, 0, 0, 0, 27, 119, 1, 0, 0, 0, 29, 31, 7, 0, 0, 0, 30, 29, 1,
		0, 0, 0, 31, 32, 1, 0, 0, 0, 32, 30, 1, 0, 0, 0, 32, 33, 1, 0, 0, 0, 33,
		2, 1, 0, 0, 0, 34, 35, 5, 47, 0, 0, 35, 36, 5, 42, 0, 0, 36, 41, 1, 0,
		0, 0, 37, 40, 3, 3, 1, 0, 38, 40, 9, 0, 0, 0, 39, 37, 1, 0, 0, 0, 39, 38,
		1, 0, 0, 0, 40, 43, 1, 0, 0, 0, 41, 42, 1, 0, 0, 0, 41, 39, 1, 0, 0, 0,
		42, 44, 1, 0, 0, 0, 43, 41, 1, 0, 0, 0, 44, 45, 5, 42, 0, 0, 45, 46, 5,
		47, 0, 0, 46, 47, 1, 0, 0, 0, 47, 48, 6, 1, 0, 0, 48, 4, 1, 0, 0, 0, 49,
		50, 5, 47, 0, 0, 50, 51, 5, 47, 0, 0, 51, 55, 1, 0, 0, 0, 52, 54, 9, 0,
		0, 0, 53, 52, 1, 0, 0, 0, 54, 57, 1, 0, 0, 0, 55, 56, 1, 0, 0, 0, 55, 53,
		1, 0, 0, 0, 56, 58, 1, 0, 0, 0, 57, 55, 1, 0, 0, 0, 58, 59, 3, 1, 0, 0,
		59, 60, 1, 0, 0, 0, 60, 61, 6, 2, 0, 0, 61, 6, 1, 0, 0, 0, 62, 64, 7, 1,
		0, 0, 63, 62, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 65, 66,
		1, 0, 0, 0, 66, 67, 1, 0, 0, 0, 67, 68, 6, 3, 0, 0, 68, 8, 1, 0, 0, 0,
		69, 70, 5, 115, 0, 0, 70, 71, 5, 101, 0, 0, 71, 72, 5, 110, 0, 0, 72, 73,
		5, 100, 0, 0, 73, 10, 1, 0, 0, 0, 74, 75, 5, 40, 0, 0, 75, 12, 1, 0, 0,
		0, 76, 77, 5, 41, 0, 0, 77, 14, 1, 0, 0, 0, 78, 79, 5, 91, 0, 0, 79, 16,
		1, 0, 0, 0, 80, 81, 5, 93, 0, 0, 81, 18, 1, 0, 0, 0, 82, 83, 5, 61, 0,
		0, 83, 20, 1, 0, 0, 0, 84, 86, 7, 2, 0, 0, 85, 84, 1, 0, 0, 0, 86, 87,
		1, 0, 0, 0, 87, 85, 1, 0, 0, 0, 87, 88, 1, 0, 0, 0, 88, 22, 1, 0, 0, 0,
		89, 91, 5, 36, 0, 0, 90, 92, 7, 3, 0, 0, 91, 90, 1, 0, 0, 0, 92, 93, 1,
		0, 0, 0, 93, 91, 1, 0, 0, 0, 93, 94, 1, 0, 0, 0, 94, 98, 1, 0, 0, 0, 95,
		97, 7, 4, 0, 0, 96, 95, 1, 0, 0, 0, 97, 100, 1, 0, 0, 0, 98, 96, 1, 0,
		0, 0, 98, 99, 1, 0, 0, 0, 99, 24, 1, 0, 0, 0, 100, 98, 1, 0, 0, 0, 101,
		103, 5, 64, 0, 0, 102, 104, 7, 5, 0, 0, 103, 102, 1, 0, 0, 0, 104, 105,
		1, 0, 0, 0, 105, 103, 1, 0, 0, 0, 105, 106, 1, 0, 0, 0, 106, 115, 1, 0,
		0, 0, 107, 109, 5, 58, 0, 0, 108, 110, 7, 5, 0, 0, 109, 108, 1, 0, 0, 0,
		110, 111, 1, 0, 0, 0, 111, 109, 1, 0, 0, 0, 111, 112, 1, 0, 0, 0, 112,
		114, 1, 0, 0, 0, 113, 107, 1, 0, 0, 0, 114, 117, 1, 0, 0, 0, 115, 113,
		1, 0, 0, 0, 115, 116, 1, 0, 0, 0, 116, 26, 1, 0, 0, 0, 117, 115, 1, 0,
		0, 0, 118, 120, 7, 6, 0, 0, 119, 118, 1, 0, 0, 0, 120, 121, 1, 0, 0, 0,
		121, 119, 1, 0, 0, 0, 121, 122, 1, 0, 0, 0, 122, 28, 1, 0, 0, 0, 13, 0,
		32, 39, 41, 55, 65, 87, 93, 98, 105, 111, 115, 121, 1, 6, 0, 0,
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

// NumscriptLexerInit initializes any static state used to implement NumscriptLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewNumscriptLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func NumscriptLexerInit() {
	staticData := &NumscriptLexerLexerStaticData
	staticData.once.Do(numscriptlexerLexerInit)
}

// NewNumscriptLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewNumscriptLexer(input antlr.CharStream) *NumscriptLexer {
	NumscriptLexerInit()
	l := new(NumscriptLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &NumscriptLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "Numscript.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// NumscriptLexer tokens.
const (
	NumscriptLexerNEWLINE           = 1
	NumscriptLexerMULTILINE_COMMENT = 2
	NumscriptLexerLINE_COMMENT      = 3
	NumscriptLexerWS                = 4
	NumscriptLexerSEND              = 5
	NumscriptLexerLPARENS           = 6
	NumscriptLexerRPARENS           = 7
	NumscriptLexerLBRACKET          = 8
	NumscriptLexerRBRACKET          = 9
	NumscriptLexerEQ                = 10
	NumscriptLexerNUMBER            = 11
	NumscriptLexerVARIABLE_NAME     = 12
	NumscriptLexerACCOUNT           = 13
	NumscriptLexerASSET             = 14
)
