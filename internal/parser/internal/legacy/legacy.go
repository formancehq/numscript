package legacy

import (
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/parser"
	antlrParser "github.com/formancehq/numscript/internal/parser/internal/legacy/antlrParser"
	"github.com/formancehq/numscript/internal/utils"

	"github.com/antlr4-go/antlr/v4"
)

type ParserError struct {
	parser.Range
	Msg string
}

func (e ParserError) Error() string {
	return e.Msg
}

type ParseResult struct {
	Source string
	Value  parser.Program
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
		Range: parser.Range{
			Start: parser.Position{Character: startC, Line: startL - 1},
			End:   parser.Position{Character: endC, Line: endL - 1},
		},
	})
}

func Parse(input string) ParseResult {
	// TODO handle lexer errors
	listener := &ErrorListener{}

	is := antlr.NewInputStream(input)
	lexer := antlrParser.NewLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := antlrParser.NewNumscriptParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(listener)

	parsed := parseProgram(p.Program())

	for _, tk := range stream.GetAllTokens() {
		if tk.GetChannel() == antlr.TokenHiddenChannel {
			parsed.Comments = append(parsed.Comments, parser.Comment{
				Content: tk.GetText()[2:],
				Range:   tokenToRange(tk),
			})
		}
	}

	return ParseResult{
		Source: input,
		Value:  parsed,
		Errors: listener.Errors,
	}
}

func ParseErrorsToString(errors []ParserError, source string) string {
	buf := "Got errors while parsing:\n"
	for _, err := range errors {
		buf += err.Msg + "\n" + err.ShowOnSource(source) + "\n"
	}
	return buf
}

func parseVarsDeclaration(varsCtx antlrParser.IVarsDeclarationContext) *parser.VarDeclarations {

	if varsCtx == nil {
		return nil
	}

	varBlock := parser.VarDeclarations{
		Range: ctxToRange(varsCtx),
	}

	for _, varDecl := range varsCtx.AllVarDeclaration() {
		decl := parseVarDeclaration(varDecl)
		if decl != nil {
			varBlock.Declarations = append(varBlock.Declarations, *decl)
		}
	}
	return &varBlock
}

func parseProgram(programCtx antlrParser.IProgramContext) parser.Program {
	vars := parseVarsDeclaration(programCtx.VarsDeclaration())

	featureCtx := programCtx.FeatureDecl()

	var flagIdents []parser.StringLiteral
	if featureCtx != nil {
		for _, flagTk := range featureCtx.GetFlag() {
			flagIdents = append(flagIdents, *parseStringTk(flagTk))
		}
	}

	var statements []parser.Statement
	for _, statementCtx := range programCtx.AllStatement() {
		statements = append(statements, parseStatement(statementCtx))
	}

	return parser.Program{
		Flags:      flagIdents,
		Statements: statements,
		Vars:       vars,
	}
}

func parseVarDeclaration(varDecl antlrParser.IVarDeclarationContext) *parser.VarDeclaration {
	if varDecl == nil {
		return nil
	}

	var origin *parser.ValueExpr
	if varDecl.VarOrigin() != nil {
		expr := parseValueExpr(varDecl.VarOrigin().ValueExpr())
		// when the origin expression is missing or malformed,
		// parseValueExpr returns nil; do not wrap it in a non-nil pointer
		if expr != nil {
			origin = &expr
		}
	}

	return &parser.VarDeclaration{
		Range:  ctxToRange(varDecl),
		Type:   parseVarType(varDecl.GetType_()),
		Name:   parseVarLiteral(varDecl.GetName()),
		Origin: origin,
	}
}

func parseVarLiteral(tk antlr.Token) *parser.Variable {
	if tk == nil || tk.GetTokenIndex() == -1 {
		return nil
	}

	name := tk.GetText()[1:]

	return &parser.Variable{
		Range: tokenToRange(tk),
		Name:  name,
	}
}

func parseVarType(tk antlr.Token) *parser.TypeDecl {
	if tk == nil || tk.GetTokenIndex() == -1 {
		return nil
	}

	return &parser.TypeDecl{
		Range: tokenToRange(tk),
		Name:  tk.GetText(),
	}
}

func parseColorConstraint(colorConstraintCtx antlrParser.IColorConstraintContext) parser.ValueExpr {
	if colorConstraintCtx == nil {
		return nil
	}

	return parseValueExpr(
		colorConstraintCtx.ValueExpr(),
	)
}

func parseSource(sourceCtx antlrParser.ISourceContext) parser.Source {
	if sourceCtx == nil {
		return nil
	}

	range_ := ctxToRange(sourceCtx)

	switch sourceCtx := sourceCtx.(type) {
	case *antlrParser.SrcAccountContext:
		return &parser.SourceAccount{
			Range:     range_,
			Color:     parseColorConstraint(sourceCtx.ColorConstraint()),
			ValueExpr: parseValueExpr(sourceCtx.ValueExpr()),
		}

	case *antlrParser.SrcInorderContext:
		var sources []parser.Source
		for _, sourceCtx := range sourceCtx.AllSource() {
			sources = append(sources, parseSource(sourceCtx))
		}
		return &parser.SourceInorder{
			Range:   range_,
			Sources: sources,
		}

	case *antlrParser.SrcOneofContext:
		var sources []parser.Source
		for _, sourceCtx := range sourceCtx.AllSource() {
			sources = append(sources, parseSource(sourceCtx))
		}
		return &parser.SourceOneof{
			Range:   range_,
			Sources: sources,
		}

	case *antlrParser.SrcAllotmentContext:
		var items []parser.SourceAllotmentItem
		for _, itemCtx := range sourceCtx.AllAllotmentClauseSrc() {
			item := parser.SourceAllotmentItem{
				Range:     ctxToRange(itemCtx),
				Allotment: parseAllotment(itemCtx.Allotment()),
				From:      parseSource(itemCtx.Source()),
			}
			items = append(items, item)
		}
		return &parser.SourceAllotment{
			Range: range_,
			Items: items,
		}

	case *antlrParser.SrcCappedContext:
		return &parser.SourceCapped{
			Range: range_,
			From:  parseSource(sourceCtx.Source()),
			Cap:   parseValueExpr(sourceCtx.GetCap_()),
		}

	case *antlrParser.SrcAccountUnboundedOverdraftContext:
		return &parser.SourceOverdraft{
			Range:   ctxToRange(sourceCtx),
			Color:   parseColorConstraint(sourceCtx.ColorConstraint()),
			Address: parseValueExpr(sourceCtx.GetAddress()),
		}

	case *antlrParser.SrcAccountWithScalingContext:
		return &parser.SourceWithScaling{
			Range:   ctxToRange(sourceCtx),
			Address: parseValueExpr(sourceCtx.GetAddress()),
			Through: parseValueExpr(sourceCtx.GetSwap()),
		}

	case *antlrParser.SrcAccountBoundedOverdraftContext:
		varMon := parseValueExpr(sourceCtx.GetMaxOvedraft())

		return &parser.SourceOverdraft{
			Range:   ctxToRange(sourceCtx),
			Color:   parseColorConstraint(sourceCtx.ColorConstraint()),
			Address: parseValueExpr(sourceCtx.GetAddress()),
			Bounded: &varMon,
		}

	case *antlrParser.SourceContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[parser.Source](sourceCtx.GetText())
	}
}

func parsePercentageRatio(source string, range_ parser.Range) *parser.PercentageLiteral {
	num, floatingDigits, err := parser.ParsePercentageRatio(source)
	if err != nil {
		// This should be unreachable for tokens emitted by the lexer.
		// Degrade gracefully to a zero-value literal instead of panicking,
		// so that a malformed token cannot crash the process (e.g. the LSP server).
		return &parser.PercentageLiteral{
			Range:  range_,
			Amount: big.NewInt(0),
		}
	}

	return &parser.PercentageLiteral{
		Range:          range_,
		Amount:         num,
		FloatingDigits: floatingDigits,
	}
}

func parseAllotment(allotmentCtx antlrParser.IAllotmentContext) parser.AllotmentValue {
	switch allotmentCtx := allotmentCtx.(type) {
	case *antlrParser.PortionedAllotmentContext:
		expr := parseValueExpr(allotmentCtx.ValueExpr())
		return &parser.ValueExprAllotment{Value: expr}

	case *antlrParser.RemainingAllotmentContext:
		return &parser.RemainingAllotment{
			Range: ctxToRange(allotmentCtx),
		}

	case *antlrParser.AllotmentContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[parser.AllotmentValue](allotmentCtx.GetText())
	}
}

func parseStringTk(stringCtx antlr.Token) *parser.StringLiteral {
	rawStr := stringCtx.GetText()
	// Remove leading and trailing '"'
	innerStr := rawStr[1 : len(rawStr)-1]
	return &parser.StringLiteral{
		Range:  tokenToRange(stringCtx),
		String: innerStr,
	}
}

func parseValueExpr(valueExprCtx antlrParser.IValueExprContext) parser.ValueExpr {
	switch valueExprCtx := valueExprCtx.(type) {
	case *antlrParser.ParenthesizedExprContext:
		return parseValueExpr(valueExprCtx.ValueExpr())

	case *antlrParser.AccountLiteralContext:
		litRng := ctxToRange(valueExprCtx)

		var parts []parser.AccountNamePart
		for _, accLit := range valueExprCtx.GetChildren() {
			switch accLit := accLit.(type) {
			case *antlrParser.AccountTextPartContext:
				varPartText := accLit.GetText()
				parts = append(parts, parser.AccountTextPart{Name: varPartText})
			case *antlrParser.AccountVarPartContext:
				v := parseVarLiteral(accLit.VARIABLE_NAME_ACC().GetSymbol())
				parts = append(parts, v)

			case antlr.TerminalNode:
				if accLit.GetSymbol().GetTokenType() == antlrParser.NumscriptParserCOLON {
					parts = append(parts, parser.AccountTextPart{Name: ":"})
				}
			}

		}

		return &parser.AccountInterpLiteral{
			Range: litRng,
			Parts: parts,
		}

	case *antlrParser.MonetaryLiteralContext:
		return parseMonetaryLit(valueExprCtx.MonetaryLit())

	case *antlrParser.AssetLiteralContext:
		return &parser.AssetLiteral{
			Range: ctxToRange(valueExprCtx),
			Asset: valueExprCtx.GetText(),
		}

	case *antlrParser.NumberLiteralContext:
		return parseNumberLiteral(valueExprCtx.NUMBER())

	case *antlrParser.PercentagePortionLiteralContext:
		return parsePercentageRatio(valueExprCtx.GetText(), ctxToRange(valueExprCtx))

	case *antlrParser.VariableExprContext:
		return variableLiteralFromCtx(valueExprCtx)

	case *antlrParser.StringLiteralContext:
		return parseStringTk(valueExprCtx.STRING().GetSymbol())

	case *antlrParser.PrefixExprContext:
		return &parser.Prefix{
			Range:    ctxToRange(valueExprCtx),
			Operator: parser.PrefixOperator(valueExprCtx.GetOp().GetText()),
			Expr:     parseValueExpr(valueExprCtx.ValueExpr()),
		}

	case *antlrParser.InfixExprContext:
		return &parser.BinaryInfix{
			Range:    ctxToRange(valueExprCtx),
			Operator: parser.InfixOperator(valueExprCtx.GetOp().GetText()),
			Left:     parseValueExpr(valueExprCtx.GetLeft()),
			Right:    parseValueExpr(valueExprCtx.GetRight()),
		}

	case nil, *antlrParser.ValueExprContext:
		return nil

	case *antlrParser.ApplicationContext:
		return parseFnCall(valueExprCtx.FunctionCall())

	default:
		return utils.NonExhaustiveMatchPanic[parser.ValueExpr](valueExprCtx.GetText())
	}
}

func variableLiteralFromCtx(ctx antlr.ParserRuleContext) *parser.Variable {
	// Discard the '$'
	name := ctx.GetText()[1:]

	return &parser.Variable{
		Range: ctxToRange(ctx),
		Name:  name,
	}
}

func parseDestination(destCtx antlrParser.IDestinationContext) parser.Destination {
	if destCtx == nil {
		return nil
	}

	range_ := ctxToRange(destCtx)

	switch destCtx := destCtx.(type) {
	case *antlrParser.DestAccountContext:
		return &parser.DestinationAccount{
			ValueExpr: parseValueExpr(destCtx.ValueExpr()),
		}

	case *antlrParser.DestInorderContext:
		var inorderClauses []parser.CappedKeptOrDestination
		for _, destInorderClause := range destCtx.AllDestinationInOrderClause() {
			inorderClauses = append(inorderClauses, parseDestinationInorderClause(destInorderClause))
		}

		return &parser.DestinationInorder{
			Range:     range_,
			Clauses:   inorderClauses,
			Remaining: parseKeptOrDestination(destCtx.KeptOrDestination()),
		}

	case *antlrParser.DestOneofContext:
		var inorderClauses []parser.CappedKeptOrDestination
		for _, destInorderClause := range destCtx.AllDestinationInOrderClause() {
			inorderClauses = append(inorderClauses, parseDestinationInorderClause(destInorderClause))
		}

		return &parser.DestinationOneof{
			Range:     range_,
			Clauses:   inorderClauses,
			Remaining: parseKeptOrDestination(destCtx.KeptOrDestination()),
		}

	case *antlrParser.DestAllotmentContext:
		var items []parser.DestinationAllotmentItem
		for _, itemCtx := range destCtx.AllAllotmentClauseDest() {
			item := parser.DestinationAllotmentItem{
				Range:     ctxToRange(itemCtx),
				Allotment: parseDestinationAllotment(itemCtx.Allotment()),
				To:        parseKeptOrDestination(itemCtx.KeptOrDestination()),
			}
			items = append(items, item)
		}
		return &parser.DestinationAllotment{
			Range: range_,
			Items: items,
		}

	case *antlrParser.DestinationContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[parser.Destination](destCtx.GetText())
	}

}

func parseDestinationInorderClause(clauseCtx antlrParser.IDestinationInOrderClauseContext) parser.CappedKeptOrDestination {
	return parser.CappedKeptOrDestination{
		Range: ctxToRange(clauseCtx),
		Cap:   parseValueExpr(clauseCtx.ValueExpr()),
		To:    parseKeptOrDestination(clauseCtx.KeptOrDestination()),
	}
}

func parseKeptOrDestination(clauseCtx antlrParser.IKeptOrDestinationContext) parser.KeptOrDestination {
	if clauseCtx == nil {
		return nil
	}

	switch clauseCtx := clauseCtx.(type) {
	case *antlrParser.DestinationToContext:
		return &parser.DestinationTo{
			Destination: parseDestination(clauseCtx.Destination()),
		}
	case *antlrParser.DestinationKeptContext:
		return &parser.DestinationKept{
			Range: ctxToRange(clauseCtx),
		}
	case *antlrParser.KeptOrDestinationContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[parser.KeptOrDestination](clauseCtx.GetText())
	}

}

func parseDestinationAllotment(allotmentCtx antlrParser.IAllotmentContext) parser.AllotmentValue {
	switch allotmentCtx := allotmentCtx.(type) {
	case *antlrParser.RemainingAllotmentContext:
		return &parser.RemainingAllotment{
			Range: ctxToRange(allotmentCtx),
		}

	case *antlrParser.PortionedAllotmentContext:
		expr := parseValueExpr(allotmentCtx.ValueExpr())
		return &parser.ValueExprAllotment{Value: expr}

	case *antlrParser.AllotmentContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[parser.AllotmentValue](allotmentCtx.GetText())
	}
}

func parseFnArgs(fnCallArgCtx antlrParser.IFunctionCallArgsContext) []parser.ValueExpr {
	if fnCallArgCtx == nil {
		return nil
	}

	var args []parser.ValueExpr
	for _, valueExpr := range fnCallArgCtx.AllValueExpr() {
		args = append(args, parseValueExpr(valueExpr))
	}
	return args
}

func parseFnCall(fnCallCtx antlrParser.IFunctionCallContext) *parser.FnCall {
	if fnCallCtx == nil {
		return nil
	}

	ident := fnCallCtx.GetFnName()
	if ident == nil {
		return nil
	}

	allArgs := fnCallCtx.FunctionCallArgs()

	return &parser.FnCall{
		Range: ctxToRange(fnCallCtx),
		Caller: &parser.FnCallIdentifier{
			Range: tokenToRange(ident),
			Name:  ident.GetText(),
		},
		Args: parseFnArgs(allArgs),
	}
}

func parseSaveStatement(saveCtx *antlrParser.SaveStatementContext) *parser.SaveStatement {
	return &parser.SaveStatement{
		Range:     ctxToRange(saveCtx),
		SentValue: parseSentValue(saveCtx.SentValue()),
		Account:   parseValueExpr(saveCtx.ValueExpr()),
	}
}

func parseStatement(statementCtx antlrParser.IStatementContext) parser.Statement {
	switch statementCtx := statementCtx.(type) {
	case *antlrParser.SendStatementContext:
		return parseSendStatement(statementCtx)

	case *antlrParser.SaveStatementContext:
		return parseSaveStatement(statementCtx)

	case *antlrParser.FnCallStatementContext:
		return parseFnCall(statementCtx.FunctionCall())

	case *antlrParser.StatementContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[parser.Statement](statementCtx.GetText())
	}
}

func parseSentValue(statementCtx antlrParser.ISentValueContext) parser.SentValue {
	switch statementCtx := statementCtx.(type) {
	case *antlrParser.SentLiteralContext:
		return &parser.SentValueLiteral{
			Range:    ctxToRange(statementCtx),
			Monetary: parseValueExpr(statementCtx.ValueExpr()),
		}
	case *antlrParser.SentAllContext:
		return &parser.SentValueAll{
			Range: ctxToRange(statementCtx),
			Asset: parseValueExpr(statementCtx.SentAllLit().GetAsset()),
		}

	case *antlrParser.SentValueContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[parser.SentValue](statementCtx.GetText())
	}

}

func parseSendStatement(statementCtx *antlrParser.SendStatementContext) *parser.SendStatement {
	return &parser.SendStatement{
		Source:      parseSource(statementCtx.Source()),
		Destination: parseDestination(statementCtx.Destination()),
		Range:       ctxToRange(statementCtx),
		SentValue:   parseSentValue(statementCtx.SentValue()),
	}
}

func parseNumberLiteral(numNode antlr.TerminalNode) *parser.NumberLiteral {
	amtStr := numNode.GetText()
	amtStr = strings.ReplaceAll(amtStr, "_", "")

	amt, ok := new(big.Int).SetString(amtStr, 10)
	if !ok {
		// This should be unreachable for tokens emitted by the lexer.
		// Degrade gracefully to a zero-value literal instead of panicking,
		// so that a malformed token cannot crash the process (e.g. the LSP server).
		amt = big.NewInt(0)
	}

	return &parser.NumberLiteral{
		Range:  tokenToRange(numNode.GetSymbol()),
		Number: amt,
	}
}

func parseMonetaryLit(monetaryLitCtx antlrParser.IMonetaryLitContext) *parser.MonetaryLiteral {
	if monetaryLitCtx.GetAmt() == nil {
		return nil
	}

	return &parser.MonetaryLiteral{
		Range:  ctxToRange(monetaryLitCtx),
		Asset:  parseValueExpr(monetaryLitCtx.GetAsset()),
		Amount: parseValueExpr(monetaryLitCtx.GetAmt()),
	}
}

func ctxToRange(ctx antlr.ParserRuleContext) parser.Range {
	startTk := ctx.GetStart()
	endTk := ctx.GetStop()

	return parser.Range{
		Start: parser.Position{
			Line:      startTk.GetLine() - 1,
			Character: startTk.GetColumn(),
		},
		End: parser.Position{
			Line: endTk.GetLine() - 1,

			// this is based on the assumption that a token cannot span multiple lines
			Character: endTk.GetColumn() + len(endTk.GetText()),
		},
	}
}

func tokenToRange(tk antlr.Token) parser.Range {
	return parser.Range{
		Start: parser.Position{
			Line:      tk.GetLine() - 1,
			Character: tk.GetColumn(),
		},
		End: parser.Position{
			Line:      tk.GetLine() - 1,
			Character: tk.GetColumn() + len(tk.GetText()),
		},
	}
}
