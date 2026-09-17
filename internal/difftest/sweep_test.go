package difftest_test

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/difftest"
)

// sweepSeeds is sized so the sweep stays well under a minute on CI while
// still covering a broad slice of the generator's output.
const sweepSeeds = 3000

// TestDifferentialSweep is the deterministic differential gate: it runs a
// fixed set of seeds through both engines and fails on any divergence.
//
// This exists alongside FuzzDiff rather than instead of it. FuzzDiff is
// coverage-guided and finds more, but it is the wrong shape for a PR gate:
// `go test -fuzz -fuzztime` reports "context deadline exceeded" as a test
// failure when its deadline lands mid-coordination, which is a red check with
// no divergence behind it. That race gets likelier the more the corpus churns,
// and the generator churns it hard (hundreds of new interesting inputs a
// minute). A fixed seed set has no deadline, no corpus, and no workers, so a
// failure here always means a real disagreement — and names the seed that
// reproduces it.
//
// Run FuzzDiff directly for exploration:
//
//	go test -run '^$' -fuzz FuzzDiff -fuzztime 5m ./internal/difftest
//
// EXPECTED TO FAIL right now. The oracle is faithful to ledger everywhere
// except the one guard in internal/oracle/DIVERGENCES.md §2, which exists to
// remove noise rather than create it, so the open numscript/ledger semantic
// disagreements in DIVERGENCES.md §3 are visible instead of hidden. They are
// pinned individually in TestKnownOpenDivergences; this sweep says how much of
// the generator's output each one reaches.
//
// It deliberately does NOT stop at the first divergence: failing on seed 24
// would say nothing about whether one class or five are in play. It sweeps
// everything, groups by divergence class, and reports counts with a
// reproducing seed each.
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

	for seed := range sweepSeeds {
		c := difftest.RunOne(context.Background(), rand.New(rand.NewSource(int64(seed))))
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

	for what, n := range tolerated {
		t.Logf("tolerated: %-28s %d/%d scripts", what, n, sweepSeeds)
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
	sb.WriteString("\nSee internal/oracle/DIVERGENCES.md §3.\n")

	first := classes[order[0]]
	fmt.Fprintf(&sb, "\nseed %d: %s\n\nvars: %v\n\nscript:\n%s",
		first.firstSeed, first.sampleWhy, first.sampleVars, first.sampleSrc)
	t.Fatal(sb.String())
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
