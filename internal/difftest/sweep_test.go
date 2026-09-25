package difftest_test

import (
	"context"
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/difftest"
	"github.com/formancehq/numscript/internal/gen"
)

// sweepSeeds is sized so the sweep stays well under a minute on CI while
// still covering a broad slice of the generator's output.
const sweepSeeds = 3000

// TestDifferentialSweep is the deterministic differential gate: a fixed set of
// seeds through both engines, failing on any divergence.
//
// It exists alongside FuzzDiff, not instead of it. FuzzDiff finds more, but
// `go test -fuzz -fuzztime` reports "context deadline exceeded" as a test
// failure when the deadline lands mid-coordination — a red check with no
// divergence behind it, and likelier the harder the corpus churns. A fixed seed
// set has no deadline, corpus or workers, so a failure here is always a real
// disagreement, and names the seed that reproduces it.
//
// Run FuzzDiff directly for exploration:
//
//	go test -run .^$. -fuzz FuzzDiff -fuzztime 5m ./internal/difftest
//
// It does not stop at the first divergence: that would say nothing about
// whether one class or five are in play. It sweeps everything, groups by class,
// and reports counts with a reproducing seed each.
func TestDifferentialSweep(t *testing.T) {
	type class struct {
		count      int
		firstSeed  int
		sampleWhy  string
		sampleVars map[string]string
		sampleSrc  string
	}
	classes := map[string]*class{}
	order := []string{}
	tolerated := map[string]int{}
	var r reach

	for seed := range sweepSeeds {
		c := difftest.RunOne(context.Background(), rand.New(rand.NewSource(int64(seed))))
		r.add(c)
		if c.OracleVsNew.Tolerated != "" {
			tolerated[c.OracleVsNew.Tolerated]++
		}
		if !c.OracleVsNew.Mismatch {
			continue
		}
		k := divergenceClass(c.OracleVsNew.Reason)
		cl, seen := classes[k]
		if !seen {
			cl = &class{firstSeed: seed, sampleWhy: c.OracleVsNew.Reason, sampleVars: c.Vars, sampleSrc: c.Script}
			classes[k] = cl
			order = append(order, k)
		}
		cl.count++
	}

	t.Log(r.table())
	for _, what := range slices.Sorted(maps.Keys(tolerated)) {
		t.Logf("tolerated: %-28s %d/%d scripts", what, tolerated[what], sweepSeeds)
	}
	if len(classes) == 0 {
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d divergence class(es) over %d seeds:\n", len(classes), sweepSeeds)
	for _, k := range order {
		cl := classes[k]
		fmt.Fprintf(&sb, "  %-44s %4d scripts (first: seed %d)\n", k, cl.count, cl.firstSeed)
	}
	sb.WriteString("\nSee internal/oracle/DIVERGENCES.md.\n")

	first := classes[order[0]]
	fmt.Fprintf(&sb, "\nseed %d: %s\n\nvars: %v\n\nscript:\n%s",
		first.firstSeed, first.sampleWhy, first.sampleVars, first.sampleSrc)
	t.Fatal(sb.String())
}

// reach counts, over the sweep, how often the generator produces the shapes a
// stateful bug needs — the numbers DIVERGENCES.md #3 quotes — and how many
// scripts were actually compared. A green sweep only means something for the
// shapes this table shows were reached.
type reach struct {
	strategy       map[gen.Strategy]int
	completed      int // both engines ran to the end: the only scripts whose postings were compared
	oracleReject   int
	save           int
	bounded        int
	sameAccount    int
	sameResource   int
	inOrder        int
	overdraws      int
	inOrderDiverg  int
	allotRemaining int
	portionVar     int
	chainedOrigin  int
}

func (r *reach) add(c difftest.Case) {
	if r.strategy == nil {
		r.strategy = map[gen.Strategy]int{}
	}
	sh := c.Shape
	r.strategy[sh.Strategy]++
	if !c.New.Failed() && !c.Oracle.Failed() {
		r.completed++
	}
	if c.Oracle.CompileErr != "" {
		r.oracleReject++
	}
	count := func(n *int, cond bool) {
		if cond {
			*n++
		}
	}
	count(&r.save, sh.HasSave)
	count(&r.bounded, sh.HasBoundedOverdraft)
	count(&r.sameAccount, sh.SaveOverdraftSameAccount)
	count(&r.sameResource, sh.SaveOverdraftSameResource)
	count(&r.inOrder, sh.SaveOverdraftSameResourceInOrder)
	count(&r.overdraws, sh.SaveOverdrawsInitial)
	count(&r.inOrderDiverg, sh.SaveOverdraftSameResourceInOrder && c.OracleVsNew.Mismatch)
	count(&r.allotRemaining, sh.HasAllotmentRemaining)
	count(&r.portionVar, sh.HasPortionVar)
	count(&r.chainedOrigin, sh.HasChainedOrigin)
}

func (r *reach) table() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "reach over %d seeds:\n", sweepSeeds)
	row := func(label string, n int) { fmt.Fprintf(&sb, "  %-44s %4d\n", label, n) }
	row("strategy: uniform", r.strategy[gen.StrategyUniform])
	row("strategy: scenario + random", r.strategy[gen.StrategyScenarioMixed])
	row("strategy: scenario only", r.strategy[gen.StrategyScenarioOnly])
	row("both engines ran to completion", r.completed)
	row("oracle rejected at compile time", r.oracleReject)
	row("scripts containing a save", r.save)
	row("containing a bounded overdraft", r.bounded)
	row("both, on the same account", r.sameAccount)
	row("both, on the same (account, asset)", r.sameResource)
	row("...in that order", r.inOrder)
	row("...where the save also overdraws", r.overdraws)
	row("...that diverged", r.inOrderDiverg)
	row("allotment with a remaining clause", r.allotRemaining)
	row("allotment portion through a var", r.portionVar)
	row("origin var chained through another", r.chainedOrigin)
	return sb.String()
}

// divergenceClass buckets a Verdict reason by its kind, discarding the
// per-script amounts and account names that follow it, so the sweep can count
// classes rather than listing thousands of individually-worded failures.
func divergenceClass(reason string) string {
	head, _, _ := strings.Cut(reason, "\n")
	for _, prefix := range []string{
		"missing-funds classification differs",
		"aggregated posting set differs",
		"aggregated amount differs",
		"tx metadata",
		"account metadata",
	} {
		if strings.HasPrefix(head, prefix) {
			return prefix
		}
	}
	if i := strings.Index(head, ":"); i > 0 {
		return head[:i]
	}
	return head
}
