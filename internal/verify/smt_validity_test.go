package verify

import (
	"context"
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/stretchr/testify/require"
)

// These tests check the SMT we emit is actually well-formed and consistent —
// independent of any particular query verdict.

var fixtures = map[string]string{
	"simple":  simple,
	"world":   world,
	"inorder": inorder,
	"capped":  capped,
}

// The base encoding (no query appended) must be:
//   - accepted by z3 with no (error ...) lines  -> the SMT is well-formed;
//   - satisfiable                                -> the encoding is NOT
//     self-contradictory. This is the key soundness guard: an unsat base would
//     make every prove-query vacuously "Proved".
func TestBaseEncodingWellFormedAndSatisfiable(t *testing.T) {
	requireZ3(t)
	for name, src := range fixtures {
		t.Run(name, func(t *testing.T) {
			enc, err := compiler.SymbolicEncodeSource(src)
			require.NoError(t, err)

			script := "(set-option :produce-models true)\n" + enc.SMTLIB + "(check-sat)\n(get-model)\n"
			sat, _, raw, err := runZ3(context.Background(), "", script)
			require.NoError(t, err, "z3 output:\n%s", raw)
			require.NotContains(t, raw, "(error ", "encoding produced invalid SMT")
			require.Equal(t, "sat", sat, "base encoding must be satisfiable (a valid execution exists); an unsat base means the encoding is contradictory")
		})
	}
}

// Every real query, in both prove and find modes, must run through z3 without
// producing invalid-SMT errors. (We assert on the absence of errors, not on the
// verdict — that's what the other tests do.)
func TestQueriesEmitValidSMT(t *testing.T) {
	requireZ3(t)
	queries := []string{
		`prove: !fail => received("dest", "USD/2") == 10`,
		`find: fail`,
		`prove: sent("src", "USD/2") == received("dest", "USD/2")`,
		`prove: end_balance("src", "USD/2") == start_balance("src", "USD/2") - sent("src", "USD/2")`,
		`prove: volumes("dest", "USD/2") >= 0`,
	}
	for _, q := range queries {
		t.Run(q, func(t *testing.T) {
			// Verify returns an error if z3 rejects the SMT as malformed
			// (a pre-verdict `(error ...)` line), so NoError here means the
			// emitted SMT was well-formed. A definite (non-Unknown) verdict
			// means z3 actually decided it.
			res, err := Verify(context.Background(), simple, q, Options{})
			require.NoError(t, err, "raw z3 output:\n%s", errRaw(res))
			require.NotEqual(t, Unknown, res.Verdict)
		})
	}
}

func errRaw(res *Result) string {
	if res == nil {
		return "<nil>"
	}
	return res.Raw
}

// Sanity check the guard itself: genuinely malformed SMT must be rejected, so a
// silent "sat" can never slip through as a false proof.
func TestMalformedSMTIsRejected(t *testing.T) {
	requireZ3(t)
	bad := "(declare-const x Int)\n(assert (= x undefined_symbol))\n(check-sat)\n"
	_, _, _, err := runZ3(context.Background(), "", bad)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "invalid"), "expected an invalid-SMT error, got: %v", err)
}

// Cross-check against a hand-computed model: for the world→{d1,d2} split there
// is exactly one outcome (world is infinite, so no free balances), and z3's
// model must match it.
func TestModelMatchesHandComputed(t *testing.T) {
	requireZ3(t)
	// received(d1)==4 && received(d2)==6 is the only possibility; asking to
	// find any OTHER split must be impossible.
	res, err := Verify(context.Background(), world,
		`find: received("d1","USD/2") != 4 || received("d2","USD/2") != 6`, Options{})
	require.NoError(t, err)
	require.Equal(t, Impossible, res.Verdict, "raw:\n%s", res.Raw)
}
