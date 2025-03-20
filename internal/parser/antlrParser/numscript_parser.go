// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package antlrParser // Numscript
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
		"'send'", "'from'", "'up'", "'to'", "'remaining'", "'allowing'", "'unbounded'",
		"'overdraft'", "'oneof'", "'kept'", "'save'", "'('", "')'", "'['", "']'",
		"'{'", "'}'", "','", "'='", "'*'", "'+'", "'-'", "'/'", "", "", "",
		"", "", "'@'", "':'",
	}
	staticData.SymbolicNames = []string{
		"", "WS", "NEWLINE", "MULTILINE_COMMENT", "LINE_COMMENT", "VARS", "MAX",
		"SOURCE", "DESTINATION", "SEND", "FROM", "UP", "TO", "REMAINING", "ALLOWING",
		"UNBOUNDED", "OVERDRAFT", "ONEOF", "KEPT", "SAVE", "LPARENS", "RPARENS",
		"LBRACKET", "RBRACKET", "LBRACE", "RBRACE", "COMMA", "EQ", "STAR", "PLUS",
		"MINUS", "DIV", "PERCENTAGE_PORTION_LITERAL", "STRING", "IDENTIFIER",
		"NUMBER", "ASSET", "ACCOUNT_START", "COLON", "ACCOUNT_TEXT", "VARIABLE_NAME_ACC",
		"VARIABLE_NAME",
	}
	staticData.RuleNames = []string{
		"monetaryLit", "accountLiteralPart", "valueExpr", "functionCallArgs",
		"functionCall", "varOrigin", "varDeclaration", "varsDeclaration", "program",
		"sentAllLit", "allotment", "source", "allotmentClauseSrc", "keptOrDestination",
		"destinationInOrderClause", "destination", "allotmentClauseDest", "sentValue",
		"statement",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 41, 253, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0,
		1, 1, 1, 1, 3, 1, 46, 8, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1,
		2, 5, 2, 56, 8, 2, 10, 2, 12, 2, 59, 9, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2,
		1, 2, 1, 2, 1, 2, 3, 2, 69, 8, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 5,
		2, 77, 8, 2, 10, 2, 12, 2, 80, 9, 2, 1, 3, 1, 3, 1, 3, 5, 3, 85, 8, 3,
		10, 3, 12, 3, 88, 9, 3, 1, 4, 1, 4, 1, 4, 3, 4, 93, 8, 4, 1, 4, 1, 4, 1,
		5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 3, 6, 103, 8, 6, 1, 7, 1, 7, 1, 7, 5,
		7, 108, 8, 7, 10, 7, 12, 7, 111, 9, 7, 1, 7, 1, 7, 1, 8, 3, 8, 116, 8,
		8, 1, 8, 5, 8, 119, 8, 8, 10, 8, 12, 8, 122, 9, 8, 1, 8, 1, 8, 1, 9, 1,
		9, 1, 9, 1, 9, 1, 9, 1, 10, 1, 10, 3, 10, 133, 8, 10, 1, 11, 1, 11, 1,
		11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11,
		1, 11, 1, 11, 4, 11, 150, 8, 11, 11, 11, 12, 11, 151, 1, 11, 1, 11, 1,
		11, 1, 11, 5, 11, 158, 8, 11, 10, 11, 12, 11, 161, 9, 11, 1, 11, 1, 11,
		1, 11, 1, 11, 4, 11, 167, 8, 11, 11, 11, 12, 11, 168, 1, 11, 1, 11, 1,
		11, 1, 11, 1, 11, 1, 11, 1, 11, 3, 11, 178, 8, 11, 1, 12, 1, 12, 1, 12,
		1, 12, 1, 13, 1, 13, 1, 13, 3, 13, 187, 8, 13, 1, 14, 1, 14, 1, 14, 1,
		14, 1, 15, 1, 15, 1, 15, 4, 15, 196, 8, 15, 11, 15, 12, 15, 197, 1, 15,
		1, 15, 1, 15, 1, 15, 5, 15, 204, 8, 15, 10, 15, 12, 15, 207, 9, 15, 1,
		15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 5, 15, 216, 8, 15, 10, 15,
		12, 15, 219, 9, 15, 1, 15, 1, 15, 1, 15, 1, 15, 3, 15, 225, 8, 15, 1, 16,
		1, 16, 1, 16, 1, 17, 1, 17, 3, 17, 232, 8, 17, 1, 18, 1, 18, 1, 18, 1,
		18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18,
		1, 18, 1, 18, 1, 18, 3, 18, 251, 8, 18, 1, 18, 0, 1, 4, 19, 0, 2, 4, 6,
		8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 0, 2, 1, 0,
		29, 30, 2, 0, 16, 16, 34, 34, 271, 0, 38, 1, 0, 0, 0, 2, 45, 1, 0, 0, 0,
		4, 68, 1, 0, 0, 0, 6, 81, 1, 0, 0, 0, 8, 89, 1, 0, 0, 0, 10, 96, 1, 0,
		0, 0, 12, 99, 1, 0, 0, 0, 14, 104, 1, 0, 0, 0, 16, 115, 1, 0, 0, 0, 18,
		125, 1, 0, 0, 0, 20, 132, 1, 0, 0, 0, 22, 177, 1, 0, 0, 0, 24, 179, 1,
		0, 0, 0, 26, 186, 1, 0, 0, 0, 28, 188, 1, 0, 0, 0, 30, 224, 1, 0, 0, 0,
		32, 226, 1, 0, 0, 0, 34, 231, 1, 0, 0, 0, 36, 250, 1, 0, 0, 0, 38, 39,
		5, 22, 0, 0, 39, 40, 3, 4, 2, 0, 40, 41, 3, 4, 2, 0, 41, 42, 5, 23, 0,
		0, 42, 1, 1, 0, 0, 0, 43, 46, 5, 39, 0, 0, 44, 46, 5, 40, 0, 0, 45, 43,
		1, 0, 0, 0, 45, 44, 1, 0, 0, 0, 46, 3, 1, 0, 0, 0, 47, 48, 6, 2, -1, 0,
		48, 69, 5, 41, 0, 0, 49, 69, 5, 36, 0, 0, 50, 69, 5, 33, 0, 0, 51, 52,
		5, 37, 0, 0, 52, 57, 3, 2, 1, 0, 53, 54, 5, 38, 0, 0, 54, 56, 3, 2, 1,
		0, 55, 53, 1, 0, 0, 0, 56, 59, 1, 0, 0, 0, 57, 55, 1, 0, 0, 0, 57, 58,
		1, 0, 0, 0, 58, 69, 1, 0, 0, 0, 59, 57, 1, 0, 0, 0, 60, 69, 5, 35, 0, 0,
		61, 69, 5, 32, 0, 0, 62, 69, 3, 0, 0, 0, 63, 64, 5, 20, 0, 0, 64, 65, 3,
		4, 2, 0, 65, 66, 5, 21, 0, 0, 66, 69, 1, 0, 0, 0, 67, 69, 3, 8, 4, 0, 68,
		47, 1, 0, 0, 0, 68, 49, 1, 0, 0, 0, 68, 50, 1, 0, 0, 0, 68, 51, 1, 0, 0,
		0, 68, 60, 1, 0, 0, 0, 68, 61, 1, 0, 0, 0, 68, 62, 1, 0, 0, 0, 68, 63,
		1, 0, 0, 0, 68, 67, 1, 0, 0, 0, 69, 78, 1, 0, 0, 0, 70, 71, 10, 4, 0, 0,
		71, 72, 5, 31, 0, 0, 72, 77, 3, 4, 2, 5, 73, 74, 10, 3, 0, 0, 74, 75, 7,
		0, 0, 0, 75, 77, 3, 4, 2, 4, 76, 70, 1, 0, 0, 0, 76, 73, 1, 0, 0, 0, 77,
		80, 1, 0, 0, 0, 78, 76, 1, 0, 0, 0, 78, 79, 1, 0, 0, 0, 79, 5, 1, 0, 0,
		0, 80, 78, 1, 0, 0, 0, 81, 86, 3, 4, 2, 0, 82, 83, 5, 26, 0, 0, 83, 85,
		3, 4, 2, 0, 84, 82, 1, 0, 0, 0, 85, 88, 1, 0, 0, 0, 86, 84, 1, 0, 0, 0,
		86, 87, 1, 0, 0, 0, 87, 7, 1, 0, 0, 0, 88, 86, 1, 0, 0, 0, 89, 90, 7, 1,
		0, 0, 90, 92, 5, 20, 0, 0, 91, 93, 3, 6, 3, 0, 92, 91, 1, 0, 0, 0, 92,
		93, 1, 0, 0, 0, 93, 94, 1, 0, 0, 0, 94, 95, 5, 21, 0, 0, 95, 9, 1, 0, 0,
		0, 96, 97, 5, 27, 0, 0, 97, 98, 3, 4, 2, 0, 98, 11, 1, 0, 0, 0, 99, 100,
		5, 34, 0, 0, 100, 102, 5, 41, 0, 0, 101, 103, 3, 10, 5, 0, 102, 101, 1,
		0, 0, 0, 102, 103, 1, 0, 0, 0, 103, 13, 1, 0, 0, 0, 104, 105, 5, 5, 0,
		0, 105, 109, 5, 24, 0, 0, 106, 108, 3, 12, 6, 0, 107, 106, 1, 0, 0, 0,
		108, 111, 1, 0, 0, 0, 109, 107, 1, 0, 0, 0, 109, 110, 1, 0, 0, 0, 110,
		112, 1, 0, 0, 0, 111, 109, 1, 0, 0, 0, 112, 113, 5, 25, 0, 0, 113, 15,
		1, 0, 0, 0, 114, 116, 3, 14, 7, 0, 115, 114, 1, 0, 0, 0, 115, 116, 1, 0,
		0, 0, 116, 120, 1, 0, 0, 0, 117, 119, 3, 36, 18, 0, 118, 117, 1, 0, 0,
		0, 119, 122, 1, 0, 0, 0, 120, 118, 1, 0, 0, 0, 120, 121, 1, 0, 0, 0, 121,
		123, 1, 0, 0, 0, 122, 120, 1, 0, 0, 0, 123, 124, 5, 0, 0, 1, 124, 17, 1,
		0, 0, 0, 125, 126, 5, 22, 0, 0, 126, 127, 3, 4, 2, 0, 127, 128, 5, 28,
		0, 0, 128, 129, 5, 23, 0, 0, 129, 19, 1, 0, 0, 0, 130, 133, 3, 4, 2, 0,
		131, 133, 5, 13, 0, 0, 132, 130, 1, 0, 0, 0, 132, 131, 1, 0, 0, 0, 133,
		21, 1, 0, 0, 0, 134, 135, 3, 4, 2, 0, 135, 136, 5, 14, 0, 0, 136, 137,
		5, 15, 0, 0, 137, 138, 5, 16, 0, 0, 138, 178, 1, 0, 0, 0, 139, 140, 3,
		4, 2, 0, 140, 141, 5, 14, 0, 0, 141, 142, 5, 16, 0, 0, 142, 143, 5, 11,
		0, 0, 143, 144, 5, 12, 0, 0, 144, 145, 3, 4, 2, 0, 145, 178, 1, 0, 0, 0,
		146, 178, 3, 4, 2, 0, 147, 149, 5, 24, 0, 0, 148, 150, 3, 24, 12, 0, 149,
		148, 1, 0, 0, 0, 150, 151, 1, 0, 0, 0, 151, 149, 1, 0, 0, 0, 151, 152,
		1, 0, 0, 0, 152, 153, 1, 0, 0, 0, 153, 154, 5, 25, 0, 0, 154, 178, 1, 0,
		0, 0, 155, 159, 5, 24, 0, 0, 156, 158, 3, 22, 11, 0, 157, 156, 1, 0, 0,
		0, 158, 161, 1, 0, 0, 0, 159, 157, 1, 0, 0, 0, 159, 160, 1, 0, 0, 0, 160,
		162, 1, 0, 0, 0, 161, 159, 1, 0, 0, 0, 162, 178, 5, 25, 0, 0, 163, 164,
		5, 17, 0, 0, 164, 166, 5, 24, 0, 0, 165, 167, 3, 22, 11, 0, 166, 165, 1,
		0, 0, 0, 167, 168, 1, 0, 0, 0, 168, 166, 1, 0, 0, 0, 168, 169, 1, 0, 0,
		0, 169, 170, 1, 0, 0, 0, 170, 171, 5, 25, 0, 0, 171, 178, 1, 0, 0, 0, 172,
		173, 5, 6, 0, 0, 173, 174, 3, 4, 2, 0, 174, 175, 5, 10, 0, 0, 175, 176,
		3, 22, 11, 0, 176, 178, 1, 0, 0, 0, 177, 134, 1, 0, 0, 0, 177, 139, 1,
		0, 0, 0, 177, 146, 1, 0, 0, 0, 177, 147, 1, 0, 0, 0, 177, 155, 1, 0, 0,
		0, 177, 163, 1, 0, 0, 0, 177, 172, 1, 0, 0, 0, 178, 23, 1, 0, 0, 0, 179,
		180, 3, 20, 10, 0, 180, 181, 5, 10, 0, 0, 181, 182, 3, 22, 11, 0, 182,
		25, 1, 0, 0, 0, 183, 184, 5, 12, 0, 0, 184, 187, 3, 30, 15, 0, 185, 187,
		5, 18, 0, 0, 186, 183, 1, 0, 0, 0, 186, 185, 1, 0, 0, 0, 187, 27, 1, 0,
		0, 0, 188, 189, 5, 6, 0, 0, 189, 190, 3, 4, 2, 0, 190, 191, 3, 26, 13,
		0, 191, 29, 1, 0, 0, 0, 192, 225, 3, 4, 2, 0, 193, 195, 5, 24, 0, 0, 194,
		196, 3, 32, 16, 0, 195, 194, 1, 0, 0, 0, 196, 197, 1, 0, 0, 0, 197, 195,
		1, 0, 0, 0, 197, 198, 1, 0, 0, 0, 198, 199, 1, 0, 0, 0, 199, 200, 5, 25,
		0, 0, 200, 225, 1, 0, 0, 0, 201, 205, 5, 24, 0, 0, 202, 204, 3, 28, 14,
		0, 203, 202, 1, 0, 0, 0, 204, 207, 1, 0, 0, 0, 205, 203, 1, 0, 0, 0, 205,
		206, 1, 0, 0, 0, 206, 208, 1, 0, 0, 0, 207, 205, 1, 0, 0, 0, 208, 209,
		5, 13, 0, 0, 209, 210, 3, 26, 13, 0, 210, 211, 5, 25, 0, 0, 211, 225, 1,
		0, 0, 0, 212, 213, 5, 17, 0, 0, 213, 217, 5, 24, 0, 0, 214, 216, 3, 28,
		14, 0, 215, 214, 1, 0, 0, 0, 216, 219, 1, 0, 0, 0, 217, 215, 1, 0, 0, 0,
		217, 218, 1, 0, 0, 0, 218, 220, 1, 0, 0, 0, 219, 217, 1, 0, 0, 0, 220,
		221, 5, 13, 0, 0, 221, 222, 3, 26, 13, 0, 222, 223, 5, 25, 0, 0, 223, 225,
		1, 0, 0, 0, 224, 192, 1, 0, 0, 0, 224, 193, 1, 0, 0, 0, 224, 201, 1, 0,
		0, 0, 224, 212, 1, 0, 0, 0, 225, 31, 1, 0, 0, 0, 226, 227, 3, 20, 10, 0,
		227, 228, 3, 26, 13, 0, 228, 33, 1, 0, 0, 0, 229, 232, 3, 4, 2, 0, 230,
		232, 3, 18, 9, 0, 231, 229, 1, 0, 0, 0, 231, 230, 1, 0, 0, 0, 232, 35,
		1, 0, 0, 0, 233, 234, 5, 9, 0, 0, 234, 235, 3, 34, 17, 0, 235, 236, 5,
		20, 0, 0, 236, 237, 5, 7, 0, 0, 237, 238, 5, 27, 0, 0, 238, 239, 3, 22,
		11, 0, 239, 240, 5, 8, 0, 0, 240, 241, 5, 27, 0, 0, 241, 242, 3, 30, 15,
		0, 242, 243, 5, 21, 0, 0, 243, 251, 1, 0, 0, 0, 244, 245, 5, 19, 0, 0,
		245, 246, 3, 34, 17, 0, 246, 247, 5, 10, 0, 0, 247, 248, 3, 4, 2, 0, 248,
		251, 1, 0, 0, 0, 249, 251, 3, 8, 4, 0, 250, 233, 1, 0, 0, 0, 250, 244,
		1, 0, 0, 0, 250, 249, 1, 0, 0, 0, 251, 37, 1, 0, 0, 0, 23, 45, 57, 68,
		76, 78, 86, 92, 102, 109, 115, 120, 132, 151, 159, 168, 177, 186, 197,
		205, 217, 224, 231, 250,
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
	NumscriptParserUP                         = 11
	NumscriptParserTO                         = 12
	NumscriptParserREMAINING                  = 13
	NumscriptParserALLOWING                   = 14
	NumscriptParserUNBOUNDED                  = 15
	NumscriptParserOVERDRAFT                  = 16
	NumscriptParserONEOF                      = 17
	NumscriptParserKEPT                       = 18
	NumscriptParserSAVE                       = 19
	NumscriptParserLPARENS                    = 20
	NumscriptParserRPARENS                    = 21
	NumscriptParserLBRACKET                   = 22
	NumscriptParserRBRACKET                   = 23
	NumscriptParserLBRACE                     = 24
	NumscriptParserRBRACE                     = 25
	NumscriptParserCOMMA                      = 26
	NumscriptParserEQ                         = 27
	NumscriptParserSTAR                       = 28
	NumscriptParserPLUS                       = 29
	NumscriptParserMINUS                      = 30
	NumscriptParserDIV                        = 31
	NumscriptParserPERCENTAGE_PORTION_LITERAL = 32
	NumscriptParserSTRING                     = 33
	NumscriptParserIDENTIFIER                 = 34
	NumscriptParserNUMBER                     = 35
	NumscriptParserASSET                      = 36
	NumscriptParserACCOUNT_START              = 37
	NumscriptParserCOLON                      = 38
	NumscriptParserACCOUNT_TEXT               = 39
	NumscriptParserVARIABLE_NAME_ACC          = 40
	NumscriptParserVARIABLE_NAME              = 41
)

// NumscriptParser rules.
const (
	NumscriptParserRULE_monetaryLit              = 0
	NumscriptParserRULE_accountLiteralPart       = 1
	NumscriptParserRULE_valueExpr                = 2
	NumscriptParserRULE_functionCallArgs         = 3
	NumscriptParserRULE_functionCall             = 4
	NumscriptParserRULE_varOrigin                = 5
	NumscriptParserRULE_varDeclaration           = 6
	NumscriptParserRULE_varsDeclaration          = 7
	NumscriptParserRULE_program                  = 8
	NumscriptParserRULE_sentAllLit               = 9
	NumscriptParserRULE_allotment                = 10
	NumscriptParserRULE_source                   = 11
	NumscriptParserRULE_allotmentClauseSrc       = 12
	NumscriptParserRULE_keptOrDestination        = 13
	NumscriptParserRULE_destinationInOrderClause = 14
	NumscriptParserRULE_destination              = 15
	NumscriptParserRULE_allotmentClauseDest      = 16
	NumscriptParserRULE_sentValue                = 17
	NumscriptParserRULE_statement                = 18
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
		p.SetState(38)
		p.Match(NumscriptParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(39)

		var _x = p.valueExpr(0)

		localctx.(*MonetaryLitContext).asset = _x
	}

	{
		p.SetState(40)

		var _x = p.valueExpr(0)

		localctx.(*MonetaryLitContext).amt = _x
	}

	{
		p.SetState(41)
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

// IAccountLiteralPartContext is an interface to support dynamic dispatch.
type IAccountLiteralPartContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsAccountLiteralPartContext differentiates from other interfaces.
	IsAccountLiteralPartContext()
}

type AccountLiteralPartContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAccountLiteralPartContext() *AccountLiteralPartContext {
	var p = new(AccountLiteralPartContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_accountLiteralPart
	return p
}

func InitEmptyAccountLiteralPartContext(p *AccountLiteralPartContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NumscriptParserRULE_accountLiteralPart
}

func (*AccountLiteralPartContext) IsAccountLiteralPartContext() {}

func NewAccountLiteralPartContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AccountLiteralPartContext {
	var p = new(AccountLiteralPartContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NumscriptParserRULE_accountLiteralPart

	return p
}

func (s *AccountLiteralPartContext) GetParser() antlr.Parser { return s.parser }

func (s *AccountLiteralPartContext) CopyAll(ctx *AccountLiteralPartContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *AccountLiteralPartContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AccountLiteralPartContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type AccountVarPartContext struct {
	AccountLiteralPartContext
}

func NewAccountVarPartContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AccountVarPartContext {
	var p = new(AccountVarPartContext)

	InitEmptyAccountLiteralPartContext(&p.AccountLiteralPartContext)
	p.parser = parser
	p.CopyAll(ctx.(*AccountLiteralPartContext))

	return p
}

func (s *AccountVarPartContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AccountVarPartContext) VARIABLE_NAME_ACC() antlr.TerminalNode {
	return s.GetToken(NumscriptParserVARIABLE_NAME_ACC, 0)
}

func (s *AccountVarPartContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterAccountVarPart(s)
	}
}

func (s *AccountVarPartContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitAccountVarPart(s)
	}
}

type AccountTextPartContext struct {
	AccountLiteralPartContext
}

func NewAccountTextPartContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AccountTextPartContext {
	var p = new(AccountTextPartContext)

	InitEmptyAccountLiteralPartContext(&p.AccountLiteralPartContext)
	p.parser = parser
	p.CopyAll(ctx.(*AccountLiteralPartContext))

	return p
}

func (s *AccountTextPartContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AccountTextPartContext) ACCOUNT_TEXT() antlr.TerminalNode {
	return s.GetToken(NumscriptParserACCOUNT_TEXT, 0)
}

func (s *AccountTextPartContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterAccountTextPart(s)
	}
}

func (s *AccountTextPartContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitAccountTextPart(s)
	}
}

func (p *NumscriptParser) AccountLiteralPart() (localctx IAccountLiteralPartContext) {
	localctx = NewAccountLiteralPartContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, NumscriptParserRULE_accountLiteralPart)
	p.SetState(45)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserACCOUNT_TEXT:
		localctx = NewAccountTextPartContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(43)
			p.Match(NumscriptParserACCOUNT_TEXT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserVARIABLE_NAME_ACC:
		localctx = NewAccountVarPartContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(44)
			p.Match(NumscriptParserVARIABLE_NAME_ACC)
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

func (s *InfixExprContext) DIV() antlr.TerminalNode {
	return s.GetToken(NumscriptParserDIV, 0)
}

func (s *InfixExprContext) PLUS() antlr.TerminalNode {
	return s.GetToken(NumscriptParserPLUS, 0)
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

type ApplicationContext struct {
	ValueExprContext
}

func NewApplicationContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ApplicationContext {
	var p = new(ApplicationContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *ApplicationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ApplicationContext) FunctionCall() IFunctionCallContext {
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

func (s *ApplicationContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.EnterApplication(s)
	}
}

func (s *ApplicationContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NumscriptListener); ok {
		listenerT.ExitApplication(s)
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

func (s *AccountLiteralContext) ACCOUNT_START() antlr.TerminalNode {
	return s.GetToken(NumscriptParserACCOUNT_START, 0)
}

func (s *AccountLiteralContext) AllAccountLiteralPart() []IAccountLiteralPartContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAccountLiteralPartContext); ok {
			len++
		}
	}

	tst := make([]IAccountLiteralPartContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAccountLiteralPartContext); ok {
			tst[i] = t.(IAccountLiteralPartContext)
			i++
		}
	}

	return tst
}

func (s *AccountLiteralContext) AccountLiteralPart(i int) IAccountLiteralPartContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAccountLiteralPartContext); ok {
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

	return t.(IAccountLiteralPartContext)
}

func (s *AccountLiteralContext) AllCOLON() []antlr.TerminalNode {
	return s.GetTokens(NumscriptParserCOLON)
}

func (s *AccountLiteralContext) COLON(i int) antlr.TerminalNode {
	return s.GetToken(NumscriptParserCOLON, i)
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
	_startState := 4
	p.EnterRecursionRule(localctx, 4, NumscriptParserRULE_valueExpr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(68)
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
			p.SetState(48)
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
			p.SetState(49)
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
			p.SetState(50)
			p.Match(NumscriptParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserACCOUNT_START:
		localctx = NewAccountLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(51)
			p.Match(NumscriptParserACCOUNT_START)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(52)
			p.AccountLiteralPart()
		}
		p.SetState(57)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(53)
					p.Match(NumscriptParserCOLON)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(54)
					p.AccountLiteralPart()
				}

			}
			p.SetState(59)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	case NumscriptParserNUMBER:
		localctx = NewNumberLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(60)
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
			p.SetState(61)
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
			p.SetState(62)
			p.MonetaryLit()
		}

	case NumscriptParserLPARENS:
		localctx = NewParenthesizedExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(63)
			p.Match(NumscriptParserLPARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(64)
			p.valueExpr(0)
		}
		{
			p.SetState(65)
			p.Match(NumscriptParserRPARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NumscriptParserOVERDRAFT, NumscriptParserIDENTIFIER:
		localctx = NewApplicationContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(67)
			p.FunctionCall()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(78)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(76)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) {
			case 1:
				localctx = NewInfixExprContext(p, NewValueExprContext(p, _parentctx, _parentState))
				localctx.(*InfixExprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_valueExpr)
				p.SetState(70)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(71)

					var _m = p.Match(NumscriptParserDIV)

					localctx.(*InfixExprContext).op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(72)

					var _x = p.valueExpr(5)

					localctx.(*InfixExprContext).right = _x
				}

			case 2:
				localctx = NewInfixExprContext(p, NewValueExprContext(p, _parentctx, _parentState))
				localctx.(*InfixExprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, NumscriptParserRULE_valueExpr)
				p.SetState(73)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(74)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*InfixExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == NumscriptParserPLUS || _la == NumscriptParserMINUS) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*InfixExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(75)

					var _x = p.valueExpr(4)

					localctx.(*InfixExprContext).right = _x
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(80)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
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
		p.SetState(81)
		p.valueExpr(0)
	}
	p.SetState(86)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NumscriptParserCOMMA {
		{
			p.SetState(82)
			p.Match(NumscriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
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
		p.SetState(89)

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
		p.SetState(90)
		p.Match(NumscriptParserLPARENS)
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

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2469611503616) != 0 {
		{
			p.SetState(91)
			p.FunctionCallArgs()
		}

	}
	{
		p.SetState(94)
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
	ValueExpr() IValueExprContext

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

func (s *VarOriginContext) ValueExpr() IValueExprContext {
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
		p.SetState(96)
		p.Match(NumscriptParserEQ)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(97)
		p.valueExpr(0)
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
		p.SetState(99)

		var _m = p.Match(NumscriptParserIDENTIFIER)

		localctx.(*VarDeclarationContext).type_ = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(100)

		var _m = p.Match(NumscriptParserVARIABLE_NAME)

		localctx.(*VarDeclarationContext).name = _m
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

	if _la == NumscriptParserEQ {
		{
			p.SetState(101)
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
		p.SetState(104)
		p.Match(NumscriptParserVARS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(105)
		p.Match(NumscriptParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(109)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NumscriptParserIDENTIFIER {
		{
			p.SetState(106)
			p.VarDeclaration()
		}

		p.SetState(111)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(112)
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
	p.SetState(115)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NumscriptParserVARS {
		{
			p.SetState(114)
			p.VarsDeclaration()
		}

	}
	p.SetState(120)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&17180459520) != 0 {
		{
			p.SetState(117)
			p.Statement()
		}

		p.SetState(122)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(123)
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
		p.SetState(125)
		p.Match(NumscriptParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

	{
		p.SetState(126)

		var _x = p.valueExpr(0)

		localctx.(*SentAllLitContext).asset = _x
	}

	{
		p.SetState(127)
		p.Match(NumscriptParserSTAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(128)
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
	p.EnterRule(localctx, 20, NumscriptParserRULE_allotment)
	p.SetState(132)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserOVERDRAFT, NumscriptParserLPARENS, NumscriptParserLBRACKET, NumscriptParserPERCENTAGE_PORTION_LITERAL, NumscriptParserSTRING, NumscriptParserIDENTIFIER, NumscriptParserNUMBER, NumscriptParserASSET, NumscriptParserACCOUNT_START, NumscriptParserVARIABLE_NAME:
		localctx = NewPortionedAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(130)
			p.valueExpr(0)
		}

	case NumscriptParserREMAINING:
		localctx = NewRemainingAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(131)
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

func (s *SrcOneofContext) ONEOF() antlr.TerminalNode {
	return s.GetToken(NumscriptParserONEOF, 0)
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
	p.EnterRule(localctx, 22, NumscriptParserRULE_source)
	var _la int

	p.SetState(177)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 15, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSrcAccountUnboundedOverdraftContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(134)

			var _x = p.valueExpr(0)

			localctx.(*SrcAccountUnboundedOverdraftContext).address = _x
		}
		{
			p.SetState(135)
			p.Match(NumscriptParserALLOWING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(136)
			p.Match(NumscriptParserUNBOUNDED)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(137)
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
			p.SetState(139)

			var _x = p.valueExpr(0)

			localctx.(*SrcAccountBoundedOverdraftContext).address = _x
		}
		{
			p.SetState(140)
			p.Match(NumscriptParserALLOWING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(141)
			p.Match(NumscriptParserOVERDRAFT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(142)
			p.Match(NumscriptParserUP)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(143)
			p.Match(NumscriptParserTO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(144)

			var _x = p.valueExpr(0)

			localctx.(*SrcAccountBoundedOverdraftContext).maxOvedraft = _x
		}

	case 3:
		localctx = NewSrcAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(146)
			p.valueExpr(0)
		}

	case 4:
		localctx = NewSrcAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(147)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(149)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2469611511808) != 0) {
			{
				p.SetState(148)
				p.AllotmentClauseSrc()
			}

			p.SetState(151)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(153)
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
			p.SetState(155)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(159)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2469628411968) != 0 {
			{
				p.SetState(156)
				p.Source()
			}

			p.SetState(161)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(162)
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
			p.SetState(163)
			p.Match(NumscriptParserONEOF)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(164)
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

		for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2469628411968) != 0) {
			{
				p.SetState(165)
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
			p.SetState(170)
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
			p.SetState(172)
			p.Match(NumscriptParserMAX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(173)

			var _x = p.valueExpr(0)

			localctx.(*SrcCappedContext).cap_ = _x
		}
		{
			p.SetState(174)
			p.Match(NumscriptParserFROM)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(175)
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
	p.EnterRule(localctx, 24, NumscriptParserRULE_allotmentClauseSrc)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(179)
		p.Allotment()
	}
	{
		p.SetState(180)
		p.Match(NumscriptParserFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(181)
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
	p.EnterRule(localctx, 26, NumscriptParserRULE_keptOrDestination)
	p.SetState(186)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserTO:
		localctx = NewDestinationToContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(183)
			p.Match(NumscriptParserTO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(184)
			p.Destination()
		}

	case NumscriptParserKEPT:
		localctx = NewDestinationKeptContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(185)
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
	p.EnterRule(localctx, 28, NumscriptParserRULE_destinationInOrderClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(188)
		p.Match(NumscriptParserMAX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(189)
		p.valueExpr(0)
	}
	{
		p.SetState(190)
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

func (s *DestOneofContext) ONEOF() antlr.TerminalNode {
	return s.GetToken(NumscriptParserONEOF, 0)
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
	p.EnterRule(localctx, 30, NumscriptParserRULE_destination)
	var _la int

	p.SetState(224)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 20, p.GetParserRuleContext()) {
	case 1:
		localctx = NewDestAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(192)
			p.valueExpr(0)
		}

	case 2:
		localctx = NewDestAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(193)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(195)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2469611511808) != 0) {
			{
				p.SetState(194)
				p.AllotmentClauseDest()
			}

			p.SetState(197)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(199)
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
			p.SetState(201)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(205)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == NumscriptParserMAX {
			{
				p.SetState(202)
				p.DestinationInOrderClause()
			}

			p.SetState(207)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(208)
			p.Match(NumscriptParserREMAINING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(209)
			p.KeptOrDestination()
		}
		{
			p.SetState(210)
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
			p.SetState(212)
			p.Match(NumscriptParserONEOF)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(213)
			p.Match(NumscriptParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(217)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == NumscriptParserMAX {
			{
				p.SetState(214)
				p.DestinationInOrderClause()
			}

			p.SetState(219)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(220)
			p.Match(NumscriptParserREMAINING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(221)
			p.KeptOrDestination()
		}
		{
			p.SetState(222)
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
	p.EnterRule(localctx, 32, NumscriptParserRULE_allotmentClauseDest)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(226)
		p.Allotment()
	}
	{
		p.SetState(227)
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
	p.EnterRule(localctx, 34, NumscriptParserRULE_sentValue)
	p.SetState(231)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSentLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(229)
			p.valueExpr(0)
		}

	case 2:
		localctx = NewSentAllContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(230)
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
	p.EnterRule(localctx, 36, NumscriptParserRULE_statement)
	p.SetState(250)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NumscriptParserSEND:
		localctx = NewSendStatementContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(233)
			p.Match(NumscriptParserSEND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(234)
			p.SentValue()
		}
		{
			p.SetState(235)
			p.Match(NumscriptParserLPARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(236)
			p.Match(NumscriptParserSOURCE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(237)
			p.Match(NumscriptParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(238)
			p.Source()
		}
		{
			p.SetState(239)
			p.Match(NumscriptParserDESTINATION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(240)
			p.Match(NumscriptParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(241)
			p.Destination()
		}
		{
			p.SetState(242)
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
			p.SetState(244)
			p.Match(NumscriptParserSAVE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(245)
			p.SentValue()
		}
		{
			p.SetState(246)
			p.Match(NumscriptParserFROM)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(247)
			p.valueExpr(0)
		}

	case NumscriptParserOVERDRAFT, NumscriptParserIDENTIFIER:
		localctx = NewFnCallStatementContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(249)
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

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *NumscriptParser) ValueExpr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 4)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 3)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
