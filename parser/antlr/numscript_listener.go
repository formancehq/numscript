// Code generated from Numscript.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Numscript

import "github.com/antlr4-go/antlr/v4"

// NumscriptListener is a complete listener for a parse tree produced by NumscriptParser.
type NumscriptListener interface {
	antlr.ParseTreeListener

	// EnterProgram is called when entering the program production.
	EnterProgram(c *ProgramContext)

	// EnterMonetaryLit is called when entering the monetaryLit production.
	EnterMonetaryLit(c *MonetaryLitContext)

	// EnterAccount is called when entering the account production.
	EnterAccount(c *AccountContext)

	// EnterVariable is called when entering the variable production.
	EnterVariable(c *VariableContext)

	// EnterStatement is called when entering the statement production.
	EnterStatement(c *StatementContext)

	// ExitProgram is called when exiting the program production.
	ExitProgram(c *ProgramContext)

	// ExitMonetaryLit is called when exiting the monetaryLit production.
	ExitMonetaryLit(c *MonetaryLitContext)

	// ExitAccount is called when exiting the account production.
	ExitAccount(c *AccountContext)

	// ExitVariable is called when exiting the variable production.
	ExitVariable(c *VariableContext)

	// ExitStatement is called when exiting the statement production.
	ExitStatement(c *StatementContext)
}
