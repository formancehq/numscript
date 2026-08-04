package vm

import (
	"fmt"
	"math/big"
)

type (
	ExecutionError interface {
		error
		execErr()
	}

	MissingFundsError struct {
		Asset  string
		Needed *big.Int
		Got    *big.Int
	}

	AssetMismatchError struct {
		Expected string
		Got      string
	}

	InvalidUncappedSource struct {
		Account string
	}

	InvalidAllotmentSum struct {
		ActualSum big.Rat
	}

	MetadataNotFoundError struct {
		Account string
		Key     string
	}

	BadMetaValueError struct {
		Account string
		Key     string
		Raw     string
	}

	InvalidAccountName struct {
		Name string
	}

	InvalidColor struct {
		Color string
	}

	InvalidScope struct {
		Scope string
	}

	NegativeBalanceError struct {
		Account string
		Amount  big.Int
	}

	DivideByZeroError struct {
		Numerator big.Int
	}

	// InternalError signals a malformed program the VM cannot execute: a bug in
	// whatever produced the bytecode, never a user-script error. Returned rather
	// than panicked so the VM never crashes its host.
	//
	// The mark violations land here rather than getting user-facing error types,
	// since both are properties a verifier can decide from the instruction stream
	// alone and neither is a legitimate outcome of a well-formed program.
	InternalError struct {
		Err error
	}

	// StoreError wraps an error returned by the host Store (balance or metadata
	// fetch): neither a script error nor a bytecode bug, so the wrapped error is
	// preserved for the host to inspect.
	StoreError struct {
		Wrapped error
	}
)

func (e MissingFundsError) Error() string {
	return fmt.Sprintf("missing funds for asset %s: needed %s, got %s", e.Asset, e.Needed, e.Got)
}

func (e AssetMismatchError) Error() string {
	return fmt.Sprintf("asset mismatch: expected %s, got %s", e.Expected, e.Got)
}

func (e InvalidUncappedSource) Error() string {
	return fmt.Sprintf("unbounded source is not allowed here: @%s", e.Account)
}

func (e InternalError) Error() string {
	return "internal error: " + e.Err.Error()
}

func (e InternalError) Unwrap() error { return e.Err }

func (e DivideByZeroError) Error() string {
	return fmt.Sprintf("cannot divide by zero (in %s/0)", e.Numerator.String())
}

func (e InvalidAccountName) Error() string {
	return fmt.Sprintf("invalid account name: %q", e.Name)
}

func (e InvalidColor) Error() string {
	return fmt.Sprintf("invalid color name: %q", e.Color)
}

func (e InvalidScope) Error() string {
	return fmt.Sprintf("invalid scope name: %q", e.Scope)
}

func (e NegativeBalanceError) Error() string {
	return fmt.Sprintf("cannot fetch negative balance from account @%s", e.Account)
}

func (e InvalidAllotmentSum) Error() string {
	return fmt.Sprintf("invalid allotment: portions must sum to 1, got %s", e.ActualSum.String())
}

func (e MetadataNotFoundError) Error() string {
	return fmt.Sprintf("metadata not found: %s[%q]", e.Account, e.Key)
}

func (e BadMetaValueError) Error() string {
	return fmt.Sprintf("invalid metadata value for %s[%q]: %q", e.Account, e.Key, e.Raw)
}

func (e StoreError) Error() string { return "store error: " + e.Wrapped.Error() }
func (e StoreError) Unwrap() error { return e.Wrapped }

func (MissingFundsError) execErr()     {}
func (AssetMismatchError) execErr()    {}
func (InvalidUncappedSource) execErr() {}
func (InvalidAllotmentSum) execErr()   {}
func (MetadataNotFoundError) execErr() {}
func (BadMetaValueError) execErr()     {}
func (InvalidAccountName) execErr()    {}
func (InvalidColor) execErr()          {}
func (InvalidScope) execErr()          {}
func (NegativeBalanceError) execErr()  {}
func (DivideByZeroError) execErr()     {}
func (InternalError) execErr()         {}
func (StoreError) execErr()            {}

var (
	_ ExecutionError = (*MissingFundsError)(nil)
	_ ExecutionError = (*AssetMismatchError)(nil)
	_ ExecutionError = (*InvalidUncappedSource)(nil)
	_ ExecutionError = (*InvalidAllotmentSum)(nil)
	_ ExecutionError = (*MetadataNotFoundError)(nil)
	_ ExecutionError = (*BadMetaValueError)(nil)
	_ ExecutionError = (*InvalidAccountName)(nil)
	_ ExecutionError = (*InvalidColor)(nil)
	_ ExecutionError = (*InvalidScope)(nil)
	_ ExecutionError = (*NegativeBalanceError)(nil)
	_ ExecutionError = (*DivideByZeroError)(nil)
	_ ExecutionError = (*InternalError)(nil)
	_ ExecutionError = (*StoreError)(nil)
)
