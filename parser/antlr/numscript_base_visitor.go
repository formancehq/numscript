// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

type BaseNumscriptVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseNumscriptVisitor) VisitProgram(ctx *ProgramContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitMonetaryLit(ctx *MonetaryLitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitSource(ctx *SourceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNumscriptVisitor) VisitStatement(ctx *StatementContext) interface{} {
	return v.VisitChildren(ctx)
}
