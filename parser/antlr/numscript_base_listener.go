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

// EnterProgram is called when production program is entered.
func (s *BaseNumscriptListener) EnterProgram(ctx *ProgramContext) {}

// ExitProgram is called when production program is exited.
func (s *BaseNumscriptListener) ExitProgram(ctx *ProgramContext) {}

// EnterMonetaryLit is called when production monetaryLit is entered.
func (s *BaseNumscriptListener) EnterMonetaryLit(ctx *MonetaryLitContext) {}

// ExitMonetaryLit is called when production monetaryLit is exited.
func (s *BaseNumscriptListener) ExitMonetaryLit(ctx *MonetaryLitContext) {}

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

// EnterDestSeq is called when production destSeq is entered.
func (s *BaseNumscriptListener) EnterDestSeq(ctx *DestSeqContext) {}

// ExitDestSeq is called when production destSeq is exited.
func (s *BaseNumscriptListener) ExitDestSeq(ctx *DestSeqContext) {}

// EnterStatement is called when production statement is entered.
func (s *BaseNumscriptListener) EnterStatement(ctx *StatementContext) {}

// ExitStatement is called when production statement is exited.
func (s *BaseNumscriptListener) ExitStatement(ctx *StatementContext) {}
