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

// Compare normalizes and diffs the two engines' results for the same
// script. It intentionally does NOT compare exact error strings (wording
// legitimately differs between the two implementations) — only whether an
// error occurred at all, and at which stage.
func Compare(newRes, oracleRes SideResult) Verdict {
	newCompileFailed := newRes.CompileErr != ""
	oracleCompileFailed := oracleRes.CompileErr != ""

	if oracleCompileFailed && !newCompileFailed {
		// Expected, not a mismatch: internal/gen's cleanup pass (a port of
		// Utils.hs's cleanupNumscript) is a best-effort heuristic, not a
		// complete guarantee — e.g. it doesn't track unboundedness that
		// propagates up through nested inorder blocks, matching a known
		// limitation of the original Haskell tool. The real reference
		// tool's own differential harness (numscript_gen/app/Main.hs)
		// explicitly buckets "oracle rejected" as an expected outcome
		// rather than a hard mismatch, for the same reason.
		return ok()
	}
	if newCompileFailed && !oracleCompileFailed {
		// The interesting direction: the generator only ever emits syntax
		// within the old machine's grammar, and the new interpreter is a
		// strict superset of it, so the new interpreter rejecting a script
		// the oracle accepted is a genuine, worth-investigating divergence.
		return mismatch(
			"new interpreter rejected a script the oracle compiled: compileErr=%q",
			newRes.CompileErr,
		)
	}
	if newCompileFailed {
		// Both rejected the script; nothing further to compare.
		return ok()
	}

	newRunFailed := newRes.RunErr != ""
	oracleRunFailed := oracleRes.RunErr != ""

	if newRunFailed != oracleRunFailed {
		return mismatch(
			"execution-error divergence: new interpreter runErr=%q, oracle runErr=%q",
			newRes.RunErr, oracleRes.RunErr,
		)
	}
	if newRunFailed {
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
	newAgg := aggregatePostings(newRes.Postings)
	oracleAgg := aggregatePostings(oracleRes.Postings)

	if len(newAgg) != len(oracleAgg) {
		return mismatch(
			"aggregated posting set differs: new interpreter has %d distinct (source,destination,asset), oracle has %d\nnew: %+v\noracle: %+v",
			len(newAgg), len(oracleAgg), newRes.Postings, oracleRes.Postings,
		)
	}

	for k, newAmount := range newAgg {
		oracleAmount, ok := oracleAgg[k]
		if !ok || newAmount.Cmp(oracleAmount) != 0 {
			return mismatch(
				"aggregated amount differs for source=%q destination=%q asset=%q: new interpreter=%v, oracle=%v\nnew: %+v\noracle: %+v",
				k.Source, k.Destination, k.Asset, newAmount, oracleAmount, newRes.Postings, oracleRes.Postings,
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
