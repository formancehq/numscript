package difftest

import (
	"fmt"
	"math/big"
)

// Verdict is the outcome of comparing both engines' results for one script.
type Verdict struct {
	Mismatch bool
	Reason   string
}

func ok() Verdict { return Verdict{} }

func mismatch(format string, args ...any) Verdict {
	return Verdict{Mismatch: true, Reason: fmt.Sprintf(format, args...)}
}

// Compare normalizes and diffs two engines' results for the same script.
// aLabel/bLabel name the two sides (e.g. "new interpreter"/"oracle",
// "vm"/"oracle") in mismatch messages only — comparison logic is symmetric
// modulo which side's rejection is treated as "expected" (see below).
// It intentionally does NOT compare exact error strings (wording
// legitimately differs between implementations) — only whether an error
// occurred at all, and at which stage.
func Compare(aRes, bRes SideResult, aLabel, bLabel string) Verdict {
	aCompileFailed := aRes.CompileErr != ""
	bCompileFailed := bRes.CompileErr != ""

	if bCompileFailed && !aCompileFailed {
		// Expected, not a mismatch: internal/gen's cleanup pass (a port of
		// Utils.hs's cleanupNumscript) is a best-effort heuristic, not a
		// complete guarantee — e.g. it doesn't track unboundedness that
		// propagates up through nested inorder blocks, matching a known
		// limitation of the original Haskell tool. The real reference
		// tool's own differential harness (numscript_gen/app/Main.hs)
		// explicitly buckets "oracle rejected" as an expected outcome
		// rather than a hard mismatch, for the same reason. The same
		// tolerance applies to any b-side (oracle, or the vm compiler)
		// rejecting something a-side accepts.
		return ok()
	}
	if aCompileFailed && !bCompileFailed {
		// The interesting direction: the generator only ever emits syntax
		// within the b-side's grammar, and a-side is expected to be a
		// strict superset of it, so a-side rejecting a script b-side
		// compiled is a genuine, worth-investigating divergence.
		return mismatch(
			"%s rejected a script %s compiled: compileErr=%q",
			aLabel, bLabel, aRes.CompileErr,
		)
	}
	if aCompileFailed {
		// Both rejected the script; nothing further to compare.
		return ok()
	}

	aRunFailed := aRes.RunErr != ""
	bRunFailed := bRes.RunErr != ""

	if aRunFailed != bRunFailed {
		return mismatch(
			"execution-error divergence: %s runErr=%q, %s runErr=%q",
			aLabel, aRes.RunErr, bLabel, bRes.RunErr,
		)
	}
	if aRunFailed {
		// Both failed at runtime (e.g. insufficient funds); not comparing
		// exact error text/category.
		return ok()
	}

	// Compare postings aggregated by (source, destination, asset) rather
	// than as an exact ordered list: the two engines can legitimately
	// split the same net transfer between two accounts into a different
	// number of posting lines (e.g. one engine emits a single merged
	// posting where the other emits two that sum to the same amount,
	// observed when an unbounded-overdraft account sits between two
	// non-adjacent draws from the same real account in a source list).
	// That's a posting-granularity difference, not a change in what moved
	// where — so it's intentionally not treated as a mismatch here.
	aAgg := aggregatePostings(aRes.Postings)
	bAgg := aggregatePostings(bRes.Postings)

	if len(aAgg) != len(bAgg) {
		return mismatch(
			"aggregated posting set differs: %s has %d distinct (source,destination,asset), %s has %d\n%s: %+v\n%s: %+v",
			aLabel, len(aAgg), bLabel, len(bAgg), aLabel, aRes.Postings, bLabel, bRes.Postings,
		)
	}

	for k, aAmount := range aAgg {
		bAmount, ok := bAgg[k]
		if !ok || aAmount.Cmp(bAmount) != 0 {
			return mismatch(
				"aggregated amount differs for source=%q destination=%q asset=%q: %s=%v, %s=%v\n%s: %+v\n%s: %+v",
				k.Source, k.Destination, k.Asset, aLabel, aAmount, bLabel, bAmount, aLabel, aRes.Postings, bLabel, bRes.Postings,
			)
		}
	}

	return ok()
}

// postingKey groups postings that move the same asset between the same
// two accounts, regardless of how many separate posting lines an engine
// happened to split that movement into.
type postingKey struct {
	Source      string
	Destination string
	Asset       string
}

func aggregatePostings(postings []Posting) map[postingKey]*big.Int {
	agg := make(map[postingKey]*big.Int, len(postings))
	for _, p := range postings {
		if p.Amount.Sign() == 0 {
			// A zero-amount posting is equivalent to no posting at all.
			// Both run_oracle.go and the new interpreter itself already
			// filter these upstream, but that makes this comparison
			// silently dependent on an invariant it doesn't enforce —
			// enforce it here too so the equivalence lives in one place.
			continue
		}
		k := postingKey{Source: p.Source, Destination: p.Destination, Asset: p.Asset}
		total, ok := agg[k]
		if !ok {
			total = new(big.Int)
			agg[k] = total
		}
		total.Add(total, p.Amount)
	}
	return agg
}
