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

		for _, leg := range c.Legs() {
			if leg.Verdict.Mismatch {
				t.Fatalf(
					"divergence (%s): %s\n\nvars: %v\n\nscript:\n%s",
					leg.Name, leg.Verdict.Reason, c.Vars, c.Script,
				)
			}
		}
	})
}

// An InternalErr means an engine broke its own contract, as opposed to
// rejecting the script. Compare tolerates a lot
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
			v := difftest.Compare(tc.a, tc.b, "new interpreter", "oracle")
			if !v.Mismatch {
				t.Fatalf("internal error was swallowed: verdict = %+v", v)
			}
		})
	}
}

// The converse: a plain b-side rejection stays tolerated, so the guard above
// didn't just turn every rejection into a mismatch.
func TestCompareStillToleratesAPlainRejection(t *testing.T) {
	v := difftest.Compare(difftest.SideResult{}, difftest.SideResult{CompileErr: "rejected"}, "new interpreter", "oracle")
	if v.Mismatch {
		t.Fatalf("a plain b-side rejection should be tolerated, got %+v", v)
	}
}

// An a-side compile rejection is a mismatch (the vm refusing a script another
// engine ran), with two named exceptions, both fail-closed capacity bounds of
// the bytecode encoding that a pathological generated script can exceed: the
// register bank (ir.ErrRegisterBankOverflow) and the instruction count a jump
// target can address (ir.ErrProgramTooLarge). Tolerated and counted; every
// other a-side rejection stays flagged.
func TestCompareToleratesOnlyTheCapacityRejections(t *testing.T) {
	overflow := difftest.SideResult{CompileErr: "register bank overflow: ...", RegisterOverflow: true}
	if v := difftest.Compare(overflow, difftest.SideResult{}, "vm", "oracle"); v.Mismatch || v.Tolerated != "vm register capacity" {
		t.Fatalf("expected the register-capacity tolerance, got %+v", v)
	}

	tooLarge := difftest.SideResult{CompileErr: "program too large: ...", ProgramTooLarge: true}
	if v := difftest.Compare(tooLarge, difftest.SideResult{}, "vm", "oracle"); v.Mismatch || v.Tolerated != "vm program size" {
		t.Fatalf("expected the program-size tolerance, got %+v", v)
	}

	plain := difftest.SideResult{CompileErr: "no such feature"}
	if v := difftest.Compare(plain, difftest.SideResult{}, "vm", "oracle"); !v.Mismatch {
		t.Fatalf("a plain a-side compile rejection must stay a mismatch, got %+v", v)
	}
}

// The vm and the interpreter must agree exactly: every tolerance that exists for
// a legacy-machine behavior is a mismatch between them, and only the vm's
// capacity bounds stay tolerated.
func TestCompareEnginesHasNoOracleTolerances(t *testing.T) {
	ran := difftest.SideResult{}

	mismatches := []struct {
		name   string
		vm, nw difftest.SideResult
	}{
		{"interpreter-only compile rejection", ran, difftest.SideResult{CompileErr: "rejected"}},
		{"interpreter-only resolve rejection", ran, difftest.SideResult{ResolveErr: "rejected"}},
		{"negative amount vs missing funds",
			difftest.SideResult{RunErr: "negative amount", NegativeAmount: true},
			difftest.SideResult{RunErr: "missing funds", MissingFunds: true}},
		{"negative max clause", ran, difftest.SideResult{RunErr: "negative max", NegativeMaxReject: true}},
	}
	for _, tc := range mismatches {
		t.Run(tc.name, func(t *testing.T) {
			if v := difftest.CompareEngines(tc.vm, tc.nw); !v.Mismatch {
				t.Fatalf("expected a mismatch, got %+v", v)
			}
		})
	}

	t.Run("vm capacity bounds stay tolerated", func(t *testing.T) {
		overflow := difftest.SideResult{CompileErr: "register bank overflow: ...", RegisterOverflow: true}
		if v := difftest.CompareEngines(overflow, ran); v.Mismatch || v.Tolerated != "vm register capacity" {
			t.Fatalf("expected the register-capacity tolerance, got %+v", v)
		}
		tooLarge := difftest.SideResult{CompileErr: "program too large: ...", ProgramTooLarge: true}
		if v := difftest.CompareEngines(tooLarge, ran); v.Mismatch || v.Tolerated != "vm program size" {
			t.Fatalf("expected the program-size tolerance, got %+v", v)
		}
	})

	t.Run("both failing at different stages is not a mismatch", func(t *testing.T) {
		if v := difftest.CompareEngines(difftest.SideResult{RunErr: "boom"}, difftest.SideResult{ResolveErr: "boom"}); v.Mismatch || v.Tolerated != "" {
			t.Fatalf("expected agreement, got %+v", v)
		}
	})
}
