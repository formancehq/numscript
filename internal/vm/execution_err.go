package vm

import (
	"fmt"
	"math/big"
)

type (
	ExecutionError interface {
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
)

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

var (
	_ ExecutionError = (*MissingFundsError)(nil)
	_ ExecutionError = (*AssetMismatchError)(nil)
	_ ExecutionError = (*InvalidUncappedSource)(nil)
	_ ExecutionError = (*InvalidAllotmentSum)(nil)
	_ ExecutionError = (*MetadataNotFoundError)(nil)
	_ ExecutionError = (*BadMetaValueError)(nil)
)
