package compiler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// countOp counts instructions matching a predicate.
func countOp(instrs []vInstr, pred func(vInstr) bool) int {
	n := 0
	for _, in := range instrs {
		if pred(in) {
			n++
		}
	}
	return n
}

func TestOptimize_FoldThenDCEToFixpoint(t *testing.T) {
	// monetaryFold drops the get_*; deadCode then drops the now-unused
	// mk_monetary. Running to a fixpoint composes both.
	out := optimize(monetaryProgram(), defaultPeepholes())

	require.Equal(t, strings.TrimSpace(`
  $r0 <- load_const("USD")
  $r1 <- load_const(10)
  set_current_asset($r0)
  check_enough_funds($r1, $r1)`), strings.TrimSpace(dump(out)))

	// no mk_monetary / get_asset / get_amount remain
	require.Zero(t, countOp(out, func(in vInstr) bool {
		if b, ok := in.(binaryOp); ok {
			_, mk := b.op.(opMakeMonetary)
			return mk
		}
		if u, ok := in.(unaryOp); ok {
			switch u.op.(type) {
			case opGetAsset, opGetAmount:
				return true
			}
		}
		return false
	}))
}

func TestOptimize_TerminatesOnEmpty(t *testing.T) {
	require.Empty(t, optimize(nil, defaultPeepholes()))
}

// stubPeephole flips changed `times` times then reports no change, to verify the
// fixpoint loop terminates rather than spinning.
type stubPeephole struct{ remaining *int }

func (stubPeephole) name() string { return "stub" }
func (s stubPeephole) run(instrs []vInstr) ([]vInstr, bool) {
	if *s.remaining > 0 {
		*s.remaining--
		return instrs, true
	}
	return instrs, false
}

func TestOptimize_FixpointTerminates(t *testing.T) {
	n := 3
	out := optimize([]vInstr{loadStr{dest: 0, value: "x"}}, []peephole{stubPeephole{&n}})
	require.Len(t, out, 1)
	require.Equal(t, 0, n) // looped exactly until the stub stopped changing
}
