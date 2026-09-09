package difftest_test

import (
	"context"
	"testing"

	"github.com/formancehq/numscript/internal/difftest"
	"github.com/formancehq/numscript/internal/gen"
)

// FuzzDiff generates a random program from the fuzz bytes, runs it against
// both the new interpreter and the vendored legacy machine, and fails if
// they disagree. Note: Go's fuzzer shrinks the *byte seed*, not the
// generated program — a shrunk failing seed isn't guaranteed to produce a
// structurally minimal script, so the actual generated script is always
// embedded in the failure message (the saved corpus file only holds the
// raw bytes, which aren't human-readable on their own).
func FuzzDiff(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte("differential-testing-seed"))

	f.Fuzz(func(t *testing.T, data []byte) {
		rng := gen.RandFromBytes(data)
		c := difftest.RunOne(context.Background(), rng)

		for _, v := range []difftest.Verdict{c.OracleVsNew, c.OracleVsVM, c.NewVsVM} {
			if v.Mismatch {
				t.Fatalf(
					"divergence: %s\n\nvars: %v\n\nscript:\n%s",
					v.Reason, c.Vars, c.Script,
				)
			}
		}
	})
}

// An InternalErr means an engine broke its own contract — today, that the
// compiler emitted bytecode the VM's verifier rejects. Compare tolerates a lot
// of legitimate disagreement, so the risk is that this gets absorbed into one
// of those tolerances and never surfaces. It must be a mismatch unconditionally,
// on either side, including when the comparison would otherwise stop early.
func TestCompareNeverSwallowsAnInternalError(t *testing.T) {
	internal := difftest.SideResult{InternalErr: "compiled program failed verification: boom"}

	cases := []struct {
		name string
		a, b difftest.SideResult
	}{
		{"a-side alone", internal, difftest.SideResult{}},
		{"b-side alone", difftest.SideResult{}, internal},
		// the dangerous one: a b-side rejection is an explicitly tolerated
		// outcome, so this is the path an internal error could vanish down
		{"a-side, with the other side rejecting", internal, difftest.SideResult{CompileErr: "rejected"}},
		{"b-side, with the other side rejecting", difftest.SideResult{CompileErr: "rejected"}, internal},
		{"both sides compile-failed too",
			difftest.SideResult{InternalErr: "boom", CompileErr: "rejected"},
			difftest.SideResult{CompileErr: "rejected"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := difftest.Compare(tc.a, tc.b, "vm", "oracle")
			if !v.Mismatch {
				t.Fatalf("internal error was swallowed: verdict = %+v", v)
			}
		})
	}
}

// The converse: a plain b-side rejection stays tolerated, so the guard above
// didn't just turn every rejection into a mismatch.
func TestCompareStillToleratesAPlainRejection(t *testing.T) {
	v := difftest.Compare(difftest.SideResult{}, difftest.SideResult{CompileErr: "rejected"}, "vm", "oracle")
	if v.Mismatch {
		t.Fatalf("a plain b-side rejection should be tolerated, got %+v", v)
	}
}
