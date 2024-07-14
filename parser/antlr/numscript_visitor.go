// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by NumscriptParser.
type NumscriptVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by NumscriptParser#ratio.
	VisitRatio(ctx *RatioContext) interface{}

	// Visit a parse tree produced by NumscriptParser#percentage.
	VisitPercentage(ctx *PercentageContext) interface{}

	// Visit a parse tree produced by NumscriptParser#varDeclaration.
	VisitVarDeclaration(ctx *VarDeclarationContext) interface{}

	// Visit a parse tree produced by NumscriptParser#varsDeclaration.
	VisitVarsDeclaration(ctx *VarsDeclarationContext) interface{}

	// Visit a parse tree produced by NumscriptParser#program.
	VisitProgram(ctx *ProgramContext) interface{}

	// Visit a parse tree produced by NumscriptParser#monetaryLit.
	VisitMonetaryLit(ctx *MonetaryLitContext) interface{}

	// Visit a parse tree produced by NumscriptParser#litCap.
	VisitLitCap(ctx *LitCapContext) interface{}

	// Visit a parse tree produced by NumscriptParser#varCap.
	VisitVarCap(ctx *VarCapContext) interface{}

	// Visit a parse tree produced by NumscriptParser#portionedAllotment.
	VisitPortionedAllotment(ctx *PortionedAllotmentContext) interface{}

	// Visit a parse tree produced by NumscriptParser#portionVariable.
	VisitPortionVariable(ctx *PortionVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#remainingAllotment.
	VisitRemainingAllotment(ctx *RemainingAllotmentContext) interface{}

	// Visit a parse tree produced by NumscriptParser#accountName.
	VisitAccountName(ctx *AccountNameContext) interface{}

	// Visit a parse tree produced by NumscriptParser#accountVariable.
	VisitAccountVariable(ctx *AccountVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcAccountUnboundedOverdraft.
	VisitSrcAccountUnboundedOverdraft(ctx *SrcAccountUnboundedOverdraftContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcAccount.
	VisitSrcAccount(ctx *SrcAccountContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcVariable.
	VisitSrcVariable(ctx *SrcVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcAllotment.
	VisitSrcAllotment(ctx *SrcAllotmentContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcSeq.
	VisitSrcSeq(ctx *SrcSeqContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcCapped.
	VisitSrcCapped(ctx *SrcCappedContext) interface{}

	// Visit a parse tree produced by NumscriptParser#allotmentClauseSrc.
	VisitAllotmentClauseSrc(ctx *AllotmentClauseSrcContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destAccount.
	VisitDestAccount(ctx *DestAccountContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destVariable.
	VisitDestVariable(ctx *DestVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destAllotment.
	VisitDestAllotment(ctx *DestAllotmentContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destSeq.
	VisitDestSeq(ctx *DestSeqContext) interface{}

	// Visit a parse tree produced by NumscriptParser#allotmentClauseDest.
	VisitAllotmentClauseDest(ctx *AllotmentClauseDestContext) interface{}

	// Visit a parse tree produced by NumscriptParser#sendMon.
	VisitSendMon(ctx *SendMonContext) interface{}

	// Visit a parse tree produced by NumscriptParser#sendVariable.
	VisitSendVariable(ctx *SendVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#statement.
	VisitStatement(ctx *StatementContext) interface{}
}
