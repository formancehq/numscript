package utils

import (
	"encoding/json"
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

func Unmarshal[T any](raw json.RawMessage) (*T, error) {
	var value T
	err := json.Unmarshal(raw, &value)
	if err != nil {
		return nil, err
	}
	return &value, err
}

func Filter[T any](slice []T, predicate func(x T) bool) []T {
	var ret []T
	for _, x := range slice {
		if predicate(x) {
			ret = append(ret, x)
		}
	}
	return ret
}
