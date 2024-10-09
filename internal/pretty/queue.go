package pretty

import "slices"

// A LIFO queue
type Queue[T any] struct {
	items []T
}

func NewQueue[T any]() Queue[T] {
	return Queue[T]{}
}

func NewQueueOf[T any](ts ...T) Queue[T] {
	slices.Reverse(ts)
	q := NewQueue[T]()
	for _, x := range ts {
		q.PushFront(x)
	}
	return q
}

func (q *Queue[T]) PushFront(x T) {
	q.items = append(q.items, x)
}

// unsafe pop() operation: this will panic on empty queue
func (q *Queue[T]) Pop() T {
	// TODO double check this is O(1)
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q Queue[any]) IsEmpty() bool {
	return len(q.items) == 0
}

func (q Queue[T]) Clone() Queue[T] {
	return Queue[T]{
		items: slices.Clone(q.items),
	}
}
