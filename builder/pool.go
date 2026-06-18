package builder

type pool[T comparable] struct {
	elems map[T]int
}

func newPool[T comparable]() pool[T] {
	return pool[T]{
		elems: make(map[T]int),
	}
}
