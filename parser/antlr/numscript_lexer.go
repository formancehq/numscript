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
		"", "", "", "", "", "'source'", "'send'", "'('", "')'", "'['", "']'",
		"'='",
	}
	staticData.SymbolicNames = []string{
		"", "WS", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "SOURCE",
		"SEND", "LPARENS", "RPARENS", "LBRACKET", "RBRACKET", "EQ", "NUMBER",
		"VARIABLE_NAME", "ACCOUNT", "ASSET",
	}
	staticData.RuleNames = []string{
		"WS", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "SOURCE", "SEND",
		"LPARENS", "RPARENS", "LBRACKET", "RBRACKET", "EQ", "NUMBER", "VARIABLE_NAME",
		"ACCOUNT", "ASSET",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 15, 132, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 1, 0,
		4, 0, 33, 8, 0, 11, 0, 12, 0, 34, 1, 0, 1, 0, 1, 1, 4, 1, 40, 8, 1, 11,
		1, 12, 1, 41, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 5, 2, 49, 8, 2, 10, 2, 12,
		2, 52, 9, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 1, 3, 5, 3,
		63, 8, 3, 10, 3, 12, 3, 66, 9, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1,
		4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1,
		7, 1, 7, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10, 1, 10, 1, 11, 4, 11, 95, 8, 11,
		11, 11, 12, 11, 96, 1, 12, 1, 12, 4, 12, 101, 8, 12, 11, 12, 12, 12, 102,
		1, 12, 5, 12, 106, 8, 12, 10, 12, 12, 12, 109, 9, 12, 1, 13, 1, 13, 4,
		13, 113, 8, 13, 11, 13, 12, 13, 114, 1, 13, 1, 13, 4, 13, 119, 8, 13, 11,
		13, 12, 13, 120, 5, 13, 123, 8, 13, 10, 13, 12, 13, 126, 9, 13, 1, 14,
		4, 14, 129, 8, 14, 11, 14, 12, 14, 130, 2, 50, 64, 0, 15, 1, 1, 3, 2, 5,
		3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19, 10, 21, 11, 23, 12, 25,
		13, 27, 14, 29, 15, 1, 0, 7, 3, 0, 9, 10, 13, 13, 32, 32, 2, 0, 10, 10,
		13, 13, 1, 0, 48, 57, 2, 0, 95, 95, 97, 122, 3, 0, 48, 57, 95, 95, 97,
		122, 5, 0, 45, 45, 48, 57, 65, 90, 95, 95, 97, 122, 2, 0, 47, 57, 65, 90,
		143, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0,
		0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0, 15, 1,
		0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0, 0, 21, 1, 0, 0, 0, 0, 23,
		1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1, 0, 0, 0, 0, 29, 1, 0, 0, 0, 1,
		32, 1, 0, 0, 0, 3, 39, 1, 0, 0, 0, 5, 43, 1, 0, 0, 0, 7, 58, 1, 0, 0, 0,
		9, 71, 1, 0, 0, 0, 11, 78, 1, 0, 0, 0, 13, 83, 1, 0, 0, 0, 15, 85, 1, 0,
		0, 0, 17, 87, 1, 0, 0, 0, 19, 89, 1, 0, 0, 0, 21, 91, 1, 0, 0, 0, 23, 94,
		1, 0, 0, 0, 25, 98, 1, 0, 0, 0, 27, 110, 1, 0, 0, 0, 29, 128, 1, 0, 0,
		0, 31, 33, 7, 0, 0, 0, 32, 31, 1, 0, 0, 0, 33, 34, 1, 0, 0, 0, 34, 32,
		1, 0, 0, 0, 34, 35, 1, 0, 0, 0, 35, 36, 1, 0, 0, 0, 36, 37, 6, 0, 0, 0,
		37, 2, 1, 0, 0, 0, 38, 40, 7, 1, 0, 0, 39, 38, 1, 0, 0, 0, 40, 41, 1, 0,
		0, 0, 41, 39, 1, 0, 0, 0, 41, 42, 1, 0, 0, 0, 42, 4, 1, 0, 0, 0, 43, 44,
		5, 47, 0, 0, 44, 45, 5, 42, 0, 0, 45, 50, 1, 0, 0, 0, 46, 49, 3, 5, 2,
		0, 47, 49, 9, 0, 0, 0, 48, 46, 1, 0, 0, 0, 48, 47, 1, 0, 0, 0, 49, 52,
		1, 0, 0, 0, 50, 51, 1, 0, 0, 0, 50, 48, 1, 0, 0, 0, 51, 53, 1, 0, 0, 0,
		52, 50, 1, 0, 0, 0, 53, 54, 5, 42, 0, 0, 54, 55, 5, 47, 0, 0, 55, 56, 1,
		0, 0, 0, 56, 57, 6, 2, 0, 0, 57, 6, 1, 0, 0, 0, 58, 59, 5, 47, 0, 0, 59,
		60, 5, 47, 0, 0, 60, 64, 1, 0, 0, 0, 61, 63, 9, 0, 0, 0, 62, 61, 1, 0,
		0, 0, 63, 66, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0, 64, 62, 1, 0, 0, 0, 65, 67,
		1, 0, 0, 0, 66, 64, 1, 0, 0, 0, 67, 68, 3, 3, 1, 0, 68, 69, 1, 0, 0, 0,
		69, 70, 6, 3, 0, 0, 70, 8, 1, 0, 0, 0, 71, 72, 5, 115, 0, 0, 72, 73, 5,
		111, 0, 0, 73, 74, 5, 117, 0, 0, 74, 75, 5, 114, 0, 0, 75, 76, 5, 99, 0,
		0, 76, 77, 5, 101, 0, 0, 77, 10, 1, 0, 0, 0, 78, 79, 5, 115, 0, 0, 79,
		80, 5, 101, 0, 0, 80, 81, 5, 110, 0, 0, 81, 82, 5, 100, 0, 0, 82, 12, 1,
		0, 0, 0, 83, 84, 5, 40, 0, 0, 84, 14, 1, 0, 0, 0, 85, 86, 5, 41, 0, 0,
		86, 16, 1, 0, 0, 0, 87, 88, 5, 91, 0, 0, 88, 18, 1, 0, 0, 0, 89, 90, 5,
		93, 0, 0, 90, 20, 1, 0, 0, 0, 91, 92, 5, 61, 0, 0, 92, 22, 1, 0, 0, 0,
		93, 95, 7, 2, 0, 0, 94, 93, 1, 0, 0, 0, 95, 96, 1, 0, 0, 0, 96, 94, 1,
		0, 0, 0, 96, 97, 1, 0, 0, 0, 97, 24, 1, 0, 0, 0, 98, 100, 5, 36, 0, 0,
		99, 101, 7, 3, 0, 0, 100, 99, 1, 0, 0, 0, 101, 102, 1, 0, 0, 0, 102, 100,
		1, 0, 0, 0, 102, 103, 1, 0, 0, 0, 103, 107, 1, 0, 0, 0, 104, 106, 7, 4,
		0, 0, 105, 104, 1, 0, 0, 0, 106, 109, 1, 0, 0, 0, 107, 105, 1, 0, 0, 0,
		107, 108, 1, 0, 0, 0, 108, 26, 1, 0, 0, 0, 109, 107, 1, 0, 0, 0, 110, 112,
		5, 64, 0, 0, 111, 113, 7, 5, 0, 0, 112, 111, 1, 0, 0, 0, 113, 114, 1, 0,
		0, 0, 114, 112, 1, 0, 0, 0, 114, 115, 1, 0, 0, 0, 115, 124, 1, 0, 0, 0,
		116, 118, 5, 58, 0, 0, 117, 119, 7, 5, 0, 0, 118, 117, 1, 0, 0, 0, 119,
		120, 1, 0, 0, 0, 120, 118, 1, 0, 0, 0, 120, 121, 1, 0, 0, 0, 121, 123,
		1, 0, 0, 0, 122, 116, 1, 0, 0, 0, 123, 126, 1, 0, 0, 0, 124, 122, 1, 0,
		0, 0, 124, 125, 1, 0, 0, 0, 125, 28, 1, 0, 0, 0, 126, 124, 1, 0, 0, 0,
		127, 129, 7, 6, 0, 0, 128, 127, 1, 0, 0, 0, 129, 130, 1, 0, 0, 0, 130,
		128, 1, 0, 0, 0, 130, 131, 1, 0, 0, 0, 131, 30, 1, 0, 0, 0, 13, 0, 34,
		41, 48, 50, 64, 96, 102, 107, 114, 120, 124, 130, 1, 6, 0, 0,
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
	NumscriptLexerWS                = 1
	NumscriptLexerNEWLINE           = 2
	NumscriptLexerMULTILINE_COMMENT = 3
	NumscriptLexerLINE_COMMENT      = 4
	NumscriptLexerSOURCE            = 5
	NumscriptLexerSEND              = 6
	NumscriptLexerLPARENS           = 7
	NumscriptLexerRPARENS           = 8
	NumscriptLexerLBRACKET          = 9
	NumscriptLexerRBRACKET          = 10
	NumscriptLexerEQ                = 11
	NumscriptLexerNUMBER            = 12
	NumscriptLexerVARIABLE_NAME     = 13
	NumscriptLexerACCOUNT           = 14
	NumscriptLexerASSET             = 15
)
