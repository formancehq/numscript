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
)

func (MissingFundsError) execErr()  {}
func (AssetMismatchError) execErr() {}

var (
	_ ExecutionError = (*MissingFundsError)(nil)
	_ ExecutionError = (*AssetMismatchError)(nil)
)
