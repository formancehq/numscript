package difftest

import "fmt"

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

	if len(newRes.Postings) != len(oracleRes.Postings) {
		return mismatch(
			"posting count differs: new interpreter has %d, oracle has %d\nnew: %+v\noracle: %+v",
			len(newRes.Postings), len(oracleRes.Postings), newRes.Postings, oracleRes.Postings,
		)
	}

	for i := range newRes.Postings {
		a, b := newRes.Postings[i], oracleRes.Postings[i]
		if a.Source != b.Source || a.Destination != b.Destination || a.Asset != b.Asset || a.Amount.Cmp(b.Amount) != 0 {
			return mismatch(
				"posting %d differs: new interpreter=%+v, oracle=%+v",
				i, a, b,
			)
		}
	}

	return ok()
}
