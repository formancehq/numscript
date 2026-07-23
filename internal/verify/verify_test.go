package verify

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func requireZ3(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("z3"); err != nil {
		t.Skip("z3 not found on PATH; skipping solver-dependent test")
	}
}

func run(t *testing.T, source, query string) *Result {
	t.Helper()
	res, err := Verify(context.Background(), source, query, Options{})
	require.NoError(t, err)
	return res
}

const simple = `
	send [USD/2 10] (
		source = @src
		destination = @dest
	)
`

func TestSimpleProved(t *testing.T) {
	requireZ3(t)
	// If it doesn't fail, @dest always receives exactly 10.
	res := run(t, simple, `prove: !fail => received("dest", "USD/2") == 10`)
	require.Equal(t, Proved, res.Verdict, "raw: %s", res.Raw)
}

func TestSimpleCounterexample(t *testing.T) {
	requireZ3(t)
	// @dest does NOT always receive 5 — there's a model where it doesn't.
	res := run(t, simple, `prove: received("dest", "USD/2") == 5`)
	require.Equal(t, Counterexample, res.Verdict, "raw: %s", res.Raw)
	require.NotEmpty(t, res.Model)
}

func TestSimpleFailPossible(t *testing.T) {
	requireZ3(t)
	// It CAN fail (when @src starts with < 10).
	res := run(t, simple, `find: fail`)
	require.Equal(t, Witness, res.Verdict, "raw: %s", res.Raw)
}

func TestSimpleConservation(t *testing.T) {
	requireZ3(t)
	// Conservation: what @src sends equals what @dest receives, always.
	res := run(t, simple, `prove: sent("src", "USD/2") == received("dest", "USD/2")`)
	require.Equal(t, Proved, res.Verdict, "raw: %s", res.Raw)
}

func TestSimpleFailImpliesNoReceipt(t *testing.T) {
	requireZ3(t)
	// On failure, nothing is received (abort discards postings).
	res := run(t, simple, `prove: fail => received("dest", "USD/2") == 0`)
	require.Equal(t, Proved, res.Verdict, "raw: %s", res.Raw)
}

const world = `
	send [USD/2 10] (
		source = @world
		destination = {
			max [USD/2 4] to @d1
			remaining to @d2
		}
	)
`

func TestWorldNeverFails(t *testing.T) {
	requireZ3(t)
	// @world is infinite, so this can never fail.
	res := run(t, world, `find: fail`)
	require.Equal(t, Impossible, res.Verdict, "raw: %s", res.Raw)
}

func TestWorldDestSplit(t *testing.T) {
	requireZ3(t)
	res := run(t, world, `prove: received("d1", "USD/2") == 4 && received("d2", "USD/2") == 6`)
	require.Equal(t, Proved, res.Verdict, "raw: %s", res.Raw)
}

const inorder = `
	send [USD/2 10] (
		source = {
			@a
			@b
			@c
		}
		destination = @dest
	)
`

func TestInorderConservation(t *testing.T) {
	requireZ3(t)
	// If it succeeds, the three sources together provide exactly 10.
	res := run(t, inorder,
		`prove: !fail => (sent("a","USD/2") + sent("b","USD/2") + sent("c","USD/2")) == 10`)
	require.Equal(t, Proved, res.Verdict, "raw: %s", res.Raw)
}

const allot = `
	send [USD/2 10] (
		source = @world
		destination = {
			1/3 to @a
			2/3 to @b
		}
	)
`

func TestAllotmentSplit(t *testing.T) {
	requireZ3(t)
	// floor(10/3)=3, +1 carry = 4 to @a; 2/3 = 6 to @b.
	res := run(t, allot, `prove: received("a","USD/2") == 4 && received("b","USD/2") == 6`)
	require.Equal(t, Proved, res.Verdict, "raw: %s", res.Raw)
}

func TestAllotmentConserves(t *testing.T) {
	requireZ3(t)
	// A split never loses or invents money: the parts always sum to the whole.
	res := run(t, allot, `prove: received("a","USD/2") + received("b","USD/2") == 10`)
	require.Equal(t, Proved, res.Verdict, "raw: %s", res.Raw)
}

func TestQueryTypeError(t *testing.T) {
	// A non-boolean top-level query is a usage error (no z3 needed).
	_, err := Verify(context.Background(), simple, `received("dest", "USD/2")`, Options{})
	require.Error(t, err)
}

func TestMismatchedAssetComparison(t *testing.T) {
	_, err := Verify(context.Background(), simple,
		`mon("USD/2", 1) == mon("EUR/2", 1)`, Options{})
	require.Error(t, err)
}
