package parser

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	antlrParser "github.com/formancehq/numscript/internal/parser/antlrParser"
	"github.com/formancehq/numscript/internal/utils"

	"github.com/antlr4-go/antlr/v4"
)

type ParserError struct {
	Range
	Msg string
}

func (e ParserError) Error() string {
	return e.Msg
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
			parsed.Comments = append(parsed.Comments, Comment{
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

func parseVarsDeclaration(varsCtx antlrParser.IVarsDeclarationContext) *VarDeclarations {

	if varsCtx == nil {
		return nil
	}

	varBlock := VarDeclarations{
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

func parseProgram(programCtx antlrParser.IProgramContext) Program {
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

func parseVarDeclaration(varDecl antlrParser.IVarDeclarationContext) *VarDeclaration {
	if varDecl == nil {
		return nil
	}

	var origin *ValueExpr
	if varDecl.VarOrigin() != nil {
		expr := parseValueExpr(varDecl.VarOrigin().ValueExpr())
		origin = &expr
	}

	return &VarDeclaration{
		Range:  ctxToRange(varDecl),
		Type:   parseVarType(varDecl.GetType_()),
		Name:   parseVarLiteral(varDecl.GetName()),
		Origin: origin,
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

func parseColorConstraint(colorConstraintCtx antlrParser.IColorConstraintContext) ValueExpr {
	if colorConstraintCtx == nil {
		return nil
	}

	return parseValueExpr(
		colorConstraintCtx.ValueExpr(),
	)
}

func parseSource(sourceCtx antlrParser.ISourceContext) Source {
	if sourceCtx == nil {
		return nil
	}

	range_ := ctxToRange(sourceCtx)

	switch sourceCtx := sourceCtx.(type) {
	case *antlrParser.SrcAccountContext:
		return &SourceAccount{
			Color:     parseColorConstraint(sourceCtx.ColorConstraint()),
			ValueExpr: parseValueExpr(sourceCtx.ValueExpr()),
		}

	case *antlrParser.SrcInorderContext:
		var sources []Source
		for _, sourceCtx := range sourceCtx.AllSource() {
			sources = append(sources, parseSource(sourceCtx))
		}
		return &SourceInorder{
			Range:   range_,
			Sources: sources,
		}

	case *antlrParser.SrcOneofContext:
		var sources []Source
		for _, sourceCtx := range sourceCtx.AllSource() {
			sources = append(sources, parseSource(sourceCtx))
		}
		return &SourceOneof{
			Range:   range_,
			Sources: sources,
		}

	case *antlrParser.SrcAllotmentContext:
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

	case *antlrParser.SrcCappedContext:
		return &SourceCapped{
			Range: range_,
			From:  parseSource(sourceCtx.Source()),
			Cap:   parseValueExpr(sourceCtx.GetCap_()),
		}

	case *antlrParser.SrcAccountUnboundedOverdraftContext:
		return &SourceOverdraft{
			Range:   ctxToRange(sourceCtx),
			Color:   parseColorConstraint(sourceCtx.ColorConstraint()),
			Address: parseValueExpr(sourceCtx.GetAddress()),
		}

	case *antlrParser.SrcAccountBoundedOverdraftContext:
		varMon := parseValueExpr(sourceCtx.GetMaxOvedraft())

		return &SourceOverdraft{
			Range:   ctxToRange(sourceCtx),
			Color:   parseColorConstraint(sourceCtx.ColorConstraint()),
			Address: parseValueExpr(sourceCtx.GetAddress()),
			Bounded: &varMon,
		}

	case *antlrParser.SourceContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[Source](sourceCtx.GetText())
	}
}

func countLeadingZeros(str string) uint16 {
	var count uint16

loop:
	for _, ch := range str {
		switch ch {
		case '0':
			count++
		default:
			break loop
		}
	}

	return count
}

func ParsePercentageRatio(source string) (*big.Int, uint16, error) {
	str := strings.TrimSuffix(source, "%")
	num, ok := new(big.Int).SetString(strings.ReplaceAll(str, ".", ""), 10)
	if !ok {
		return nil, 0, fmt.Errorf("unepexcted invalid string literal: %s", source)
	}

	var floatingDigits uint16
	split := strings.Split(str, ".")

	if len(split) > 1 {
		floatingPart := split[1]
		floatingDigits = uint16(len(floatingPart)) + countLeadingZeros(floatingPart)
	} else {
		floatingDigits = 0
	}

	return num, floatingDigits, nil
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

func parseAllotment(allotmentCtx antlrParser.IAllotmentContext) AllotmentValue {
	switch allotmentCtx := allotmentCtx.(type) {
	case *antlrParser.PortionedAllotmentContext:
		expr := parseValueExpr(allotmentCtx.ValueExpr())
		return &ValueExprAllotment{expr}

	case *antlrParser.RemainingAllotmentContext:
		return &RemainingAllotment{
			Range: ctxToRange(allotmentCtx),
		}

	case *antlrParser.AllotmentContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[AllotmentValue](allotmentCtx.GetText())
	}
}

func parseStringLiteralCtx(stringCtx *antlrParser.StringLiteralContext) *StringLiteral {
	rawStr := stringCtx.GetText()
	// Remove leading and trailing '"'
	innerStr := rawStr[1 : len(rawStr)-1]
	return &StringLiteral{
		Range:  ctxToRange(stringCtx),
		String: innerStr,
	}
}

func parseValueExpr(valueExprCtx antlrParser.IValueExprContext) ValueExpr {
	switch valueExprCtx := valueExprCtx.(type) {
	case *antlrParser.ParenthesizedExprContext:
		return parseValueExpr(valueExprCtx.ValueExpr())

	case *antlrParser.AccountLiteralContext:
		litRng := ctxToRange(valueExprCtx)

		var parts []AccountNamePart
		for _, accLit := range valueExprCtx.AllAccountLiteralPart() {
			varPartText := accLit.GetText()
			switch accLit := accLit.(type) {
			case *antlrParser.AccountTextPartContext:
				parts = append(parts, AccountTextPart{Name: varPartText})
			case *antlrParser.AccountVarPartContext:
				v := parseVarLiteral(accLit.VARIABLE_NAME_ACC().GetSymbol())
				parts = append(parts, v)
			}
		}

		return &AccountInterpLiteral{
			Range: litRng,
			Parts: parts,
		}

	case *antlrParser.MonetaryLiteralContext:
		return parseMonetaryLit(valueExprCtx.MonetaryLit())

	case *antlrParser.AssetLiteralContext:
		return &AssetLiteral{
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
		return parseStringLiteralCtx(valueExprCtx)

	case *antlrParser.PrefixExprContext:
		return &Prefix{
			Range:    ctxToRange(valueExprCtx),
			Operator: PrefixOperator(valueExprCtx.GetOp().GetText()),
			Expr:     parseValueExpr(valueExprCtx.ValueExpr()),
		}

	case *antlrParser.InfixExprContext:
		return &BinaryInfix{
			Range:    ctxToRange(valueExprCtx),
			Operator: InfixOperator(valueExprCtx.GetOp().GetText()),
			Left:     parseValueExpr(valueExprCtx.GetLeft()),
			Right:    parseValueExpr(valueExprCtx.GetRight()),
		}

	case nil, *antlrParser.ValueExprContext:
		return nil

	case *antlrParser.ApplicationContext:
		return parseFnCall(valueExprCtx.FunctionCall())

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

func parseDestination(destCtx antlrParser.IDestinationContext) Destination {
	if destCtx == nil {
		return nil
	}

	range_ := ctxToRange(destCtx)

	switch destCtx := destCtx.(type) {
	case *antlrParser.DestAccountContext:
		return &DestinationAccount{
			ValueExpr: parseValueExpr(destCtx.ValueExpr()),
		}

	case *antlrParser.DestInorderContext:
		var inorderClauses []CappedKeptOrDestination
		for _, destInorderClause := range destCtx.AllDestinationInOrderClause() {
			inorderClauses = append(inorderClauses, parseDestinationInorderClause(destInorderClause))
		}

		return &DestinationInorder{
			Range:     range_,
			Clauses:   inorderClauses,
			Remaining: parseKeptOrDestination(destCtx.KeptOrDestination()),
		}

	case *antlrParser.DestOneofContext:
		var inorderClauses []CappedKeptOrDestination
		for _, destInorderClause := range destCtx.AllDestinationInOrderClause() {
			inorderClauses = append(inorderClauses, parseDestinationInorderClause(destInorderClause))
		}

		return &DestinationOneof{
			Range:     range_,
			Clauses:   inorderClauses,
			Remaining: parseKeptOrDestination(destCtx.KeptOrDestination()),
		}

	case *antlrParser.DestAllotmentContext:
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

	case *antlrParser.DestinationContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[Destination](destCtx.GetText())
	}

}

func parseDestinationInorderClause(clauseCtx antlrParser.IDestinationInOrderClauseContext) CappedKeptOrDestination {
	return CappedKeptOrDestination{
		Range: ctxToRange(clauseCtx),
		Cap:   parseValueExpr(clauseCtx.ValueExpr()),
		To:    parseKeptOrDestination(clauseCtx.KeptOrDestination()),
	}
}

func parseKeptOrDestination(clauseCtx antlrParser.IKeptOrDestinationContext) KeptOrDestination {
	if clauseCtx == nil {
		return nil
	}

	switch clauseCtx := clauseCtx.(type) {
	case *antlrParser.DestinationToContext:
		return &DestinationTo{
			Destination: parseDestination(clauseCtx.Destination()),
		}
	case *antlrParser.DestinationKeptContext:
		return &DestinationKept{
			Range: ctxToRange(clauseCtx),
		}
	case *antlrParser.KeptOrDestinationContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[KeptOrDestination](clauseCtx.GetText())
	}

}

func parseDestinationAllotment(allotmentCtx antlrParser.IAllotmentContext) AllotmentValue {
	switch allotmentCtx := allotmentCtx.(type) {
	case *antlrParser.RemainingAllotmentContext:
		return &RemainingAllotment{
			Range: ctxToRange(allotmentCtx),
		}

	case *antlrParser.PortionedAllotmentContext:
		expr := parseValueExpr(allotmentCtx.ValueExpr())
		return &ValueExprAllotment{Value: expr}

	case *antlrParser.AllotmentContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[AllotmentValue](allotmentCtx.GetText())
	}
}

func parseFnArgs(fnCallArgCtx antlrParser.IFunctionCallArgsContext) []ValueExpr {
	if fnCallArgCtx == nil {
		return nil
	}

	var args []ValueExpr
	for _, valueExpr := range fnCallArgCtx.AllValueExpr() {
		args = append(args, parseValueExpr(valueExpr))
	}
	return args
}

func parseFnCall(fnCallCtx antlrParser.IFunctionCallContext) *FnCall {
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

func parseSaveStatement(saveCtx *antlrParser.SaveStatementContext) *SaveStatement {
	return &SaveStatement{
		Range:     ctxToRange(saveCtx),
		SentValue: parseSentValue(saveCtx.SentValue()),
		Amount:    parseValueExpr(saveCtx.ValueExpr()),
	}
}

func parseStatement(statementCtx antlrParser.IStatementContext) Statement {
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
		return utils.NonExhaustiveMatchPanic[Statement](statementCtx.GetText())
	}
}

func parseSentValue(statementCtx antlrParser.ISentValueContext) SentValue {
	switch statementCtx := statementCtx.(type) {
	case *antlrParser.SentLiteralContext:
		return &SentValueLiteral{
			Range:    ctxToRange(statementCtx),
			Monetary: parseValueExpr(statementCtx.ValueExpr()),
		}
	case *antlrParser.SentAllContext:
		return &SentValueAll{
			Range: ctxToRange(statementCtx),
			Asset: parseValueExpr(statementCtx.SentAllLit().GetAsset()),
		}

	case *antlrParser.SentValueContext:
		return nil

	default:
		return utils.NonExhaustiveMatchPanic[SentValue](statementCtx.GetText())
	}

}

func parseSendStatement(statementCtx *antlrParser.SendStatementContext) *SendStatement {
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

	amt, ok := new(big.Int).SetString(amtStr, 10)
	if !ok {
		panic("Invalid number: " + amtStr)
	}

	return &NumberLiteral{
		Range:  tokenToRange(numNode.GetSymbol()),
		Number: amt,
	}
}

func parseMonetaryLit(monetaryLitCtx antlrParser.IMonetaryLitContext) *MonetaryLiteral {
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
