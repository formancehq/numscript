package parser

import (
	"math"
	parser "numscript/parser/antlr"
	"strconv"
	"strings"

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

func parseSource(sourceCtx parser.ISourceContext) Source {
	range_ := ctxToRange(sourceCtx)

	switch sourceCtx := sourceCtx.(type) {
	case *parser.SrcAccountContext:
		// Discard the '@'
		name := sourceCtx.GetText()[1:]

		return &AccountLiteral{
			Range: range_,
			Name:  name,
		}

	case *parser.SrcVariableContext:
		// Discard the '$'
		name := sourceCtx.GetText()[1:]

		return &VariableLiteral{
			Range: range_,
			Name:  name,
		}

	case *parser.SrcSeqContext:
		var sources []Source
		for _, sourceCtx := range sourceCtx.AllSource() {
			sources = append(sources, parseSource(sourceCtx))
		}
		return &SourceSeq{
			Range:   range_,
			Sources: sources,
		}

	case *parser.SrcAllotmentContext:
		var items []SourceAllotmentItem
		for _, itemCtx := range sourceCtx.AllAllotmentClauseSrc() {
			item := SourceAllotmentItem{
				Range:     ctxToRange(itemCtx),
				Allotment: parseSourceAllotment(itemCtx.Portion()),
				From:      parseSource(itemCtx.Source()),
			}
			items = append(items, item)
		}
		return &SourceAllotment{
			Range: range_,
			Items: items,
		}

	case *parser.SourceContext:
		panic("Invalid source context" + sourceCtx.GetText())

	default:
		panic("unhandled context: " + sourceCtx.GetText())
	}
}

func parseRatio(source string, range_ Range) *RatioLiteral {
	split := strings.Split(source, "/")

	num, err := strconv.ParseUint(split[0], 0, 64)
	if err != nil {
		panic(err)
	}

	den, err := strconv.ParseUint(split[1], 0, 64)
	if err != nil {
		panic(err)
	}

	return &RatioLiteral{
		Range:       range_,
		Numerator:   num,
		Denominator: den,
	}
}

func parsePercentageRatio(source string, range_ Range) *RatioLiteral {
	str := strings.TrimSuffix(source, "%")
	num, err := strconv.ParseUint(strings.Replace(str, ".", "", -1), 0, 64)
	if err != nil {
		panic(err)
	}

	var denominator uint64
	split := strings.Split(str, ".")
	if len(split) > 1 {
		// TODO verify this is always correct
		floatingDigits := len(split[1])
		denominator = (uint64)(math.Pow10(2 + floatingDigits))
	} else {
		denominator = 100
	}

	return &RatioLiteral{
		Range:       range_,
		Numerator:   num,
		Denominator: denominator,
	}
}

func parseSourceAllotment(portionCtx parser.IPortionContext) SourceAllotmentValue {
	switch portionCtx.(type) {
	case *parser.RatioContext:
		return parseRatio(portionCtx.GetText(), ctxToRange(portionCtx))

	case *parser.PercentageContext:
		return parsePercentageRatio(portionCtx.GetText(), ctxToRange(portionCtx))

	default:
		panic("unhandled portion ctx")
	}
}

func parseDestination(destCtx parser.IDestinationContext) Destination {
	range_ := ctxToRange(destCtx)

	switch destCtx := destCtx.(type) {
	case *parser.DestAccountContext:
		// Discard the '@'
		name := destCtx.GetText()[1:]

		return &AccountLiteral{
			Range: range_,
			Name:  name,
		}

	case *parser.DestVariableContext:
		// Discard the '$'
		name := destCtx.GetText()[1:]

		return &VariableLiteral{
			Range: range_,
			Name:  name,
		}

	case *parser.DestSeqContext:
		var destinations []Destination
		for _, destCtx := range destCtx.AllDestination() {
			destinations = append(destinations, parseDestination(destCtx))
		}
		return &DestinationSeq{
			Range:        range_,
			Destinations: destinations,
		}

	case *parser.DestAllotmentContext:
		var items []DestinationAllotmentItem
		for _, itemCtx := range destCtx.AllAllotmentClauseDest() {
			item := DestinationAllotmentItem{
				Range:     ctxToRange(itemCtx),
				Allotment: parseDestinationAllotment(itemCtx.Portion()),
				To:        parseDestination(itemCtx.Destination()),
			}
			items = append(items, item)
		}
		return &DestinationAllotment{
			Range: range_,
			Items: items,
		}

	default:
		panic("Unhandled dest" + destCtx.GetText())

	}

}

func parseDestinationAllotment(portionCtx parser.IPortionContext) DestinationAllotmentValue {
	switch portionCtx.(type) {
	case *parser.RatioContext:
		return parseRatio(portionCtx.GetText(), ctxToRange(portionCtx))

	case *parser.PercentageContext:
		return parsePercentageRatio(portionCtx.GetText(), ctxToRange(portionCtx))

	default:
		panic("unhandled portion ctx")
	}
}

func parseStatement(statementCtx parser.IStatementContext) Statement {
	return &SendStatement{
		Source:      parseSource(statementCtx.Source()),
		Destination: parseDestination(statementCtx.Destination()),
		Range:       ctxToRange(statementCtx),
		Monetary:    parseMonetaryLit(statementCtx.MonetaryLit()),
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
