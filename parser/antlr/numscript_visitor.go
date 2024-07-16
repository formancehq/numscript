// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by NumscriptParser.
type NumscriptVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by NumscriptParser#assetLiteral.
	VisitAssetLiteral(ctx *AssetLiteralContext) interface{}

	// Visit a parse tree produced by NumscriptParser#stringLiteral.
	VisitStringLiteral(ctx *StringLiteralContext) interface{}

	// Visit a parse tree produced by NumscriptParser#monetaryLiteral.
	VisitMonetaryLiteral(ctx *MonetaryLiteralContext) interface{}

	// Visit a parse tree produced by NumscriptParser#accountLiteral.
	VisitAccountLiteral(ctx *AccountLiteralContext) interface{}

	// Visit a parse tree produced by NumscriptParser#variableLiteral.
	VisitVariableLiteral(ctx *VariableLiteralContext) interface{}

	// Visit a parse tree produced by NumscriptParser#portionLiteral.
	VisitPortionLiteral(ctx *PortionLiteralContext) interface{}

	// Visit a parse tree produced by NumscriptParser#numberLiteral.
	VisitNumberLiteral(ctx *NumberLiteralContext) interface{}

	// Visit a parse tree produced by NumscriptParser#ratio.
	VisitRatio(ctx *RatioContext) interface{}

	// Visit a parse tree produced by NumscriptParser#percentage.
	VisitPercentage(ctx *PercentageContext) interface{}

	// Visit a parse tree produced by NumscriptParser#functionCallArgs.
	VisitFunctionCallArgs(ctx *FunctionCallArgsContext) interface{}

	// Visit a parse tree produced by NumscriptParser#functionCall.
	VisitFunctionCall(ctx *FunctionCallContext) interface{}

	// Visit a parse tree produced by NumscriptParser#varOrigin.
	VisitVarOrigin(ctx *VarOriginContext) interface{}

	// Visit a parse tree produced by NumscriptParser#varDeclaration.
	VisitVarDeclaration(ctx *VarDeclarationContext) interface{}

	// Visit a parse tree produced by NumscriptParser#varsDeclaration.
	VisitVarsDeclaration(ctx *VarsDeclarationContext) interface{}

	// Visit a parse tree produced by NumscriptParser#program.
	VisitProgram(ctx *ProgramContext) interface{}

	// Visit a parse tree produced by NumscriptParser#number.
	VisitNumber(ctx *NumberContext) interface{}

	// Visit a parse tree produced by NumscriptParser#numberVariable.
	VisitNumberVariable(ctx *NumberVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#asset.
	VisitAsset(ctx *AssetContext) interface{}

	// Visit a parse tree produced by NumscriptParser#assetVariable.
	VisitAssetVariable(ctx *AssetVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#accountName.
	VisitAccountName(ctx *AccountNameContext) interface{}

	// Visit a parse tree produced by NumscriptParser#accountVariable.
	VisitAccountVariable(ctx *AccountVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#monetary.
	VisitMonetary(ctx *MonetaryContext) interface{}

	// Visit a parse tree produced by NumscriptParser#monetaryVariable.
	VisitMonetaryVariable(ctx *MonetaryVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#monetaryLit.
	VisitMonetaryLit(ctx *MonetaryLitContext) interface{}

	// Visit a parse tree produced by NumscriptParser#sentAllLit.
	VisitSentAllLit(ctx *SentAllLitContext) interface{}

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

	// Visit a parse tree produced by NumscriptParser#srcAccountUnboundedOverdraft.
	VisitSrcAccountUnboundedOverdraft(ctx *SrcAccountUnboundedOverdraftContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcAccountBoundedOverdraft.
	VisitSrcAccountBoundedOverdraft(ctx *SrcAccountBoundedOverdraftContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcAccount.
	VisitSrcAccount(ctx *SrcAccountContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcVariable.
	VisitSrcVariable(ctx *SrcVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcAllotment.
	VisitSrcAllotment(ctx *SrcAllotmentContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcInorder.
	VisitSrcInorder(ctx *SrcInorderContext) interface{}

	// Visit a parse tree produced by NumscriptParser#srcCapped.
	VisitSrcCapped(ctx *SrcCappedContext) interface{}

	// Visit a parse tree produced by NumscriptParser#allotmentClauseSrc.
	VisitAllotmentClauseSrc(ctx *AllotmentClauseSrcContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destinationTo.
	VisitDestinationTo(ctx *DestinationToContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destinationKept.
	VisitDestinationKept(ctx *DestinationKeptContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destinationInOrderClause.
	VisitDestinationInOrderClause(ctx *DestinationInOrderClauseContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destAccount.
	VisitDestAccount(ctx *DestAccountContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destVariable.
	VisitDestVariable(ctx *DestVariableContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destAllotment.
	VisitDestAllotment(ctx *DestAllotmentContext) interface{}

	// Visit a parse tree produced by NumscriptParser#destInorder.
	VisitDestInorder(ctx *DestInorderContext) interface{}

	// Visit a parse tree produced by NumscriptParser#allotmentClauseDest.
	VisitAllotmentClauseDest(ctx *AllotmentClauseDestContext) interface{}

	// Visit a parse tree produced by NumscriptParser#sentLiteral.
	VisitSentLiteral(ctx *SentLiteralContext) interface{}

	// Visit a parse tree produced by NumscriptParser#sentAll.
	VisitSentAll(ctx *SentAllContext) interface{}

	// Visit a parse tree produced by NumscriptParser#sendStatement.
	VisitSendStatement(ctx *SendStatementContext) interface{}

	// Visit a parse tree produced by NumscriptParser#fnCallStatement.
	VisitFnCallStatement(ctx *FnCallStatementContext) interface{}
}
