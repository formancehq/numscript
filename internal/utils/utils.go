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

func NonNeg(a *big.Int) *big.Int {
	return MaxBigInt(a, big.NewInt(0))
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

func Map[T any, U any](slice []T, f func(x T) U) []U {
	// TODO make
	var ret []U
	for _, x := range slice {
		ret = append(ret, f(x))
	}
	return ret
}

func MapGetOrPutDefault[T any](m map[string]T, key string, getDefault func() T) T {
	lookup, ok := m[key]
	if !ok {
		default_ := getDefault()
		m[key] = default_
		return default_
	}
	return lookup
}

func NestedMapGetOrPutDefault[T any](m map[string]map[string]T, key1 string, key2 string, getDefault func() T) T {
	m1 := MapGetOrPutDefault(m, key1, func() map[string]T {
		return map[string]T{}
	})

	return MapGetOrPutDefault(m1, key2, getDefault)
}
