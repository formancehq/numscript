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

	// CannotCastToString is reported when an account-interpolation part has a
	// type that has no string representation (monetary, asset, portion). Mirrors
	// the interpreter's runtime error of the same name; the compiler catches it
	// statically since the part's type is known at compile time.
	CannotCastToString struct {
		parser.Range
		Type typecheck.Type
	}
)

func (UnboundVar) compileError()            {}
func (TypeError) compileError()             {}
func (InvalidUncappedSource) compileError() {}
func (DuplicateRemaining) compileError()    {}
func (InvalidMetaPosition) compileError()   {}
func (CannotCastToString) compileError()    {}

func (e TypeError) Error() string { return e.Kind.Message() }
func (InvalidMetaPosition) Error() string {
	return "meta() is only allowed as a variable origin"
}
func (e CannotCastToString) Error() string {
	return "cannot cast a value of type " + string(e.Type) + " to string"
}

var (
	_ CompilerError = (*UnboundVar)(nil)
	_ CompilerError = (*TypeError)(nil)
	_ CompilerError = (*InvalidUncappedSource)(nil)
	_ CompilerError = (*DuplicateRemaining)(nil)
	_ CompilerError = (*InvalidMetaPosition)(nil)
	_ CompilerError = (*CannotCastToString)(nil)
)
