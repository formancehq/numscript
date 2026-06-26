package vm

import "math/big"

type (
	ExecutionError interface {
		execErr()
	}

	MissingFundsError struct {
		Asset    string
		Expected *big.Int
		Got      *big.Int
	}

	AssetMismatchError struct {
		Expected string
		Got      string
	}

	InvalidUncappedSource struct {
		Account string
	}
)

func (MissingFundsError) execErr()     {}
func (AssetMismatchError) execErr()    {}
func (InvalidUncappedSource) execErr() {}

var (
	_ ExecutionError = (*MissingFundsError)(nil)
	_ ExecutionError = (*AssetMismatchError)(nil)
	_ ExecutionError = (*InvalidUncappedSource)(nil)
)
