package syntax

import (
	"strconv"

	"github.com/formancehq/numscript/internal/parser"

	"github.com/antlr4-go/antlr/v4"
	antlrParser "github.com/formancehq/numscript/internal/ir/internal/syntax/antlrParser"
)

// ParserError is a parse error with range information.
type ParserError struct {
	Range parser.Range
	Msg   string
}

func (e ParserError) Error() string {
	return e.Msg
}

type ParseResult struct {
	Value  Program
	Errors []ParserError
}

type errorListener struct {
	antlr.DefaultErrorListener
	Errors []ParserError
}

func (l *errorListener) SyntaxError(_ antlr.Recognizer, offendingSymbol any, startL, startC int, msg string, _ antlr.RecognitionException) {
	length := 1
	if token, ok := offendingSymbol.(antlr.Token); ok {
		length = len(token.GetText())
	}
	endL := startL
	endC := startC + length - 1
	l.Errors = append(l.Errors, ParserError{
		Msg: msg,
		Range: parser.Range{
			Start: parser.Position{Character: startC, Line: startL - 1},
			End:   parser.Position{Character: endC, Line: endL - 1},
		},
	})
}

// Parse parses an IR textual program and returns the AST.
func Parse(input string) ParseResult {
	listener := &errorListener{}

	is := antlr.NewInputStream(input)
	lexer := antlrParser.NewIRLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := antlrParser.NewIRParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(listener)

	tree := p.Program()

	// On a syntax error ANTLR's recovery leaves partial nodes behind (a call
	// without its parens, an assignment without its rhs). Walking those means
	// dereferencing tokens that were never matched, so don't: the errors are
	// what the caller needs anyway.
	if len(listener.Errors) > 0 {
		return ParseResult{
			Errors: listener.Errors,
		}
	}

	return ParseResult{
		Value: buildAST(tree),
	}
}

func tokenToRange(tok antlr.Token) parser.Range {
	startL := tok.GetLine() - 1
	startC := tok.GetColumn()
	endC := startC + len(tok.GetText()) - 1
	return parser.Range{
		Start: parser.Position{Character: startC, Line: startL},
		End:   parser.Position{Character: endC, Line: startL},
	}
}

// ---- AST builder (walks the ANTLR parse tree) ----

func buildAST(tree antlrParser.IProgramContext) Program {
	if tree == nil {
		return Program{}
	}
	lines := tree.AllLine()
	stmts := make([]Stmt, 0, len(lines))
	for _, l := range lines {
		if s := buildStmt(l); s != nil {
			stmts = append(stmts, s)
		}
	}
	return Program{Stmts: stmts}
}

func buildStmt(ctx antlrParser.ILineContext) Stmt {
	if ctx == nil {
		return nil
	}

	// label marker
	if lm := ctx.LabelMarker(); lm != nil {
		return buildLabelMarker(lm)
	}

	// instruction
	if instr := ctx.Instruction(); instr != nil {
		return buildInstruction(instr)
	}

	return nil
}

func buildLabelMarker(ctx antlrParser.ILabelMarkerContext) *LabelStmt {
	tok := ctx.LABEL().GetSymbol()
	name := tok.GetText()[1:] // strip '#'
	return &LabelStmt{
		Range: tokenToRange(tok),
		Name:  name,
	}
}

func buildInstruction(ctx antlrParser.IInstructionContext) *InstrStmt {
	switch c := ctx.(type) {
	case *antlrParser.InstrWithDestContext:
		return buildInstrWithDest(c)
	case *antlrParser.InstrNoDestContext:
		return buildInstrNoDest(c)
	case *antlrParser.ConstAssignContext:
		return buildConstAssign(c)
	case *antlrParser.InfixInstrContext:
		return buildInfixInstr(c)
	case *antlrParser.CompoundAssignInstrContext:
		return buildCompoundAssignInstr(c)
	}
	return nil
}

func buildInstrWithDest(ctx *antlrParser.InstrWithDestContext) *InstrStmt {
	dest := buildDest(ctx.Dest())
	call := buildInstrCall(ctx.InstrCall())
	rng := mergeRanges(dest.Range, call.Range)
	return &InstrStmt{
		Range: rng,
		Dest:  &dest,
		Call:  &call,
	}
}

func buildInstrNoDest(ctx *antlrParser.InstrNoDestContext) *InstrStmt {
	call := buildInstrCall(ctx.InstrCall())
	return &InstrStmt{
		Range: call.Range,
		Call:  &call,
	}
}

func buildConstAssign(ctx *antlrParser.ConstAssignContext) *InstrStmt {
	dest := buildDest(ctx.Dest())
	c := buildConst(ctx.Const_())
	rng := mergeRanges(dest.Range, c.Range)
	return &InstrStmt{
		Range: rng,
		Dest:  &dest,
		Const: &c,
	}
}

func buildInfixInstr(ctx *antlrParser.InfixInstrContext) *InstrStmt {
	dest := buildDest(ctx.Dest())
	left := buildRegRef(ctx.GetLeft())
	right := buildRegRef(ctx.GetRight())
	op := ctx.GetOp().GetText()
	rng := mergeRanges(dest.Range, right.Range)
	return &InstrStmt{
		Range: rng,
		Dest:  &dest,
		Infix: &Infix{
			Range: rng,
			Op:    op,
			Left:  left,
			Right: right,
		},
	}
}

func buildCompoundAssignInstr(ctx *antlrParser.CompoundAssignInstrContext) *InstrStmt {
	left := buildRegRef(ctx.GetLeft())
	right := buildRegRef(ctx.GetRight())
	op := ctx.GetOp().GetText()
	// strip the trailing '='
	infixOp := op[:len(op)-1]
	rng := mergeRanges(left.Range, right.Range)
	return &InstrStmt{
		Range: rng,
		Dest:  &Dest{Kind: DestReg, Regs: []RegRef{left}, Range: left.Range},
		CompoundAssign: &Infix{
			Range: rng,
			Op:    infixOp,
			Left:  left,
			Right: right,
		},
	}
}

func buildDest(ctx antlrParser.IDestContext) Dest {
	switch d := ctx.(type) {
	case *antlrParser.DestRegContext:
		reg := buildRegRef(d.Reg())
		return Dest{Kind: DestReg, Regs: []RegRef{reg}, Range: reg.Range}
	case *antlrParser.DestDiscardContext:
		tok := d.UNDERSCORE().GetSymbol()
		return Dest{Kind: DestDiscard, Range: tokenToRange(tok)}
	case *antlrParser.DestListContext:
		regs := buildRegList(d.RegList())
		if len(regs) == 0 {
			return Dest{Kind: DestList, Range: tokenToRange(d.LBRACKET().GetSymbol())}
		}
		rng := mergeRanges(
			tokenToRange(d.LBRACKET().GetSymbol()),
			tokenToRange(d.RBRACKET().GetSymbol()),
		)
		return Dest{Kind: DestList, Regs: regs, Range: rng}
	}
	return Dest{}
}

func buildRegList(ctx antlrParser.IRegListContext) []RegRef {
	if ctx == nil {
		return nil
	}
	allRegs := ctx.AllReg()
	regs := make([]RegRef, len(allRegs))
	for i, r := range allRegs {
		regs[i] = buildRegRef(r)
	}
	return regs
}

func buildInstrCall(ctx antlrParser.IInstrCallContext) InstrCall {
	nameCtx := ctx.InstrName()
	name, typeParam := buildInstrName(nameCtx)
	args := buildArgs(ctx.Args())

	rngStart := tokenToRange(ctx.LPAREN().GetSymbol())
	rngEnd := tokenToRange(ctx.RPAREN().GetSymbol())

	return InstrCall{
		Range:     mergeRanges(rngStart, rngEnd),
		Name:      name,
		TypeParam: typeParam,
		Args:      args,
	}
}

func buildInstrName(ctx antlrParser.IInstrNameContext) (name string, typeParam string) {
	name = ctx.IDENTIFIER().GetText()
	if tn := ctx.TypeName(); tn != nil {
		// a type name is either a TYPE_KEYWORD or a plain IDENTIFIER (see IR.g4)
		typeParam = tn.GetText()
	}
	return
}

func buildArgs(ctx antlrParser.IArgsContext) []Arg {
	if ctx == nil {
		return nil
	}
	allArgs := ctx.AllArg()
	args := make([]Arg, len(allArgs))
	for i, a := range allArgs {
		args[i] = buildArg(a)
	}
	return args
}

func buildArg(ctx antlrParser.IArgContext) Arg {
	switch a := ctx.(type) {
	case *antlrParser.PositionalArgContext:
		val := buildValue(a.Value())
		return Arg{Range: val.Range, Value: val}
	case *antlrParser.LabeledArgContext:
		label := a.IDENTIFIER().GetText()
		val := buildValue(a.Value())
		rng := mergeRanges(tokenToRange(a.IDENTIFIER().GetSymbol()), val.Range)
		return Arg{Range: rng, Label: label, Value: val}
	}
	return Arg{}
}

func buildValue(ctx antlrParser.IValueContext) Value {
	switch v := ctx.(type) {
	case *antlrParser.ValRegContext:
		reg := buildRegRef(v.Reg())
		return Value{Range: reg.Range, Kind: ValReg, Reg: &reg}
	case *antlrParser.ValLabelContext:
		tok := v.LABEL().GetSymbol()
		name := tok.GetText()[1:] // strip '#'
		return Value{Range: tokenToRange(tok), Kind: ValLabel, Label: &name}
	case *antlrParser.ValIntContext:
		tok := v.INT().GetSymbol()
		s := tok.GetText()
		return Value{Range: tokenToRange(tok), Kind: ValInt, Int: &s}
	case *antlrParser.ValRegListContext:
		regs := buildRegList(v.RegList())
		lTok := v.LBRACKET().GetSymbol()
		rTok := v.RBRACKET().GetSymbol()
		rng := mergeRanges(tokenToRange(lTok), tokenToRange(rTok))
		return Value{Range: rng, Kind: ValRegList, Regs: &regs}
	}
	return Value{}
}

func buildConst(ctx antlrParser.IConst_Context) Const {
	switch c := ctx.(type) {
	case *antlrParser.ConstStringContext:
		tok := c.STRING().GetSymbol()
		raw := tok.GetText()
		// strip surrounding quotes
		s, err := strconv.Unquote(raw)
		if err != nil {
			s = raw[1 : len(raw)-1]
		}
		return Const{Range: tokenToRange(tok), Kind: ConstString, StrVal: &s}
	case *antlrParser.ConstIntContext:
		tok := c.INT().GetSymbol()
		s := tok.GetText()
		return Const{Range: tokenToRange(tok), Kind: ConstInt, IntVal: &s}
	}
	return Const{}
}

func buildRegRef(ctx antlrParser.IRegContext) RegRef {
	tok := ctx.REG().GetSymbol()
	name := tok.GetText()
	return RegRef{Range: tokenToRange(tok), Name: name}
}

func mergeRanges(a, b parser.Range) parser.Range {
	return parser.Range{Start: a.Start, End: b.End}
}
