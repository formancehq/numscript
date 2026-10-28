package ir

import "slices"

// deadCredit drops the destination credit of a post_account when no later
// instruction may read that destination's balance: a run's result carries its
// postings, not its balances, and a credit is a delta that touches no store.
//
// A later bounded pull or take, a save or a balance() reads a balance. It is
// one of the destination's when its account is the destination, or is not a
// constant, or the destination itself is not a constant. Jumps only go forward,
// so the instructions after the post in program order are all it can reach.
func deadCredit(instrs []Instr) []Instr {
	accounts := constAccounts(instrs)
	destinations := map[int]string{}
	walkConsts(instrs, func(i int, env constEnv) {
		p, ok := instrs[i].(PostAccount)
		if !ok || env == nil {
			return
		}
		if v, ok := env[p.Destination].(string); ok {
			destinations[i] = v
		}
	})

	var out []Instr
	for i, in := range instrs {
		p, ok := in.(PostAccount)
		if !ok || p.NoCredit {
			continue
		}
		dst, known := destinations[i]
		if balanceReadAfter(instrs, accounts, i, dst, known) {
			continue
		}
		if out == nil {
			out = slices.Clone(instrs)
		}
		p.NoCredit = true
		out[i] = p
	}
	return out
}

func balanceReadAfter(instrs []Instr, accounts map[int]string, from int, dst string, known bool) bool {
	for j := from + 1; j < len(instrs); j++ {
		reads := false
		switch in := instrs[j].(type) {
		case PullAccount:
			reads = in.Overdraft != nil
		case TakeAccount:
			reads = in.Overdraft != nil
		case Save, FetchBalance:
			reads = true
		}
		if !reads {
			continue
		}
		if v, ok := accounts[j]; !known || !ok || v == dst {
			return true
		}
	}
	return false
}
