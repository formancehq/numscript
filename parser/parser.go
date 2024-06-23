package parser

import (
	parser "numscript/parser/antlr"
	"strconv"

	"github.com/antlr4-go/antlr/v4"
)

type LexerError struct{}
type ParserError struct{}

type ParseResult[T any] struct {
	Value        T
	ParserErrors []ParserError
	LexerErrors  []LexerError
}

func Parse(input string) ParseResult[Program] {
	var parserErrors []ParserError
	var lexerErrors []LexerError

	is := antlr.NewInputStream(input)
	lexer := parser.NewNumscriptLexer(is)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	parser := parser.NewNumscriptParser(stream)

	return ParseResult[Program]{
		Value:        parseProgram(parser.Program()),
		ParserErrors: parserErrors,
		LexerErrors:  lexerErrors,
	}
}

func parseProgram(programCtx parser.IProgramContext) Program {
	var statements []Statement

	for _, statementCtx := range programCtx.AllStatement() {
		statements = append(statements, parseStatement(statementCtx))
	}

	return Program{statements}
}

func parseStatement(statementCtx parser.IStatementContext) Statement {
	return &SendStatement{
		Range:    ctxToRange(statementCtx),
		Monetary: parseMonetaryLit(statementCtx.MonetaryLit()),
	}
}

func parseMonetaryLit(monetaryLitCtx parser.IMonetaryLitContext) Literal {
	amtStr := monetaryLitCtx.GetAmt().GetText()

	amt, err := strconv.Atoi(amtStr)
	if err != nil {
		panic("Invalid amt: " + amtStr)
	}

	return &MonetaryLiteral{
		Range:  ctxToRange(monetaryLitCtx),
		Asset:  monetaryLitCtx.GetAsset().GetText(),
		Amount: amt,
	}
}

func ctxToRange(ctx antlr.ParserRuleContext) Range {
	startTk := ctx.GetStart()
	endTk := ctx.GetStop()

	return Range{
		Start: Position{
			Line:      startTk.GetLine() - 1,
			Character: startTk.GetColumn(),
		},
		End: Position{
			Line: endTk.GetLine() - 1,

			// this is based on the assumption that a token cannot span multiple lines
			Character: endTk.GetColumn() + len(endTk.GetText()),
		},
	}
}
