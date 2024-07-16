// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

// NumscriptListener is a complete listener for a parse tree produced by NumscriptParser.
type NumscriptListener interface {
	antlr.ParseTreeListener

	// EnterAssetLiteral is called when entering the assetLiteral production.
	EnterAssetLiteral(c *AssetLiteralContext)

	// EnterStringLiteral is called when entering the stringLiteral production.
	EnterStringLiteral(c *StringLiteralContext)

	// EnterMonetaryLiteral is called when entering the monetaryLiteral production.
	EnterMonetaryLiteral(c *MonetaryLiteralContext)

	// EnterAccountLiteral is called when entering the accountLiteral production.
	EnterAccountLiteral(c *AccountLiteralContext)

	// EnterVariableLiteral is called when entering the variableLiteral production.
	EnterVariableLiteral(c *VariableLiteralContext)

	// EnterPortionLiteral is called when entering the portionLiteral production.
	EnterPortionLiteral(c *PortionLiteralContext)

	// EnterNumberLiteral is called when entering the numberLiteral production.
	EnterNumberLiteral(c *NumberLiteralContext)

	// EnterRatio is called when entering the ratio production.
	EnterRatio(c *RatioContext)

	// EnterPercentage is called when entering the percentage production.
	EnterPercentage(c *PercentageContext)

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

	// EnterNumber is called when entering the number production.
	EnterNumber(c *NumberContext)

	// EnterNumberVariable is called when entering the numberVariable production.
	EnterNumberVariable(c *NumberVariableContext)

	// EnterAsset is called when entering the asset production.
	EnterAsset(c *AssetContext)

	// EnterAssetVariable is called when entering the assetVariable production.
	EnterAssetVariable(c *AssetVariableContext)

	// EnterAccountName is called when entering the accountName production.
	EnterAccountName(c *AccountNameContext)

	// EnterAccountVariable is called when entering the accountVariable production.
	EnterAccountVariable(c *AccountVariableContext)

	// EnterMonetary is called when entering the monetary production.
	EnterMonetary(c *MonetaryContext)

	// EnterMonetaryVariable is called when entering the monetaryVariable production.
	EnterMonetaryVariable(c *MonetaryVariableContext)

	// EnterMonetaryLit is called when entering the monetaryLit production.
	EnterMonetaryLit(c *MonetaryLitContext)

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

	// EnterSrcVariable is called when entering the srcVariable production.
	EnterSrcVariable(c *SrcVariableContext)

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

	// EnterDestAccount is called when entering the destAccount production.
	EnterDestAccount(c *DestAccountContext)

	// EnterDestVariable is called when entering the destVariable production.
	EnterDestVariable(c *DestVariableContext)

	// EnterDestAllotment is called when entering the destAllotment production.
	EnterDestAllotment(c *DestAllotmentContext)

	// EnterDestInorder is called when entering the destInorder production.
	EnterDestInorder(c *DestInorderContext)

	// EnterAllotmentClauseDest is called when entering the allotmentClauseDest production.
	EnterAllotmentClauseDest(c *AllotmentClauseDestContext)

	// EnterSentLiteral is called when entering the sentLiteral production.
	EnterSentLiteral(c *SentLiteralContext)

	// EnterSentAll is called when entering the sentAll production.
	EnterSentAll(c *SentAllContext)

	// EnterSendStatement is called when entering the sendStatement production.
	EnterSendStatement(c *SendStatementContext)

	// EnterFnCallStatement is called when entering the fnCallStatement production.
	EnterFnCallStatement(c *FnCallStatementContext)

	// ExitAssetLiteral is called when exiting the assetLiteral production.
	ExitAssetLiteral(c *AssetLiteralContext)

	// ExitStringLiteral is called when exiting the stringLiteral production.
	ExitStringLiteral(c *StringLiteralContext)

	// ExitMonetaryLiteral is called when exiting the monetaryLiteral production.
	ExitMonetaryLiteral(c *MonetaryLiteralContext)

	// ExitAccountLiteral is called when exiting the accountLiteral production.
	ExitAccountLiteral(c *AccountLiteralContext)

	// ExitVariableLiteral is called when exiting the variableLiteral production.
	ExitVariableLiteral(c *VariableLiteralContext)

	// ExitPortionLiteral is called when exiting the portionLiteral production.
	ExitPortionLiteral(c *PortionLiteralContext)

	// ExitNumberLiteral is called when exiting the numberLiteral production.
	ExitNumberLiteral(c *NumberLiteralContext)

	// ExitRatio is called when exiting the ratio production.
	ExitRatio(c *RatioContext)

	// ExitPercentage is called when exiting the percentage production.
	ExitPercentage(c *PercentageContext)

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

	// ExitNumber is called when exiting the number production.
	ExitNumber(c *NumberContext)

	// ExitNumberVariable is called when exiting the numberVariable production.
	ExitNumberVariable(c *NumberVariableContext)

	// ExitAsset is called when exiting the asset production.
	ExitAsset(c *AssetContext)

	// ExitAssetVariable is called when exiting the assetVariable production.
	ExitAssetVariable(c *AssetVariableContext)

	// ExitAccountName is called when exiting the accountName production.
	ExitAccountName(c *AccountNameContext)

	// ExitAccountVariable is called when exiting the accountVariable production.
	ExitAccountVariable(c *AccountVariableContext)

	// ExitMonetary is called when exiting the monetary production.
	ExitMonetary(c *MonetaryContext)

	// ExitMonetaryVariable is called when exiting the monetaryVariable production.
	ExitMonetaryVariable(c *MonetaryVariableContext)

	// ExitMonetaryLit is called when exiting the monetaryLit production.
	ExitMonetaryLit(c *MonetaryLitContext)

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

	// ExitSrcVariable is called when exiting the srcVariable production.
	ExitSrcVariable(c *SrcVariableContext)

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

	// ExitDestAccount is called when exiting the destAccount production.
	ExitDestAccount(c *DestAccountContext)

	// ExitDestVariable is called when exiting the destVariable production.
	ExitDestVariable(c *DestVariableContext)

	// ExitDestAllotment is called when exiting the destAllotment production.
	ExitDestAllotment(c *DestAllotmentContext)

	// ExitDestInorder is called when exiting the destInorder production.
	ExitDestInorder(c *DestInorderContext)

	// ExitAllotmentClauseDest is called when exiting the allotmentClauseDest production.
	ExitAllotmentClauseDest(c *AllotmentClauseDestContext)

	// ExitSentLiteral is called when exiting the sentLiteral production.
	ExitSentLiteral(c *SentLiteralContext)

	// ExitSentAll is called when exiting the sentAll production.
	ExitSentAll(c *SentAllContext)

	// ExitSendStatement is called when exiting the sendStatement production.
	ExitSendStatement(c *SendStatementContext)

	// ExitFnCallStatement is called when exiting the fnCallStatement production.
	ExitFnCallStatement(c *FnCallStatementContext)
}
