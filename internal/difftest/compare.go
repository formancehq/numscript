package difftest

import (
	"fmt"
	"math/big"
)

// Verdict is the outcome of comparing both engines' results for one script.
type Verdict struct {
	Mismatch bool
	Reason   string

	// Tolerated names the tolerance that suppressed a difference, empty when the
	// two sides agreed outright, so a sweep can count how often each one fires. A
	// tolerance absorbing most of the corpus has become a blind spot.
	Tolerated string
}

func ok() Verdict { return Verdict{} }

func tolerated(what string) Verdict { return Verdict{Tolerated: what} }

func mismatch(format string, args ...any) Verdict {
	return Verdict{Mismatch: true, Reason: fmt.Sprintf(format, args...)}
}

// Compare normalizes and diffs two engines' results for the same script.
// aLabel/bLabel appear in mismatch messages only.
//
// Error strings are never compared, only whether an error occurred and at which
// stage: wording legitimately differs between implementations.
func Compare(script string, aRes, bRes SideResult, aLabel, bLabel string) Verdict {
	// Checked before anything else, and symmetrically: an engine breaking its
	// own contract is never an expected outcome, and every tolerance below is
	// about the two engines legitimately disagreeing. Ordering matters — the
	// next branch tolerates a b-side rejection, which would otherwise hide this.
	if aRes.InternalErr != "" {
		return mismatch("%s reported an internal error: %s", aLabel, aRes.InternalErr)
	}
	if bRes.InternalErr != "" {
		return mismatch("%s reported an internal error: %s", bLabel, bRes.InternalErr)
	}

	aCompileFailed := aRes.CompileErr != ""
	bCompileFailed := bRes.CompileErr != ""

	if bCompileFailed && !aCompileFailed {
		// Expected, not a mismatch: internal/gen's cleanup pass is best-effort, not
		// a guarantee — it does not track unboundedness propagating up through
		// nested inorder blocks, so it can still emit a script the b-side rejects.
		return ok()
	}
	if aCompileFailed && !bCompileFailed {
		// The interesting direction: the generator stays within the b-side's
		// grammar and a-side should be a strict superset, so a-side rejecting what
		// b-side compiled is a genuine divergence.
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

	// MissingFunds classification, not raw fail/succeed, is what must agree.
	//
	// A negative source-side `max ... from` amount is a hard reject on the oracle
	// but contributes zero on the interpreter, which can let a different source
	// cover the shortfall (oracle/DIVERGENCES.md §3.3). Comparing fail-vs-succeed
	// would flag that as a false positive; funds adequacy is the real invariant.
	if aRes.MissingFunds != bRes.MissingFunds {
		return mismatch(
			"missing-funds classification differs: %s missingFunds=%v (runErr=%q), %s missingFunds=%v (runErr=%q)",
			aLabel, aRes.MissingFunds, aRes.RunErr, bLabel, bRes.MissingFunds, bRes.RunErr,
		)
	}
	if aRunFailed != bRunFailed {
		// Neither side's (possible) failure was a missing-funds one — see
		// above. Tolerated.
		return ok()
	}
	if aRunFailed {
		// Both failed at runtime with matching missing-funds classification;
		// not comparing exact error text/category beyond that.
		return ok()
	}

	// Aggregated by (source, destination, asset) rather than compared as an
	// ordered list: the engines can split the same net transfer into a different
	// number of posting lines, seen when an unbounded-overdraft account sits
	// between two non-adjacent draws from the same account. That is granularity,
	// not a change in what moved where.
	aAgg := aggregatePostings(aRes.Postings)
	bAgg := aggregatePostings(bRes.Postings)

	if postingsDiffer(aAgg, bAgg) {
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
	}

	// Metadata is compared too: set_tx_meta/set_account_meta are the only
	// observable effect of some generated statements (a number or string value
	// never reaches a posting), so comparing postings alone would let a wrong key,
	// a wrong value or a dropped write pass unnoticed.
	if v := compareMeta("tx metadata", aRes.TxMeta, bRes.TxMeta, aLabel, bLabel); v.Mismatch {
		return v
	}
	if v := compareMeta("account metadata", aRes.AccountMeta, bRes.AccountMeta, aLabel, bLabel); v.Mismatch {
		return v
	}

	return ok()
}

// compareMeta compares one normalized metadata map as a set of key/value pairs.
func compareMeta(what string, a, b map[string]string, aLabel, bLabel string) Verdict {
	for k, av := range a {
		bv, present := b[k]
		if !present {
			return mismatch("%s: %s wrote %q=%q, %s wrote nothing for that key\n%s: %v\n%s: %v",
				what, aLabel, k, av, bLabel, aLabel, a, bLabel, b)
		}
		if av != bv {
			return mismatch("%s: %s wrote %q=%q, %s wrote %q\n%s: %v\n%s: %v",
				what, aLabel, k, av, bLabel, bv, aLabel, a, bLabel, b)
		}
	}
	for k, bv := range b {
		if _, present := a[k]; !present {
			return mismatch("%s: %s wrote %q=%q, %s wrote nothing for that key\n%s: %v\n%s: %v",
				what, bLabel, k, bv, aLabel, aLabel, a, bLabel, b)
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

func postingsDiffer(a, b map[postingKey]*big.Int) bool {
	if len(a) != len(b) {
		return true
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok || av.Cmp(bv) != 0 {
			return true
		}
	}
	return false
}
