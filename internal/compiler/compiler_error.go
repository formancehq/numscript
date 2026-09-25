package compiler

import (
	"fmt"

	"github.com/formancehq/numscript/internal/flags"
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

	// CannotStoreScopedAccountInMeta is reported when a scoped account (the
	// result of scoped()) is used as the *value* stored by set_tx_meta or
	// set_account_meta. Mirrors the interpreter's runtime error of the same name,
	// but caught at compile time since the compiler already knows an expression's
	// scopedness statically.
	CannotStoreScopedAccountInMeta struct {
		parser.Range
	}

	// InvalidScopedAccountPosition is reported when scoped() is reached from a
	// position that only wants a plain string/generic value (e.g. account
	// interpolation, or any other non-account context) — a defensive check that
	// should be unreachable given the compiler's other call sites already route
	// account-typed expressions through compileAccountExpr.
	InvalidScopedAccountPosition struct {
		parser.Range
	}

	// FeatureNotImplemented is returned (never panicked) when the compiler meets a
	// construct it does not support yet — e.g. colors or scoped accounts — so the
	// host gets an error instead of a crash.
	FeatureNotImplemented struct {
		parser.Range
		Feature string
	}

	// ExperimentalFeature is reported when the script uses a construct gated
	// behind a feature flag that wasn't enabled. Mirrors the interpreter's
	// interpreter.ExperimentalFeature.
	ExperimentalFeature struct {
		parser.Range
		FlagName flags.FeatureFlag
	}

	// InvalidFeature is reported when a #![feature(..)] declaration names a flag
	// that doesn't exist.
	InvalidFeature struct {
		parser.Range
		Feature string
	}
)

func (UnboundVar) compileError()                     {}
func (TypeError) compileError()                      {}
func (InvalidUncappedSource) compileError()          {}
func (DuplicateRemaining) compileError()             {}
func (InvalidMetaPosition) compileError()            {}
func (CannotCastToString) compileError()             {}
func (CannotStoreScopedAccountInMeta) compileError() {}
func (InvalidScopedAccountPosition) compileError()   {}
func (FeatureNotImplemented) compileError()          {}
func (ExperimentalFeature) compileError()            {}
func (InvalidFeature) compileError()                 {}

func (e FeatureNotImplemented) Error() string {
	return "internal error: feature not implemented: " + e.Feature
}
func (e UnboundVar) Error() string {
	return fmt.Sprintf("the variable '$%s' was not declared", e.Var)
}
func (InvalidUncappedSource) Error() string {
	return "cannot take all balance of an unbounded source"
}
func (DuplicateRemaining) Error() string {
	return "a 'remaining' clause should be the last in an allotment expression"
}
func (e TypeError) Error() string { return e.Kind.Message() }
func (InvalidMetaPosition) Error() string {
	return "meta() is only allowed as a variable origin"
}
func (e CannotCastToString) Error() string {
	return "cannot cast a value of type " + string(e.Type) + " to string"
}
func (CannotStoreScopedAccountInMeta) Error() string {
	return "cannot store a scoped account as a metadata value"
}
func (InvalidScopedAccountPosition) Error() string {
	return "a scoped account cannot be used here"
}
func (e ExperimentalFeature) Error() string {
	return fmt.Sprintf("this feature is experimental. You need the '%s' feature flag to enable it", e.FlagName)
}
func (e InvalidFeature) Error() string {
	return fmt.Sprintf("Invalid feature: %s", e.Feature)
}

var (
	_ CompilerError = (*UnboundVar)(nil)
	_ CompilerError = (*TypeError)(nil)
	_ CompilerError = (*InvalidUncappedSource)(nil)
	_ CompilerError = (*DuplicateRemaining)(nil)
	_ CompilerError = (*InvalidMetaPosition)(nil)
	_ CompilerError = (*CannotCastToString)(nil)
	_ CompilerError = (*CannotStoreScopedAccountInMeta)(nil)
	_ CompilerError = (*InvalidScopedAccountPosition)(nil)
	_ CompilerError = (*FeatureNotImplemented)(nil)
	_ CompilerError = (*ExperimentalFeature)(nil)
	_ CompilerError = (*InvalidFeature)(nil)
)
