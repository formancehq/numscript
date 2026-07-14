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

	// InvalidMetaPosition is reported when meta() appears anywhere other than as
	// a top-level variable origin (the only place it's supported).
	InvalidMetaPosition struct {
		parser.Range
	}
)

func (UnboundVar) compileError()            {}
func (TypeError) compileError()             {}
func (InvalidUncappedSource) compileError() {}
func (DuplicateRemaining) compileError()    {}
func (InvalidMetaPosition) compileError()   {}

func (e TypeError) Error() string { return e.Kind.Message() }
func (InvalidMetaPosition) Error() string {
	return "meta() is only allowed as a variable origin"
}

var (
	_ CompilerError = (*UnboundVar)(nil)
	_ CompilerError = (*TypeError)(nil)
	_ CompilerError = (*InvalidUncappedSource)(nil)
	_ CompilerError = (*DuplicateRemaining)(nil)
	_ CompilerError = (*InvalidMetaPosition)(nil)
)
