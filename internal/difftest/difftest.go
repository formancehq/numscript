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

// Case is one generated script plus both engines' results and the verdict
// between them.
type Case struct {
	Script string
	Vars   map[string]string
	Shape  gen.Shape
	New    SideResult
	Oracle SideResult

	// OracleVsNew compares this repo's interpreter against the vendored
	// legacy machine, which is the only independent ground truth available:
	// the oracle is a separate implementation, maintained elsewhere, that
	// this repo's engine must agree with.
	//
	// The harness is built to carry more engines than this — Compare is
	// deliberately engine-agnostic and takes labels. When internal/vm lands,
	// re-add a VM leg here (runVM, Case.VM, OracleVsVM and NewVsVM) so the
	// compiler+VM is checked against both the oracle and the interpreter.
	OracleVsNew Verdict
}

// RunOne generates one program from rng, runs it against both engines, and
// compares the results.
//
// A script the generator marked oracle-incompatible (numscript-only shapes:
// oneof, colors, division portions, wrong-asset caps) never reaches the legacy
// machine: its oracle leg is recorded as a named tolerance so the sweep counts
// it. Until the compiler+VM leg lands, such a script is only executed by the
// interpreter (a panic still fails the fuzz target); the VM leg is what will
// compare those shapes engine-against-engine.
func RunOne(ctx context.Context, rng *rand.Rand) Case {
	g := gen.Generate(rng)

	newRes := runNew(ctx, g.Script, g.Vars, g.Balances, g.Metadata, g.Flags)

	c := Case{
		Script: g.Script,
		Vars:   g.Vars,
		Shape:  g.Shape,
		New:    newRes,
	}

	if g.OracleCompatible {
		c.Oracle = runOracle(ctx, g.Script, g.Vars, g.Balances, g.Metadata)
		c.OracleVsNew = Compare(newRes, c.Oracle, "new interpreter", "oracle")
	} else {
		c.OracleVsNew = tolerated("numscript-only script, oracle skipped")
	}

	return c
}

// Legs returns the pairwise verdicts with stable names, for callers that
// report per leg. One leg today; the compiler+VM adds two more.
func (c Case) Legs() []struct {
	Name    string
	Verdict Verdict
} {
	return []struct {
		Name    string
		Verdict Verdict
	}{
		{"oracle vs new", c.OracleVsNew},
	}
}

// AnyMismatch reports whether the comparison found a divergence.
func (c Case) AnyMismatch() bool {
	return c.OracleVsNew.Mismatch
}
