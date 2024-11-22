// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

// NumscriptListener is a complete listener for a parse tree produced by NumscriptParser.
type NumscriptListener interface {
	antlr.ParseTreeListener

	// EnterMonetaryLit is called when entering the monetaryLit production.
	EnterMonetaryLit(c *MonetaryLitContext)

	// EnterRatio is called when entering the ratio production.
	EnterRatio(c *RatioContext)

	// EnterPercentage is called when entering the percentage production.
	EnterPercentage(c *PercentageContext)

	// EnterInfixCompExpr is called when entering the infixCompExpr production.
	EnterInfixCompExpr(c *InfixCompExprContext)

	// EnterAccountLiteral is called when entering the accountLiteral production.
	EnterAccountLiteral(c *AccountLiteralContext)

	// EnterParensExpr is called when entering the parensExpr production.
	EnterParensExpr(c *ParensExprContext)

	// EnterMonetaryLiteral is called when entering the monetaryLiteral production.
	EnterMonetaryLiteral(c *MonetaryLiteralContext)

	// EnterInfixEqExpr is called when entering the infixEqExpr production.
	EnterInfixEqExpr(c *InfixEqExprContext)

	// EnterVariableExpr is called when entering the variableExpr production.
	EnterVariableExpr(c *VariableExprContext)

	// EnterPortionLiteral is called when entering the portionLiteral production.
	EnterPortionLiteral(c *PortionLiteralContext)

	// EnterInfixAndExpr is called when entering the infixAndExpr production.
	EnterInfixAndExpr(c *InfixAndExprContext)

	// EnterAssetLiteral is called when entering the assetLiteral production.
	EnterAssetLiteral(c *AssetLiteralContext)

	// EnterStringLiteral is called when entering the stringLiteral production.
	EnterStringLiteral(c *StringLiteralContext)

	// EnterInfixOrExpr is called when entering the infixOrExpr production.
	EnterInfixOrExpr(c *InfixOrExprContext)

	// EnterInfixAddSubExpr is called when entering the infixAddSubExpr production.
	EnterInfixAddSubExpr(c *InfixAddSubExprContext)

	// EnterNumberLiteral is called when entering the numberLiteral production.
	EnterNumberLiteral(c *NumberLiteralContext)

	// EnterFunctionCallArgs is called when entering the functionCallArgs production.
	EnterFunctionCallArgs(c *FunctionCallArgsContext)

	// EnterFunctionCall is called when entering the functionCall production.
	EnterFunctionCall(c *FunctionCallContext)

	// EnterVarOrigin is called when entering the varOrigin production.
	EnterVarOrigin(c *VarOriginContext)

	// EnterVarDeclaration is called when entering the varDeclaration production.
	EnterVarDeclaration(c *VarDeclarationContext)

	// EnterVarsDeclaration is called when entering the varsDeclaration production.
	EnterVarsDeclaration(c *VarsDeclarationContext)

	// EnterProgram is called when entering the program production.
	EnterProgram(c *ProgramContext)

	// EnterSentAllLit is called when entering the sentAllLit production.
	EnterSentAllLit(c *SentAllLitContext)

	// EnterLitCap is called when entering the litCap production.
	EnterLitCap(c *LitCapContext)

	// EnterVarCap is called when entering the varCap production.
	EnterVarCap(c *VarCapContext)

	// EnterPortionedAllotment is called when entering the portionedAllotment production.
	EnterPortionedAllotment(c *PortionedAllotmentContext)

	// EnterPortionVariable is called when entering the portionVariable production.
	EnterPortionVariable(c *PortionVariableContext)

	// EnterRemainingAllotment is called when entering the remainingAllotment production.
	EnterRemainingAllotment(c *RemainingAllotmentContext)

	// EnterSrcAccountUnboundedOverdraft is called when entering the srcAccountUnboundedOverdraft production.
	EnterSrcAccountUnboundedOverdraft(c *SrcAccountUnboundedOverdraftContext)

	// EnterSrcAccountBoundedOverdraft is called when entering the srcAccountBoundedOverdraft production.
	EnterSrcAccountBoundedOverdraft(c *SrcAccountBoundedOverdraftContext)

	// EnterSrcAccount is called when entering the srcAccount production.
	EnterSrcAccount(c *SrcAccountContext)

	// EnterSrcAllotment is called when entering the srcAllotment production.
	EnterSrcAllotment(c *SrcAllotmentContext)

	// EnterSrcInorder is called when entering the srcInorder production.
	EnterSrcInorder(c *SrcInorderContext)

	// EnterSrcCapped is called when entering the srcCapped production.
	EnterSrcCapped(c *SrcCappedContext)

	// EnterAllotmentClauseSrc is called when entering the allotmentClauseSrc production.
	EnterAllotmentClauseSrc(c *AllotmentClauseSrcContext)

	// EnterDestinationTo is called when entering the destinationTo production.
	EnterDestinationTo(c *DestinationToContext)

	// EnterDestinationKept is called when entering the destinationKept production.
	EnterDestinationKept(c *DestinationKeptContext)

	// EnterDestinationInOrderClause is called when entering the destinationInOrderClause production.
	EnterDestinationInOrderClause(c *DestinationInOrderClauseContext)

	// EnterDestInorder is called when entering the destInorder production.
	EnterDestInorder(c *DestInorderContext)

	// EnterDestIf is called when entering the destIf production.
	EnterDestIf(c *DestIfContext)

	// EnterDestAccount is called when entering the destAccount production.
	EnterDestAccount(c *DestAccountContext)

	// EnterDestAllotment is called when entering the destAllotment production.
	EnterDestAllotment(c *DestAllotmentContext)

	// EnterAllotmentClauseDest is called when entering the allotmentClauseDest production.
	EnterAllotmentClauseDest(c *AllotmentClauseDestContext)

	// EnterSentLiteral is called when entering the sentLiteral production.
	EnterSentLiteral(c *SentLiteralContext)

	// EnterSentAll is called when entering the sentAll production.
	EnterSentAll(c *SentAllContext)

	// EnterSendStatement is called when entering the sendStatement production.
	EnterSendStatement(c *SendStatementContext)

	// EnterSaveStatement is called when entering the saveStatement production.
	EnterSaveStatement(c *SaveStatementContext)

	// EnterFnCallStatement is called when entering the fnCallStatement production.
	EnterFnCallStatement(c *FnCallStatementContext)

	// ExitMonetaryLit is called when exiting the monetaryLit production.
	ExitMonetaryLit(c *MonetaryLitContext)

	// ExitRatio is called when exiting the ratio production.
	ExitRatio(c *RatioContext)

	// ExitPercentage is called when exiting the percentage production.
	ExitPercentage(c *PercentageContext)

	// ExitInfixCompExpr is called when exiting the infixCompExpr production.
	ExitInfixCompExpr(c *InfixCompExprContext)

	// ExitAccountLiteral is called when exiting the accountLiteral production.
	ExitAccountLiteral(c *AccountLiteralContext)

	// ExitParensExpr is called when exiting the parensExpr production.
	ExitParensExpr(c *ParensExprContext)

	// ExitMonetaryLiteral is called when exiting the monetaryLiteral production.
	ExitMonetaryLiteral(c *MonetaryLiteralContext)

	// ExitInfixEqExpr is called when exiting the infixEqExpr production.
	ExitInfixEqExpr(c *InfixEqExprContext)

	// ExitVariableExpr is called when exiting the variableExpr production.
	ExitVariableExpr(c *VariableExprContext)

	// ExitPortionLiteral is called when exiting the portionLiteral production.
	ExitPortionLiteral(c *PortionLiteralContext)

	// ExitInfixAndExpr is called when exiting the infixAndExpr production.
	ExitInfixAndExpr(c *InfixAndExprContext)

	// ExitAssetLiteral is called when exiting the assetLiteral production.
	ExitAssetLiteral(c *AssetLiteralContext)

	// ExitStringLiteral is called when exiting the stringLiteral production.
	ExitStringLiteral(c *StringLiteralContext)

	// ExitInfixOrExpr is called when exiting the infixOrExpr production.
	ExitInfixOrExpr(c *InfixOrExprContext)

	// ExitInfixAddSubExpr is called when exiting the infixAddSubExpr production.
	ExitInfixAddSubExpr(c *InfixAddSubExprContext)

	// ExitNumberLiteral is called when exiting the numberLiteral production.
	ExitNumberLiteral(c *NumberLiteralContext)

	// ExitFunctionCallArgs is called when exiting the functionCallArgs production.
	ExitFunctionCallArgs(c *FunctionCallArgsContext)

	// ExitFunctionCall is called when exiting the functionCall production.
	ExitFunctionCall(c *FunctionCallContext)

	// ExitVarOrigin is called when exiting the varOrigin production.
	ExitVarOrigin(c *VarOriginContext)

	// ExitVarDeclaration is called when exiting the varDeclaration production.
	ExitVarDeclaration(c *VarDeclarationContext)

	// ExitVarsDeclaration is called when exiting the varsDeclaration production.
	ExitVarsDeclaration(c *VarsDeclarationContext)

	// ExitProgram is called when exiting the program production.
	ExitProgram(c *ProgramContext)

	// ExitSentAllLit is called when exiting the sentAllLit production.
	ExitSentAllLit(c *SentAllLitContext)

	// ExitLitCap is called when exiting the litCap production.
	ExitLitCap(c *LitCapContext)

	// ExitVarCap is called when exiting the varCap production.
	ExitVarCap(c *VarCapContext)

	// ExitPortionedAllotment is called when exiting the portionedAllotment production.
	ExitPortionedAllotment(c *PortionedAllotmentContext)

	// ExitPortionVariable is called when exiting the portionVariable production.
	ExitPortionVariable(c *PortionVariableContext)

	// ExitRemainingAllotment is called when exiting the remainingAllotment production.
	ExitRemainingAllotment(c *RemainingAllotmentContext)

	// ExitSrcAccountUnboundedOverdraft is called when exiting the srcAccountUnboundedOverdraft production.
	ExitSrcAccountUnboundedOverdraft(c *SrcAccountUnboundedOverdraftContext)

	// ExitSrcAccountBoundedOverdraft is called when exiting the srcAccountBoundedOverdraft production.
	ExitSrcAccountBoundedOverdraft(c *SrcAccountBoundedOverdraftContext)

	// ExitSrcAccount is called when exiting the srcAccount production.
	ExitSrcAccount(c *SrcAccountContext)

	// ExitSrcAllotment is called when exiting the srcAllotment production.
	ExitSrcAllotment(c *SrcAllotmentContext)

	// ExitSrcInorder is called when exiting the srcInorder production.
	ExitSrcInorder(c *SrcInorderContext)

	// ExitSrcCapped is called when exiting the srcCapped production.
	ExitSrcCapped(c *SrcCappedContext)

	// ExitAllotmentClauseSrc is called when exiting the allotmentClauseSrc production.
	ExitAllotmentClauseSrc(c *AllotmentClauseSrcContext)

	// ExitDestinationTo is called when exiting the destinationTo production.
	ExitDestinationTo(c *DestinationToContext)

	// ExitDestinationKept is called when exiting the destinationKept production.
	ExitDestinationKept(c *DestinationKeptContext)

	// ExitDestinationInOrderClause is called when exiting the destinationInOrderClause production.
	ExitDestinationInOrderClause(c *DestinationInOrderClauseContext)

	// ExitDestInorder is called when exiting the destInorder production.
	ExitDestInorder(c *DestInorderContext)

	// ExitDestIf is called when exiting the destIf production.
	ExitDestIf(c *DestIfContext)

	// ExitDestAccount is called when exiting the destAccount production.
	ExitDestAccount(c *DestAccountContext)

	// ExitDestAllotment is called when exiting the destAllotment production.
	ExitDestAllotment(c *DestAllotmentContext)

	// ExitAllotmentClauseDest is called when exiting the allotmentClauseDest production.
	ExitAllotmentClauseDest(c *AllotmentClauseDestContext)

	// ExitSentLiteral is called when exiting the sentLiteral production.
	ExitSentLiteral(c *SentLiteralContext)

	// ExitSentAll is called when exiting the sentAll production.
	ExitSentAll(c *SentAllContext)

	// ExitSendStatement is called when exiting the sendStatement production.
	ExitSendStatement(c *SendStatementContext)

	// ExitSaveStatement is called when exiting the saveStatement production.
	ExitSaveStatement(c *SaveStatementContext)

	// ExitFnCallStatement is called when exiting the fnCallStatement production.
	ExitFnCallStatement(c *FnCallStatementContext)
}
