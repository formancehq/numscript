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

	NegativeBalanceError struct {
		Account string
		Amount  big.Int
	}

	DivideByZeroError struct {
		Numerator big.Int
	}

	// InternalError signals a malformed program the VM cannot execute (e.g. an
	// unknown opcode). It is a bug in whatever produced the bytecode, never a
	// user-script error, but it is returned rather than panicked so the VM never
	// crashes its host.
	InternalError struct {
		Opcode byte
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
	return fmt.Sprintf("internal error: unknown opcode %d", e.Opcode)
}

func (e DivideByZeroError) Error() string {
	return fmt.Sprintf("cannot divide by zero (in %s/0)", e.Numerator.String())
}

func (e InvalidAccountName) Error() string {
	return fmt.Sprintf("invalid account name: %q", e.Name)
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

func (MissingFundsError) execErr()     {}
func (AssetMismatchError) execErr()    {}
func (InvalidUncappedSource) execErr() {}
func (InvalidAllotmentSum) execErr()   {}
func (MetadataNotFoundError) execErr() {}
func (BadMetaValueError) execErr()     {}
func (InvalidAccountName) execErr()    {}
func (NegativeBalanceError) execErr()  {}
func (DivideByZeroError) execErr()     {}
func (InternalError) execErr()         {}

var (
	_ ExecutionError = (*MissingFundsError)(nil)
	_ ExecutionError = (*AssetMismatchError)(nil)
	_ ExecutionError = (*InvalidUncappedSource)(nil)
	_ ExecutionError = (*InvalidAllotmentSum)(nil)
	_ ExecutionError = (*MetadataNotFoundError)(nil)
	_ ExecutionError = (*BadMetaValueError)(nil)
	_ ExecutionError = (*InvalidAccountName)(nil)
	_ ExecutionError = (*NegativeBalanceError)(nil)
	_ ExecutionError = (*DivideByZeroError)(nil)
	_ ExecutionError = (*InternalError)(nil)
)
