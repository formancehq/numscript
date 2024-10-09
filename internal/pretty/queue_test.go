package pretty_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/pretty"

	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	q := pretty.NewQueue[int]()
	q.PushFront(1)
	q.PushFront(2)
	q.PushFront(3)

	require.False(t, q.IsEmpty())

	require.Equal(t, 1, q.Pop())
	require.Equal(t, 2, q.Pop())
	require.Equal(t, 3, q.Pop())

	require.True(t, q.IsEmpty())
}
