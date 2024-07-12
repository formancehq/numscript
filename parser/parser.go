package parser

import (
	"math"
	parser "numscript/parser/antlr"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

type LexerError struct{}
type ParserError struct {
	Range Range
	Msg   string
}

type ParseResult[T any] struct {
	Value       T
	Errors      []ParserError
	LexerErrors []LexerError
}

type ErrorListener struct {
	antlr.DefaultErrorListener
	Errors []ParserError
}

func (l *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, startL, startC int, msg string, e antlr.RecognitionException) {
	length := 1
	if token, ok := offendingSymbol.(antlr.Token); ok {
		length = len(token.GetText())
	}
	endL := startL
	endC := startC + length - 1 // -1 so that end character is inside the offending token
	l.Errors = append(l.Errors, ParserError{
		Msg: msg,
		Range: Range{
			Start: Position{Character: startC, Line: startL - 1},
			End:   Position{Character: endC, Line: endL - 1},
		},
	})
}

func Parse(input string) ParseResult[Program] {
	// TODO handle lexer errors
	listener := &ErrorListener{}

	is := antlr.NewInputStream(input)
	lexer := parser.NewNumscriptLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	parser := parser.NewNumscriptParser(stream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)

	parsed := parseProgram(parser.Program())

	return ParseResult[Program]{
		Value:  parsed,
		Errors: listener.Errors,
	}
}

func parseProgram(programCtx parser.IProgramContext) Program {
	var statements []Statement
	for _, statementCtx := range programCtx.AllStatement() {
		statements = append(statements, parseStatement(statementCtx))
	}

	var vars []VarDeclaration
	if declarationCtx := programCtx.VarsDeclaration(); declarationCtx != nil {
		for _, varDecl := range declarationCtx.AllVarDeclaration() {
			decl := parseVarDeclaration(varDecl)
			if decl != nil {
				vars = append(vars, *decl)
			}
		}
	}

	return Program{
		Statements: statements,
		Vars:       vars,
	}
}

func parseVarDeclaration(varDecl parser.IVarDeclarationContext) *VarDeclaration {
	if varDecl == nil {
		return nil
	}

	return &VarDeclaration{
		Range: ctxToRange(varDecl),
		Type:  parseVarType(varDecl.GetType_()),
		Name:  parseVarLiteral(varDecl.GetName()),
	}
}

func parseVarLiteral(tk antlr.Token) *VariableLiteral {
	if tk == nil || tk.GetTokenIndex() == -1 {
		return nil
	}

	name := tk.GetText()[1:]

	return &VariableLiteral{
		Range: tokenToRange(tk),
		Name:  name,
	}
}

func parseVarType(tk antlr.Token) *TypeDecl {
	if tk == nil || tk.GetTokenIndex() == -1 {
		return nil
	}

	return &TypeDecl{
		Range: tokenToRange(tk),
		Name:  tk.GetText(),
	}
}

func parseCapLit(capCtx parser.ICapContext) Literal {
	switch capCtx := capCtx.(type) {
	case *parser.LitCapContext:
		return parseMonetaryLit(capCtx.MonetaryLit())

	case *parser.VarCapContext:
		// Discard the '$'
		name := capCtx.GetText()[1:]
		return &VariableLiteral{
			Range: ctxToRange(capCtx),
			Name:  name,
		}

	case *parser.CapContext:
		return nil

	default:
		panic("Invalid ctx")
	}
}

func parseSource(sourceCtx parser.ISourceContext) Source {
	if sourceCtx == nil {
		return nil
	}

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

	case *parser.SrcCappedContext:
		return &SourceCapped{
			Range: range_,
			From:  parseSource(sourceCtx.Source()),
			Cap:   parseCapLit(sourceCtx.Cap_()),
		}

	case *parser.SourceContext:
		return nil

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

	case *parser.PortionContext:
		return nil

	default:
		panic("unhandled portion ctx")
	}
}

func parseDestination(destCtx parser.IDestinationContext) Destination {
	if destCtx == nil {
		return nil
	}

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

	case *parser.DestinationContext:
		return nil

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

	case *parser.PortionContext:
		return nil

	default:
		panic("unhandled portion ctx")
	}
}

func parseStatement(statementCtx parser.IStatementContext) Statement {
	return &SendStatement{
		Source:      parseSource(statementCtx.Source()),
		Destination: parseDestination(statementCtx.Destination()),
		Range:       ctxToRange(statementCtx),
		Monetary:    parseSendExpr(statementCtx.SendExpr()),
	}
}

func parseSendExpr(sendExpr parser.ISendExprContext) Literal {
	switch sendExpr := sendExpr.(type) {
	case *parser.SendVariableContext:
		return parseVarLiteral(sendExpr.VARIABLE_NAME().GetSymbol())

	case *parser.SendMonContext:
		return parseMonetaryLit(sendExpr.MonetaryLit())

	case *parser.SendExprContext:
		return nil
	default:
		panic("Unhandled SendExprContext")
	}
}

func parseMonetaryLit(monetaryLitCtx parser.IMonetaryLitContext) *MonetaryLiteral {

	if monetaryLitCtx.GetAmt() == nil {
		return nil
	}

	// TODO better err handling
	if monetaryLitCtx.GetAmt().GetTokenIndex() == -1 {
		return nil
	}

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

func tokenToRange(tk antlr.Token) Range {
	return Range{
		Start: Position{
			Line:      tk.GetLine() - 1,
			Character: tk.GetColumn(),
		},
		End: Position{
			Line:      tk.GetLine() - 1,
			Character: tk.GetColumn() + len(tk.GetText()),
		},
	}
}
