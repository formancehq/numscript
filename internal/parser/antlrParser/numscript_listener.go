// Code generated from Numscript.g4 by ANTLR 4.13.2. DO NOT EDIT.

package antlrParser // Numscript
import "github.com/antlr4-go/antlr/v4"

// NumscriptListener is a complete listener for a parse tree produced by NumscriptParser.
type NumscriptListener interface {
	antlr.ParseTreeListener

	// EnterMonetaryLit is called when entering the monetaryLit production.
	EnterMonetaryLit(c *MonetaryLitContext)

	// EnterAccountTextPart is called when entering the accountTextPart production.
	EnterAccountTextPart(c *AccountTextPartContext)

	// EnterAccountVarPart is called when entering the accountVarPart production.
	EnterAccountVarPart(c *AccountVarPartContext)

	// EnterVariableExpr is called when entering the variableExpr production.
	EnterVariableExpr(c *VariableExprContext)

	// EnterInfixExpr is called when entering the infixExpr production.
	EnterInfixExpr(c *InfixExprContext)

	// EnterApplication is called when entering the application production.
	EnterApplication(c *ApplicationContext)

	// EnterAssetLiteral is called when entering the assetLiteral production.
	EnterAssetLiteral(c *AssetLiteralContext)

	// EnterStringLiteral is called when entering the stringLiteral production.
	EnterStringLiteral(c *StringLiteralContext)

	// EnterParenthesizedExpr is called when entering the parenthesizedExpr production.
	EnterParenthesizedExpr(c *ParenthesizedExprContext)

	// EnterAccountLiteral is called when entering the accountLiteral production.
	EnterAccountLiteral(c *AccountLiteralContext)

	// EnterMonetaryLiteral is called when entering the monetaryLiteral production.
	EnterMonetaryLiteral(c *MonetaryLiteralContext)

	// EnterNumberLiteral is called when entering the numberLiteral production.
	EnterNumberLiteral(c *NumberLiteralContext)

	// EnterPercentagePortionLiteral is called when entering the percentagePortionLiteral production.
	EnterPercentagePortionLiteral(c *PercentagePortionLiteralContext)

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

	// EnterPortionedAllotment is called when entering the portionedAllotment production.
	EnterPortionedAllotment(c *PortionedAllotmentContext)

	// EnterRemainingAllotment is called when entering the remainingAllotment production.
	EnterRemainingAllotment(c *RemainingAllotmentContext)

	// EnterColorConstraint is called when entering the colorConstraint production.
	EnterColorConstraint(c *ColorConstraintContext)

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

	// EnterSrcOneof is called when entering the srcOneof production.
	EnterSrcOneof(c *SrcOneofContext)

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

	// EnterDestAllotment is called when entering the destAllotment production.
	EnterDestAllotment(c *DestAllotmentContext)

	// EnterDestInorder is called when entering the destInorder production.
	EnterDestInorder(c *DestInorderContext)

	// EnterDestOneof is called when entering the destOneof production.
	EnterDestOneof(c *DestOneofContext)

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

	// ExitAccountTextPart is called when exiting the accountTextPart production.
	ExitAccountTextPart(c *AccountTextPartContext)

	// ExitAccountVarPart is called when exiting the accountVarPart production.
	ExitAccountVarPart(c *AccountVarPartContext)

	// ExitVariableExpr is called when exiting the variableExpr production.
	ExitVariableExpr(c *VariableExprContext)

	// ExitInfixExpr is called when exiting the infixExpr production.
	ExitInfixExpr(c *InfixExprContext)

	// ExitApplication is called when exiting the application production.
	ExitApplication(c *ApplicationContext)

	// ExitAssetLiteral is called when exiting the assetLiteral production.
	ExitAssetLiteral(c *AssetLiteralContext)

	// ExitStringLiteral is called when exiting the stringLiteral production.
	ExitStringLiteral(c *StringLiteralContext)

	// ExitParenthesizedExpr is called when exiting the parenthesizedExpr production.
	ExitParenthesizedExpr(c *ParenthesizedExprContext)

	// ExitAccountLiteral is called when exiting the accountLiteral production.
	ExitAccountLiteral(c *AccountLiteralContext)

	// ExitMonetaryLiteral is called when exiting the monetaryLiteral production.
	ExitMonetaryLiteral(c *MonetaryLiteralContext)

	// ExitNumberLiteral is called when exiting the numberLiteral production.
	ExitNumberLiteral(c *NumberLiteralContext)

	// ExitPercentagePortionLiteral is called when exiting the percentagePortionLiteral production.
	ExitPercentagePortionLiteral(c *PercentagePortionLiteralContext)

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

	// ExitPortionedAllotment is called when exiting the portionedAllotment production.
	ExitPortionedAllotment(c *PortionedAllotmentContext)

	// ExitRemainingAllotment is called when exiting the remainingAllotment production.
	ExitRemainingAllotment(c *RemainingAllotmentContext)

	// ExitColorConstraint is called when exiting the colorConstraint production.
	ExitColorConstraint(c *ColorConstraintContext)

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

	// ExitSrcOneof is called when exiting the srcOneof production.
	ExitSrcOneof(c *SrcOneofContext)

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

	// ExitDestAllotment is called when exiting the destAllotment production.
	ExitDestAllotment(c *DestAllotmentContext)

	// ExitDestInorder is called when exiting the destInorder production.
	ExitDestInorder(c *DestInorderContext)

	// ExitDestOneof is called when exiting the destOneof production.
	ExitDestOneof(c *DestOneofContext)

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
