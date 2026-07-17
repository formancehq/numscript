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

func countType[T vInstr](instrs []vInstr) int {
	n := 0
	for _, in := range instrs {
		if _, ok := in.(T); ok {
			n++
		}
	}
	return n
}

// fused asserts the region was bypassed: no pull/send survive, take+post appear.
func fused(t *testing.T, out []vInstr) {
	t.Helper()
	require.False(t, hasOp[pullAccount](out), "pull_account should be gone:\n%s", dump(out))
	require.False(t, hasOp[sendToAccount](out), "send_to_account should be gone:\n%s", dump(out))
	require.True(t, hasOp[takeAccount](out), "expected take_account")
	require.True(t, hasOp[postAccount](out), "expected post_account")
}

// notFused asserts the queue was kept: pull/send survive, no take/post.
func notFused(t *testing.T, out []vInstr) {
	t.Helper()
	require.False(t, hasOp[takeAccount](out), "should NOT fuse:\n%s", dump(out))
	require.False(t, hasOp[postAccount](out))
	require.True(t, hasOp[pullAccount](out))
}

// TestFundsBypass_SingleToSingle: a plain-account 1-source/1-dest send fuses to
// take+post, with take before the check and post after the dst load (no reorder).
func TestFundsBypass_SingleToSingle(t *testing.T) {
	out, changed := fundsBypass{}.run(virtualInstrs(t, `send [USD 10] (source = @a destination = @b)`))
	require.True(t, changed)
	fused(t, out)

	d := dump(out)
	takeAt := strings.Index(d, "take_account")
	checkAt := strings.Index(d, "check_enough_funds")
	postAt := strings.Index(d, "post_account")
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
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, changed := fundsBypass{}.run(virtualInstrs(t, tc.src))
			require.True(t, changed, "should fuse")
			fused(t, out)
		})
	}
}

// TestFundsBypass_FanOut: 1 source -> M destinations fuses to one take + one post
// per destination.
func TestFundsBypass_FanOut(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		posts int
	}{
		{"allotment", `send [USD 100] (source = @a destination = { 1/2 to @x 1/2 to @y })`, 2},
		{"allotment-3", `send [USD 100] (source = @a destination = { 1/4 to @x 1/4 to @y remaining to @z })`, 3},
		{"send-all", `send [USD *] (source = @a destination = { 1/2 to @x 1/2 to @y })`, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, changed := fundsBypass{}.run(virtualInstrs(t, tc.src))
			require.True(t, changed, "should fuse")
			fused(t, out)
			require.Equal(t, 1, countType[takeAccount](out), "one take (single source)")
			require.Equal(t, tc.posts, countType[postAccount](out), "one post per destination")
		})
	}
}

// TestFundsBypass_FanIn: N sources -> 1 destination fuses to one take per source
// + one post per source. Fires for allotment sources and send-all inorder (no
// jump); NOT for capped inorder (early-exit jump).
func TestFundsBypass_FanIn(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		takes int
	}{
		{"allotment-source", `send [USD 100] (source = { 1/2 from @a 1/2 from @b } destination = @dest)`, 2},
		{"send-all-inorder", `send [USD *] (source = { @a @b } destination = @dest)`, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, changed := fundsBypass{}.run(virtualInstrs(t, tc.src))
			require.True(t, changed, "should fuse")
			fused(t, out)
			require.Equal(t, tc.takes, countType[takeAccount](out), "one take per source")
			require.Equal(t, tc.takes, countType[postAccount](out), "one post per source")
		})
	}
}

// TestFundsBypass_NotFired: shapes that must keep the runtime queue.
func TestFundsBypass_NotFired(t *testing.T) {
	cases := []struct{ name, src string }{
		{"n-by-m", `send [USD 10] (source = { @a @b } destination = { 1/2 to @x 1/2 to @y })`},
		{"inorder-dest", `send [USD 100] (source = @a destination = { max [USD 30] to @x remaining to @y })`}, // negative-max unsafe
		{"kept-fanout", `send [USD 10] (source = @a destination = { 1/2 to @x remaining kept })`},
		{"single-kept", `send [USD 10] (source = @a destination = { remaining kept })`},
		{"capped-inorder-source", `send [USD 10] (source = { @a @b } destination = @dest)`}, // early-exit jmp
		{"aliasing-fanin", `send [USD 10] (source = { 1/2 from @a 1/2 from @a } destination = @dest)`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, changed := fundsBypass{}.run(virtualInstrs(t, tc.src))
			require.False(t, changed, "should NOT fuse:\n%s", strings.TrimSpace(dump(out)))
			notFused(t, out)
		})
	}
}

// TestFundsBypass_TwoStatements: consecutive sends both fuse (per-statement
// regioning at set_current_asset).
func TestFundsBypass_TwoStatements(t *testing.T) {
	src := `
		send [USD 10] (source = @a destination = @b)
		send [USD 100] (source = @c destination = { 1/2 to @x 1/2 to @y })
	`
	out, changed := fundsBypass{}.run(virtualInstrs(t, src))
	require.True(t, changed)
	require.False(t, hasOp[pullAccount](out))
	require.False(t, hasOp[sendToAccount](out))
	require.Equal(t, 2, countType[takeAccount](out), "one take per statement")
	require.Equal(t, 3, countType[postAccount](out), "1 (single) + 2 (fan-out) posts")
}
