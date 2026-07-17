package compiler

// fundsBypass fuses a `send` whose source/destination pairing is static into
// direct debit/post ops that skip the runtime funds queue.
//
// The queue exists to pair N funding sources with M destinations when the pairing
// is data-dependent. It's redundant unless BOTH sides branch — so three shapes
// are bypassable:
//
//	single->single  (1 source, 1 destination)
//	fan-out         (1 source, M destinations)
//	fan-in          (N sources, 1 destination)
//
// while the true N×M case (both branch) keeps the queue.
//
// The rewrites reuse two vInstrs: take_account (pull's amount math + debit, no
// queue) and post_account (posting + credit, no debit). Because a numscript send
// lowers as all source pulls (the source phase) followed by all destination sends
// (the destination phase), a statement region is classified by pull count P and
// send count S:
//
//	P==1, S==1 (drain-all)     single->single: pull->take, send->post(A, B, got)
//	P==1, S>1  (allotment)     fan-out: pull->take, each send->post(A, dst, share)
//	P>1,  S==1 (drain-all)     fan-in:  each pull->take, send-> P posts(src_i, dst, got_i)
//
// Correctness notes:
//   - fan-out posts use the send's cap register (the destination share) as the
//     amount, WITHOUT clamping to the source. That is only sound when the shares
//     are provably non-negative and sum to `got` — i.e. an ALLOTMENT destination
//     (portions are in [0,1] and sum to 1, so shares are in [0,got] and sum to
//     got, and take's single `got` debit balances the posts). An inorder
//     destination is NOT eligible: a negative `max` clause inflates the
//     compiler's `remaining` above `got`, and the queue's Send clamps to the
//     actual funds where a raw post would over-pay. So fan-out requires every
//     send's cap to be a makeAllotment output. A `kept`/refund send (nil account)
//     would also leave the source short after the debit, so bail on those too.
//   - fan-in emits one post per source in pull order (matching the queue's FIFO),
//     each reading that pull's `got` register. Two guards: the source accounts
//     must be statically distinct (else runtime `compactAt` coalesces adjacent
//     same-account funds into one posting, changing the output); and the region
//     must contain no jmp (an inorder early-exit would skip pulls, leaving their
//     `got` stale and forcing the posts to reference garbage). So fan-in fires for
//     allotment sources and send-all inorder, not capped inorder.
//
// Regions are delimited by set_current_asset (which starts every send statement,
// at which point the queue is provably empty). Anything the pass can't prove is
// left as pull/send and still runs through the queue.
type fundsBypass struct{}

func (fundsBypass) name() string { return "funds-bypass" }

func (fundsBypass) run(instrs []vInstr) ([]vInstr, bool) {
	// edits[i] replaces instrs[i] with a (possibly multi-instruction) slice; an
	// index absent from edits is kept as-is.
	edits := map[int][]vInstr{}

	regionStart := -1
	flush := func(end int) {
		if regionStart >= 0 {
			bypassRegion(instrs, regionStart, end, edits)
		}
	}
	for i, in := range instrs {
		if _, ok := in.(setCurrentAsset); ok {
			flush(i)
			regionStart = i
		}
	}
	flush(len(instrs))

	if len(edits) == 0 {
		return instrs, false
	}

	out := make([]vInstr, 0, len(instrs))
	for i, in := range instrs {
		if repl, ok := edits[i]; ok {
			out = append(out, repl...)
			continue
		}
		out = append(out, in)
	}
	return out, true
}

// bypassRegion classifies the send statement in instrs[start:end] and records its
// rewrite into edits, or leaves it untouched if it isn't a bypassable shape.
func bypassRegion(instrs []vInstr, start, end int, edits map[int][]vInstr) {
	var pulls, sends []int
	hasJmp := false
	for i := start; i < end; i++ {
		switch instrs[i].(type) {
		case pullAccount:
			pulls = append(pulls, i)
		case sendToAccount:
			sends = append(sends, i)
		case jmpIfZero:
			hasJmp = true
		}
	}
	P, S := len(pulls), len(sends)
	if P == 0 || S == 0 {
		return
	}

	toTake := func(pi int) takeAccount {
		p := instrs[pi].(pullAccount)
		return takeAccount{
			dest:        p.dest,
			account:     p.account,
			cap:         p.cap,
			overdraft:   p.overdraft,
			color:       p.color,
			boundedZero: p.boundedZero,
		}
	}

	if P == 1 {
		p := instrs[pulls[0]].(pullAccount)
		// every send must route to a real account: a kept/refund send would leave
		// the source short after take's single full debit.
		for _, si := range sends {
			if instrs[si].(sendToAccount).account == nil {
				return
			}
		}

		if S == 1 {
			sd := instrs[sends[0]].(sendToAccount)
			if sd.cap != nil { // single source+dest is always a drain-all send
				return
			}
			edits[pulls[0]] = []vInstr{toTake(pulls[0])}
			edits[sends[0]] = []vInstr{postAccount{srcAccount: p.account, dstAccount: *sd.account, amount: p.dest, color: p.color}}
			return
		}

		// fan-out: the single source feeds every send. Sound only for an allotment
		// destination — every send's cap must be a makeAllotment output (a
		// non-negative share summing to got). Inorder dests can carry a negative
		// `max` and are left to the queue.
		allotOut := allotmentOutputs(instrs, start, end)
		for _, si := range sends {
			sd := instrs[si].(sendToAccount)
			if sd.cap == nil || !allotOut[*sd.cap] {
				return
			}
		}
		edits[pulls[0]] = []vInstr{toTake(pulls[0])}
		for _, si := range sends {
			sd := instrs[si].(sendToAccount)
			edits[si] = []vInstr{postAccount{srcAccount: p.account, dstAccount: *sd.account, amount: *sd.cap, color: p.color}}
		}
		return
	}

	// fan-in: P>1 sources drained by a single drain-all send to one account.
	if S != 1 {
		return // N×M: keep the queue
	}
	sd := instrs[sends[0]].(sendToAccount)
	if sd.cap != nil || sd.account == nil {
		return
	}
	if hasJmp {
		return // inorder early-exit: a skipped pull would leave its `got` stale
	}
	if !distinctConstAccounts(instrs, pulls) {
		return // aliasing sources would coalesce into one posting in the queue
	}

	posts := make([]vInstr, 0, P)
	for _, pi := range pulls {
		p := instrs[pi].(pullAccount)
		edits[pi] = []vInstr{toTake(pi)}
		posts = append(posts, postAccount{srcAccount: p.account, dstAccount: *sd.account, amount: p.dest, color: p.color})
	}
	edits[sends[0]] = posts // one post per source, in pull (FIFO) order
}

// allotmentOutputs collects the destination registers written by every
// makeAllotment in instrs[start:end]. In a single-source (P==1) region the only
// allotment is the destination split, so a send whose cap is one of these regs is
// a provably non-negative allotment share.
func allotmentOutputs(instrs []vInstr, start, end int) map[reg]bool {
	out := map[reg]bool{}
	for i := start; i < end; i++ {
		if ma, ok := instrs[i].(makeAllotment); ok {
			for _, d := range ma.dest {
				out[d] = true
			}
		}
	}
	return out
}

// distinctConstAccounts reports whether every pull's account is a compile-time
// constant string (a loadStr) and all are pairwise distinct — the condition under
// which per-source direct posts reproduce the queue's postings exactly (no
// compactAt coalescing of same-account funds).
func distinctConstAccounts(instrs []vInstr, pulls []int) bool {
	seen := make(map[string]bool, len(pulls))
	for _, pi := range pulls {
		v, ok := constStrOf(instrs, instrs[pi].(pullAccount).account)
		if !ok || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}

// constStrOf returns the constant value of r if r is defined by a loadStr.
func constStrOf(instrs []vInstr, r reg) (string, bool) {
	for _, in := range instrs {
		if ls, ok := in.(loadStr); ok && ls.dest == r {
			return ls.value, true
		}
	}
	return "", false
}
