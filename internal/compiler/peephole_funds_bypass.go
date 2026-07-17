package compiler

// fundsBypass fuses a 1-source/1-destination send into a single direct move that
// skips the runtime funds queue.
//
// The queue exists to pair N funding sources with M destinations when the pairing
// is data-dependent. When a send has exactly one source feeding exactly one
// destination, the pair is static: the source is pulled and then immediately
// drained in full to that one destination. That is a pull whose queued funds are
// consumed by a single drain-all send, with nothing else touching the queue in
// between — so the queue round-trip (a queued-source allocation + the Send drain)
// is pure overhead. This pass rewrites that shape:
//
//	got = pull_account(account: A, cap, overdraft, color)
//	...queue-neutral ops (check_enough_funds, loads, ...)...
//	send_to_account(account: B)              // drain-all: no cap
//	=>
//	got = take_account(account: A, cap, overdraft, color)   // debit, no queue
//	...queue-neutral ops...
//	post_account(src: A, dst: B, amount: got)               // posting, no debit
//
// The pull becomes a take (debit at the source site); the send becomes a post
// (posting at the destination site). Each stays in its original position, so
// nothing is reordered — in particular the check_enough_funds between them still
// reads `got` after the take writes it and before the post. The VM runs these as
// Take/PostDirect, skipping the queued-source allocation and the Send drain.
//
// Detection is a whole-list abstract interpretation of the queue: it tracks the
// number of pending pulls and only fires when a drain-all send finds the queue
// holding exactly one entry. Fan-out (one pull, several capped sends), fan-in
// (several pulls, one drain-all send) and the general N×M case are left as
// pull/send — they still run through the runtime queue. Only single->single is
// rewritten here, which is guard-free (no source-aliasing / posting-coalescing
// concerns, unlike fan-in).
//
// Soundness of the abstract queue:
//   - The queue is empty at every statement boundary. set_current_asset starts a
//     send statement, so tracking resets there (the queue is provably empty).
//   - A capped send may leave the front partially consumed — a statically-unknown
//     remainder — so it taints tracking until the next reset.
//   - A jump/label means control flow we don't model, so it also taints tracking.
//   - Only a drain-all send over a depth-1 queue is matched; anything else just
//     updates (or taints) the tracked state without rewriting.
type fundsBypass struct{}

func (fundsBypass) name() string { return "funds-bypass" }

func (fundsBypass) run(instrs []vInstr) ([]vInstr, bool) {
	type match struct {
		pullIdx int
		sendIdx int
		dst     reg
	}
	var matches []match

	depth := 0        // number of pending pulls in the queue (valid when !tainted)
	pendingPull := -1 // index of the sole pending pull when depth == 1
	tainted := false  // a capped send / control flow left the queue state unknown

	for i, in := range instrs {
		switch v := in.(type) {
		case setCurrentAsset:
			// start of a send statement: the queue is provably empty here
			depth, pendingPull, tainted = 0, -1, false

		case pullAccount:
			if tainted {
				break
			}
			depth++
			if depth == 1 {
				pendingPull = i
			} else {
				pendingPull = -1
			}

		case sendToAccount:
			if v.cap == nil {
				// drain-all: empties the queue regardless of prior state
				if !tainted && depth == 1 && v.account != nil {
					matches = append(matches, match{pullIdx: pendingPull, sendIdx: i, dst: *v.account})
				}
				depth, pendingPull, tainted = 0, -1, false
			} else {
				// capped: unknown remainder left at the front
				tainted = true
			}

		case takeAccount, postAccount:
			// already fused (defensive: neither queues nor drains)

		case jmpIfZero, labelMarker:
			tainted = true // control flow we don't model

		default:
			// queue-neutral: loads, arithmetic, asserts, check_enough_funds, save,
			// make_allotment, fetch_balance, meta — none touch the source queue
		}
	}

	if len(matches) == 0 {
		return instrs, false
	}

	replace := make(map[int]vInstr, 2*len(matches))
	for _, m := range matches {
		p := instrs[m.pullIdx].(pullAccount)
		replace[m.pullIdx] = takeAccount{
			dest:        p.dest,
			account:     p.account,
			cap:         p.cap,
			overdraft:   p.overdraft,
			color:       p.color,
			boundedZero: p.boundedZero,
		}
		replace[m.sendIdx] = postAccount{
			srcAccount: p.account,
			dstAccount: m.dst,
			amount:     p.dest,
			color:      p.color,
		}
	}

	out := make([]vInstr, 0, len(instrs))
	for i, in := range instrs {
		if r, ok := replace[i]; ok {
			out = append(out, r)
			continue
		}
		out = append(out, in)
	}
	return out, true
}
