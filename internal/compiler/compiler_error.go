package compiler

import "github.com/formancehq/numscript/internal/parser"

type (
	CompilerError interface {
		parser.Ranged
		compileError()
	}

	UnboundVar struct {
		parser.Range
		Var string
	}

	TypeMismatch struct {
		parser.Range
		Expected string
		Got      string
	}

	InvalidUncappedSource struct {
		parser.Range
	}
)

func (UnboundVar) compileError()            {}
func (TypeMismatch) compileError()          {}
func (InvalidUncappedSource) compileError() {}

var (
	_ CompilerError = (*UnboundVar)(nil)
	_ CompilerError = (*TypeMismatch)(nil)
	_ CompilerError = (*InvalidUncappedSource)(nil)
)
