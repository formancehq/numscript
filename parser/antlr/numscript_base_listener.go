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

// EnterProgram is called when production program is entered.
func (s *BaseNumscriptListener) EnterProgram(ctx *ProgramContext) {}

// ExitProgram is called when production program is exited.
func (s *BaseNumscriptListener) ExitProgram(ctx *ProgramContext) {}

// EnterMonetaryLit is called when production monetaryLit is entered.
func (s *BaseNumscriptListener) EnterMonetaryLit(ctx *MonetaryLitContext) {}

// ExitMonetaryLit is called when production monetaryLit is exited.
func (s *BaseNumscriptListener) ExitMonetaryLit(ctx *MonetaryLitContext) {}

// EnterSource is called when production source is entered.
func (s *BaseNumscriptListener) EnterSource(ctx *SourceContext) {}

// ExitSource is called when production source is exited.
func (s *BaseNumscriptListener) ExitSource(ctx *SourceContext) {}

// EnterStatement is called when production statement is entered.
func (s *BaseNumscriptListener) EnterStatement(ctx *StatementContext) {}

// ExitStatement is called when production statement is exited.
func (s *BaseNumscriptListener) ExitStatement(ctx *StatementContext) {}
