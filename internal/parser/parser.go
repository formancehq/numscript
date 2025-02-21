package parser

import (
	"math/big"
	"strconv"
	"strings"

	parser "github.com/formancehq/numscript/internal/parser/antlr"
	"github.com/formancehq/numscript/internal/utils"

	"github.com/antlr4-go/antlr/v4"
)

type ParserError struct {
	Range Range
	Msg   string
}

type ParseResult struct {
	Source string
	Value  Program
	Errors []ParserError
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

func Parse(input string) ParseResult {
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

	return ParseResult{
		Source: input,
		Value:  parsed,
		Errors: listener.Errors,
	}
}

func ParseErrorsToString(errors []ParserError, source string) string {
	buf := "Got errors while parsing:\n"
	for _, err := range errors {
		buf += err.Msg + "\n" + err.Range.ShowOnSource(source) + "\n"
	}
	return buf
}

func parseVarsDeclaration(varsCtx parser.IVarsDeclarationContext) []VarDeclaration {
	if varsCtx == nil {
		return nil
	}

	var vars []VarDeclaration
	for _, varDecl := range varsCtx.AllVarDeclaration() {
		decl := parseVarDeclaration(varDecl)
		if decl != nil {
			vars = append(vars, *decl)
		}
	}
	return vars
}

func parseProgram(programCtx parser.IProgramContext) Program {
	vars := parseVarsDeclaration(programCtx.VarsDeclaration())

	var statements []Statement
	for _, statementCtx := range programCtx.AllStatement() {
		statements = append(statements, parseStatement(statementCtx))
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

	var fnCallStatement *FnCall
	if varDecl.VarOrigin() != nil {
		fnCallStatement = parseFnCall(varDecl.VarOrigin().FunctionCall())
	}

	return &VarDeclaration{
		Range:  ctxToRange(varDecl),
		Type:   parseVarType(varDecl.GetType_()),
		Name:   parseVarLiteral(varDecl.GetName()),
		Origin: fnCallStatement,
	}
}

func parseVarLiteral(tk antlr.Token) *Variable {
	if tk == nil || tk.GetTokenIndex() == -1 {
		return nil
	}

	name := tk.GetText()[1:]

	return &Variable{
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

func parseSource(sourceCtx parser.ISourceContext) Source {
	if sourceCtx == nil {
		return nil
	}

	range_ := ctxToRange(sourceCtx)

	switch sourceCtx := sourceCtx.(type) {
	case *parser.SrcAccountContext:
		return &SourceAccount{
			ValueExpr: parseValueExpr(sourceCtx.ValueExpr()),
		}

	case *parser.SrcInorderContext:
		var sources []Source
		for _, sourceCtx := range sourceCtx.AllSource() {
			sources = append(sources, parseSource(sourceCtx))
		}
		return &SourceInorder{
			Range:   range_,
			Sources: sources,
		}

	case *parser.SrcOneofContext:
		var sources []Source
		for _, sourceCtx := range sourceCtx.AllSource() {
			sources = append(sources, parseSource(sourceCtx))
		}
		return &SourceOneof{
			Range:   range_,
			Sources: sources,
		}

	case *parser.SrcAllotmentContext:
		var items []SourceAllotmentItem
		for _, itemCtx := range sourceCtx.AllAllotmentClauseSrc() {
			item := SourceAllotmentItem{
				Range:     ctxToRange(itemCtx),
				Allotment: parseAllotment(itemCtx.Allotment()),
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
			Cap:   parseValueExpr(sourceCtx.GetCap_()),
		}

	case *parser.SrcAccountUnboundedOverdraftContext:
		return &SourceOverdraft{
			Range:   ctxToRange(sourceCtx),
			Address: parseValueExpr(sourceCtx.GetAddress()),
		}

	case *parser.SrcAccountBoundedOverdraftContext:
		varMon := parseValueExpr(sourceCtx.GetMaxOvedraft())

		return &SourceOverdraft{
			Range:   ctxToRange(sourceCtx),
			Address: parseValueExpr(sourceCtx.GetAddress()),
			Bounded: &varMon,
		}

	case *parser.SourceContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[Source](sourceCtx.GetText())
	}
}

// TODO actually handle big int
func ParsePercentageRatio(source string) (*big.Int, uint16, error) {
	str := strings.TrimSuffix(source, "%")
	num, err := strconv.ParseUint(strings.Replace(str, ".", "", -1), 0, 64)
	if err != nil {
		return nil, 0, err
	}

	var floatingDigits uint16
	split := strings.Split(str, ".")
	if len(split) > 1 {
		floatingDigits = uint16(len(split[1]))
	} else {
		floatingDigits = 0
	}

	return big.NewInt(int64(num)), floatingDigits, nil
}

func parsePercentageRatio(source string, range_ Range) *PercentageLiteral {
	num, floatingDigits, err := ParsePercentageRatio(source)
	if err != nil {
		panic(err)
	}

	return &PercentageLiteral{
		Range:          range_,
		Amount:         num,
		FloatingDigits: floatingDigits,
	}
}

func parseAllotment(allotmentCtx parser.IAllotmentContext) AllotmentValue {
	switch allotmentCtx := allotmentCtx.(type) {
	case *parser.PortionedAllotmentContext:
		expr := parseValueExpr(allotmentCtx.ValueExpr())
		return &ValueExprAllotment{expr}

	case *parser.RemainingAllotmentContext:
		return &RemainingAllotment{
			Range: ctxToRange(allotmentCtx),
		}

	case *parser.AllotmentContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[AllotmentValue](allotmentCtx.GetText())
	}
}

func parseStringLiteralCtx(stringCtx *parser.StringLiteralContext) *StringLiteral {
	rawStr := stringCtx.GetText()
	// Remove leading and trailing '"'
	innerStr := rawStr[1 : len(rawStr)-1]
	return &StringLiteral{
		Range:  ctxToRange(stringCtx),
		String: innerStr,
	}
}

func parseValueExpr(valueExprCtx parser.IValueExprContext) ValueExpr {
	switch valueExprCtx := valueExprCtx.(type) {
	case *parser.ParenthesizedExprContext:
		return parseValueExpr(valueExprCtx.ValueExpr())

	case *parser.AccountLiteralContext:
		return &AccountLiteral{
			Range: ctxToRange(valueExprCtx),
			// Discard the '@'
			Name: valueExprCtx.GetText()[1:],
		}

	case *parser.MonetaryLiteralContext:
		return parseMonetaryLit(valueExprCtx.MonetaryLit())

	case *parser.AssetLiteralContext:
		return &AssetLiteral{
			Range: ctxToRange(valueExprCtx),
			Asset: valueExprCtx.GetText(),
		}

	case *parser.NumberLiteralContext:
		return parseNumberLiteral(valueExprCtx.NUMBER())

	case *parser.PercentagePortionLiteralContext:
		return parsePercentageRatio(valueExprCtx.GetText(), ctxToRange(valueExprCtx))

	case *parser.VariableExprContext:
		return variableLiteralFromCtx(valueExprCtx)

	case *parser.StringLiteralContext:
		return parseStringLiteralCtx(valueExprCtx)

	case *parser.InfixExprContext:
		return &BinaryInfix{
			Range:    ctxToRange(valueExprCtx),
			Operator: InfixOperator(valueExprCtx.GetOp().GetText()),
			Left:     parseValueExpr(valueExprCtx.GetLeft()),
			Right:    parseValueExpr(valueExprCtx.GetRight()),
		}

	case nil, *parser.ValueExprContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[ValueExpr](valueExprCtx.GetText())
	}
}

func variableLiteralFromCtx(ctx antlr.ParserRuleContext) *Variable {
	// Discard the '$'
	name := ctx.GetText()[1:]

	return &Variable{
		Range: ctxToRange(ctx),
		Name:  name,
	}
}

func parseDestination(destCtx parser.IDestinationContext) Destination {
	if destCtx == nil {
		return nil
	}

	range_ := ctxToRange(destCtx)

	switch destCtx := destCtx.(type) {
	case *parser.DestAccountContext:
		return &DestinationAccount{
			ValueExpr: parseValueExpr(destCtx.ValueExpr()),
		}

	case *parser.DestInorderContext:
		var inorderClauses []CappedKeptOrDestination
		for _, destInorderClause := range destCtx.AllDestinationInOrderClause() {
			inorderClauses = append(inorderClauses, parseDestinationInorderClause(destInorderClause))
		}

		return &DestinationInorder{
			Range:     range_,
			Clauses:   inorderClauses,
			Remaining: parseKeptOrDestination(destCtx.KeptOrDestination()),
		}

	case *parser.DestOneofContext:
		var inorderClauses []CappedKeptOrDestination
		for _, destInorderClause := range destCtx.AllDestinationInOrderClause() {
			inorderClauses = append(inorderClauses, parseDestinationInorderClause(destInorderClause))
		}

		return &DestinationOneof{
			Range:     range_,
			Clauses:   inorderClauses,
			Remaining: parseKeptOrDestination(destCtx.KeptOrDestination()),
		}

	case *parser.DestAllotmentContext:
		var items []DestinationAllotmentItem
		for _, itemCtx := range destCtx.AllAllotmentClauseDest() {
			item := DestinationAllotmentItem{
				Range:     ctxToRange(itemCtx),
				Allotment: parseDestinationAllotment(itemCtx.Allotment()),
				To:        parseKeptOrDestination(itemCtx.KeptOrDestination()),
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
		return utils.NonExhaustiveMatchPanic[Destination](destCtx.GetText())
	}

}

func parseDestinationInorderClause(clauseCtx parser.IDestinationInOrderClauseContext) CappedKeptOrDestination {
	return CappedKeptOrDestination{
		Range: ctxToRange(clauseCtx),
		Cap:   parseValueExpr(clauseCtx.ValueExpr()),
		To:    parseKeptOrDestination(clauseCtx.KeptOrDestination()),
	}
}

func parseKeptOrDestination(clauseCtx parser.IKeptOrDestinationContext) KeptOrDestination {
	if clauseCtx == nil {
		return nil
	}

	switch clauseCtx := clauseCtx.(type) {
	case *parser.DestinationToContext:
		return &DestinationTo{
			Destination: parseDestination(clauseCtx.Destination()),
		}
	case *parser.DestinationKeptContext:
		return &DestinationKept{
			Range: ctxToRange(clauseCtx),
		}
	case *parser.KeptOrDestinationContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[KeptOrDestination](clauseCtx.GetText())
	}

}

func parseDestinationAllotment(allotmentCtx parser.IAllotmentContext) AllotmentValue {
	switch allotmentCtx := allotmentCtx.(type) {
	case *parser.RemainingAllotmentContext:
		return &RemainingAllotment{
			Range: ctxToRange(allotmentCtx),
		}

	case *parser.PortionedAllotmentContext:
		expr := parseValueExpr(allotmentCtx.ValueExpr())
		return &ValueExprAllotment{Value: expr}

	case *parser.AllotmentContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[AllotmentValue](allotmentCtx.GetText())
	}
}

func parseFnArgs(fnCallArgCtx parser.IFunctionCallArgsContext) []ValueExpr {
	if fnCallArgCtx == nil {
		return nil
	}

	var args []ValueExpr
	for _, valueExpr := range fnCallArgCtx.AllValueExpr() {
		args = append(args, parseValueExpr(valueExpr))
	}
	return args
}

func parseFnCall(fnCallCtx parser.IFunctionCallContext) *FnCall {
	if fnCallCtx == nil {
		return nil
	}

	ident := fnCallCtx.GetFnName()
	if ident == nil {
		return nil
	}

	allArgs := fnCallCtx.FunctionCallArgs()

	return &FnCall{
		Range: ctxToRange(fnCallCtx),
		Caller: &FnCallIdentifier{
			Range: tokenToRange(ident),
			Name:  ident.GetText(),
		},
		Args: parseFnArgs(allArgs),
	}
}

func parseSaveStatement(saveCtx *parser.SaveStatementContext) *SaveStatement {
	return &SaveStatement{
		Range:     ctxToRange(saveCtx),
		SentValue: parseSentValue(saveCtx.SentValue()),
		Amount:    parseValueExpr(saveCtx.ValueExpr()),
	}
}

func parseStatement(statementCtx parser.IStatementContext) Statement {
	switch statementCtx := statementCtx.(type) {
	case *parser.SendStatementContext:
		return parseSendStatement(statementCtx)

	case *parser.SaveStatementContext:
		return parseSaveStatement(statementCtx)

	case *parser.FnCallStatementContext:
		return parseFnCall(statementCtx.FunctionCall())

	case *parser.StatementContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[Statement](statementCtx.GetText())
	}
}

func parseSentValue(statementCtx parser.ISentValueContext) SentValue {
	switch statementCtx := statementCtx.(type) {
	case *parser.SentLiteralContext:
		return &SentValueLiteral{
			Range:    ctxToRange(statementCtx),
			Monetary: parseValueExpr(statementCtx.ValueExpr()),
		}
	case *parser.SentAllContext:
		return &SentValueAll{
			Range: ctxToRange(statementCtx),
			Asset: parseValueExpr(statementCtx.SentAllLit().GetAsset()),
		}

	case *parser.SentValueContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[SentValue](statementCtx.GetText())
	}

}

func parseSendStatement(statementCtx *parser.SendStatementContext) *SendStatement {
	return &SendStatement{
		Source:      parseSource(statementCtx.Source()),
		Destination: parseDestination(statementCtx.Destination()),
		Range:       ctxToRange(statementCtx),
		SentValue:   parseSentValue(statementCtx.SentValue()),
	}
}

func parseNumberLiteral(numNode antlr.TerminalNode) *NumberLiteral {
	amtStr := numNode.GetText()
	amtStr = strings.ReplaceAll(amtStr, "_", "")

	amt, err := strconv.Atoi(amtStr)
	if err != nil {
		panic("Invalid number: " + amtStr)
	}

	return &NumberLiteral{
		Range:  tokenToRange(numNode.GetSymbol()),
		Number: amt,
	}
}

func parseMonetaryLit(monetaryLitCtx parser.IMonetaryLitContext) *MonetaryLiteral {
	if monetaryLitCtx.GetAmt() == nil {
		return nil
	}

	return &MonetaryLiteral{
		Range:  ctxToRange(monetaryLitCtx),
		Asset:  parseValueExpr(monetaryLitCtx.GetAsset()),
		Amount: parseValueExpr(monetaryLitCtx.GetAmt()),
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
