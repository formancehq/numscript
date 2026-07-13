package compiler

import (
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/typecheck"
)

type (
	CompilerError interface {
		parser.Ranged
		compileError()
	}

	UnboundVar struct {
		parser.Range
		Var string
	}

	TypeError struct {
		parser.Range
		Kind typecheck.ErrorKind
	}

	InvalidUncappedSource struct {
		parser.Range
	}

	DuplicateRemaining struct {
		parser.Range
	}
)

func (UnboundVar) compileError()            {}
func (TypeError) compileError()             {}
func (InvalidUncappedSource) compileError() {}
func (DuplicateRemaining) compileError()    {}

func (e TypeError) Error() string { return e.Kind.Message() }

var (
	_ CompilerError = (*UnboundVar)(nil)
	_ CompilerError = (*TypeError)(nil)
	_ CompilerError = (*InvalidUncappedSource)(nil)
	_ CompilerError = (*DuplicateRemaining)(nil)
)
