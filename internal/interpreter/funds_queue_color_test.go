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

// compactTop() relies on (Name, Color) equality: adjacent senders that match
// on both dimensions collapse, those that disagree on either don't. Color
// must NEVER be collapsed across — that would silently merge buckets that
// the source-level segregation is meant to keep apart.
func TestFundsQueueCompactDoesNotMergeAcrossColors(t *testing.T) {
	t.Parallel()

	queue := newFundsQueue([]Sender{
		{Name: "a", Color: "RED", Amount: big.NewInt(10)},
		{Name: "a", Color: "RED", Amount: big.NewInt(5)},
		{Name: "a", Color: "BLUE", Amount: big.NewInt(7)},
	})

	out := queue.PullAnything(big.NewInt(22))
	require.Equal(t, []Sender{
		// (a, RED) entries compact into a single 15 ✓
		{Name: "a", Color: "RED", Amount: big.NewInt(15)},
		// (a, BLUE) stays distinct ✓
		{Name: "a", Color: "BLUE", Amount: big.NewInt(7)},
	}, out)
}
