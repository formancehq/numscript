package compiler

// worldAccount mirrors vm.worldAccount: the implicit unbounded source. Kept as a
// local literal to avoid a compiler->vm dependency just for this constant.
const worldAccount = "world"

// postFromUnbounded is the aggressive follow-up to fundsBypass. Where fundsBypass
// turns a single->single send into `take + checkEnoughFunds + post`, this pass
// collapses that triple into a single `post_from_unbounded` when the source is
// UNBOUNDED — @world or `allowing unbounded overdraft`.
//
// An unbounded pull makes exactly `cap` available regardless of the balance and
// can never be short, so two pieces of work the general take path still does are
// pure overhead here:
//
//   - the enough-funds check (got == cap == needed, always passes), and
//   - the source-balance debit (an entryFor map hit + a big.Int subtract),
//     recorded only so a later balance() read of the source stays correct.
//
// Dropping the debit is only sound when the source balance is never observed
// afterward. Two guards ensure that:
//
//   - the program contains NO balance() read at all (no fetchBalance). This is the
//     only way a script observes the running balance directly, and it is the
//     common case (balance() is rare). If any exists, the pass bails entirely
//     rather than reason about which account it reads.
//   - a non-@world source must not be pulled/taken anywhere else in the program:
//     a later *bounded* pull of the same account would fold the (now-missing)
//     debit into its available-funds math and change the postings. @world is
//     exempt because the VM always treats @world as unbounded, so its debit is
//     never folded into any later pull.
//
// The postings the queue would have produced are unchanged (the destination is
// still credited by the emitted posting); only the unobservable source debit and
// the always-true check are removed.
type postFromUnboundedPass struct{}

func (postFromUnboundedPass) name() string { return "post-from-unbounded" }

func (postFromUnboundedPass) run(instrs []vInstr) ([]vInstr, bool) {
	// Guard 1: any balance() read disables the pass wholesale.
	for _, in := range instrs {
		if _, ok := in.(fetchBalance); ok {
			return instrs, false
		}
	}

	// Count how many pulls/takes reference each compile-time-constant account, to
	// enforce the "non-world source used once" guard.
	pullCount := map[string]int{}
	// nonLeaf = accounts whose balance is READ somewhere: a pull/take source (a
	// later bounded pull folds in accumulated deltas) or a save target. A dst that
	// is NONE of these is a leaf — its credit is never observed and can be dropped
	// (balance() reads are already excluded wholesale above). Reads via balance()
	// go through fetchBalance, handled by the guard; postings/sends only WRITE the
	// destination balance, so they are not observers.
	nonLeaf := map[string]bool{}
	for _, in := range instrs {
		switch v := in.(type) {
		case pullAccount:
			if acc, ok := constStrOf(instrs, v.account); ok {
				pullCount[acc]++
				nonLeaf[acc] = true
			}
		case takeAccount:
			if acc, ok := constStrOf(instrs, v.account); ok {
				pullCount[acc]++
				nonLeaf[acc] = true
			}
		case save:
			if acc, ok := constStrOf(instrs, v.account); ok {
				nonLeaf[acc] = true
			}
		}
	}

	// readers[r] = indices of instructions that read register r.
	readers := map[reg][]int{}
	for i, in := range instrs {
		for _, r := range in.sources() {
			readers[r] = append(readers[r], i)
		}
	}

	// edits[i] replaces instrs[i]; a nil slice deletes it.
	edits := map[int][]vInstr{}

	for i, in := range instrs {
		t, ok := in.(takeAccount)
		if !ok {
			continue
		}
		// Only the plain capped, uncolored case: an unbounded source always carries
		// a cap (an uncapped unbounded source is a compile error), and colors are
		// unimplemented.
		if t.cap == nil || t.color != nil {
			continue
		}

		acc, ok := constStrOf(instrs, t.account)
		if !ok {
			continue
		}
		isWorld := acc == worldAccount
		unbounded := isWorld || (t.overdraft == nil && !t.boundedZero)
		if !unbounded {
			continue
		}
		if !isWorld && pullCount[acc] != 1 {
			continue // a later (possibly bounded) pull would observe the dropped debit
		}

		// t.dest (the pulled amount) must be read by exactly the paired check and
		// post, and nothing else — otherwise removing them would strand a live use.
		rs := readers[t.dest]
		if len(rs) != 2 {
			continue
		}
		var checkIdx, postIdx = -1, -1
		for _, j := range rs {
			switch v := instrs[j].(type) {
			case checkEnoughFunds:
				if v.got == t.dest {
					checkIdx = j
				}
			case postAccount:
				if v.amount == t.dest && v.color == nil {
					postIdx = j
				}
			}
		}
		if checkIdx < 0 || postIdx < 0 {
			continue
		}
		post := instrs[postIdx].(postAccount)

		// credit is droppable when dst is a compile-time-constant LEAF account: its
		// balance is never read (no pull/take source, no save), and balance() reads
		// are already excluded. A dynamic (non-const) dst can't be proven a leaf.
		credit := true
		if dst, ok := constStrOf(instrs, post.dstAccount); ok && !nonLeaf[dst] {
			credit = false
		}

		edits[i] = nil        // drop the take
		edits[checkIdx] = nil // drop the always-true funds check
		edits[postIdx] = []vInstr{postFromUnbounded{
			srcAccount: post.srcAccount,
			dstAccount: post.dstAccount,
			cap:        *t.cap,
			color:      nil,
			credit:     credit,
		}}
	}

	if len(edits) == 0 {
		return instrs, false
	}

	out := make([]vInstr, 0, len(instrs))
	for i, in := range instrs {
		if repl, ok := edits[i]; ok {
			out = append(out, repl...) // nil slice => deletion
			continue
		}
		out = append(out, in)
	}
	return out, true
}
