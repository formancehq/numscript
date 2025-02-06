package utils

import (
	"fmt"
	"math/big"
)

func NonExhaustiveMatchPanic[T any](value any) T {
	panic(fmt.Sprintf("Non exhaustive match (got '%#v')", value))
}

func MinBigInt(a *big.Int, b *big.Int) *big.Int {
	var min big.Int

	if a.Cmp(b) == -1 {
		min.Set(a)
	} else {
		min.Set(b)
	}

	return &min
}

func MaxBigInt(a *big.Int, b *big.Int) *big.Int {
	var max big.Int

	if a.Cmp(b) == 1 {
		max.Set(a)
	} else {
		max.Set(b)
	}

	return &max
}
