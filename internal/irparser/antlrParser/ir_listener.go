// Code generated from IR.g4 by ANTLR 4.13.2. DO NOT EDIT.

package antlrParser // IR
import "github.com/antlr4-go/antlr/v4"

// IRListener is a complete listener for a parse tree produced by IRParser.
type IRListener interface {
	antlr.ParseTreeListener

	// EnterProgram is called when entering the program production.
	EnterProgram(c *ProgramContext)

	// EnterLine is called when entering the line production.
	EnterLine(c *LineContext)

	// EnterLabelMarker is called when entering the labelMarker production.
	EnterLabelMarker(c *LabelMarkerContext)

	// EnterInstrWithDest is called when entering the instrWithDest production.
	EnterInstrWithDest(c *InstrWithDestContext)

	// EnterInstrNoDest is called when entering the instrNoDest production.
	EnterInstrNoDest(c *InstrNoDestContext)

	// EnterConstAssign is called when entering the constAssign production.
	EnterConstAssign(c *ConstAssignContext)

	// EnterInfixInstr is called when entering the infixInstr production.
	EnterInfixInstr(c *InfixInstrContext)

	// EnterCompoundAssignInstr is called when entering the compoundAssignInstr production.
	EnterCompoundAssignInstr(c *CompoundAssignInstrContext)

	// EnterDestReg is called when entering the destReg production.
	EnterDestReg(c *DestRegContext)

	// EnterDestDiscard is called when entering the destDiscard production.
	EnterDestDiscard(c *DestDiscardContext)

	// EnterDestList is called when entering the destList production.
	EnterDestList(c *DestListContext)

	// EnterRegList is called when entering the regList production.
	EnterRegList(c *RegListContext)

	// EnterInstrCall is called when entering the instrCall production.
	EnterInstrCall(c *InstrCallContext)

	// EnterInstrName is called when entering the instrName production.
	EnterInstrName(c *InstrNameContext)

	// EnterTypeName is called when entering the typeName production.
	EnterTypeName(c *TypeNameContext)

	// EnterArgs is called when entering the args production.
	EnterArgs(c *ArgsContext)

	// EnterPositionalArg is called when entering the positionalArg production.
	EnterPositionalArg(c *PositionalArgContext)

	// EnterLabeledArg is called when entering the labeledArg production.
	EnterLabeledArg(c *LabeledArgContext)

	// EnterValReg is called when entering the valReg production.
	EnterValReg(c *ValRegContext)

	// EnterValLabel is called when entering the valLabel production.
	EnterValLabel(c *ValLabelContext)

	// EnterValInt is called when entering the valInt production.
	EnterValInt(c *ValIntContext)

	// EnterValBool is called when entering the valBool production.
	EnterValBool(c *ValBoolContext)

	// EnterValRegList is called when entering the valRegList production.
	EnterValRegList(c *ValRegListContext)

	// EnterConstString is called when entering the constString production.
	EnterConstString(c *ConstStringContext)

	// EnterConstInt is called when entering the constInt production.
	EnterConstInt(c *ConstIntContext)

	// EnterReg is called when entering the reg production.
	EnterReg(c *RegContext)

	// ExitProgram is called when exiting the program production.
	ExitProgram(c *ProgramContext)

	// ExitLine is called when exiting the line production.
	ExitLine(c *LineContext)

	// ExitLabelMarker is called when exiting the labelMarker production.
	ExitLabelMarker(c *LabelMarkerContext)

	// ExitInstrWithDest is called when exiting the instrWithDest production.
	ExitInstrWithDest(c *InstrWithDestContext)

	// ExitInstrNoDest is called when exiting the instrNoDest production.
	ExitInstrNoDest(c *InstrNoDestContext)

	// ExitConstAssign is called when exiting the constAssign production.
	ExitConstAssign(c *ConstAssignContext)

	// ExitInfixInstr is called when exiting the infixInstr production.
	ExitInfixInstr(c *InfixInstrContext)

	// ExitCompoundAssignInstr is called when exiting the compoundAssignInstr production.
	ExitCompoundAssignInstr(c *CompoundAssignInstrContext)

	// ExitDestReg is called when exiting the destReg production.
	ExitDestReg(c *DestRegContext)

	// ExitDestDiscard is called when exiting the destDiscard production.
	ExitDestDiscard(c *DestDiscardContext)

	// ExitDestList is called when exiting the destList production.
	ExitDestList(c *DestListContext)

	// ExitRegList is called when exiting the regList production.
	ExitRegList(c *RegListContext)

	// ExitInstrCall is called when exiting the instrCall production.
	ExitInstrCall(c *InstrCallContext)

	// ExitInstrName is called when exiting the instrName production.
	ExitInstrName(c *InstrNameContext)

	// ExitTypeName is called when exiting the typeName production.
	ExitTypeName(c *TypeNameContext)

	// ExitArgs is called when exiting the args production.
	ExitArgs(c *ArgsContext)

	// ExitPositionalArg is called when exiting the positionalArg production.
	ExitPositionalArg(c *PositionalArgContext)

	// ExitLabeledArg is called when exiting the labeledArg production.
	ExitLabeledArg(c *LabeledArgContext)

	// ExitValReg is called when exiting the valReg production.
	ExitValReg(c *ValRegContext)

	// ExitValLabel is called when exiting the valLabel production.
	ExitValLabel(c *ValLabelContext)

	// ExitValInt is called when exiting the valInt production.
	ExitValInt(c *ValIntContext)

	// ExitValBool is called when exiting the valBool production.
	ExitValBool(c *ValBoolContext)

	// ExitValRegList is called when exiting the valRegList production.
	ExitValRegList(c *ValRegListContext)

	// ExitConstString is called when exiting the constString production.
	ExitConstString(c *ConstStringContext)

	// ExitConstInt is called when exiting the constInt production.
	ExitConstInt(c *ConstIntContext)

	// ExitReg is called when exiting the reg production.
	ExitReg(c *RegContext)
}
