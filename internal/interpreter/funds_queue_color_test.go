package interpreter

// White-box tests for the color-awareness of the internal funds queue.
// These live in `package interpreter` because they reach into unexported
// types (Sender, newFundsQueue). The public color semantics — what the
// numscript ↔ ledger contract actually exposes — are covered in
// color_semantics_test.go (black-box).

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// compactTop() relies on (Name, Color) equality. Adjacent senders that
// match on both dimensions collapse; those that disagree on either don't.
func TestFundsQueueCompactRespectsColor(t *testing.T) {
	t.Parallel()

	queue := newFundsQueue([]Sender{
		{Name: "a", Color: "RED", Amount: big.NewInt(10)},
		{Name: "a", Color: "RED", Amount: big.NewInt(5)},
		{Name: "a", Color: "BLUE", Amount: big.NewInt(7)},
	})

	out := queue.PullAnything(big.NewInt(22))
	require.Equal(t, []Sender{
		{Name: "a", Color: "RED", Amount: big.NewInt(15)},
		{Name: "a", Color: "BLUE", Amount: big.NewInt(7)},
	}, out)
}

// PullColored only pulls senders that match the requested color, leaving
// the rest of the queue untouched and still drainable on subsequent pulls.
func TestFundsQueuePullColoredIsSelective(t *testing.T) {
	t.Parallel()

	queue := newFundsQueue([]Sender{
		{Name: "a", Color: "RED", Amount: big.NewInt(10)},
		{Name: "b", Color: "BLUE", Amount: big.NewInt(20)},
		{Name: "c", Color: "RED", Amount: big.NewInt(30)},
	})

	out := queue.PullColored(big.NewInt(35), "RED")
	require.Equal(t, []Sender{
		{Name: "a", Color: "RED", Amount: big.NewInt(10)},
		{Name: "c", Color: "RED", Amount: big.NewInt(25)},
	}, out)

	remaining := queue.PullColored(big.NewInt(20), "BLUE")
	require.Equal(t, []Sender{
		{Name: "b", Color: "BLUE", Amount: big.NewInt(20)},
	}, remaining)
}

// PullUncolored is just PullColored("") — the empty bucket is the same as
// any other color from the queue's perspective.
func TestFundsQueuePullUncoloredIgnoresColored(t *testing.T) {
	t.Parallel()

	queue := newFundsQueue([]Sender{
		{Name: "a", Color: "RED", Amount: big.NewInt(100)},
		{Name: "b", Color: "", Amount: big.NewInt(40)},
	})

	out := queue.PullUncolored(big.NewInt(40))
	require.Equal(t, []Sender{
		{Name: "b", Color: "", Amount: big.NewInt(40)},
	}, out)

	// the RED sender is still there
	remaining := queue.PullColored(big.NewInt(50), "RED")
	require.Equal(t, []Sender{
		{Name: "a", Color: "RED", Amount: big.NewInt(50)},
	}, remaining)
}
