// Package difftest compares this repo's interpreter against the vendored legacy
// ledger machine (internal/oracle) on programs from internal/gen.
//
// Its own Go module, wired with local `replace` directives, so it can import
// both the root module and the oracle without either depending on this one.
package difftest

import (
	"context"
	"math/rand"

	"github.com/formancehq/numscript/internal/gen"
)

// Case is one generated script plus all three engines' results and the
// pairwise verdicts between them.
type Case struct {
	Script string
	Vars   map[string]string
	Shape  gen.Shape
	New    SideResult
	Oracle SideResult
	VM     SideResult

	// OracleVsNew compares this repo's interpreter against the vendored
	// legacy machine, which is the only independent ground truth available:
	// the oracle is a separate implementation, maintained elsewhere, that
	// this repo's engine must agree with. OracleVsVM extends the same check
	// to the compiler+VM engine; NewVsVM has no independent ground truth
	// (both engines live in this repo), so a mismatch there means the
	// interpreter and the compiler+VM disagree with each other, not just
	// with the legacy oracle.
	//
	// The VM sits on the a-side of its two verdicts so that Compare's b-side
	// tolerances keep their meaning: an oracle-side rejection stays an
	// expected outcome, while the VM rejecting a script another engine ran is
	// a mismatch.
	OracleVsNew Verdict
	OracleVsVM  Verdict
	NewVsVM     Verdict
}

// RunOne generates one program from rng, runs it against all three engines,
// and compares the results pairwise.
//
// A script the generator marked oracle-incompatible (numscript-only shapes:
// oneof, colors, division portions, wrong-asset caps) never reaches the legacy
// machine: its oracle legs are recorded as a named tolerance so the sweep
// counts them, and only NewVsVM is compared — which is the point of those
// shapes, since both engines live in this repo and must agree exactly.
func RunOne(ctx context.Context, rng *rand.Rand) Case {
	g := gen.Generate(rng)

	newRes := runNew(ctx, g.Script, g.Vars, g.Balances, g.Metadata, g.Flags)
	vmRes := runVM(ctx, g.Script, g.Vars, g.Balances, g.Metadata, g.Flags)

	c := Case{
		Script: g.Script,
		Vars:   g.Vars,
		Shape:  g.Shape,
		New:    newRes,
		VM:     vmRes,

		NewVsVM: Compare(vmRes, newRes, "vm", "new interpreter"),
	}

	if g.OracleCompatible {
		c.Oracle = runOracle(ctx, g.Script, g.Vars, g.Balances, g.Metadata)
		c.OracleVsNew = Compare(newRes, c.Oracle, "new interpreter", "oracle")
		c.OracleVsVM = Compare(vmRes, c.Oracle, "vm", "oracle")
	} else {
		c.OracleVsNew = tolerated("numscript-only script, oracle skipped")
		c.OracleVsVM = tolerated("numscript-only script, oracle skipped")
	}

	return c
}

// Legs returns the three pairwise verdicts with stable names, for callers
// that report per leg.
func (c Case) Legs() []struct {
	Name    string
	Verdict Verdict
} {
	return []struct {
		Name    string
		Verdict Verdict
	}{
		{"oracle vs new", c.OracleVsNew},
		{"oracle vs vm", c.OracleVsVM},
		{"new vs vm", c.NewVsVM},
	}
}

// AnyMismatch reports whether any of the three pairwise verdicts found a
// divergence.
func (c Case) AnyMismatch() bool {
	return c.OracleVsNew.Mismatch || c.OracleVsVM.Mismatch || c.NewVsVM.Mismatch
}
