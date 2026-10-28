package ir

import "maps"

// deadCode removes a pure instruction whose dests no path reads. Store reads,
// pulls and every check stay, as they can fail or have effects.
func deadCode(instrs []Instr) []Instr {
	// a mk_portion fails on a zero denominator, so it is pure only where the
	// denominator is known not to be zero
	safePortions := map[int]bool{}
	walkConsts(instrs, func(i int, env constEnv) {
		b, ok := instrs[i].(BinaryOp)
		if !ok || env == nil {
			return
		}
		if _, mk := b.Op.(OpMakePortion); mk {
			if _, ok := evalPure(b, env); ok {
				safePortions[i] = true
			}
		}
	})

	pure := func(i int) bool {
		switch in := instrs[i].(type) {
		case LoadInt, LoadStr, ConstBool, LoadVar, UnaryOp:
			return true
		case BinaryOp:
			if _, mk := in.Op.(OpMakePortion); mk {
				return safePortions[i]
			}
			return true
		}
		return false
	}

	// backwards liveness: jumps only go forward, so a label's live set is
	// known before any jump to it is reached
	labelLive := map[Label]map[Reg]bool{}
	live := map[Reg]bool{}
	dead := map[int]bool{}
	for i := len(instrs) - 1; i >= 0; i-- {
		in := instrs[i]
		switch j := in.(type) {
		case LabelMarker:
			labelLive[j.Label] = maps.Clone(live)
			continue
		case Jmp:
			live = maps.Clone(labelLive[j.Target])
			continue
		case JmpIfFalse:
			maps.Copy(live, labelLive[j.Target])
			live[j.Cond] = true
			continue
		case JmpIfTrue:
			maps.Copy(live, labelLive[j.Target])
			live[j.Cond] = true
			continue
		}

		dests := in.dests()
		if len(dests) != 0 && pure(i) && !anyLive(live, dests) {
			dead[i] = true
			continue
		}
		for _, d := range dests {
			delete(live, d)
		}
		for _, s := range in.sources() {
			live[s] = true
		}
	}

	if len(dead) == 0 {
		return nil
	}
	out := make([]Instr, 0, len(instrs)-len(dead))
	for i, in := range instrs {
		if !dead[i] {
			out = append(out, in)
		}
	}
	return out
}

func anyLive(live map[Reg]bool, regs []Reg) bool {
	for _, r := range regs {
		if live[r] {
			return true
		}
	}
	return false
}
