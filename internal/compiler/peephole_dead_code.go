package compiler

// deadCode removes pure instructions whose destination register is never read.
// Only side-effect-free kinds are eligible (loads and arithmetic); instructions
// that mutate runtime state — pulls, sends, asset checks, jumps, labels,
// allotments, variable fetches — are always kept, even if a result is unused.
//
// This cleans up after passes like monetaryFold, which leave a mk_monetary whose
// result is no longer consumed.
type deadCode struct{}

func (deadCode) name() string { return "dead-code" }

func (deadCode) run(instrs []vInstr) ([]vInstr, bool) {
	used := usedRegs(instrs)

	out := make([]vInstr, 0, len(instrs))
	changed := false
	for _, in := range instrs {
		if isPure(in) && !anyUsed(in.dests(), used) {
			changed = true
			continue // drop: pure and its result is dead
		}
		out = append(out, in)
	}
	if !changed {
		return instrs, false
	}
	return out, true
}

func anyUsed(regs []reg, used map[reg]bool) bool {
	for _, r := range regs {
		if used[r] {
			return true
		}
	}
	return false
}

// isPure reports whether an instruction's only effect is writing its dests, so
// it can be dropped when those dests are unused. Conservative: unknown kinds are
// treated as impure (kept).
func isPure(in vInstr) bool {
	switch in.(type) {
	case loadInt, loadStr, binaryOp, unaryOp:
		return true
	default:
		return false
	}
}
