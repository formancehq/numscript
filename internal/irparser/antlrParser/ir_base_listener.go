// Code generated from IR.g4 by ANTLR 4.13.2. DO NOT EDIT.

package antlrParser // IR
import "github.com/antlr4-go/antlr/v4"

// BaseIRListener is a complete listener for a parse tree produced by IRParser.
type BaseIRListener struct{}

var _ IRListener = &BaseIRListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseIRListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseIRListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseIRListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseIRListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterProgram is called when production program is entered.
func (s *BaseIRListener) EnterProgram(ctx *ProgramContext) {}

// ExitProgram is called when production program is exited.
func (s *BaseIRListener) ExitProgram(ctx *ProgramContext) {}

// EnterLine is called when production line is entered.
func (s *BaseIRListener) EnterLine(ctx *LineContext) {}

// ExitLine is called when production line is exited.
func (s *BaseIRListener) ExitLine(ctx *LineContext) {}

// EnterLabelMarker is called when production labelMarker is entered.
func (s *BaseIRListener) EnterLabelMarker(ctx *LabelMarkerContext) {}

// ExitLabelMarker is called when production labelMarker is exited.
func (s *BaseIRListener) ExitLabelMarker(ctx *LabelMarkerContext) {}

// EnterInstrWithDest is called when production instrWithDest is entered.
func (s *BaseIRListener) EnterInstrWithDest(ctx *InstrWithDestContext) {}

// ExitInstrWithDest is called when production instrWithDest is exited.
func (s *BaseIRListener) ExitInstrWithDest(ctx *InstrWithDestContext) {}

// EnterInstrNoDest is called when production instrNoDest is entered.
func (s *BaseIRListener) EnterInstrNoDest(ctx *InstrNoDestContext) {}

// ExitInstrNoDest is called when production instrNoDest is exited.
func (s *BaseIRListener) ExitInstrNoDest(ctx *InstrNoDestContext) {}

// EnterConstAssign is called when production constAssign is entered.
func (s *BaseIRListener) EnterConstAssign(ctx *ConstAssignContext) {}

// ExitConstAssign is called when production constAssign is exited.
func (s *BaseIRListener) ExitConstAssign(ctx *ConstAssignContext) {}

// EnterInfixInstr is called when production infixInstr is entered.
func (s *BaseIRListener) EnterInfixInstr(ctx *InfixInstrContext) {}

// ExitInfixInstr is called when production infixInstr is exited.
func (s *BaseIRListener) ExitInfixInstr(ctx *InfixInstrContext) {}

// EnterCompoundAssignInstr is called when production compoundAssignInstr is entered.
func (s *BaseIRListener) EnterCompoundAssignInstr(ctx *CompoundAssignInstrContext) {}

// ExitCompoundAssignInstr is called when production compoundAssignInstr is exited.
func (s *BaseIRListener) ExitCompoundAssignInstr(ctx *CompoundAssignInstrContext) {}

// EnterDestReg is called when production destReg is entered.
func (s *BaseIRListener) EnterDestReg(ctx *DestRegContext) {}

// ExitDestReg is called when production destReg is exited.
func (s *BaseIRListener) ExitDestReg(ctx *DestRegContext) {}

// EnterDestDiscard is called when production destDiscard is entered.
func (s *BaseIRListener) EnterDestDiscard(ctx *DestDiscardContext) {}

// ExitDestDiscard is called when production destDiscard is exited.
func (s *BaseIRListener) ExitDestDiscard(ctx *DestDiscardContext) {}

// EnterDestList is called when production destList is entered.
func (s *BaseIRListener) EnterDestList(ctx *DestListContext) {}

// ExitDestList is called when production destList is exited.
func (s *BaseIRListener) ExitDestList(ctx *DestListContext) {}

// EnterRegList is called when production regList is entered.
func (s *BaseIRListener) EnterRegList(ctx *RegListContext) {}

// ExitRegList is called when production regList is exited.
func (s *BaseIRListener) ExitRegList(ctx *RegListContext) {}

// EnterInstrCall is called when production instrCall is entered.
func (s *BaseIRListener) EnterInstrCall(ctx *InstrCallContext) {}

// ExitInstrCall is called when production instrCall is exited.
func (s *BaseIRListener) ExitInstrCall(ctx *InstrCallContext) {}

// EnterInstrName is called when production instrName is entered.
func (s *BaseIRListener) EnterInstrName(ctx *InstrNameContext) {}

// ExitInstrName is called when production instrName is exited.
func (s *BaseIRListener) ExitInstrName(ctx *InstrNameContext) {}

// EnterTypeName is called when production typeName is entered.
func (s *BaseIRListener) EnterTypeName(ctx *TypeNameContext) {}

// ExitTypeName is called when production typeName is exited.
func (s *BaseIRListener) ExitTypeName(ctx *TypeNameContext) {}

// EnterArgs is called when production args is entered.
func (s *BaseIRListener) EnterArgs(ctx *ArgsContext) {}

// ExitArgs is called when production args is exited.
func (s *BaseIRListener) ExitArgs(ctx *ArgsContext) {}

// EnterPositionalArg is called when production positionalArg is entered.
func (s *BaseIRListener) EnterPositionalArg(ctx *PositionalArgContext) {}

// ExitPositionalArg is called when production positionalArg is exited.
func (s *BaseIRListener) ExitPositionalArg(ctx *PositionalArgContext) {}

// EnterLabeledArg is called when production labeledArg is entered.
func (s *BaseIRListener) EnterLabeledArg(ctx *LabeledArgContext) {}

// ExitLabeledArg is called when production labeledArg is exited.
func (s *BaseIRListener) ExitLabeledArg(ctx *LabeledArgContext) {}

// EnterValReg is called when production valReg is entered.
func (s *BaseIRListener) EnterValReg(ctx *ValRegContext) {}

// ExitValReg is called when production valReg is exited.
func (s *BaseIRListener) ExitValReg(ctx *ValRegContext) {}

// EnterValLabel is called when production valLabel is entered.
func (s *BaseIRListener) EnterValLabel(ctx *ValLabelContext) {}

// ExitValLabel is called when production valLabel is exited.
func (s *BaseIRListener) ExitValLabel(ctx *ValLabelContext) {}

// EnterValInt is called when production valInt is entered.
func (s *BaseIRListener) EnterValInt(ctx *ValIntContext) {}

// ExitValInt is called when production valInt is exited.
func (s *BaseIRListener) ExitValInt(ctx *ValIntContext) {}

// EnterValBool is called when production valBool is entered.
func (s *BaseIRListener) EnterValBool(ctx *ValBoolContext) {}

// ExitValBool is called when production valBool is exited.
func (s *BaseIRListener) ExitValBool(ctx *ValBoolContext) {}

// EnterValRegList is called when production valRegList is entered.
func (s *BaseIRListener) EnterValRegList(ctx *ValRegListContext) {}

// ExitValRegList is called when production valRegList is exited.
func (s *BaseIRListener) ExitValRegList(ctx *ValRegListContext) {}

// EnterConstString is called when production constString is entered.
func (s *BaseIRListener) EnterConstString(ctx *ConstStringContext) {}

// ExitConstString is called when production constString is exited.
func (s *BaseIRListener) ExitConstString(ctx *ConstStringContext) {}

// EnterConstInt is called when production constInt is entered.
func (s *BaseIRListener) EnterConstInt(ctx *ConstIntContext) {}

// ExitConstInt is called when production constInt is exited.
func (s *BaseIRListener) ExitConstInt(ctx *ConstIntContext) {}

// EnterReg is called when production reg is entered.
func (s *BaseIRListener) EnterReg(ctx *RegContext) {}

// ExitReg is called when production reg is exited.
func (s *BaseIRListener) ExitReg(ctx *RegContext) {}
