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
		"", "'/'", "'+'", "'oneof'", "", "", "", "", "'vars'", "'max'", "'source'",
		"'destination'", "'send'", "'from'", "'up'", "'to'", "'remaining'",
		"'allowing'", "'unbounded'", "'overdraft'", "'kept'", "'save'", "'('",
		"')'", "'['", "']'", "'{'", "'}'", "','", "'='", "'*'", "'-'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "WS", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT",
		"VARS", "MAX", "SOURCE", "DESTINATION", "SEND", "FROM", "UP", "TO",
		"REMAINING", "ALLOWING", "UNBOUNDED", "OVERDRAFT", "KEPT", "SAVE", "LPARENS",
		"RPARENS", "LBRACKET", "RBRACKET", "LBRACE", "RBRACE", "COMMA", "EQ",
		"STAR", "MINUS", "PERCENTAGE_PORTION_LITERAL", "STRING", "IDENTIFIER",
		"NUMBER", "VARIABLE_NAME", "ACCOUNT", "ASSET",
	}
	staticData.RuleNames = []string{
		"monetaryLit", "valueExpr", "functionCallArgs", "functionCall", "varOrigin",
		"varDeclaration", "varsDeclaration", "program", "sentAllLit", "allotment",
		"source", "allotmentClauseSrc", "keptOrDestination", "destinationInOrderClause",
		"destination", "allotmentClauseDest", "sentValue", "statement",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 38, 238, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 54, 8, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 5, 1, 62, 8, 1, 10, 1, 12, 1, 65, 9,
		1, 1, 2, 1, 2, 1, 2, 5, 2, 70, 8, 2, 10, 2, 12, 2, 73, 9, 2, 1, 3, 1, 3,
		1, 3, 3, 3, 78, 8, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 3,
		5, 88, 8, 5, 1, 6, 1, 6, 1, 6, 5, 6, 93, 8, 6, 10, 6, 12, 6, 96, 9, 6,
		1, 6, 1, 6, 1, 7, 3, 7, 101, 8, 7, 1, 7, 5, 7, 104, 8, 7, 10, 7, 12, 7,
		107, 9, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 3, 9,
		118, 8, 9, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10,
		1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 4, 10, 135, 8, 10, 11, 10, 12,
		10, 136, 1, 10, 1, 10, 1, 10, 1, 10, 5, 10, 143, 8, 10, 10, 10, 12, 10,
		146, 9, 10, 1, 10, 1, 10, 1, 10, 1, 10, 4, 10, 152, 8, 10, 11, 10, 12,
		10, 153, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 3, 10, 163, 8,
		10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 3, 12, 172, 8, 12,
		1, 13, 1, 13, 1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 4, 14, 181, 8, 14, 11,
		14, 12, 14, 182, 1, 14, 1, 14, 1, 14, 1, 14, 5, 14, 189, 8, 14, 10, 14,
		12, 14, 192, 9, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 5,
		14, 201, 8, 14, 10, 14, 12, 14, 204, 9, 14, 1, 14, 1, 14, 1, 14, 1, 14,
		3, 14, 210, 8, 14, 1, 15, 1, 15, 1, 15, 1, 16, 1, 16, 3, 16, 217, 8, 16,
		1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1,
		17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 3, 17, 236, 8, 17, 1, 17,
		0, 1, 2, 18, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30,
		32, 34, 0, 2, 2, 0, 2, 2, 31, 31, 2, 0, 19, 19, 34, 34, 254, 0, 36, 1,
		0, 0, 0, 2, 53, 1, 0, 0, 0, 4, 66, 1, 0, 0, 0, 6, 74, 1, 0, 0, 0, 8, 81,
		1, 0, 0, 0, 10, 84, 1, 0, 0, 0, 12, 89, 1, 0, 0, 0, 14, 100, 1, 0, 0, 0,
		16, 110, 1, 0, 0, 0, 18, 117, 1, 0, 0, 0, 20, 162, 1, 0, 0, 0, 22, 164,
		1, 0, 0, 0, 24, 171, 1, 0, 0, 0, 26, 173, 1, 0, 0, 0, 28, 209, 1, 0, 0,
		0, 30, 211, 1, 0, 0, 0, 32, 216, 1, 0, 0, 0, 34, 235, 1, 0, 0, 0, 36, 37,
		5, 24, 0, 0, 37, 38, 3, 2, 1, 0, 38, 39, 3, 2, 1, 0, 39, 40, 5, 25, 0,
		0, 40, 1, 1, 0, 0, 0, 41, 42, 6, 1, -1, 0, 42, 54, 5, 36, 0, 0, 43, 54,
		5, 38, 0, 0, 44, 54, 5, 33, 0, 0, 45, 54, 5, 37, 0, 0, 46, 54, 5, 35, 0,
		0, 47, 54, 5, 32, 0, 0, 48, 54, 3, 0, 0, 0, 49, 50, 5, 22, 0, 0, 50, 51,
		3, 2, 1, 0, 51, 52, 5, 23, 0, 0, 52, 54, 1, 0, 0, 0, 53, 41, 1, 0, 0, 0,
		53, 43, 1, 0, 0, 0, 53, 44, 1, 0, 0, 0, 53, 45, 1, 0, 0, 0, 53, 46, 1,
		0, 0, 0, 53, 47, 1, 0, 0, 0, 53, 48, 1, 0, 0, 0, 53, 49, 1, 0, 0, 0, 54,
		63, 1, 0, 0, 0, 55, 56, 10, 3, 0, 0, 56, 57, 5, 1, 0, 0, 57, 62, 3, 2,
		1, 4, 58, 59, 10, 2, 0, 0, 59, 60, 7, 0, 0, 0, 60, 62, 3, 2, 1, 3, 61,
		55, 1, 0, 0, 0, 61, 58, 1, 0, 0, 0, 62, 65, 1, 0, 0, 0, 63, 61, 1, 0, 0,
		0, 63, 64, 1, 0, 0, 0, 64, 3, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 66, 71, 3,
		2, 1, 0, 67, 68, 5, 28, 0, 0, 68, 70, 3, 2, 1, 0, 69, 67, 1, 0, 0, 0, 70,
		73, 1, 0, 0, 0, 71, 69, 1, 0, 0, 0, 71, 72, 1, 0, 0, 0, 72, 5, 1, 0, 0,
		0, 73, 71, 1, 0, 0, 0, 74, 75, 7, 1, 0, 0, 75, 77, 5, 22, 0, 0, 76, 78,
		3, 4, 2, 0, 77, 76, 1, 0, 0, 0, 77, 78, 1, 0, 0, 0, 78, 79, 1, 0, 0, 0,
		79, 80, 5, 23, 0, 0, 80, 7, 1, 0, 0, 0, 81, 82, 5, 29, 0, 0, 82, 83, 3,
		6, 3, 0, 83, 9, 1, 0, 0, 0, 84, 85, 5, 34, 0, 0, 85, 87, 5, 36, 0, 0, 86,
		88, 3, 8, 4, 0, 87, 86, 1, 0, 0, 0, 87, 88, 1, 0, 0, 0, 88, 11, 1, 0, 0,
		0, 89, 90, 5, 8, 0, 0, 90, 94, 5, 26, 0, 0, 91, 93, 3, 10, 5, 0, 92, 91,
		1, 0, 0, 0, 93, 96, 1, 0, 0, 0, 94, 92, 1, 0, 0, 0, 94, 95, 1, 0, 0, 0,
		95, 97, 1, 0, 0, 0, 96, 94, 1, 0, 0, 0, 97, 98, 5, 27, 0, 0, 98, 13, 1,
		0, 0, 0, 99, 101, 3, 12, 6, 0, 100, 99, 1, 0, 0, 0, 100, 101, 1, 0, 0,
		0, 101, 105, 1, 0, 0, 0, 102, 104, 3, 34, 17, 0, 103, 102, 1, 0, 0, 0,
		104, 107, 1, 0, 0, 0, 105, 103, 1, 0, 0, 0, 105, 106, 1, 0, 0, 0, 106,
		108, 1, 0, 0, 0, 107, 105, 1, 0, 0, 0, 108, 109, 5, 0, 0, 1, 109, 15, 1,
		0, 0, 0, 110, 111, 5, 24, 0, 0, 111, 112, 3, 2, 1, 0, 112, 113, 5, 30,
		0, 0, 113, 114, 5, 25, 0, 0, 114, 17, 1, 0, 0, 0, 115, 118, 3, 2, 1, 0,
		116, 118, 5, 16, 0, 0, 117, 115, 1, 0, 0, 0, 117, 116, 1, 0, 0, 0, 118,
		19, 1, 0, 0, 0, 119, 120, 3, 2, 1, 0, 120, 121, 5, 17, 0, 0, 121, 122,
		5, 18, 0, 0, 122, 123, 5, 19, 0, 0, 123, 163, 1, 0, 0, 0, 124, 125, 3,
		2, 1, 0, 125, 126, 5, 17, 0, 0, 126, 127, 5, 19, 0, 0, 127, 128, 5, 14,
		0, 0, 128, 129, 5, 15, 0, 0, 129, 130, 3, 2, 1, 0, 130, 163, 1, 0, 0, 0,
		131, 163, 3, 2, 1, 0, 132, 134, 5, 26, 0, 0, 133, 135, 3, 22, 11, 0, 134,
		133, 1, 0, 0, 0, 135, 136, 1, 0, 0, 0, 136, 134, 1, 0, 0, 0, 136, 137,
		1, 0, 0, 0, 137, 138, 1, 0, 0, 0, 138, 139, 5, 27, 0, 0, 139, 163, 1, 0,
		0, 0, 140, 144, 5, 26, 0, 0, 141, 143, 3, 20, 10, 0, 142, 141, 1, 0, 0,
		0, 143, 146, 1, 0, 0, 0, 144, 142, 1, 0, 0, 0, 144, 145, 1, 0, 0, 0, 145,
		147, 1, 0, 0, 0, 146, 144, 1, 0, 0, 0, 147, 163, 5, 27, 0, 0, 148, 149,
		5, 3, 0, 0, 149, 151, 5, 26, 0, 0, 150, 152, 3, 20, 10, 0, 151, 150, 1,
		0, 0, 0, 152, 153, 1, 0, 0, 0, 153, 151, 1, 0, 0, 0, 153, 154, 1, 0, 0,
		0, 154, 155, 1, 0, 0, 0, 155, 156, 5, 27, 0, 0, 156, 163, 1, 0, 0, 0, 157,
		158, 5, 9, 0, 0, 158, 159, 3, 2, 1, 0, 159, 160, 5, 13, 0, 0, 160, 161,
		3, 20, 10, 0, 161, 163, 1, 0, 0, 0, 162, 119, 1, 0, 0, 0, 162, 124, 1,
		0, 0, 0, 162, 131, 1, 0, 0, 0, 162, 132, 1, 0, 0, 0, 162, 140, 1, 0, 0,
		0, 162, 148, 1, 0, 0, 0, 162, 157, 1, 0, 0, 0, 163, 21, 1, 0, 0, 0, 164,
		165, 3, 18, 9, 0, 165, 166, 5, 13, 0, 0, 166, 167, 3, 20, 10, 0, 167, 23,
		1, 0, 0, 0, 168, 169, 5, 15, 0, 0, 169, 172, 3, 28, 14, 0, 170, 172, 5,
		20, 0, 0, 171, 168, 1, 0, 0, 0, 171, 170, 1, 0, 0, 0, 172, 25, 1, 0, 0,
		0, 173, 174, 5, 9, 0, 0, 174, 175, 3, 2, 1, 0, 175, 176, 3, 24, 12, 0,
		176, 27, 1, 0, 0, 0, 177, 210, 3, 2, 1, 0, 178, 180, 5, 26, 0, 0, 179,
		181, 3, 30, 15, 0, 180, 179, 1, 0, 0, 0, 181, 182, 1, 0, 0, 0, 182, 180,
		1, 0, 0, 0, 182, 183, 1, 0, 0, 0, 183, 184, 1, 0, 0, 0, 184, 185, 5, 27,
		0, 0, 185, 210, 1, 0, 0, 0, 186, 190, 5, 26, 0, 0, 187, 189, 3, 26, 13,
		0, 188, 187, 1, 0, 0, 0, 189, 192, 1, 0, 0, 0, 190, 188, 1, 0, 0, 0, 190,
		191, 1, 0, 0, 0, 191, 193, 1, 0, 0, 0, 192, 190, 1, 0, 0, 0, 193, 194,
		5, 16, 0, 0, 194, 195, 3, 24, 12, 0, 195, 196, 5, 27, 0, 0, 196, 210, 1,
		0, 0, 0, 197, 198, 5, 3, 0, 0, 198, 202, 5, 26, 0, 0, 199, 201, 3, 26,
		13, 0, 200, 199, 1, 0, 0, 0, 201, 204, 1, 0, 0, 0, 202, 200, 1, 0, 0, 0,
		202, 203, 1, 0, 0, 0, 203, 205, 1, 0, 0, 0, 204, 202, 1, 0, 0, 0, 205,
		206, 5, 16, 0, 0, 206, 207, 3, 24, 12, 0, 207, 208, 5, 27, 0, 0, 208, 210,
		1, 0, 0, 0, 209, 177, 1, 0, 0, 0, 209, 178, 1, 0, 0, 0, 209, 186, 1, 0,
		0, 0, 209, 197, 1, 0, 0, 0, 210, 29, 1, 0, 0, 0, 211, 212, 3, 18, 9, 0,
		212, 213, 3, 24, 12, 0, 213, 31, 1, 0, 0, 0, 214, 217, 3, 2, 1, 0, 215,
		217, 3, 16, 8, 0, 216, 214, 1, 0, 0, 0, 216, 215, 1, 0, 0, 0, 217, 33,
		1, 0, 0, 0, 218, 219, 5, 12, 0, 0, 219, 220, 3, 32, 16, 0, 220, 221, 5,
		22, 0, 0, 221, 222, 5, 10, 0, 0, 222, 223, 5, 29, 0, 0, 223, 224, 3, 20,
		10, 0, 224, 225, 5, 11, 0, 0, 225, 226, 5, 29, 0, 0, 226, 227, 3, 28, 14,
		0, 227, 228, 5, 23, 0, 0, 228, 236, 1, 0, 0, 0, 229, 230, 5, 21, 0, 0,
		230, 231, 3, 32, 16, 0, 231, 232, 5, 13, 0, 0, 232, 233, 3, 2, 1, 0, 233,
		236, 1, 0, 0, 0, 234, 236, 3, 6, 3, 0, 235, 218, 1, 0, 0, 0, 235, 229,
		1, 0, 0, 0, 235, 234, 1, 0, 0, 0, 236, 35, 1, 0, 0, 0, 21, 53, 61, 63,
		71, 77, 87, 94, 100, 105, 117, 136, 144, 153, 162, 171, 182, 190, 202,
		209, 216, 235,
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
	NumscriptParserWS                         = 4
	NumscriptParserNEWLINE                    = 5
	NumscriptParserMULTILINE_COMMENT          = 6
	NumscriptParserLINE_COMMENT               = 7
	NumscriptParserVARS                       = 8
	NumscriptParserMAX                        = 9
	NumscriptParserSOURCE                     = 10
	NumscriptParserDESTINATION                = 11
	NumscriptParserSEND                       = 12
	NumscriptParserFROM                       = 13
	NumscriptParserUP                         = 14
	NumscriptParserTO                         = 15
	NumscriptParserREMAINING                  = 16
	NumscriptParserALLOWING                   = 17
	NumscriptParserUNBOUNDED                  = 18
	NumscriptParserOVERDRAFT                  = 19
	NumscriptParserKEPT                       = 20
	NumscriptParserSAVE                       = 21
	NumscriptParserLPARENS                    = 22
	NumscriptParserRPARENS                    = 23
	NumscriptParserLBRACKET                   = 24
	NumscriptParserRBRACKET                   = 25
	NumscriptParserLBRACE                     = 26
	NumscriptParserRBRACE                     = 27
	NumscriptParserCOMMA                      = 28
	NumscriptParserEQ                         = 29
	NumscriptParserSTAR                       = 30
	NumscriptParserMINUS                      = 31
	NumscriptParserPERCENTAGE_PORTION_LITERAL = 32
	NumscriptParserSTRING                     = 33
	NumscriptParserIDENTIFIER                 = 34
	NumscriptParserNUMBER                     = 35
	NumscriptParserVARIABLE_NAME              = 36
	NumscriptParserACCOUNT                    = 37
	NumscriptParserASSET                      = 38
)

// NumscriptParser rules.
const (
	NumscriptParserRULE_monetaryLit              = 0
	NumscriptParserRULE_valueExpr                = 1
	NumscriptParserRULE_functionCallArgs         = 2
	NumscriptParserRULE_functionCall             = 3
	NumscriptParserRULE_varOrigin                = 4
	NumscriptParserRULE_varDeclaration           = 5
	NumscriptParserRULE_varsDeclaration          = 6
	NumscriptParserRULE_program                  = 7
	NumscriptParserRULE_sentAllLit               = 8
	NumscriptParserRULE_allotment                = 9
	NumscriptParserRULE_source                   = 10
	NumscriptParserRULE_allotmentClauseSrc       = 11
	NumscriptParserRULE_keptOrDestination        = 12
	NumscriptParserRULE_destinationInOrderClause = 13
	NumscriptParserRULE_destination              = 14
	NumscriptParserRULE_allotmentClauseDest      = 15
	NumscriptParserRULE_sentValue                = 16
	NumscriptParserRULE_statement                = 17
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
		p.SetState(36)
		p.Match(NumscriptParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(37)

		var _x = p.valueExpr(0)

		localctx.(*MonetaryLitContext).asset = _x
	}

	{
		p.SetState(38)

		var _x = p.valueExpr(0)

		localctx.(*MonetaryLitContext).amt = _x
	}

	{
		p.SetState(39)
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

type InfixExprContext struct {
	ValueExprContext
	left  IValueExprContext
	op    antlr.Token
	right IValueExprContext
}

func NewInfixExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InfixExprContext {
	var p = new(InfixExprContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *InfixExprContext) GetOp() antlr.Token { return s.op }

func (s *InfixExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *InfixExprContext) GetLeft() IValueExprContext { return s.left }

func (s *InfixExprContext) GetRight() IValueExprContext { return s.right }

func (s *InfixExprContext) SetLeft(v IValueExprContext) { s.left = v }

func (s *InfixExprContext) SetRight(v IValueExprContext) { s.right = v }

func (s *InfixExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InfixExprContext) AllValueExpr() []IValueExprContext {
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

func (s *InfixExprContext) ValueExpr(i int) IValueExprContext {
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

func (s *InfixExprContext) MINUS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserMINUS, 0)
}

func (s *InfixExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterInfixExpr(s)
	}
}

func (s *InfixExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitInfixExpr(s)
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

type ParenthesizedExprContext struct {
	ValueExprContext
}

func NewParenthesizedExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ParenthesizedExprContext {
	var p = new(ParenthesizedExprContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *ParenthesizedExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParenthesizedExprContext) LPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLPARENS, 0)
}

func (s *ParenthesizedExprContext) ValueExpr() IValueExprContext {
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

func (s *ParenthesizedExprContext) RPARENS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRPARENS, 0)
}

func (s *ParenthesizedExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterParenthesizedExpr(s)
	}
}

func (s *ParenthesizedExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitParenthesizedExpr(s)
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

type PercentagePortionLiteralContext struct {
	ValueExprContext
}

func NewPercentagePortionLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PercentagePortionLiteralContext {
	var p = new(PercentagePortionLiteralContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *PercentagePortionLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PercentagePortionLiteralContext) PERCENTAGE_PORTION_LITERAL() antlr.TerminalNode {
	return s.GetToken(NumscriptParserPERCENTAGE_PORTION_LITERAL, 0)
}

func (s *PercentagePortionLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterPercentagePortionLiteral(s)
	}
}

func (s *PercentagePortionLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitPercentagePortionLiteral(s)
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
	_startState := 2
	p.EnterRecursionRule(localctx, 2, NumscriptParserRULE_valueExpr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(53)
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
			p.SetState(42)
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
			p.SetState(43)
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
			p.SetState(44)
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
			p.SetState(45)
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
			p.SetState(46)
			p.Match(NumscriptParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserPERCENTAGE_PORTION_LITERAL:
		localctx = NewPercentagePortionLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(47)
			p.Match(NumscriptParserPERCENTAGE_PORTION_LITERAL)
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
			p.SetState(48)
			p.MonetaryLit()
		}

	case NumscriptParserLPARENS:
		localctx = NewParenthesizedExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(49)
			p.Match(NumscriptParserLPARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(50)
			p.valueExpr(0)
		}
		{
			p.SetState(51)
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
	p.SetState(63)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(61)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext()) {
			case 1:
				localctx = NewInfixExprContext(p, NewValueExprContext(p, _parentctx, _parentState))
				localctx.(*InfixExprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_valueExpr)
				p.SetState(55)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(56)

					var _m = p.Match(NumscriptParserT__0)

					localctx.(*InfixExprContext).op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(57)

					var _x = p.valueExpr(4)

					localctx.(*InfixExprContext).right = _x
				}

			case 2:
				localctx = NewInfixExprContext(p, NewValueExprContext(p, _parentctx, _parentState))
				localctx.(*InfixExprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_valueExpr)
				p.SetState(58)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(59)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*InfixExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == NumscriptParserT__1 || _la == NumscriptParserMINUS) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*InfixExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(60)

					var _x = p.valueExpr(3)

					localctx.(*InfixExprContext).right = _x
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(65)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
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
	p.EnterRule(localctx, 4, NumscriptParserRULE_functionCallArgs)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(66)
		p.valueExpr(0)
	}
	p.SetState(71)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NumscriptParserCOMMA {
		{
			p.SetState(67)
			p.Match(NumscriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(68)
			p.valueExpr(0)
		}

		p.SetState(73)
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
	p.EnterRule(localctx, 6, NumscriptParserRULE_functionCall)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(74)

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
		p.SetState(75)
		p.Match(NumscriptParserLPARENS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(77)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&528301948928) != 0 {
		{
			p.SetState(76)
			p.FunctionCallArgs()
		}

	}
	{
		p.SetState(79)
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
	p.EnterRule(localctx, 8, NumscriptParserRULE_varOrigin)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(81)
		p.Match(NumscriptParserEQ)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(82)
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
	p.EnterRule(localctx, 10, NumscriptParserRULE_varDeclaration)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(84)

		var _m = p.Match(NumscriptParserIDENTIFIER)

		localctx.(*VarDeclarationContext).type_ = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(85)

		var _m = p.Match(NumscriptParserVARIABLE_NAME)

		localctx.(*VarDeclarationContext).name = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(87)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NumscriptParserEQ {
		{
			p.SetState(86)
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
	p.EnterRule(localctx, 12, NumscriptParserRULE_varsDeclaration)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(89)
		p.Match(NumscriptParserVARS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(90)
		p.Match(NumscriptParserLBRACE)
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

	for _la == NumscriptParserIDENTIFIER {
		{
			p.SetState(91)
			p.VarDeclaration()
		}

		p.SetState(96)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(97)
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
	p.EnterRule(localctx, 14, NumscriptParserRULE_program)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(100)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NumscriptParserVARS {
		{
			p.SetState(99)
			p.VarsDeclaration()
		}

	}
	p.SetState(105)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&17182494720) != 0 {
		{
			p.SetState(102)
			p.Statement()
		}

		p.SetState(107)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(108)
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
	p.EnterRule(localctx, 16, NumscriptParserRULE_sentAllLit)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(110)
		p.Match(NumscriptParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(111)

		var _x = p.valueExpr(0)

		localctx.(*SentAllLitContext).asset = _x
	}

	{
		p.SetState(112)
		p.Match(NumscriptParserSTAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(113)
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

func (s *PortionedAllotmentContext) ValueExpr() IValueExprContext {
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

func (p *NumscriptParser) Allotment() (localctx IAllotmentContext) {
	localctx = NewAllotmentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, NumscriptParserRULE_allotment)
	p.SetState(117)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserLPARENS, NumscriptParserLBRACKET, NumscriptParserPERCENTAGE_PORTION_LITERAL, NumscriptParserSTRING, NumscriptParserNUMBER, NumscriptParserVARIABLE_NAME, NumscriptParserACCOUNT, NumscriptParserASSET:
		localctx = NewPortionedAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(115)
			p.valueExpr(0)
		}

	case NumscriptParserREMAINING:
		localctx = NewRemainingAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(116)
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

type SrcOneofContext struct {
	SourceContext
}

func NewSrcOneofContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcOneofContext {
	var p = new(SrcOneofContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcOneofContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcOneofContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACE, 0)
}

func (s *SrcOneofContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACE, 0)
}

func (s *SrcOneofContext) AllSource() []ISourceContext {
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

func (s *SrcOneofContext) Source(i int) ISourceContext {
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

func (s *SrcOneofContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterSrcOneof(s)
	}
}

func (s *SrcOneofContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitSrcOneof(s)
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
	cap_ IValueExprContext
}

func NewSrcCappedContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcCappedContext {
	var p = new(SrcCappedContext)

	InitEmptySourceContext(&p.SourceContext)
	p.parser = parser
	p.CopyAll(ctx.(*SourceContext))

	return p
}

func (s *SrcCappedContext) GetCap_() IValueExprContext { return s.cap_ }

func (s *SrcCappedContext) SetCap_(v IValueExprContext) { s.cap_ = v }

func (s *SrcCappedContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcCappedContext) MAX() antlr.TerminalNode {
	return s.GetToken(NumscriptParserMAX, 0)
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

func (s *SrcCappedContext) ValueExpr() IValueExprContext {
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
	p.EnterRule(localctx, 20, NumscriptParserRULE_source)
	var _la int

	p.SetState(162)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSrcAccountUnboundedOverdraftContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(119)

			var _x = p.valueExpr(0)

			localctx.(*SrcAccountUnboundedOverdraftContext).address = _x
		}
		{
			p.SetState(120)
			p.Match(NumscriptParserALLOWING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(121)
			p.Match(NumscriptParserUNBOUNDED)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(122)
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
			p.SetState(124)

			var _x = p.valueExpr(0)

			localctx.(*SrcAccountBoundedOverdraftContext).address = _x
		}
		{
			p.SetState(125)
			p.Match(NumscriptParserALLOWING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(126)
			p.Match(NumscriptParserOVERDRAFT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(127)
			p.Match(NumscriptParserUP)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(128)
			p.Match(NumscriptParserTO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(129)

			var _x = p.valueExpr(0)

			localctx.(*SrcAccountBoundedOverdraftContext).maxOvedraft = _x
		}

	case 3:
		localctx = NewSrcAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(131)
			p.valueExpr(0)
		}

	case 4:
		localctx = NewSrcAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(132)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(134)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&528302014464) != 0) {
			{
				p.SetState(133)
				p.AllotmentClauseSrc()
			}

			p.SetState(136)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(138)
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
			p.SetState(140)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(144)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&528369058312) != 0 {
			{
				p.SetState(141)
				p.Source()
			}

			p.SetState(146)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(147)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		localctx = NewSrcOneofContext(p, localctx)
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(148)
			p.Match(NumscriptParserT__2)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(149)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(151)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&528369058312) != 0) {
			{
				p.SetState(150)
				p.Source()
			}

			p.SetState(153)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(155)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		localctx = NewSrcCappedContext(p, localctx)
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(157)
			p.Match(NumscriptParserMAX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(158)

			var _x = p.valueExpr(0)

			localctx.(*SrcCappedContext).cap_ = _x
		}
		{
			p.SetState(159)
			p.Match(NumscriptParserFROM)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(160)
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
	p.EnterRule(localctx, 22, NumscriptParserRULE_allotmentClauseSrc)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(164)
		p.Allotment()
	}
	{
		p.SetState(165)
		p.Match(NumscriptParserFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(166)
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
	p.EnterRule(localctx, 24, NumscriptParserRULE_keptOrDestination)
	p.SetState(171)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserTO:
		localctx = NewDestinationToContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(168)
			p.Match(NumscriptParserTO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(169)
			p.Destination()
		}

	case NumscriptParserKEPT:
		localctx = NewDestinationKeptContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(170)
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
	p.EnterRule(localctx, 26, NumscriptParserRULE_destinationInOrderClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(173)
		p.Match(NumscriptParserMAX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(174)
		p.valueExpr(0)
	}
	{
		p.SetState(175)
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

type DestOneofContext struct {
	DestinationContext
}

func NewDestOneofContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestOneofContext {
	var p = new(DestOneofContext)

	InitEmptyDestinationContext(&p.DestinationContext)
	p.parser = parser
	p.CopyAll(ctx.(*DestinationContext))

	return p
}

func (s *DestOneofContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestOneofContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserLBRACE, 0)
}

func (s *DestOneofContext) REMAINING() antlr.TerminalNode {
	return s.GetToken(NumscriptParserREMAINING, 0)
}

func (s *DestOneofContext) KeptOrDestination() IKeptOrDestinationContext {
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

func (s *DestOneofContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NumscriptParserRBRACE, 0)
}

func (s *DestOneofContext) AllDestinationInOrderClause() []IDestinationInOrderClauseContext {
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

func (s *DestOneofContext) DestinationInOrderClause(i int) IDestinationInOrderClauseContext {
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

func (s *DestOneofContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterDestOneof(s)
	}
}

func (s *DestOneofContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitDestOneof(s)
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
	localctx = NewDestinationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, NumscriptParserRULE_destination)
	var _la int

	p.SetState(209)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 18, p.GetParserRuleContext()) {
	case 1:
		localctx = NewDestAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(177)
			p.valueExpr(0)
		}

	case 2:
		localctx = NewDestAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(178)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(180)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&528302014464) != 0) {
			{
				p.SetState(179)
				p.AllotmentClauseDest()
			}

			p.SetState(182)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(184)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewDestInorderContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(186)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(190)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == NumscriptParserMAX {
			{
				p.SetState(187)
				p.DestinationInOrderClause()
			}

			p.SetState(192)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(193)
			p.Match(NumscriptParserREMAINING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(194)
			p.KeptOrDestination()
		}
		{
			p.SetState(195)
			p.Match(NumscriptParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		localctx = NewDestOneofContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(197)
			p.Match(NumscriptParserT__2)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(198)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(202)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == NumscriptParserMAX {
			{
				p.SetState(199)
				p.DestinationInOrderClause()
			}

			p.SetState(204)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(205)
			p.Match(NumscriptParserREMAINING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(206)
			p.KeptOrDestination()
		}
		{
			p.SetState(207)
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
	p.EnterRule(localctx, 30, NumscriptParserRULE_allotmentClauseDest)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(211)
		p.Allotment()
	}
	{
		p.SetState(212)
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
	p.EnterRule(localctx, 32, NumscriptParserRULE_sentValue)
	p.SetState(216)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 19, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSentLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(214)
			p.valueExpr(0)
		}

	case 2:
		localctx = NewSentAllContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(215)
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
	p.EnterRule(localctx, 34, NumscriptParserRULE_statement)
	p.SetState(235)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserSEND:
		localctx = NewSendStatementContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(218)
			p.Match(NumscriptParserSEND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(219)
			p.SentValue()
		}
		{
			p.SetState(220)
			p.Match(NumscriptParserLPARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(221)
			p.Match(NumscriptParserSOURCE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(222)
			p.Match(NumscriptParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(223)
			p.Source()
		}
		{
			p.SetState(224)
			p.Match(NumscriptParserDESTINATION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(225)
			p.Match(NumscriptParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(226)
			p.Destination()
		}
		{
			p.SetState(227)
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
			p.SetState(229)
			p.Match(NumscriptParserSAVE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(230)
			p.SentValue()
		}
		{
			p.SetState(231)
			p.Match(NumscriptParserFROM)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(232)
			p.valueExpr(0)
		}

	case NumscriptParserOVERDRAFT, NumscriptParserIDENTIFIER:
		localctx = NewFnCallStatementContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(234)
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
	case 1:
		var t *ValueExprContext = nil
		if localctx != nil {
			t = localctx.(*ValueExprContext)
		}
		return p.ValueExpr_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *NumscriptParser) ValueExpr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 3)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 2)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
