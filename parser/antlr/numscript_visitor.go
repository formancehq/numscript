// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by NumscriptParser.
type NumscriptVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by NumscriptParser#program.
	VisitProgram(ctx *ProgramContext) interface{}

	// Visit a parse tree produced by NumscriptParser#monetaryLit.
	VisitMonetaryLit(ctx *MonetaryLitContext) interface{}

	// Visit a parse tree produced by NumscriptParser#statement.
	VisitStatement(ctx *StatementContext) interface{}
}
