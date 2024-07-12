// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

type BaseNumscriptVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseNumscriptVisitor) VisitRatio(ctx *RatioContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitPercentage(ctx *PercentageContext) interface{} {
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

func (v *BaseNumscriptVisitor) VisitMonetaryLit(ctx *MonetaryLitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitLitCap(ctx *LitCapContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitVarCap(ctx *VarCapContext) interface{} {
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

func (v *BaseNumscriptVisitor) VisitSrcSeq(ctx *SrcSeqContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSrcCapped(ctx *SrcCappedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitAllotmentClauseSrc(ctx *AllotmentClauseSrcContext) interface{} {
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

func (v *BaseNumscriptVisitor) VisitDestSeq(ctx *DestSeqContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitAllotmentClauseDest(ctx *AllotmentClauseDestContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitStatement(ctx *StatementContext) interface{} {
	return v.VisitChildren(ctx)
}
