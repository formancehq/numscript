// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

type BaseNumscriptVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseNumscriptVisitor) VisitAssetLiteral(ctx *AssetLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitStringLiteral(ctx *StringLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitMonetaryLiteral(ctx *MonetaryLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitAccountLiteral(ctx *AccountLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitVariableLiteral(ctx *VariableLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitPortionLiteral(ctx *PortionLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitNumberLiteral(ctx *NumberLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitRatio(ctx *RatioContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitPercentage(ctx *PercentageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitFunctionCallArgs(ctx *FunctionCallArgsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitVarOrigin(ctx *VarOriginContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitVarDeclaration(ctx *VarDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitVarsDeclaration(ctx *VarsDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitProgram(ctx *ProgramContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitNumber(ctx *NumberContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitNumberVariable(ctx *NumberVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitAsset(ctx *AssetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitAssetVariable(ctx *AssetVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitAccountName(ctx *AccountNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitAccountVariable(ctx *AccountVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitMonetary(ctx *MonetaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitMonetaryVariable(ctx *MonetaryVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitMonetaryLit(ctx *MonetaryLitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSentAllLit(ctx *SentAllLitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitLitCap(ctx *LitCapContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitVarCap(ctx *VarCapContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitPortionedAllotment(ctx *PortionedAllotmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitPortionVariable(ctx *PortionVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitRemainingAllotment(ctx *RemainingAllotmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSrcAccountUnboundedOverdraft(ctx *SrcAccountUnboundedOverdraftContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSrcAccountBoundedOverdraft(ctx *SrcAccountBoundedOverdraftContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSrcAccount(ctx *SrcAccountContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSrcVariable(ctx *SrcVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSrcAllotment(ctx *SrcAllotmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSrcInorder(ctx *SrcInorderContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSrcCapped(ctx *SrcCappedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitAllotmentClauseSrc(ctx *AllotmentClauseSrcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitDestinationTo(ctx *DestinationToContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitDestinationKept(ctx *DestinationKeptContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitDestinationInOrderClause(ctx *DestinationInOrderClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitDestAccount(ctx *DestAccountContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitDestVariable(ctx *DestVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitDestAllotment(ctx *DestAllotmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitDestInorder(ctx *DestInorderContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitAllotmentClauseDest(ctx *AllotmentClauseDestContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSentLiteral(ctx *SentLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSentAll(ctx *SentAllContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSendStatement(ctx *SendStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitFnCallStatement(ctx *FnCallStatementContext) interface{} {
	return v.VisitChildren(ctx)
}
