package difftest_test

import (
	"context"
	"math/rand"
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
func TestDifferentialSweep(t *testing.T) {
	for seed := range sweepSeeds {
		c := difftest.RunOne(context.Background(), rand.New(rand.NewSource(int64(seed))))
		if c.OracleVsNew.Mismatch {
			t.Fatalf(
				"seed %d: divergence: %s\n\nvars: %v\n\nscript:\n%s",
				seed, c.OracleVsNew.Reason, c.Vars, c.Script,
			)
		}
	}
}
