package compiler

import (
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

// virtualInstrs compiles source straight to the virtual-instruction stream.
func virtualInstrs(t *testing.T, src string) []vInstr {
	t.Helper()
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	compiled, cErr := compileProgramToVirtual(parsed.Value)
	require.Nil(t, cErr)
	return compiled.instructions
}

func hasOp[T vInstr](instrs []vInstr) bool {
	for _, in := range instrs {
		if _, ok := in.(T); ok {
			return true
		}
	}
	return false
}

// TestFundsBypass_SingleToSingle: a plain-account 1-source/1-dest send fuses to
// take+post, and no pull/send survives.
func TestFundsBypass_SingleToSingle(t *testing.T) {
	in := virtualInstrs(t, `send [USD 10] (source = @a destination = @b)`)
	out, changed := fundsBypass{}.run(in)
	require.True(t, changed)

	require.True(t, hasOp[takeAccount](out), "expected a take_account")
	require.True(t, hasOp[postAccount](out), "expected a post_account")
	require.False(t, hasOp[pullAccount](out), "pull_account should be gone")
	require.False(t, hasOp[sendToAccount](out), "send_to_account should be gone")

	// the take sits where the pull was (before the check); the post where the
	// send was (after the check, after the dst load) — nothing is reordered.
	d := dump(out)
	takeAt := strings.Index(d, "take_account")
	checkAt := strings.Index(d, "check_enough_funds")
	postAt := strings.Index(d, "post_account")
	require.True(t, takeAt >= 0 && checkAt >= 0 && postAt >= 0)
	require.Less(t, takeAt, checkAt, "take must precede the funds check")
	require.Less(t, checkAt, postAt, "the funds check must precede the post")
}

// TestFundsBypass_Variants: every single-leaf-source / single-dest shape fuses.
func TestFundsBypass_Variants(t *testing.T) {
	cases := []struct{ name, src string }{
		{"world", `send [USD 10] (source = @world destination = @b)`},
		{"unbounded-overdraft", `send [USD 10] (source = @a allowing unbounded overdraft destination = @b)`},
		{"bounded-overdraft", `send [USD 10] (source = @a allowing overdraft up to [USD 5] destination = @b)`},
		{"capped", `send [USD 10] (source = max [USD 5] from @a destination = @b)`},
		{"send-all", `send [USD *] (source = @a destination = @b)`},
		{"kept-is-not-fused", `send [USD 10] (source = @a destination = @b)`}, // sanity dup
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, changed := fundsBypass{}.run(virtualInstrs(t, tc.src))
			require.True(t, changed, "should fuse")
			require.True(t, hasOp[takeAccount](out))
			require.True(t, hasOp[postAccount](out))
			require.False(t, hasOp[pullAccount](out))
			require.False(t, hasOp[sendToAccount](out))
		})
	}
}

// TestFundsBypass_NotFired: branching on either side keeps the queue (pull/send
// survive, no take/post produced).
func TestFundsBypass_NotFired(t *testing.T) {
	cases := []struct{ name, src string }{
		{"fan-out-allotment", `send [USD 10] (source = @a destination = { 1/2 to @x 1/2 to @y })`},
		{"fan-out-inorder", `send [USD 10] (source = @a destination = { max [USD 3] to @x remaining to @y })`},
		{"fan-in", `send [USD 10] (source = { @a @b } destination = @dest)`},
		{"n-by-m", `send [USD 10] (source = { @a @b } destination = { 1/2 to @x 1/2 to @y })`},
		{"kept", `send [USD 10] (source = @a destination = { remaining kept })`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, changed := fundsBypass{}.run(virtualInstrs(t, tc.src))
			require.False(t, changed, "should NOT fuse: %s", strings.TrimSpace(dump(out)))
			require.False(t, hasOp[takeAccount](out))
			require.False(t, hasOp[postAccount](out))
		})
	}
}

// TestFundsBypass_TwoStatements: consecutive single->single sends both fuse (the
// per-statement reset at set_current_asset lets the second match).
func TestFundsBypass_TwoStatements(t *testing.T) {
	src := `
		send [USD 10] (source = @a destination = @b)
		send [USD 20] (source = @c destination = @d)
	`
	out, changed := fundsBypass{}.run(virtualInstrs(t, src))
	require.True(t, changed)
	require.Equal(t, 2, countOp(out, func(in vInstr) bool { _, ok := in.(takeAccount); return ok }))
	require.Equal(t, 2, countOp(out, func(in vInstr) bool { _, ok := in.(postAccount); return ok }))
}

// TestFundsBypass_StatementAfterFanout: a single->single after a (non-fusable)
// fan-out still fuses — the capped-send taint is cleared at the next statement.
func TestFundsBypass_StatementAfterFanout(t *testing.T) {
	src := `
		send [USD 10] (source = @a destination = { 1/2 to @x 1/2 to @y })
		send [USD 20] (source = @c destination = @d)
	`
	out, changed := fundsBypass{}.run(virtualInstrs(t, src))
	require.True(t, changed)
	require.True(t, hasOp[takeAccount](out))
	// the fan-out send/pull are untouched
	require.True(t, hasOp[sendToAccount](out))
}
