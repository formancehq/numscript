// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

// BaseNumscriptListener is a complete listener for a parse tree produced by NumscriptParser.
type BaseNumscriptListener struct{}

var _ NumscriptListener = &BaseNumscriptListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseNumscriptListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseNumscriptListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseNumscriptListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseNumscriptListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterRatio is called when production ratio is entered.
func (s *BaseNumscriptListener) EnterRatio(ctx *RatioContext) {}

// ExitRatio is called when production ratio is exited.
func (s *BaseNumscriptListener) ExitRatio(ctx *RatioContext) {}

// EnterPercentage is called when production percentage is entered.
func (s *BaseNumscriptListener) EnterPercentage(ctx *PercentageContext) {}

// ExitPercentage is called when production percentage is exited.
func (s *BaseNumscriptListener) ExitPercentage(ctx *PercentageContext) {}

// EnterVarDeclaration is called when production varDeclaration is entered.
func (s *BaseNumscriptListener) EnterVarDeclaration(ctx *VarDeclarationContext) {}

// ExitVarDeclaration is called when production varDeclaration is exited.
func (s *BaseNumscriptListener) ExitVarDeclaration(ctx *VarDeclarationContext) {}

// EnterVarsDeclaration is called when production varsDeclaration is entered.
func (s *BaseNumscriptListener) EnterVarsDeclaration(ctx *VarsDeclarationContext) {}

// ExitVarsDeclaration is called when production varsDeclaration is exited.
func (s *BaseNumscriptListener) ExitVarsDeclaration(ctx *VarsDeclarationContext) {}

// EnterProgram is called when production program is entered.
func (s *BaseNumscriptListener) EnterProgram(ctx *ProgramContext) {}

// ExitProgram is called when production program is exited.
func (s *BaseNumscriptListener) ExitProgram(ctx *ProgramContext) {}

// EnterMonetaryLit is called when production monetaryLit is entered.
func (s *BaseNumscriptListener) EnterMonetaryLit(ctx *MonetaryLitContext) {}

// ExitMonetaryLit is called when production monetaryLit is exited.
func (s *BaseNumscriptListener) ExitMonetaryLit(ctx *MonetaryLitContext) {}

// EnterLitCap is called when production litCap is entered.
func (s *BaseNumscriptListener) EnterLitCap(ctx *LitCapContext) {}

// ExitLitCap is called when production litCap is exited.
func (s *BaseNumscriptListener) ExitLitCap(ctx *LitCapContext) {}

// EnterVarCap is called when production varCap is entered.
func (s *BaseNumscriptListener) EnterVarCap(ctx *VarCapContext) {}

// ExitVarCap is called when production varCap is exited.
func (s *BaseNumscriptListener) ExitVarCap(ctx *VarCapContext) {}

// EnterPortionedAllotment is called when production portionedAllotment is entered.
func (s *BaseNumscriptListener) EnterPortionedAllotment(ctx *PortionedAllotmentContext) {}

// ExitPortionedAllotment is called when production portionedAllotment is exited.
func (s *BaseNumscriptListener) ExitPortionedAllotment(ctx *PortionedAllotmentContext) {}

// EnterPortionVariable is called when production portionVariable is entered.
func (s *BaseNumscriptListener) EnterPortionVariable(ctx *PortionVariableContext) {}

// ExitPortionVariable is called when production portionVariable is exited.
func (s *BaseNumscriptListener) ExitPortionVariable(ctx *PortionVariableContext) {}

// EnterRemainingAllotment is called when production remainingAllotment is entered.
func (s *BaseNumscriptListener) EnterRemainingAllotment(ctx *RemainingAllotmentContext) {}

// ExitRemainingAllotment is called when production remainingAllotment is exited.
func (s *BaseNumscriptListener) ExitRemainingAllotment(ctx *RemainingAllotmentContext) {}

// EnterAccountName is called when production accountName is entered.
func (s *BaseNumscriptListener) EnterAccountName(ctx *AccountNameContext) {}

// ExitAccountName is called when production accountName is exited.
func (s *BaseNumscriptListener) ExitAccountName(ctx *AccountNameContext) {}

// EnterAccountVariable is called when production accountVariable is entered.
func (s *BaseNumscriptListener) EnterAccountVariable(ctx *AccountVariableContext) {}

// ExitAccountVariable is called when production accountVariable is exited.
func (s *BaseNumscriptListener) ExitAccountVariable(ctx *AccountVariableContext) {}

// EnterSrcAccountUnboundedOverdraft is called when production srcAccountUnboundedOverdraft is entered.
func (s *BaseNumscriptListener) EnterSrcAccountUnboundedOverdraft(ctx *SrcAccountUnboundedOverdraftContext) {
}

// ExitSrcAccountUnboundedOverdraft is called when production srcAccountUnboundedOverdraft is exited.
func (s *BaseNumscriptListener) ExitSrcAccountUnboundedOverdraft(ctx *SrcAccountUnboundedOverdraftContext) {
}

// EnterSrcAccount is called when production srcAccount is entered.
func (s *BaseNumscriptListener) EnterSrcAccount(ctx *SrcAccountContext) {}

// ExitSrcAccount is called when production srcAccount is exited.
func (s *BaseNumscriptListener) ExitSrcAccount(ctx *SrcAccountContext) {}

// EnterSrcVariable is called when production srcVariable is entered.
func (s *BaseNumscriptListener) EnterSrcVariable(ctx *SrcVariableContext) {}

// ExitSrcVariable is called when production srcVariable is exited.
func (s *BaseNumscriptListener) ExitSrcVariable(ctx *SrcVariableContext) {}

// EnterSrcAllotment is called when production srcAllotment is entered.
func (s *BaseNumscriptListener) EnterSrcAllotment(ctx *SrcAllotmentContext) {}

// ExitSrcAllotment is called when production srcAllotment is exited.
func (s *BaseNumscriptListener) ExitSrcAllotment(ctx *SrcAllotmentContext) {}

// EnterSrcSeq is called when production srcSeq is entered.
func (s *BaseNumscriptListener) EnterSrcSeq(ctx *SrcSeqContext) {}

// ExitSrcSeq is called when production srcSeq is exited.
func (s *BaseNumscriptListener) ExitSrcSeq(ctx *SrcSeqContext) {}

// EnterSrcCapped is called when production srcCapped is entered.
func (s *BaseNumscriptListener) EnterSrcCapped(ctx *SrcCappedContext) {}

// ExitSrcCapped is called when production srcCapped is exited.
func (s *BaseNumscriptListener) ExitSrcCapped(ctx *SrcCappedContext) {}

// EnterAllotmentClauseSrc is called when production allotmentClauseSrc is entered.
func (s *BaseNumscriptListener) EnterAllotmentClauseSrc(ctx *AllotmentClauseSrcContext) {}

// ExitAllotmentClauseSrc is called when production allotmentClauseSrc is exited.
func (s *BaseNumscriptListener) ExitAllotmentClauseSrc(ctx *AllotmentClauseSrcContext) {}

// EnterDestAccount is called when production destAccount is entered.
func (s *BaseNumscriptListener) EnterDestAccount(ctx *DestAccountContext) {}

// ExitDestAccount is called when production destAccount is exited.
func (s *BaseNumscriptListener) ExitDestAccount(ctx *DestAccountContext) {}

// EnterDestVariable is called when production destVariable is entered.
func (s *BaseNumscriptListener) EnterDestVariable(ctx *DestVariableContext) {}

// ExitDestVariable is called when production destVariable is exited.
func (s *BaseNumscriptListener) ExitDestVariable(ctx *DestVariableContext) {}

// EnterDestAllotment is called when production destAllotment is entered.
func (s *BaseNumscriptListener) EnterDestAllotment(ctx *DestAllotmentContext) {}

// ExitDestAllotment is called when production destAllotment is exited.
func (s *BaseNumscriptListener) ExitDestAllotment(ctx *DestAllotmentContext) {}

// EnterDestSeq is called when production destSeq is entered.
func (s *BaseNumscriptListener) EnterDestSeq(ctx *DestSeqContext) {}

// ExitDestSeq is called when production destSeq is exited.
func (s *BaseNumscriptListener) ExitDestSeq(ctx *DestSeqContext) {}

// EnterAllotmentClauseDest is called when production allotmentClauseDest is entered.
func (s *BaseNumscriptListener) EnterAllotmentClauseDest(ctx *AllotmentClauseDestContext) {}

// ExitAllotmentClauseDest is called when production allotmentClauseDest is exited.
func (s *BaseNumscriptListener) ExitAllotmentClauseDest(ctx *AllotmentClauseDestContext) {}

// EnterSendMon is called when production sendMon is entered.
func (s *BaseNumscriptListener) EnterSendMon(ctx *SendMonContext) {}

// ExitSendMon is called when production sendMon is exited.
func (s *BaseNumscriptListener) ExitSendMon(ctx *SendMonContext) {}

// EnterSendVariable is called when production sendVariable is entered.
func (s *BaseNumscriptListener) EnterSendVariable(ctx *SendVariableContext) {}

// ExitSendVariable is called when production sendVariable is exited.
func (s *BaseNumscriptListener) ExitSendVariable(ctx *SendVariableContext) {}

// EnterStatement is called when production statement is entered.
func (s *BaseNumscriptListener) EnterStatement(ctx *StatementContext) {}

// ExitStatement is called when production statement is exited.
func (s *BaseNumscriptListener) ExitStatement(ctx *StatementContext) {}
