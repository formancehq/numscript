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
		"", "'+'", "'=='", "'!='", "'<'", "'<='", "'>'", "'>='", "'||'", "'&&'",
		"", "", "", "", "'vars'", "'max'", "'source'", "'destination'", "'send'",
		"'from'", "'up'", "'to'", "'remaining'", "'allowing'", "'unbounded'",
		"'overdraft'", "'if'", "'else'", "'kept'", "'save'", "'('", "')'", "'['",
		"']'", "'{'", "'}'", "','", "'='", "'*'", "'-'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "", "", "WS", "NEWLINE", "MULTILINE_COMMENT",
		"LINE_COMMENT", "VARS", "MAX", "SOURCE", "DESTINATION", "SEND", "FROM",
		"UP", "TO", "REMAINING", "ALLOWING", "UNBOUNDED", "OVERDRAFT", "IF",
		"ELSE", "KEPT", "SAVE", "LPARENS", "RPARENS", "LBRACKET", "RBRACKET",
		"LBRACE", "RBRACE", "COMMA", "EQ", "STAR", "MINUS", "RATIO_PORTION_LITERAL",
		"PERCENTAGE_PORTION_LITERAL", "STRING", "IDENTIFIER", "NUMBER", "VARIABLE_NAME",
		"ACCOUNT", "ASSET",
	}
	staticData.RuleNames = []string{
		"monetaryLit", "portion", "valueExpr", "functionCallArgs", "functionCall",
		"varOrigin", "varDeclaration", "varsDeclaration", "program", "sentAllLit",
		"cap", "allotment", "source", "allotmentClauseSrc", "keptOrDestination",
		"destinationInOrderClause", "destination", "allotmentClauseDest", "sentValue",
		"statement",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 47, 251, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 1, 0, 1, 0, 1,
		0, 1, 0, 1, 0, 1, 1, 1, 1, 3, 1, 48, 8, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2,
		1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 3, 2, 62, 8, 2, 1, 2, 1, 2, 1,
		2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1,
		2, 5, 2, 79, 8, 2, 10, 2, 12, 2, 82, 9, 2, 1, 3, 1, 3, 1, 3, 5, 3, 87,
		8, 3, 10, 3, 12, 3, 90, 9, 3, 1, 4, 1, 4, 1, 4, 3, 4, 95, 8, 4, 1, 4, 1,
		4, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 3, 6, 105, 8, 6, 1, 7, 1, 7, 1,
		7, 5, 7, 110, 8, 7, 10, 7, 12, 7, 113, 9, 7, 1, 7, 1, 7, 1, 8, 3, 8, 118,
		8, 8, 1, 8, 5, 8, 121, 8, 8, 10, 8, 12, 8, 124, 9, 8, 1, 8, 1, 8, 1, 9,
		1, 9, 1, 9, 1, 9, 1, 9, 1, 10, 1, 10, 3, 10, 135, 8, 10, 1, 11, 1, 11,
		1, 11, 3, 11, 140, 8, 11, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1,
		12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 4, 12, 157,
		8, 12, 11, 12, 12, 12, 158, 1, 12, 1, 12, 1, 12, 1, 12, 5, 12, 165, 8,
		12, 10, 12, 12, 12, 168, 9, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12,
		3, 12, 176, 8, 12, 1, 13, 1, 13, 1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 3,
		14, 185, 8, 14, 1, 15, 1, 15, 1, 15, 1, 15, 1, 16, 1, 16, 1, 16, 1, 16,
		4, 16, 195, 8, 16, 11, 16, 12, 16, 196, 1, 16, 1, 16, 1, 16, 1, 16, 5,
		16, 203, 8, 16, 10, 16, 12, 16, 206, 9, 16, 1, 16, 1, 16, 1, 16, 1, 16,
		3, 16, 212, 8, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 220,
		8, 16, 10, 16, 12, 16, 223, 9, 16, 1, 17, 1, 17, 1, 17, 1, 18, 1, 18, 3,
		18, 230, 8, 18, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19,
		1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 3, 19, 249,
		8, 19, 1, 19, 0, 2, 4, 32, 20, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22,
		24, 26, 28, 30, 32, 34, 36, 38, 0, 4, 2, 0, 1, 1, 39, 39, 1, 0, 2, 3, 1,
		0, 4, 7, 2, 0, 25, 25, 43, 43, 268, 0, 40, 1, 0, 0, 0, 2, 47, 1, 0, 0,
		0, 4, 61, 1, 0, 0, 0, 6, 83, 1, 0, 0, 0, 8, 91, 1, 0, 0, 0, 10, 98, 1,
		0, 0, 0, 12, 101, 1, 0, 0, 0, 14, 106, 1, 0, 0, 0, 16, 117, 1, 0, 0, 0,
		18, 127, 1, 0, 0, 0, 20, 134, 1, 0, 0, 0, 22, 139, 1, 0, 0, 0, 24, 175,
		1, 0, 0, 0, 26, 177, 1, 0, 0, 0, 28, 184, 1, 0, 0, 0, 30, 186, 1, 0, 0,
		0, 32, 211, 1, 0, 0, 0, 34, 224, 1, 0, 0, 0, 36, 229, 1, 0, 0, 0, 38, 248,
		1, 0, 0, 0, 40, 41, 5, 32, 0, 0, 41, 42, 3, 4, 2, 0, 42, 43, 3, 4, 2, 0,
		43, 44, 5, 33, 0, 0, 44, 1, 1, 0, 0, 0, 45, 48, 5, 40, 0, 0, 46, 48, 5,
		41, 0, 0, 47, 45, 1, 0, 0, 0, 47, 46, 1, 0, 0, 0, 48, 3, 1, 0, 0, 0, 49,
		50, 6, 2, -1, 0, 50, 62, 5, 45, 0, 0, 51, 62, 5, 47, 0, 0, 52, 62, 5, 42,
		0, 0, 53, 62, 5, 46, 0, 0, 54, 62, 5, 44, 0, 0, 55, 62, 3, 0, 0, 0, 56,
		62, 3, 2, 1, 0, 57, 58, 5, 30, 0, 0, 58, 59, 3, 4, 2, 0, 59, 60, 5, 31,
		0, 0, 60, 62, 1, 0, 0, 0, 61, 49, 1, 0, 0, 0, 61, 51, 1, 0, 0, 0, 61, 52,
		1, 0, 0, 0, 61, 53, 1, 0, 0, 0, 61, 54, 1, 0, 0, 0, 61, 55, 1, 0, 0, 0,
		61, 56, 1, 0, 0, 0, 61, 57, 1, 0, 0, 0, 62, 80, 1, 0, 0, 0, 63, 64, 10,
		6, 0, 0, 64, 65, 7, 0, 0, 0, 65, 79, 3, 4, 2, 7, 66, 67, 10, 5, 0, 0, 67,
		68, 7, 1, 0, 0, 68, 79, 3, 4, 2, 6, 69, 70, 10, 4, 0, 0, 70, 71, 7, 2,
		0, 0, 71, 79, 3, 4, 2, 5, 72, 73, 10, 3, 0, 0, 73, 74, 5, 8, 0, 0, 74,
		79, 3, 4, 2, 4, 75, 76, 10, 2, 0, 0, 76, 77, 5, 9, 0, 0, 77, 79, 3, 4,
		2, 3, 78, 63, 1, 0, 0, 0, 78, 66, 1, 0, 0, 0, 78, 69, 1, 0, 0, 0, 78, 72,
		1, 0, 0, 0, 78, 75, 1, 0, 0, 0, 79, 82, 1, 0, 0, 0, 80, 78, 1, 0, 0, 0,
		80, 81, 1, 0, 0, 0, 81, 5, 1, 0, 0, 0, 82, 80, 1, 0, 0, 0, 83, 88, 3, 4,
		2, 0, 84, 85, 5, 36, 0, 0, 85, 87, 3, 4, 2, 0, 86, 84, 1, 0, 0, 0, 87,
		90, 1, 0, 0, 0, 88, 86, 1, 0, 0, 0, 88, 89, 1, 0, 0, 0, 89, 7, 1, 0, 0,
		0, 90, 88, 1, 0, 0, 0, 91, 92, 7, 3, 0, 0, 92, 94, 5, 30, 0, 0, 93, 95,
		3, 6, 3, 0, 94, 93, 1, 0, 0, 0, 94, 95, 1, 0, 0, 0, 95, 96, 1, 0, 0, 0,
		96, 97, 5, 31, 0, 0, 97, 9, 1, 0, 0, 0, 98, 99, 5, 37, 0, 0, 99, 100, 3,
		8, 4, 0, 100, 11, 1, 0, 0, 0, 101, 102, 5, 43, 0, 0, 102, 104, 5, 45, 0,
		0, 103, 105, 3, 10, 5, 0, 104, 103, 1, 0, 0, 0, 104, 105, 1, 0, 0, 0, 105,
		13, 1, 0, 0, 0, 106, 107, 5, 14, 0, 0, 107, 111, 5, 34, 0, 0, 108, 110,
		3, 12, 6, 0, 109, 108, 1, 0, 0, 0, 110, 113, 1, 0, 0, 0, 111, 109, 1, 0,
		0, 0, 111, 112, 1, 0, 0, 0, 112, 114, 1, 0, 0, 0, 113, 111, 1, 0, 0, 0,
		114, 115, 5, 35, 0, 0, 115, 15, 1, 0, 0, 0, 116, 118, 3, 14, 7, 0, 117,
		116, 1, 0, 0, 0, 117, 118, 1, 0, 0, 0, 118, 122, 1, 0, 0, 0, 119, 121,
		3, 38, 19, 0, 120, 119, 1, 0, 0, 0, 121, 124, 1, 0, 0, 0, 122, 120, 1,
		0, 0, 0, 122, 123, 1, 0, 0, 0, 123, 125, 1, 0, 0, 0, 124, 122, 1, 0, 0,
		0, 125, 126, 5, 0, 0, 1, 126, 17, 1, 0, 0, 0, 127, 128, 5, 32, 0, 0, 128,
		129, 3, 4, 2, 0, 129, 130, 5, 38, 0, 0, 130, 131, 5, 33, 0, 0, 131, 19,
		1, 0, 0, 0, 132, 135, 3, 0, 0, 0, 133, 135, 5, 45, 0, 0, 134, 132, 1, 0,
		0, 0, 134, 133, 1, 0, 0, 0, 135, 21, 1, 0, 0, 0, 136, 140, 3, 2, 1, 0,
		137, 140, 5, 45, 0, 0, 138, 140, 5, 22, 0, 0, 139, 136, 1, 0, 0, 0, 139,
		137, 1, 0, 0, 0, 139, 138, 1, 0, 0, 0, 140, 23, 1, 0, 0, 0, 141, 142, 3,
		4, 2, 0, 142, 143, 5, 23, 0, 0, 143, 144, 5, 24, 0, 0, 144, 145, 5, 25,
		0, 0, 145, 176, 1, 0, 0, 0, 146, 147, 3, 4, 2, 0, 147, 148, 5, 23, 0, 0,
		148, 149, 5, 25, 0, 0, 149, 150, 5, 20, 0, 0, 150, 151, 5, 21, 0, 0, 151,
		152, 3, 4, 2, 0, 152, 176, 1, 0, 0, 0, 153, 176, 3, 4, 2, 0, 154, 156,
		5, 34, 0, 0, 155, 157, 3, 26, 13, 0, 156, 155, 1, 0, 0, 0, 157, 158, 1,
		0, 0, 0, 158, 156, 1, 0, 0, 0, 158, 159, 1, 0, 0, 0, 159, 160, 1, 0, 0,
		0, 160, 161, 5, 35, 0, 0, 161, 176, 1, 0, 0, 0, 162, 166, 5, 34, 0, 0,
		163, 165, 3, 24, 12, 0, 164, 163, 1, 0, 0, 0, 165, 168, 1, 0, 0, 0, 166,
		164, 1, 0, 0, 0, 166, 167, 1, 0, 0, 0, 167, 169, 1, 0, 0, 0, 168, 166,
		1, 0, 0, 0, 169, 176, 5, 35, 0, 0, 170, 171, 5, 15, 0, 0, 171, 172, 3,
		20, 10, 0, 172, 173, 5, 19, 0, 0, 173, 174, 3, 24, 12, 0, 174, 176, 1,
		0, 0, 0, 175, 141, 1, 0, 0, 0, 175, 146, 1, 0, 0, 0, 175, 153, 1, 0, 0,
		0, 175, 154, 1, 0, 0, 0, 175, 162, 1, 0, 0, 0, 175, 170, 1, 0, 0, 0, 176,
		25, 1, 0, 0, 0, 177, 178, 3, 22, 11, 0, 178, 179, 5, 19, 0, 0, 179, 180,
		3, 24, 12, 0, 180, 27, 1, 0, 0, 0, 181, 182, 5, 21, 0, 0, 182, 185, 3,
		32, 16, 0, 183, 185, 5, 28, 0, 0, 184, 181, 1, 0, 0, 0, 184, 183, 1, 0,
		0, 0, 185, 29, 1, 0, 0, 0, 186, 187, 5, 15, 0, 0, 187, 188, 3, 4, 2, 0,
		188, 189, 3, 28, 14, 0, 189, 31, 1, 0, 0, 0, 190, 191, 6, 16, -1, 0, 191,
		212, 3, 4, 2, 0, 192, 194, 5, 34, 0, 0, 193, 195, 3, 34, 17, 0, 194, 193,
		1, 0, 0, 0, 195, 196, 1, 0, 0, 0, 196, 194, 1, 0, 0, 0, 196, 197, 1, 0,
		0, 0, 197, 198, 1, 0, 0, 0, 198, 199, 5, 35, 0, 0, 199, 212, 1, 0, 0, 0,
		200, 204, 5, 34, 0, 0, 201, 203, 3, 30, 15, 0, 202, 201, 1, 0, 0, 0, 203,
		206, 1, 0, 0, 0, 204, 202, 1, 0, 0, 0, 204, 205, 1, 0, 0, 0, 205, 207,
		1, 0, 0, 0, 206, 204, 1, 0, 0, 0, 207, 208, 5, 22, 0, 0, 208, 209, 3, 28,
		14, 0, 209, 210, 5, 35, 0, 0, 210, 212, 1, 0, 0, 0, 211, 190, 1, 0, 0,
		0, 211, 192, 1, 0, 0, 0, 211, 200, 1, 0, 0, 0, 212, 221, 1, 0, 0, 0, 213,
		214, 10, 3, 0, 0, 214, 215, 5, 26, 0, 0, 215, 216, 3, 4, 2, 0, 216, 217,
		5, 27, 0, 0, 217, 218, 3, 32, 16, 4, 218, 220, 1, 0, 0, 0, 219, 213, 1,
		0, 0, 0, 220, 223, 1, 0, 0, 0, 221, 219, 1, 0, 0, 0, 221, 222, 1, 0, 0,
		0, 222, 33, 1, 0, 0, 0, 223, 221, 1, 0, 0, 0, 224, 225, 3, 22, 11, 0, 225,
		226, 3, 28, 14, 0, 226, 35, 1, 0, 0, 0, 227, 230, 3, 4, 2, 0, 228, 230,
		3, 18, 9, 0, 229, 227, 1, 0, 0, 0, 229, 228, 1, 0, 0, 0, 230, 37, 1, 0,
		0, 0, 231, 232, 5, 18, 0, 0, 232, 233, 3, 36, 18, 0, 233, 234, 5, 30, 0,
		0, 234, 235, 5, 16, 0, 0, 235, 236, 5, 37, 0, 0, 236, 237, 3, 24, 12, 0,
		237, 238, 5, 17, 0, 0, 238, 239, 5, 37, 0, 0, 239, 240, 3, 32, 16, 0, 240,
		241, 5, 31, 0, 0, 241, 249, 1, 0, 0, 0, 242, 243, 5, 29, 0, 0, 243, 244,
		3, 36, 18, 0, 244, 245, 5, 19, 0, 0, 245, 246, 3, 4, 2, 0, 246, 249, 1,
		0, 0, 0, 247, 249, 3, 8, 4, 0, 248, 231, 1, 0, 0, 0, 248, 242, 1, 0, 0,
		0, 248, 247, 1, 0, 0, 0, 249, 39, 1, 0, 0, 0, 22, 47, 61, 78, 80, 88, 94,
		104, 111, 117, 122, 134, 139, 158, 166, 175, 184, 196, 204, 211, 221, 229,
		248,
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
	NumscriptParserT__0                       = 1
	NumscriptParserT__1                       = 2
	NumscriptParserT__2                       = 3
	NumscriptParserT__3                       = 4
	NumscriptParserT__4                       = 5
	NumscriptParserT__5                       = 6
	NumscriptParserT__6                       = 7
	NumscriptParserT__7                       = 8
	NumscriptParserT__8                       = 9
	NumscriptParserWS                         = 10
	NumscriptParserNEWLINE                    = 11
	NumscriptParserMULTILINE_COMMENT          = 12
	NumscriptParserLINE_COMMENT               = 13
	NumscriptParserVARS                       = 14
	NumscriptParserMAX                        = 15
	NumscriptParserSOURCE                     = 16
	NumscriptParserDESTINATION                = 17
	NumscriptParserSEND                       = 18
	NumscriptParserFROM                       = 19
	NumscriptParserUP                         = 20
	NumscriptParserTO                         = 21
	NumscriptParserREMAINING                  = 22
	NumscriptParserALLOWING                   = 23
	NumscriptParserUNBOUNDED                  = 24
	NumscriptParserOVERDRAFT                  = 25
	NumscriptParserIF                         = 26
	NumscriptParserELSE                       = 27
	NumscriptParserKEPT                       = 28
	NumscriptParserSAVE                       = 29
	NumscriptParserLPARENS                    = 30
	NumscriptParserRPARENS                    = 31
	NumscriptParserLBRACKET                   = 32
	NumscriptParserRBRACKET                   = 33
	NumscriptParserLBRACE                     = 34
	NumscriptParserRBRACE                     = 35
	NumscriptParserCOMMA                      = 36
	NumscriptParserEQ                         = 37
	NumscriptParserSTAR                       = 38
	NumscriptParserMINUS                      = 39
	NumscriptParserRATIO_PORTION_LITERAL      = 40
	NumscriptParserPERCENTAGE_PORTION_LITERAL = 41
	NumscriptParserSTRING                     = 42
	NumscriptParserIDENTIFIER                 = 43
	NumscriptParserNUMBER                     = 44
	NumscriptParserVARIABLE_NAME              = 45
	NumscriptParserACCOUNT                    = 46
	NumscriptParserASSET                      = 47
)

// NumscriptParser rules.
const (
	NumscriptParserRULE_monetaryLit              = 0
	NumscriptParserRULE_portion                  = 1
	NumscriptParserRULE_valueExpr                = 2
	NumscriptParserRULE_functionCallArgs         = 3
	NumscriptParserRULE_functionCall             = 4
	NumscriptParserRULE_varOrigin                = 5
	NumscriptParserRULE_varDeclaration           = 6
	NumscriptParserRULE_varsDeclaration          = 7
	NumscriptParserRULE_program                  = 8
	NumscriptParserRULE_sentAllLit               = 9
	NumscriptParserRULE_cap                      = 10
	NumscriptParserRULE_allotment                = 11
	NumscriptParserRULE_source                   = 12
	NumscriptParserRULE_allotmentClauseSrc       = 13
	NumscriptParserRULE_keptOrDestination        = 14
	NumscriptParserRULE_destinationInOrderClause = 15
	NumscriptParserRULE_destination              = 16
	NumscriptParserRULE_allotmentClauseDest      = 17
	NumscriptParserRULE_sentValue                = 18
	NumscriptParserRULE_statement                = 19
)

// IMonetaryLitContext is an interface to support dynamic dispatch.
type IMonetaryLitContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetAsset returns the asset rule contexts.
	GetAsset() IValueExprContext

	// GetAmt returns the amt rule contexts.
	GetAmt() IValueExprContext

	// SetAsset sets the asset rule contexts.
	SetAsset(IValueExprContext)

	// SetAmt sets the amt rule contexts.
	SetAmt(IValueExprContext)

	// Getter signatures
	LBRACKET() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode
	AllValueExpr() []IValueExprContext
	ValueExpr(i int) IValueExprContext

	// IsMonetaryLitContext differentiates from other interfaces.
	IsMonetaryLitContext()
}

type MonetaryLitContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	asset  IValueExprContext
	amt    IValueExprContext
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

func (s *MonetaryLitContext) GetAsset() IValueExprContext { return s.asset }

func (s *MonetaryLitContext) GetAmt() IValueExprContext { return s.amt }

func (s *MonetaryLitContext) SetAsset(v IValueExprContext) { s.asset = v }

func (s *MonetaryLitContext) SetAmt(v IValueExprContext) { s.amt = v }

func (s *MonetaryLitContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACKET, 0)
}

func (s *MonetaryLitContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACKET, 0)
}

func (s *MonetaryLitContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *MonetaryLitContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
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

	return t.(IValueExprContext)
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

func (p *NumscriptParser) MonetaryLit() (localctx IMonetaryLitContext) {
	localctx = NewMonetaryLitContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, NumscriptParserRULE_monetaryLit)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(40)
		p.Match(NumscriptParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(41)

		var _x = p.valueExpr(0)

		localctx.(*MonetaryLitContext).asset = _x
	}

	{
		p.SetState(42)

		var _x = p.valueExpr(0)

		localctx.(*MonetaryLitContext).amt = _x
	}

	{
		p.SetState(43)
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

func (p *NumscriptParser) Portion() (localctx IPortionContext) {
	localctx = NewPortionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, NumscriptParserRULE_portion)
	p.SetState(47)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserRATIO_PORTION_LITERAL:
		localctx = NewRatioContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(45)
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
			p.SetState(46)
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

// IValueExprContext is an interface to support dynamic dispatch.
type IValueExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsValueExprContext differentiates from other interfaces.
	IsValueExprContext()
}

type ValueExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueExprContext() *ValueExprContext {
	var p = new(ValueExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_valueExpr
	return p
}

func InitEmptyValueExprContext(p *ValueExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_valueExpr
}

func (*ValueExprContext) IsValueExprContext() {}

func NewValueExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueExprContext {
	var p = new(ValueExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_valueExpr

	return p
}

func (s *ValueExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueExprContext) CopyAll(ctx *ValueExprContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ValueExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type InfixCompExprContext struct {
	ValueExprContext
	left  IValueExprContext
	op    antlr.Token
	right IValueExprContext
}

func NewInfixCompExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InfixCompExprContext {
	var p = new(InfixCompExprContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *InfixCompExprContext) GetOp() antlr.Token { return s.op }

func (s *InfixCompExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *InfixCompExprContext) GetLeft() IValueExprContext { return s.left }

func (s *InfixCompExprContext) GetRight() IValueExprContext { return s.right }

func (s *InfixCompExprContext) SetLeft(v IValueExprContext) { s.left = v }

func (s *InfixCompExprContext) SetRight(v IValueExprContext) { s.right = v }

func (s *InfixCompExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InfixCompExprContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *InfixCompExprContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
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

	return t.(IValueExprContext)
}

func (s *InfixCompExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterInfixCompExpr(s)
	}
}

func (s *InfixCompExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitInfixCompExpr(s)
	}
}

type AccountLiteralContext struct {
	ValueExprContext
}

func NewAccountLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AccountLiteralContext {
	var p = new(AccountLiteralContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *AccountLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AccountLiteralContext) ACCOUNT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserACCOUNT, 0)
}

func (s *AccountLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterAccountLiteral(s)
	}
}

func (s *AccountLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitAccountLiteral(s)
	}
}

type ParensExprContext struct {
	ValueExprContext
}

func NewParensExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ParensExprContext {
	var p = new(ParensExprContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *ParensExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParensExprContext) LPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLPARENS, 0)
}

func (s *ParensExprContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *ParensExprContext) RPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRPARENS, 0)
}

func (s *ParensExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterParensExpr(s)
	}
}

func (s *ParensExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitParensExpr(s)
	}
}

type MonetaryLiteralContext struct {
	ValueExprContext
}

func NewMonetaryLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MonetaryLiteralContext {
	var p = new(MonetaryLiteralContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *MonetaryLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MonetaryLiteralContext) MonetaryLit() IMonetaryLitContext {
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

func (s *MonetaryLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterMonetaryLiteral(s)
	}
}

func (s *MonetaryLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitMonetaryLiteral(s)
	}
}

type InfixEqExprContext struct {
	ValueExprContext
	left  IValueExprContext
	op    antlr.Token
	right IValueExprContext
}

func NewInfixEqExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InfixEqExprContext {
	var p = new(InfixEqExprContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *InfixEqExprContext) GetOp() antlr.Token { return s.op }

func (s *InfixEqExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *InfixEqExprContext) GetLeft() IValueExprContext { return s.left }

func (s *InfixEqExprContext) GetRight() IValueExprContext { return s.right }

func (s *InfixEqExprContext) SetLeft(v IValueExprContext) { s.left = v }

func (s *InfixEqExprContext) SetRight(v IValueExprContext) { s.right = v }

func (s *InfixEqExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InfixEqExprContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *InfixEqExprContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
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

	return t.(IValueExprContext)
}

func (s *InfixEqExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterInfixEqExpr(s)
	}
}

func (s *InfixEqExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitInfixEqExpr(s)
	}
}

type VariableExprContext struct {
	ValueExprContext
}

func NewVariableExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VariableExprContext {
	var p = new(VariableExprContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *VariableExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VariableExprContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARIABLE_NAME, 0)
}

func (s *VariableExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterVariableExpr(s)
	}
}

func (s *VariableExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitVariableExpr(s)
	}
}

type PortionLiteralContext struct {
	ValueExprContext
}

func NewPortionLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PortionLiteralContext {
	var p = new(PortionLiteralContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *PortionLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PortionLiteralContext) Portion() IPortionContext {
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

func (s *PortionLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterPortionLiteral(s)
	}
}

func (s *PortionLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitPortionLiteral(s)
	}
}

type InfixAndExprContext struct {
	ValueExprContext
	left  IValueExprContext
	op    antlr.Token
	right IValueExprContext
}

func NewInfixAndExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InfixAndExprContext {
	var p = new(InfixAndExprContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *InfixAndExprContext) GetOp() antlr.Token { return s.op }

func (s *InfixAndExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *InfixAndExprContext) GetLeft() IValueExprContext { return s.left }

func (s *InfixAndExprContext) GetRight() IValueExprContext { return s.right }

func (s *InfixAndExprContext) SetLeft(v IValueExprContext) { s.left = v }

func (s *InfixAndExprContext) SetRight(v IValueExprContext) { s.right = v }

func (s *InfixAndExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InfixAndExprContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *InfixAndExprContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
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

	return t.(IValueExprContext)
}

func (s *InfixAndExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterInfixAndExpr(s)
	}
}

func (s *InfixAndExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitInfixAndExpr(s)
	}
}

type AssetLiteralContext struct {
	ValueExprContext
}

func NewAssetLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AssetLiteralContext {
	var p = new(AssetLiteralContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *AssetLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AssetLiteralContext) ASSET() antlr.TerminalNode {
	return s.GetToken(NumscriptParserASSET, 0)
}

func (s *AssetLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterAssetLiteral(s)
	}
}

func (s *AssetLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitAssetLiteral(s)
	}
}

type StringLiteralContext struct {
	ValueExprContext
}

func NewStringLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *StringLiteralContext {
	var p = new(StringLiteralContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *StringLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StringLiteralContext) STRING() antlr.TerminalNode {
	return s.GetToken(NumscriptParserSTRING, 0)
}

func (s *StringLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterStringLiteral(s)
	}
}

func (s *StringLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitStringLiteral(s)
	}
}

type InfixOrExprContext struct {
	ValueExprContext
	left  IValueExprContext
	op    antlr.Token
	right IValueExprContext
}

func NewInfixOrExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InfixOrExprContext {
	var p = new(InfixOrExprContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *InfixOrExprContext) GetOp() antlr.Token { return s.op }

func (s *InfixOrExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *InfixOrExprContext) GetLeft() IValueExprContext { return s.left }

func (s *InfixOrExprContext) GetRight() IValueExprContext { return s.right }

func (s *InfixOrExprContext) SetLeft(v IValueExprContext) { s.left = v }

func (s *InfixOrExprContext) SetRight(v IValueExprContext) { s.right = v }

func (s *InfixOrExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InfixOrExprContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *InfixOrExprContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
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

	return t.(IValueExprContext)
}

func (s *InfixOrExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterInfixOrExpr(s)
	}
}

func (s *InfixOrExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitInfixOrExpr(s)
	}
}

type InfixAddSubExprContext struct {
	ValueExprContext
	left  IValueExprContext
	op    antlr.Token
	right IValueExprContext
}

func NewInfixAddSubExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InfixAddSubExprContext {
	var p = new(InfixAddSubExprContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *InfixAddSubExprContext) GetOp() antlr.Token { return s.op }

func (s *InfixAddSubExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *InfixAddSubExprContext) GetLeft() IValueExprContext { return s.left }

func (s *InfixAddSubExprContext) GetRight() IValueExprContext { return s.right }

func (s *InfixAddSubExprContext) SetLeft(v IValueExprContext) { s.left = v }

func (s *InfixAddSubExprContext) SetRight(v IValueExprContext) { s.right = v }

func (s *InfixAddSubExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InfixAddSubExprContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *InfixAddSubExprContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
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

	return t.(IValueExprContext)
}

func (s *InfixAddSubExprContext) MINUS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserMINUS, 0)
}

func (s *InfixAddSubExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterInfixAddSubExpr(s)
	}
}

func (s *InfixAddSubExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitInfixAddSubExpr(s)
	}
}

type NumberLiteralContext struct {
	ValueExprContext
}

func NewNumberLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NumberLiteralContext {
	var p = new(NumberLiteralContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *NumberLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NumberLiteralContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(NumscriptParserNUMBER, 0)
}

func (s *NumberLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterNumberLiteral(s)
	}
}

func (s *NumberLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitNumberLiteral(s)
	}
}

func (p *NumscriptParser) ValueExpr() (localctx IValueExprContext) {
	return p.valueExpr(0)
}

func (p *NumscriptParser) valueExpr(_p int) (localctx IValueExprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewValueExprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IValueExprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 4
	p.EnterRecursionRule(localctx, 4, NumscriptParserRULE_valueExpr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(61)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserVARIABLE_NAME:
		localctx = NewVariableExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(50)
			p.Match(NumscriptParserVARIABLE_NAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserASSET:
		localctx = NewAssetLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(51)
			p.Match(NumscriptParserASSET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserSTRING:
		localctx = NewStringLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(52)
			p.Match(NumscriptParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserACCOUNT:
		localctx = NewAccountLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(53)
			p.Match(NumscriptParserACCOUNT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserNUMBER:
		localctx = NewNumberLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(54)
			p.Match(NumscriptParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserLBRACKET:
		localctx = NewMonetaryLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(55)
			p.MonetaryLit()
		}

	case NumscriptParserRATIO_PORTION_LITERAL, NumscriptParserPERCENTAGE_PORTION_LITERAL:
		localctx = NewPortionLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(56)
			p.Portion()
		}

	case NumscriptParserLPARENS:
		localctx = NewParensExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(57)
			p.Match(NumscriptParserLPARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(58)
			p.valueExpr(0)
		}
		{
			p.SetState(59)
			p.Match(NumscriptParserRPARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(80)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(78)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext()) {
			case 1:
				localctx = NewInfixAddSubExprContext(p, NewValueExprContext(p, _parentctx, _parentState))
				localctx.(*InfixAddSubExprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_valueExpr)
				p.SetState(63)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
					goto errorExit
				}
				{
					p.SetState(64)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*InfixAddSubExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == NumscriptParserT__0 || _la == NumscriptParserMINUS) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*InfixAddSubExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(65)

					var _x = p.valueExpr(7)

					localctx.(*InfixAddSubExprContext).right = _x
				}

			case 2:
				localctx = NewInfixEqExprContext(p, NewValueExprContext(p, _parentctx, _parentState))
				localctx.(*InfixEqExprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_valueExpr)
				p.SetState(66)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(67)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*InfixEqExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == NumscriptParserT__1 || _la == NumscriptParserT__2) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*InfixEqExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(68)

					var _x = p.valueExpr(6)

					localctx.(*InfixEqExprContext).right = _x
				}

			case 3:
				localctx = NewInfixCompExprContext(p, NewValueExprContext(p, _parentctx, _parentState))
				localctx.(*InfixCompExprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_valueExpr)
				p.SetState(69)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(70)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*InfixCompExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&240) != 0) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*InfixCompExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(71)

					var _x = p.valueExpr(5)

					localctx.(*InfixCompExprContext).right = _x
				}

			case 4:
				localctx = NewInfixOrExprContext(p, NewValueExprContext(p, _parentctx, _parentState))
				localctx.(*InfixOrExprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_valueExpr)
				p.SetState(72)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(73)

					var _m = p.Match(NumscriptParserT__7)

					localctx.(*InfixOrExprContext).op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(74)

					var _x = p.valueExpr(4)

					localctx.(*InfixOrExprContext).right = _x
				}

			case 5:
				localctx = NewInfixAndExprContext(p, NewValueExprContext(p, _parentctx, _parentState))
				localctx.(*InfixAndExprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_valueExpr)
				p.SetState(75)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(76)

					var _m = p.Match(NumscriptParserT__8)

					localctx.(*InfixAndExprContext).op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(77)

					var _x = p.valueExpr(3)

					localctx.(*InfixAndExprContext).right = _x
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(82)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext())
		if p.HasError() {
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
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFunctionCallArgsContext is an interface to support dynamic dispatch.
type IFunctionCallArgsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllValueExpr() []IValueExprContext
	ValueExpr(i int) IValueExprContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsFunctionCallArgsContext differentiates from other interfaces.
	IsFunctionCallArgsContext()
}

type FunctionCallArgsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunctionCallArgsContext() *FunctionCallArgsContext {
	var p = new(FunctionCallArgsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_functionCallArgs
	return p
}

func InitEmptyFunctionCallArgsContext(p *FunctionCallArgsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_functionCallArgs
}

func (*FunctionCallArgsContext) IsFunctionCallArgsContext() {}

func NewFunctionCallArgsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FunctionCallArgsContext {
	var p = new(FunctionCallArgsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_functionCallArgs

	return p
}

func (s *FunctionCallArgsContext) GetParser() antlr.Parser { return s.parser }

func (s *FunctionCallArgsContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *FunctionCallArgsContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
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

	return t.(IValueExprContext)
}

func (s *FunctionCallArgsContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(NumscriptParserCOMMA)
}

func (s *FunctionCallArgsContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(NumscriptParserCOMMA, i)
}

func (s *FunctionCallArgsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionCallArgsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FunctionCallArgsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterFunctionCallArgs(s)
	}
}

func (s *FunctionCallArgsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitFunctionCallArgs(s)
	}
}

func (p *NumscriptParser) FunctionCallArgs() (localctx IFunctionCallArgsContext) {
	localctx = NewFunctionCallArgsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, NumscriptParserRULE_functionCallArgs)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(83)
		p.valueExpr(0)
	}
	p.SetState(88)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NumscriptParserCOMMA {
		{
			p.SetState(84)
			p.Match(NumscriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(85)
			p.valueExpr(0)
		}

		p.SetState(90)
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

// IFunctionCallContext is an interface to support dynamic dispatch.
type IFunctionCallContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetFnName returns the fnName token.
	GetFnName() antlr.Token

	// SetFnName sets the fnName token.
	SetFnName(antlr.Token)

	// Getter signatures
	LPARENS() antlr.TerminalNode
	RPARENS() antlr.TerminalNode
	OVERDRAFT() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode
	FunctionCallArgs() IFunctionCallArgsContext

	// IsFunctionCallContext differentiates from other interfaces.
	IsFunctionCallContext()
}

type FunctionCallContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	fnName antlr.Token
}

func NewEmptyFunctionCallContext() *FunctionCallContext {
	var p = new(FunctionCallContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_functionCall
	return p
}

func InitEmptyFunctionCallContext(p *FunctionCallContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_functionCall
}

func (*FunctionCallContext) IsFunctionCallContext() {}

func NewFunctionCallContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FunctionCallContext {
	var p = new(FunctionCallContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_functionCall

	return p
}

func (s *FunctionCallContext) GetParser() antlr.Parser { return s.parser }

func (s *FunctionCallContext) GetFnName() antlr.Token { return s.fnName }

func (s *FunctionCallContext) SetFnName(v antlr.Token) { s.fnName = v }

func (s *FunctionCallContext) LPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLPARENS, 0)
}

func (s *FunctionCallContext) RPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRPARENS, 0)
}

func (s *FunctionCallContext) OVERDRAFT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserOVERDRAFT, 0)
}

func (s *FunctionCallContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NumscriptParserIDENTIFIER, 0)
}

func (s *FunctionCallContext) FunctionCallArgs() IFunctionCallArgsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionCallArgsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionCallArgsContext)
}

func (s *FunctionCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionCallContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FunctionCallContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterFunctionCall(s)
	}
}

func (s *FunctionCallContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitFunctionCall(s)
	}
}

func (p *NumscriptParser) FunctionCall() (localctx IFunctionCallContext) {
	localctx = NewFunctionCallContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, NumscriptParserRULE_functionCall)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(91)

		var _lt = p.GetTokenStream().LT(1)

		localctx.(*FunctionCallContext).fnName = _lt

		_la = p.GetTokenStream().LA(1)

		if !(_la == NumscriptParserOVERDRAFT || _la == NumscriptParserIDENTIFIER) {
			var _ri = p.GetErrorHandler().RecoverInline(p)

			localctx.(*FunctionCallContext).fnName = _ri
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(92)
		p.Match(NumscriptParserLPARENS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(94)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&271584740769792) != 0 {
		{
			p.SetState(93)
			p.FunctionCallArgs()
		}

	}
	{
		p.SetState(96)
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

// IVarOriginContext is an interface to support dynamic dispatch.
type IVarOriginContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EQ() antlr.TerminalNode
	FunctionCall() IFunctionCallContext

	// IsVarOriginContext differentiates from other interfaces.
	IsVarOriginContext()
}

type VarOriginContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyVarOriginContext() *VarOriginContext {
	var p = new(VarOriginContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_varOrigin
	return p
}

func InitEmptyVarOriginContext(p *VarOriginContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_varOrigin
}

func (*VarOriginContext) IsVarOriginContext() {}

func NewVarOriginContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *VarOriginContext {
	var p = new(VarOriginContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_varOrigin

	return p
}

func (s *VarOriginContext) GetParser() antlr.Parser { return s.parser }

func (s *VarOriginContext) EQ() antlr.TerminalNode {
	return s.GetToken(NumscriptParserEQ, 0)
}

func (s *VarOriginContext) FunctionCall() IFunctionCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionCallContext)
}

func (s *VarOriginContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VarOriginContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *VarOriginContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterVarOrigin(s)
	}
}

func (s *VarOriginContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitVarOrigin(s)
	}
}

func (p *NumscriptParser) VarOrigin() (localctx IVarOriginContext) {
	localctx = NewVarOriginContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, NumscriptParserRULE_varOrigin)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(98)
		p.Match(NumscriptParserEQ)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(99)
		p.FunctionCall()
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
	IDENTIFIER() antlr.TerminalNode
	VARIABLE_NAME() antlr.TerminalNode
	VarOrigin() IVarOriginContext

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

func (s *VarDeclarationContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NumscriptParserIDENTIFIER, 0)
}

func (s *VarDeclarationContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARIABLE_NAME, 0)
}

func (s *VarDeclarationContext) VarOrigin() IVarOriginContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVarOriginContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVarOriginContext)
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

func (p *NumscriptParser) VarDeclaration() (localctx IVarDeclarationContext) {
	localctx = NewVarDeclarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, NumscriptParserRULE_varDeclaration)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(101)

		var _m = p.Match(NumscriptParserIDENTIFIER)

		localctx.(*VarDeclarationContext).type_ = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(102)

		var _m = p.Match(NumscriptParserVARIABLE_NAME)

		localctx.(*VarDeclarationContext).name = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(104)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NumscriptParserEQ {
		{
			p.SetState(103)
			p.VarOrigin()
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

func (p *NumscriptParser) VarsDeclaration() (localctx IVarsDeclarationContext) {
	localctx = NewVarsDeclarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, NumscriptParserRULE_varsDeclaration)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(106)
		p.Match(NumscriptParserVARS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(107)
		p.Match(NumscriptParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(111)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NumscriptParserIDENTIFIER {
		{
			p.SetState(108)
			p.VarDeclaration()
		}

		p.SetState(113)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(114)
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
	EOF() antlr.TerminalNode
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

func (s *ProgramContext) EOF() antlr.TerminalNode {
	return s.GetToken(NumscriptParserEOF, 0)
}

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

func (p *NumscriptParser) Program() (localctx IProgramContext) {
	localctx = NewProgramContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, NumscriptParserRULE_program)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(117)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NumscriptParserVARS {
		{
			p.SetState(116)
			p.VarsDeclaration()
		}

	}
	p.SetState(122)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&8796663709696) != 0 {
		{
			p.SetState(119)
			p.Statement()
		}

		p.SetState(124)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(125)
		p.Match(NumscriptParserEOF)
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

// ISentAllLitContext is an interface to support dynamic dispatch.
type ISentAllLitContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetAsset returns the asset rule contexts.
	GetAsset() IValueExprContext

	// SetAsset sets the asset rule contexts.
	SetAsset(IValueExprContext)

	// Getter signatures
	LBRACKET() antlr.TerminalNode
	STAR() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode
	ValueExpr() IValueExprContext

	// IsSentAllLitContext differentiates from other interfaces.
	IsSentAllLitContext()
}

type SentAllLitContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	asset  IValueExprContext
}

func NewEmptySentAllLitContext() *SentAllLitContext {
	var p = new(SentAllLitContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_sentAllLit
	return p
}

func InitEmptySentAllLitContext(p *SentAllLitContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_sentAllLit
}

func (*SentAllLitContext) IsSentAllLitContext() {}

func NewSentAllLitContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SentAllLitContext {
	var p = new(SentAllLitContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_sentAllLit

	return p
}

func (s *SentAllLitContext) GetParser() antlr.Parser { return s.parser }

func (s *SentAllLitContext) GetAsset() IValueExprContext { return s.asset }

func (s *SentAllLitContext) SetAsset(v IValueExprContext) { s.asset = v }

func (s *SentAllLitContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACKET, 0)
}

func (s *SentAllLitContext) STAR() antlr.TerminalNode {
	return s.GetToken(NumscriptParserSTAR, 0)
}

func (s *SentAllLitContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACKET, 0)
}

func (s *SentAllLitContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *SentAllLitContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SentAllLitContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SentAllLitContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSentAllLit(s)
	}
}

func (s *SentAllLitContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSentAllLit(s)
	}
}

func (p *NumscriptParser) SentAllLit() (localctx ISentAllLitContext) {
	localctx = NewSentAllLitContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, NumscriptParserRULE_sentAllLit)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(127)
		p.Match(NumscriptParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(128)

		var _x = p.valueExpr(0)

		localctx.(*SentAllLitContext).asset = _x
	}

	{
		p.SetState(129)
		p.Match(NumscriptParserSTAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(130)
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

func (p *NumscriptParser) Cap_() (localctx ICapContext) {
	localctx = NewCapContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, NumscriptParserRULE_cap)
	p.SetState(134)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserLBRACKET:
		localctx = NewLitCapContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(132)
			p.MonetaryLit()
		}

	case NumscriptParserVARIABLE_NAME:
		localctx = NewVarCapContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(133)
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

// IAllotmentContext is an interface to support dynamic dispatch.
type IAllotmentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsAllotmentContext differentiates from other interfaces.
	IsAllotmentContext()
}

type AllotmentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAllotmentContext() *AllotmentContext {
	var p = new(AllotmentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_allotment
	return p
}

func InitEmptyAllotmentContext(p *AllotmentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_allotment
}

func (*AllotmentContext) IsAllotmentContext() {}

func NewAllotmentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AllotmentContext {
	var p = new(AllotmentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_allotment

	return p
}

func (s *AllotmentContext) GetParser() antlr.Parser { return s.parser }

func (s *AllotmentContext) CopyAll(ctx *AllotmentContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *AllotmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AllotmentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type RemainingAllotmentContext struct {
	AllotmentContext
}

func NewRemainingAllotmentContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *RemainingAllotmentContext {
	var p = new(RemainingAllotmentContext)

	InitEmptyAllotmentContext(&p.AllotmentContext)
	p.parser = parser
	p.CopyAll(ctx.(*AllotmentContext))

	return p
}

func (s *RemainingAllotmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RemainingAllotmentContext) REMAINING() antlr.TerminalNode {
	return s.GetToken(NumscriptParserREMAINING, 0)
}

func (s *RemainingAllotmentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterRemainingAllotment(s)
	}
}

func (s *RemainingAllotmentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitRemainingAllotment(s)
	}
}

type PortionedAllotmentContext struct {
	AllotmentContext
}

func NewPortionedAllotmentContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PortionedAllotmentContext {
	var p = new(PortionedAllotmentContext)

	InitEmptyAllotmentContext(&p.AllotmentContext)
	p.parser = parser
	p.CopyAll(ctx.(*AllotmentContext))

	return p
}

func (s *PortionedAllotmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PortionedAllotmentContext) Portion() IPortionContext {
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

func (s *PortionedAllotmentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterPortionedAllotment(s)
	}
}

func (s *PortionedAllotmentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitPortionedAllotment(s)
	}
}

type PortionVariableContext struct {
	AllotmentContext
}

func NewPortionVariableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PortionVariableContext {
	var p = new(PortionVariableContext)

	InitEmptyAllotmentContext(&p.AllotmentContext)
	p.parser = parser
	p.CopyAll(ctx.(*AllotmentContext))

	return p
}

func (s *PortionVariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PortionVariableContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARIABLE_NAME, 0)
}

func (s *PortionVariableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterPortionVariable(s)
	}
}

func (s *PortionVariableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitPortionVariable(s)
	}
}

func (p *NumscriptParser) Allotment() (localctx IAllotmentContext) {
	localctx = NewAllotmentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, NumscriptParserRULE_allotment)
	p.SetState(139)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserRATIO_PORTION_LITERAL, NumscriptParserPERCENTAGE_PORTION_LITERAL:
		localctx = NewPortionedAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(136)
			p.Portion()
		}

	case NumscriptParserVARIABLE_NAME:
		localctx = NewPortionVariableContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(137)
			p.Match(NumscriptParserVARIABLE_NAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserREMAINING:
		localctx = NewRemainingAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(138)
			p.Match(NumscriptParserREMAINING)
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

type SrcAccountBoundedOverdraftContext struct {
	SourceContext
	address     IValueExprContext
	maxOvedraft IValueExprContext
}

func NewSrcAccountBoundedOverdraftContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcAccountBoundedOverdraftContext {
	var p = new(SrcAccountBoundedOverdraftContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcAccountBoundedOverdraftContext) GetAddress() IValueExprContext { return s.address }

func (s *SrcAccountBoundedOverdraftContext) GetMaxOvedraft() IValueExprContext { return s.maxOvedraft }

func (s *SrcAccountBoundedOverdraftContext) SetAddress(v IValueExprContext) { s.address = v }

func (s *SrcAccountBoundedOverdraftContext) SetMaxOvedraft(v IValueExprContext) { s.maxOvedraft = v }

func (s *SrcAccountBoundedOverdraftContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcAccountBoundedOverdraftContext) ALLOWING() antlr.TerminalNode {
	return s.GetToken(NumscriptParserALLOWING, 0)
}

func (s *SrcAccountBoundedOverdraftContext) OVERDRAFT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserOVERDRAFT, 0)
}

func (s *SrcAccountBoundedOverdraftContext) UP() antlr.TerminalNode {
	return s.GetToken(NumscriptParserUP, 0)
}

func (s *SrcAccountBoundedOverdraftContext) TO() antlr.TerminalNode {
	return s.GetToken(NumscriptParserTO, 0)
}

func (s *SrcAccountBoundedOverdraftContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *SrcAccountBoundedOverdraftContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
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

	return t.(IValueExprContext)
}

func (s *SrcAccountBoundedOverdraftContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSrcAccountBoundedOverdraft(s)
	}
}

func (s *SrcAccountBoundedOverdraftContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSrcAccountBoundedOverdraft(s)
	}
}

type SrcAccountUnboundedOverdraftContext struct {
	SourceContext
	address IValueExprContext
}

func NewSrcAccountUnboundedOverdraftContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcAccountUnboundedOverdraftContext {
	var p = new(SrcAccountUnboundedOverdraftContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcAccountUnboundedOverdraftContext) GetAddress() IValueExprContext { return s.address }

func (s *SrcAccountUnboundedOverdraftContext) SetAddress(v IValueExprContext) { s.address = v }

func (s *SrcAccountUnboundedOverdraftContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcAccountUnboundedOverdraftContext) ALLOWING() antlr.TerminalNode {
	return s.GetToken(NumscriptParserALLOWING, 0)
}

func (s *SrcAccountUnboundedOverdraftContext) UNBOUNDED() antlr.TerminalNode {
	return s.GetToken(NumscriptParserUNBOUNDED, 0)
}

func (s *SrcAccountUnboundedOverdraftContext) OVERDRAFT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserOVERDRAFT, 0)
}

func (s *SrcAccountUnboundedOverdraftContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *SrcAccountUnboundedOverdraftContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSrcAccountUnboundedOverdraft(s)
	}
}

func (s *SrcAccountUnboundedOverdraftContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSrcAccountUnboundedOverdraft(s)
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

type SrcInorderContext struct {
	SourceContext
}

func NewSrcInorderContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcInorderContext {
	var p = new(SrcInorderContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcInorderContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcInorderContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACE, 0)
}

func (s *SrcInorderContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACE, 0)
}

func (s *SrcInorderContext) AllSource() []ISourceContext {
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

func (s *SrcInorderContext) Source(i int) ISourceContext {
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

func (s *SrcInorderContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSrcInorder(s)
	}
}

func (s *SrcInorderContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSrcInorder(s)
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

func (s *SrcAccountContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
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

func (p *NumscriptParser) Source() (localctx ISourceContext) {
	localctx = NewSourceContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, NumscriptParserRULE_source)
	var _la int

	p.SetState(175)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 14, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSrcAccountUnboundedOverdraftContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(141)

			var _x = p.valueExpr(0)

			localctx.(*SrcAccountUnboundedOverdraftContext).address = _x
		}
		{
			p.SetState(142)
			p.Match(NumscriptParserALLOWING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(143)
			p.Match(NumscriptParserUNBOUNDED)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(144)
			p.Match(NumscriptParserOVERDRAFT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewSrcAccountBoundedOverdraftContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(146)

			var _x = p.valueExpr(0)

			localctx.(*SrcAccountBoundedOverdraftContext).address = _x
		}
		{
			p.SetState(147)
			p.Match(NumscriptParserALLOWING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(148)
			p.Match(NumscriptParserOVERDRAFT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(149)
			p.Match(NumscriptParserUP)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(150)
			p.Match(NumscriptParserTO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(151)

			var _x = p.valueExpr(0)

			localctx.(*SrcAccountBoundedOverdraftContext).maxOvedraft = _x
		}

	case 3:
		localctx = NewSrcAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(153)
			p.valueExpr(0)
		}

	case 4:
		localctx = NewSrcAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(154)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(156)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&38482911166464) != 0) {
			{
				p.SetState(155)
				p.AllotmentClauseSrc()
			}

			p.SetState(158)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(160)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		localctx = NewSrcInorderContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(162)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(166)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&271601920671744) != 0 {
			{
				p.SetState(163)
				p.Source()
			}

			p.SetState(168)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(169)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		localctx = NewSrcCappedContext(p, localctx)
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(170)
			p.Match(NumscriptParserMAX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(171)
			p.Cap_()
		}
		{
			p.SetState(172)
			p.Match(NumscriptParserFROM)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(173)
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
	Allotment() IAllotmentContext
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

func (s *AllotmentClauseSrcContext) Allotment() IAllotmentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAllotmentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAllotmentContext)
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

func (p *NumscriptParser) AllotmentClauseSrc() (localctx IAllotmentClauseSrcContext) {
	localctx = NewAllotmentClauseSrcContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, NumscriptParserRULE_allotmentClauseSrc)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(177)
		p.Allotment()
	}
	{
		p.SetState(178)
		p.Match(NumscriptParserFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(179)
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

// IKeptOrDestinationContext is an interface to support dynamic dispatch.
type IKeptOrDestinationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsKeptOrDestinationContext differentiates from other interfaces.
	IsKeptOrDestinationContext()
}

type KeptOrDestinationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyKeptOrDestinationContext() *KeptOrDestinationContext {
	var p = new(KeptOrDestinationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_keptOrDestination
	return p
}

func InitEmptyKeptOrDestinationContext(p *KeptOrDestinationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_keptOrDestination
}

func (*KeptOrDestinationContext) IsKeptOrDestinationContext() {}

func NewKeptOrDestinationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *KeptOrDestinationContext {
	var p = new(KeptOrDestinationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_keptOrDestination

	return p
}

func (s *KeptOrDestinationContext) GetParser() antlr.Parser { return s.parser }

func (s *KeptOrDestinationContext) CopyAll(ctx *KeptOrDestinationContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *KeptOrDestinationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *KeptOrDestinationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type DestinationKeptContext struct {
	KeptOrDestinationContext
}

func NewDestinationKeptContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestinationKeptContext {
	var p = new(DestinationKeptContext)

	InitEmptyKeptOrDestinationContext(&p.KeptOrDestinationContext)
	p.parser = parser
	p.CopyAll(ctx.(*KeptOrDestinationContext))

	return p
}

func (s *DestinationKeptContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestinationKeptContext) KEPT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserKEPT, 0)
}

func (s *DestinationKeptContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestinationKept(s)
	}
}

func (s *DestinationKeptContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestinationKept(s)
	}
}

type DestinationToContext struct {
	KeptOrDestinationContext
}

func NewDestinationToContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestinationToContext {
	var p = new(DestinationToContext)

	InitEmptyKeptOrDestinationContext(&p.KeptOrDestinationContext)
	p.parser = parser
	p.CopyAll(ctx.(*KeptOrDestinationContext))

	return p
}

func (s *DestinationToContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestinationToContext) TO() antlr.TerminalNode {
	return s.GetToken(NumscriptParserTO, 0)
}

func (s *DestinationToContext) Destination() IDestinationContext {
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

func (s *DestinationToContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestinationTo(s)
	}
}

func (s *DestinationToContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestinationTo(s)
	}
}

func (p *NumscriptParser) KeptOrDestination() (localctx IKeptOrDestinationContext) {
	localctx = NewKeptOrDestinationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, NumscriptParserRULE_keptOrDestination)
	p.SetState(184)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserTO:
		localctx = NewDestinationToContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(181)
			p.Match(NumscriptParserTO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(182)
			p.destination(0)
		}

	case NumscriptParserKEPT:
		localctx = NewDestinationKeptContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(183)
			p.Match(NumscriptParserKEPT)
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

// IDestinationInOrderClauseContext is an interface to support dynamic dispatch.
type IDestinationInOrderClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	MAX() antlr.TerminalNode
	ValueExpr() IValueExprContext
	KeptOrDestination() IKeptOrDestinationContext

	// IsDestinationInOrderClauseContext differentiates from other interfaces.
	IsDestinationInOrderClauseContext()
}

type DestinationInOrderClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDestinationInOrderClauseContext() *DestinationInOrderClauseContext {
	var p = new(DestinationInOrderClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_destinationInOrderClause
	return p
}

func InitEmptyDestinationInOrderClauseContext(p *DestinationInOrderClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_destinationInOrderClause
}

func (*DestinationInOrderClauseContext) IsDestinationInOrderClauseContext() {}

func NewDestinationInOrderClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DestinationInOrderClauseContext {
	var p = new(DestinationInOrderClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_destinationInOrderClause

	return p
}

func (s *DestinationInOrderClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *DestinationInOrderClauseContext) MAX() antlr.TerminalNode {
	return s.GetToken(NumscriptParserMAX, 0)
}

func (s *DestinationInOrderClauseContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *DestinationInOrderClauseContext) KeptOrDestination() IKeptOrDestinationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IKeptOrDestinationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IKeptOrDestinationContext)
}

func (s *DestinationInOrderClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestinationInOrderClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DestinationInOrderClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestinationInOrderClause(s)
	}
}

func (s *DestinationInOrderClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestinationInOrderClause(s)
	}
}

func (p *NumscriptParser) DestinationInOrderClause() (localctx IDestinationInOrderClauseContext) {
	localctx = NewDestinationInOrderClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, NumscriptParserRULE_destinationInOrderClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(186)
		p.Match(NumscriptParserMAX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(187)
		p.valueExpr(0)
	}
	{
		p.SetState(188)
		p.KeptOrDestination()
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

type DestInorderContext struct {
	DestinationContext
}

func NewDestInorderContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestInorderContext {
	var p = new(DestInorderContext)

	InitEmptyDestinationContext(&p.DestinationContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestinationContext))

	return p
}

func (s *DestInorderContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestInorderContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACE, 0)
}

func (s *DestInorderContext) REMAINING() antlr.TerminalNode {
	return s.GetToken(NumscriptParserREMAINING, 0)
}

func (s *DestInorderContext) KeptOrDestination() IKeptOrDestinationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IKeptOrDestinationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IKeptOrDestinationContext)
}

func (s *DestInorderContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACE, 0)
}

func (s *DestInorderContext) AllDestinationInOrderClause() []IDestinationInOrderClauseContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IDestinationInOrderClauseContext); ok {
			len++
		}
	}

	tst := make([]IDestinationInOrderClauseContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IDestinationInOrderClauseContext); ok {
			tst[i] = t.(IDestinationInOrderClauseContext)
			i++
		}
	}

	return tst
}

func (s *DestInorderContext) DestinationInOrderClause(i int) IDestinationInOrderClauseContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDestinationInOrderClauseContext); ok {
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

	return t.(IDestinationInOrderClauseContext)
}

func (s *DestInorderContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestInorder(s)
	}
}

func (s *DestInorderContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestInorder(s)
	}
}

type DestIfContext struct {
	DestinationContext
	ifBranch   IDestinationContext
	elseBranch IDestinationContext
}

func NewDestIfContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestIfContext {
	var p = new(DestIfContext)

	InitEmptyDestinationContext(&p.DestinationContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestinationContext))

	return p
}

func (s *DestIfContext) GetIfBranch() IDestinationContext { return s.ifBranch }

func (s *DestIfContext) GetElseBranch() IDestinationContext { return s.elseBranch }

func (s *DestIfContext) SetIfBranch(v IDestinationContext) { s.ifBranch = v }

func (s *DestIfContext) SetElseBranch(v IDestinationContext) { s.elseBranch = v }

func (s *DestIfContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestIfContext) IF() antlr.TerminalNode {
	return s.GetToken(NumscriptParserIF, 0)
}

func (s *DestIfContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *DestIfContext) ELSE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserELSE, 0)
}

func (s *DestIfContext) AllDestination() []IDestinationContext {
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

func (s *DestIfContext) Destination(i int) IDestinationContext {
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

func (s *DestIfContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestIf(s)
	}
}

func (s *DestIfContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestIf(s)
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

func (s *DestAccountContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
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

func (p *NumscriptParser) Destination() (localctx IDestinationContext) {
	return p.destination(0)
}

func (p *NumscriptParser) destination(_p int) (localctx IDestinationContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewDestinationContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IDestinationContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 32
	p.EnterRecursionRule(localctx, 32, NumscriptParserRULE_destination, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(211)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 18, p.GetParserRuleContext()) {
	case 1:
		localctx = NewDestAccountContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(191)
			p.valueExpr(0)
		}

	case 2:
		localctx = NewDestAllotmentContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(192)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(194)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&38482911166464) != 0) {
			{
				p.SetState(193)
				p.AllotmentClauseDest()
			}

			p.SetState(196)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(198)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewDestInorderContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(200)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(204)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == NumscriptParserMAX {
			{
				p.SetState(201)
				p.DestinationInOrderClause()
			}

			p.SetState(206)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(207)
			p.Match(NumscriptParserREMAINING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(208)
			p.KeptOrDestination()
		}
		{
			p.SetState(209)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(221)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 19, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			localctx = NewDestIfContext(p, NewDestinationContext(p, _parentctx, _parentState))
			localctx.(*DestIfContext).ifBranch = _prevctx

			p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_destination)
			p.SetState(213)

			if !(p.Precpred(p.GetParserRuleContext(), 3)) {
				p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
				goto errorExit
			}
			{
				p.SetState(214)
				p.Match(NumscriptParserIF)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(215)
				p.valueExpr(0)
			}
			{
				p.SetState(216)
				p.Match(NumscriptParserELSE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(217)

				var _x = p.destination(4)

				localctx.(*DestIfContext).elseBranch = _x
			}

		}
		p.SetState(223)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 19, p.GetParserRuleContext())
		if p.HasError() {
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
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAllotmentClauseDestContext is an interface to support dynamic dispatch.
type IAllotmentClauseDestContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Allotment() IAllotmentContext
	KeptOrDestination() IKeptOrDestinationContext

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

func (s *AllotmentClauseDestContext) Allotment() IAllotmentContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAllotmentContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAllotmentContext)
}

func (s *AllotmentClauseDestContext) KeptOrDestination() IKeptOrDestinationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IKeptOrDestinationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IKeptOrDestinationContext)
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

func (p *NumscriptParser) AllotmentClauseDest() (localctx IAllotmentClauseDestContext) {
	localctx = NewAllotmentClauseDestContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, NumscriptParserRULE_allotmentClauseDest)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(224)
		p.Allotment()
	}
	{
		p.SetState(225)
		p.KeptOrDestination()
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

// ISentValueContext is an interface to support dynamic dispatch.
type ISentValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsSentValueContext differentiates from other interfaces.
	IsSentValueContext()
}

type SentValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySentValueContext() *SentValueContext {
	var p = new(SentValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_sentValue
	return p
}

func InitEmptySentValueContext(p *SentValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_sentValue
}

func (*SentValueContext) IsSentValueContext() {}

func NewSentValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SentValueContext {
	var p = new(SentValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_sentValue

	return p
}

func (s *SentValueContext) GetParser() antlr.Parser { return s.parser }

func (s *SentValueContext) CopyAll(ctx *SentValueContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *SentValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SentValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type SentAllContext struct {
	SentValueContext
}

func NewSentAllContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SentAllContext {
	var p = new(SentAllContext)

	InitEmptySentValueContext(&p.SentValueContext)
	p.parser = parser
	p.CopyAll(ctx.(*SentValueContext))

	return p
}

func (s *SentAllContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SentAllContext) SentAllLit() ISentAllLitContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISentAllLitContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISentAllLitContext)
}

func (s *SentAllContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSentAll(s)
	}
}

func (s *SentAllContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSentAll(s)
	}
}

type SentLiteralContext struct {
	SentValueContext
}

func NewSentLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SentLiteralContext {
	var p = new(SentLiteralContext)

	InitEmptySentValueContext(&p.SentValueContext)
	p.parser = parser
	p.CopyAll(ctx.(*SentValueContext))

	return p
}

func (s *SentLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SentLiteralContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *SentLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSentLiteral(s)
	}
}

func (s *SentLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSentLiteral(s)
	}
}

func (p *NumscriptParser) SentValue() (localctx ISentValueContext) {
	localctx = NewSentValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, NumscriptParserRULE_sentValue)
	p.SetState(229)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 20, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSentLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(227)
			p.valueExpr(0)
		}

	case 2:
		localctx = NewSentAllContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(228)
			p.SentAllLit()
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

// IStatementContext is an interface to support dynamic dispatch.
type IStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
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

func (s *StatementContext) CopyAll(ctx *StatementContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *StatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type SendStatementContext struct {
	StatementContext
}

func NewSendStatementContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SendStatementContext {
	var p = new(SendStatementContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *SendStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SendStatementContext) SEND() antlr.TerminalNode {
	return s.GetToken(NumscriptParserSEND, 0)
}

func (s *SendStatementContext) SentValue() ISentValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISentValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISentValueContext)
}

func (s *SendStatementContext) LPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLPARENS, 0)
}

func (s *SendStatementContext) SOURCE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserSOURCE, 0)
}

func (s *SendStatementContext) AllEQ() []antlr.TerminalNode {
	return s.GetTokens(NumscriptParserEQ)
}

func (s *SendStatementContext) EQ(i int) antlr.TerminalNode {
	return s.GetToken(NumscriptParserEQ, i)
}

func (s *SendStatementContext) Source() ISourceContext {
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

func (s *SendStatementContext) DESTINATION() antlr.TerminalNode {
	return s.GetToken(NumscriptParserDESTINATION, 0)
}

func (s *SendStatementContext) Destination() IDestinationContext {
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

func (s *SendStatementContext) RPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRPARENS, 0)
}

func (s *SendStatementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSendStatement(s)
	}
}

func (s *SendStatementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSendStatement(s)
	}
}

type SaveStatementContext struct {
	StatementContext
}

func NewSaveStatementContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SaveStatementContext {
	var p = new(SaveStatementContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *SaveStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SaveStatementContext) SAVE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserSAVE, 0)
}

func (s *SaveStatementContext) SentValue() ISentValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISentValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISentValueContext)
}

func (s *SaveStatementContext) FROM() antlr.TerminalNode {
	return s.GetToken(NumscriptParserFROM, 0)
}

func (s *SaveStatementContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *SaveStatementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSaveStatement(s)
	}
}

func (s *SaveStatementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSaveStatement(s)
	}
}

type FnCallStatementContext struct {
	StatementContext
}

func NewFnCallStatementContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FnCallStatementContext {
	var p = new(FnCallStatementContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *FnCallStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FnCallStatementContext) FunctionCall() IFunctionCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionCallContext)
}

func (s *FnCallStatementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterFnCallStatement(s)
	}
}

func (s *FnCallStatementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitFnCallStatement(s)
	}
}

func (p *NumscriptParser) Statement() (localctx IStatementContext) {
	localctx = NewStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, NumscriptParserRULE_statement)
	p.SetState(248)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserSEND:
		localctx = NewSendStatementContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(231)
			p.Match(NumscriptParserSEND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(232)
			p.SentValue()
		}
		{
			p.SetState(233)
			p.Match(NumscriptParserLPARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(234)
			p.Match(NumscriptParserSOURCE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(235)
			p.Match(NumscriptParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(236)
			p.Source()
		}
		{
			p.SetState(237)
			p.Match(NumscriptParserDESTINATION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(238)
			p.Match(NumscriptParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(239)
			p.destination(0)
		}
		{
			p.SetState(240)
			p.Match(NumscriptParserRPARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserSAVE:
		localctx = NewSaveStatementContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(242)
			p.Match(NumscriptParserSAVE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(243)
			p.SentValue()
		}
		{
			p.SetState(244)
			p.Match(NumscriptParserFROM)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(245)
			p.valueExpr(0)
		}

	case NumscriptParserOVERDRAFT, NumscriptParserIDENTIFIER:
		localctx = NewFnCallStatementContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(247)
			p.FunctionCall()
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

func (p *NumscriptParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 2:
		var t *ValueExprContext = nil
		if localctx != nil {
			t = localctx.(*ValueExprContext)
		}
		return p.ValueExpr_Sempred(t, predIndex)

	case 16:
		var t *DestinationContext = nil
		if localctx != nil {
			t = localctx.(*DestinationContext)
		}
		return p.Destination_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *NumscriptParser) ValueExpr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 6)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 5)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 4)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 3)

	case 4:
		return p.Precpred(p.GetParserRuleContext(), 2)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *NumscriptParser) Destination_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 5:
		return p.Precpred(p.GetParserRuleContext(), 3)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
