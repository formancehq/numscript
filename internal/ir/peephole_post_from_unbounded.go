package ir

import "slices"

// postFromUnbounded turns a send statement with a single unbounded source and a
// single destination into a direct posting, skipping the funds queue:
//
//	$pulled = pull_account(account: $src, cap: $cap)    ->  $pulled = max_zero($cap)
//	check_enough_funds($pulled, $needed)                     (unchanged)
//	..                                                       (unchanged)
//	send_to_account(account: $dst)                      ->  post_account(source: $src, destination: $dst, amount: $pulled)
//
// Nothing moves: each instruction is rewritten in place, so every check runs
// in the order it did. An unbounded pull makes max(0, cap) available without
// reading any balance, and a drain-all send posts exactly what the one pull
// queued, so the two rewrites together are the pull and the send minus the
// queue — minus the source debit too, which is what makes the rewrite worth
// anything: for an unbounded source the debit is a delta that only a later read
// of that balance would fold in, so it is dropped only when no instruction may
// read it (see sourceBalanceObserved).
//
// A statement is the instructions between two set_current_asset. The pass
// relies on the funds queue being empty when a statement starts, which compiled
// numscript guarantees: every statement drains what it pulled.
func postFromUnbounded(instrs []Instr) []Instr {
	accounts := constAccounts(instrs)

	var out []Instr
	start := 0
	for i := 0; i <= len(instrs); i++ {
		if i < len(instrs) {
			if _, ok := instrs[i].(SetCurrentAsset); !ok {
				continue
			}
		}
		if edits, ok := fuseStatement(instrs, accounts, start, i); ok {
			if out == nil {
				out = slices.Clone(instrs)
			}
			for at, in := range edits {
				out[at] = in
			}
		}
		start = i
	}
	return out
}

// fuseStatement decides whether the statement in instrs[start:end] is the
// single-unbounded-source, single-destination shape, and returns the two
// replacements keyed by index.
func fuseStatement(instrs []Instr, accounts map[int]string, start, end int) (map[int]Instr, bool) {
	pull, send := -1, -1
	for i := start; i < end; i++ {
		switch instrs[i].(type) {
		case PullAccount:
			if pull >= 0 {
				return nil, false
			}
			pull = i
		case SendToAccount:
			if send >= 0 {
				return nil, false
			}
			send = i
		case PostAccount, MarkPush, MarkEnd:
			return nil, false
		}
	}
	if pull < 0 || send < 0 || send < pull {
		return nil, false
	}

	p := instrs[pull].(PullAccount)
	s := instrs[send].(SendToAccount)
	if p.Cap == nil || p.Overdraft != nil {
		return nil, false // bounded, or the uncapped-unbounded error case
	}
	if s.Account == nil || s.Cap != nil {
		return nil, false // kept, or a partial drain
	}

	src, ok := accounts[pull]
	if !ok || sourceBalanceObserved(instrs, accounts, src) {
		return nil, false
	}

	// post_account reads these where the send was; nothing in between may
	// rewrite them, and no jump may land or leave in between.
	held := present(&p.Account, &p.Dest, p.Color, p.Scope)
	for i := pull + 1; i < send; i++ {
		switch instrs[i].(type) {
		case LabelMarker, Jmp, JmpIfFalse, JmpIfTrue:
			return nil, false
		}
		for _, d := range instrs[i].dests() {
			if slices.Contains(held, d) {
				return nil, false
			}
		}
	}

	return map[int]Instr{
		pull: UnaryOp{Op: OpMaxZero{}, Dest: p.Dest, Arg: *p.Cap},
		send: PostAccount{
			Source:           p.Account,
			Destination:      *s.Account,
			Amount:           p.Dest,
			Color:            p.Color,
			SourceScope:      p.Scope,
			DestinationScope: s.Scope,
		},
	}, true
}

// sourceBalanceObserved reports whether some instruction may read src's balance
// entry, which is what the dropped debit would have changed: a bounded pull, a
// save or a balance() read whose account is src, or is not a constant and so
// could be src at run time. An unbounded pull never reads a balance.
//
// The account name alone decides it, whatever the scope, color or asset, which
// is more conservative than the balance entry's key.
func sourceBalanceObserved(instrs []Instr, accounts map[int]string, src string) bool {
	for i, in := range instrs {
		mayBeSrc := func() bool {
			v, ok := accounts[i]
			return !ok || v == src
		}
		switch in := in.(type) {
		case PullAccount:
			if in.Overdraft != nil && mayBeSrc() {
				return true
			}
		case Save:
			if mayBeSrc() {
				return true
			}
		case FetchBalance:
			if mayBeSrc() {
				return true
			}
		}
	}
	return false
}

// constAccounts maps each pull, save and balance read to its account, when the
// account is the same constant on every path reaching it.
func constAccounts(instrs []Instr) map[int]string {
	accounts := map[int]string{}
	walkConsts(instrs, func(i int, env constEnv) {
		if env == nil {
			return
		}
		var account Reg
		switch in := instrs[i].(type) {
		case PullAccount:
			account = in.Account
		case Save:
			account = in.Account
		case FetchBalance:
			account = in.Account
		default:
			return
		}
		if v, ok := env[account].(string); ok {
			accounts[i] = v
		}
	})
	return accounts
}
