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
		"", "'+'", "'-'",
	}
	staticData.RuleNames = []string{
		"T__0", "T__1",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 2, 9, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 1, 0, 1, 0, 1, 1, 1, 1, 0, 0,
		2, 1, 1, 3, 2, 1, 0, 0, 8, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 1, 5, 1,
		0, 0, 0, 3, 7, 1, 0, 0, 0, 5, 6, 5, 43, 0, 0, 6, 2, 1, 0, 0, 0, 7, 8, 5,
		45, 0, 0, 8, 4, 1, 0, 0, 0, 1, 0, 0,
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
	NumscriptLexerT__0 = 1
	NumscriptLexerT__1 = 2
)
