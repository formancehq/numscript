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

func NonNeg(a *big.Int) *big.Int {
	return MaxBigInt(a, big.NewInt(0))
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

func MapCmp[T any](m1, m2 map[string]T, cmp func(x1 T, x2 T) bool) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k1, v1 := range m1 {
		v2, ok := m2[k1]
		if !ok || !cmp(v1, v2) {
			return false
		}
	}

	return true
}

func Map2Cmp[T any](m1, m2 map[string]map[string]T, cmp func(x1 T, x2 T) bool) bool {
	return MapCmp(m1, m2, func(nested1, nested2 map[string]T) bool {
		return MapCmp(nested1, nested2, cmp)
	})
}
