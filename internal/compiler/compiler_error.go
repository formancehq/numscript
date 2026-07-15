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

	// CannotCastToString is reported for an interpolation part whose type has no
	// string form (monetary, asset, portion).
	CannotCastToString struct {
		parser.Range
		Type typecheck.Type
	}

	// FeatureNotImplemented is returned (never panicked) when the compiler meets a
	// construct it does not support yet — e.g. colors or scoped accounts — so the
	// host gets an error instead of a crash.
	FeatureNotImplemented struct {
		parser.Range
		Feature string
	}
)

func (UnboundVar) compileError()            {}
func (TypeError) compileError()             {}
func (InvalidUncappedSource) compileError() {}
func (DuplicateRemaining) compileError()    {}
func (InvalidMetaPosition) compileError()   {}
func (CannotCastToString) compileError()    {}
func (FeatureNotImplemented) compileError() {}

func (e FeatureNotImplemented) Error() string {
	return "internal error: feature not implemented: " + e.Feature
}
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
	_ CompilerError = (*FeatureNotImplemented)(nil)
)
