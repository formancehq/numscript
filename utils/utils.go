package utils

import "fmt"

func NonExhaustiveMatchPanic[T any](value any) T {
	panic(fmt.Sprintf("Non exhaustive match (got '%#v')", value))
}
