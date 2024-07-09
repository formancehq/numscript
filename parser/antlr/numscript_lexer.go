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
		"", "", "", "", "", "'source'", "'destination'", "'send'", "'from'",
		"'to'", "'('", "')'", "'['", "']'", "'{'", "'}'", "'='",
	}
	staticData.SymbolicNames = []string{
		"", "WS", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "SOURCE",
		"DESTINATION", "SEND", "FROM", "TO", "LPARENS", "RPARENS", "LBRACKET",
		"RBRACKET", "LBRACE", "RBRACE", "EQ", "RATIO_PORTION_LITERAL", "PERCENTAGE_PORTION_LITERAL",
		"NUMBER", "VARIABLE_NAME", "ACCOUNT", "ASSET",
	}
	staticData.RuleNames = []string{
		"WS", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "SOURCE", "DESTINATION",
		"SEND", "FROM", "TO", "LPARENS", "RPARENS", "LBRACKET", "RBRACKET",
		"LBRACE", "RBRACE", "EQ", "RATIO_PORTION_LITERAL", "PERCENTAGE_PORTION_LITERAL",
		"NUMBER", "VARIABLE_NAME", "ACCOUNT", "ASSET",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 22, 202, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7,
		20, 2, 21, 7, 21, 1, 0, 4, 0, 47, 8, 0, 11, 0, 12, 0, 48, 1, 0, 1, 0, 1,
		1, 4, 1, 54, 8, 1, 11, 1, 12, 1, 55, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 5, 2,
		63, 8, 2, 10, 2, 12, 2, 66, 9, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 3, 1,
		3, 1, 3, 1, 3, 5, 3, 77, 8, 3, 10, 3, 12, 3, 80, 9, 3, 1, 3, 1, 3, 1, 3,
		1, 3, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5,
		1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6,
		1, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10,
		1, 10, 1, 11, 1, 11, 1, 12, 1, 12, 1, 13, 1, 13, 1, 14, 1, 14, 1, 15, 1,
		15, 1, 16, 4, 16, 133, 8, 16, 11, 16, 12, 16, 134, 1, 16, 3, 16, 138, 8,
		16, 1, 16, 1, 16, 3, 16, 142, 8, 16, 1, 16, 4, 16, 145, 8, 16, 11, 16,
		12, 16, 146, 1, 17, 4, 17, 150, 8, 17, 11, 17, 12, 17, 151, 1, 17, 1, 17,
		4, 17, 156, 8, 17, 11, 17, 12, 17, 157, 3, 17, 160, 8, 17, 1, 17, 1, 17,
		1, 18, 4, 18, 165, 8, 18, 11, 18, 12, 18, 166, 1, 19, 1, 19, 4, 19, 171,
		8, 19, 11, 19, 12, 19, 172, 1, 19, 5, 19, 176, 8, 19, 10, 19, 12, 19, 179,
		9, 19, 1, 20, 1, 20, 4, 20, 183, 8, 20, 11, 20, 12, 20, 184, 1, 20, 1,
		20, 4, 20, 189, 8, 20, 11, 20, 12, 20, 190, 5, 20, 193, 8, 20, 10, 20,
		12, 20, 196, 9, 20, 1, 21, 4, 21, 199, 8, 21, 11, 21, 12, 21, 200, 2, 64,
		78, 0, 22, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19,
		10, 21, 11, 23, 12, 25, 13, 27, 14, 29, 15, 31, 16, 33, 17, 35, 18, 37,
		19, 39, 20, 41, 21, 43, 22, 1, 0, 8, 3, 0, 9, 10, 13, 13, 32, 32, 2, 0,
		10, 10, 13, 13, 1, 0, 48, 57, 1, 0, 32, 32, 2, 0, 95, 95, 97, 122, 3, 0,
		48, 57, 95, 95, 97, 122, 5, 0, 45, 45, 48, 57, 65, 90, 95, 95, 97, 122,
		2, 0, 47, 57, 65, 90, 220, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1,
		0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13,
		1, 0, 0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0, 0,
		21, 1, 0, 0, 0, 0, 23, 1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1, 0, 0, 0,
		0, 29, 1, 0, 0, 0, 0, 31, 1, 0, 0, 0, 0, 33, 1, 0, 0, 0, 0, 35, 1, 0, 0,
		0, 0, 37, 1, 0, 0, 0, 0, 39, 1, 0, 0, 0, 0, 41, 1, 0, 0, 0, 0, 43, 1, 0,
		0, 0, 1, 46, 1, 0, 0, 0, 3, 53, 1, 0, 0, 0, 5, 57, 1, 0, 0, 0, 7, 72, 1,
		0, 0, 0, 9, 85, 1, 0, 0, 0, 11, 92, 1, 0, 0, 0, 13, 104, 1, 0, 0, 0, 15,
		109, 1, 0, 0, 0, 17, 114, 1, 0, 0, 0, 19, 117, 1, 0, 0, 0, 21, 119, 1,
		0, 0, 0, 23, 121, 1, 0, 0, 0, 25, 123, 1, 0, 0, 0, 27, 125, 1, 0, 0, 0,
		29, 127, 1, 0, 0, 0, 31, 129, 1, 0, 0, 0, 33, 132, 1, 0, 0, 0, 35, 149,
		1, 0, 0, 0, 37, 164, 1, 0, 0, 0, 39, 168, 1, 0, 0, 0, 41, 180, 1, 0, 0,
		0, 43, 198, 1, 0, 0, 0, 45, 47, 7, 0, 0, 0, 46, 45, 1, 0, 0, 0, 47, 48,
		1, 0, 0, 0, 48, 46, 1, 0, 0, 0, 48, 49, 1, 0, 0, 0, 49, 50, 1, 0, 0, 0,
		50, 51, 6, 0, 0, 0, 51, 2, 1, 0, 0, 0, 52, 54, 7, 1, 0, 0, 53, 52, 1, 0,
		0, 0, 54, 55, 1, 0, 0, 0, 55, 53, 1, 0, 0, 0, 55, 56, 1, 0, 0, 0, 56, 4,
		1, 0, 0, 0, 57, 58, 5, 47, 0, 0, 58, 59, 5, 42, 0, 0, 59, 64, 1, 0, 0,
		0, 60, 63, 3, 5, 2, 0, 61, 63, 9, 0, 0, 0, 62, 60, 1, 0, 0, 0, 62, 61,
		1, 0, 0, 0, 63, 66, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0, 64, 62, 1, 0, 0, 0,
		65, 67, 1, 0, 0, 0, 66, 64, 1, 0, 0, 0, 67, 68, 5, 42, 0, 0, 68, 69, 5,
		47, 0, 0, 69, 70, 1, 0, 0, 0, 70, 71, 6, 2, 0, 0, 71, 6, 1, 0, 0, 0, 72,
		73, 5, 47, 0, 0, 73, 74, 5, 47, 0, 0, 74, 78, 1, 0, 0, 0, 75, 77, 9, 0,
		0, 0, 76, 75, 1, 0, 0, 0, 77, 80, 1, 0, 0, 0, 78, 79, 1, 0, 0, 0, 78, 76,
		1, 0, 0, 0, 79, 81, 1, 0, 0, 0, 80, 78, 1, 0, 0, 0, 81, 82, 3, 3, 1, 0,
		82, 83, 1, 0, 0, 0, 83, 84, 6, 3, 0, 0, 84, 8, 1, 0, 0, 0, 85, 86, 5, 115,
		0, 0, 86, 87, 5, 111, 0, 0, 87, 88, 5, 117, 0, 0, 88, 89, 5, 114, 0, 0,
		89, 90, 5, 99, 0, 0, 90, 91, 5, 101, 0, 0, 91, 10, 1, 0, 0, 0, 92, 93,
		5, 100, 0, 0, 93, 94, 5, 101, 0, 0, 94, 95, 5, 115, 0, 0, 95, 96, 5, 116,
		0, 0, 96, 97, 5, 105, 0, 0, 97, 98, 5, 110, 0, 0, 98, 99, 5, 97, 0, 0,
		99, 100, 5, 116, 0, 0, 100, 101, 5, 105, 0, 0, 101, 102, 5, 111, 0, 0,
		102, 103, 5, 110, 0, 0, 103, 12, 1, 0, 0, 0, 104, 105, 5, 115, 0, 0, 105,
		106, 5, 101, 0, 0, 106, 107, 5, 110, 0, 0, 107, 108, 5, 100, 0, 0, 108,
		14, 1, 0, 0, 0, 109, 110, 5, 102, 0, 0, 110, 111, 5, 114, 0, 0, 111, 112,
		5, 111, 0, 0, 112, 113, 5, 109, 0, 0, 113, 16, 1, 0, 0, 0, 114, 115, 5,
		116, 0, 0, 115, 116, 5, 111, 0, 0, 116, 18, 1, 0, 0, 0, 117, 118, 5, 40,
		0, 0, 118, 20, 1, 0, 0, 0, 119, 120, 5, 41, 0, 0, 120, 22, 1, 0, 0, 0,
		121, 122, 5, 91, 0, 0, 122, 24, 1, 0, 0, 0, 123, 124, 5, 93, 0, 0, 124,
		26, 1, 0, 0, 0, 125, 126, 5, 123, 0, 0, 126, 28, 1, 0, 0, 0, 127, 128,
		5, 125, 0, 0, 128, 30, 1, 0, 0, 0, 129, 130, 5, 61, 0, 0, 130, 32, 1, 0,
		0, 0, 131, 133, 7, 2, 0, 0, 132, 131, 1, 0, 0, 0, 133, 134, 1, 0, 0, 0,
		134, 132, 1, 0, 0, 0, 134, 135, 1, 0, 0, 0, 135, 137, 1, 0, 0, 0, 136,
		138, 7, 3, 0, 0, 137, 136, 1, 0, 0, 0, 137, 138, 1, 0, 0, 0, 138, 139,
		1, 0, 0, 0, 139, 141, 5, 47, 0, 0, 140, 142, 7, 3, 0, 0, 141, 140, 1, 0,
		0, 0, 141, 142, 1, 0, 0, 0, 142, 144, 1, 0, 0, 0, 143, 145, 7, 2, 0, 0,
		144, 143, 1, 0, 0, 0, 145, 146, 1, 0, 0, 0, 146, 144, 1, 0, 0, 0, 146,
		147, 1, 0, 0, 0, 147, 34, 1, 0, 0, 0, 148, 150, 7, 2, 0, 0, 149, 148, 1,
		0, 0, 0, 150, 151, 1, 0, 0, 0, 151, 149, 1, 0, 0, 0, 151, 152, 1, 0, 0,
		0, 152, 159, 1, 0, 0, 0, 153, 155, 5, 46, 0, 0, 154, 156, 7, 2, 0, 0, 155,
		154, 1, 0, 0, 0, 156, 157, 1, 0, 0, 0, 157, 155, 1, 0, 0, 0, 157, 158,
		1, 0, 0, 0, 158, 160, 1, 0, 0, 0, 159, 153, 1, 0, 0, 0, 159, 160, 1, 0,
		0, 0, 160, 161, 1, 0, 0, 0, 161, 162, 5, 37, 0, 0, 162, 36, 1, 0, 0, 0,
		163, 165, 7, 2, 0, 0, 164, 163, 1, 0, 0, 0, 165, 166, 1, 0, 0, 0, 166,
		164, 1, 0, 0, 0, 166, 167, 1, 0, 0, 0, 167, 38, 1, 0, 0, 0, 168, 170, 5,
		36, 0, 0, 169, 171, 7, 4, 0, 0, 170, 169, 1, 0, 0, 0, 171, 172, 1, 0, 0,
		0, 172, 170, 1, 0, 0, 0, 172, 173, 1, 0, 0, 0, 173, 177, 1, 0, 0, 0, 174,
		176, 7, 5, 0, 0, 175, 174, 1, 0, 0, 0, 176, 179, 1, 0, 0, 0, 177, 175,
		1, 0, 0, 0, 177, 178, 1, 0, 0, 0, 178, 40, 1, 0, 0, 0, 179, 177, 1, 0,
		0, 0, 180, 182, 5, 64, 0, 0, 181, 183, 7, 6, 0, 0, 182, 181, 1, 0, 0, 0,
		183, 184, 1, 0, 0, 0, 184, 182, 1, 0, 0, 0, 184, 185, 1, 0, 0, 0, 185,
		194, 1, 0, 0, 0, 186, 188, 5, 58, 0, 0, 187, 189, 7, 6, 0, 0, 188, 187,
		1, 0, 0, 0, 189, 190, 1, 0, 0, 0, 190, 188, 1, 0, 0, 0, 190, 191, 1, 0,
		0, 0, 191, 193, 1, 0, 0, 0, 192, 186, 1, 0, 0, 0, 193, 196, 1, 0, 0, 0,
		194, 192, 1, 0, 0, 0, 194, 195, 1, 0, 0, 0, 195, 42, 1, 0, 0, 0, 196, 194,
		1, 0, 0, 0, 197, 199, 7, 7, 0, 0, 198, 197, 1, 0, 0, 0, 199, 200, 1, 0,
		0, 0, 200, 198, 1, 0, 0, 0, 200, 201, 1, 0, 0, 0, 201, 44, 1, 0, 0, 0,
		20, 0, 48, 55, 62, 64, 78, 134, 137, 141, 146, 151, 157, 159, 166, 172,
		177, 184, 190, 194, 200, 1, 6, 0, 0,
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
	NumscriptLexerWS                         = 1
	NumscriptLexerNEWLINE                    = 2
	NumscriptLexerMULTILINE_COMMENT          = 3
	NumscriptLexerLINE_COMMENT               = 4
	NumscriptLexerSOURCE                     = 5
	NumscriptLexerDESTINATION                = 6
	NumscriptLexerSEND                       = 7
	NumscriptLexerFROM                       = 8
	NumscriptLexerTO                         = 9
	NumscriptLexerLPARENS                    = 10
	NumscriptLexerRPARENS                    = 11
	NumscriptLexerLBRACKET                   = 12
	NumscriptLexerRBRACKET                   = 13
	NumscriptLexerLBRACE                     = 14
	NumscriptLexerRBRACE                     = 15
	NumscriptLexerEQ                         = 16
	NumscriptLexerRATIO_PORTION_LITERAL      = 17
	NumscriptLexerPERCENTAGE_PORTION_LITERAL = 18
	NumscriptLexerNUMBER                     = 19
	NumscriptLexerVARIABLE_NAME              = 20
	NumscriptLexerACCOUNT                    = 21
	NumscriptLexerASSET                      = 22
)
