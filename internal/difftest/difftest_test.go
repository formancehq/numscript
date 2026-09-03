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
