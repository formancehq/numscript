package compiler

// A peephole is an optional optimization pass over the virtual instruction
// stream. run returns the (possibly) rewritten program and whether it changed
// anything. Each concrete pass lives in its own peephole_*.go file and is
// independently testable; optimize composes them and runs to a fixpoint.
type peephole interface {
	name() string
	run(instrs []vInstr) ([]vInstr, bool)
}

// defaultPeepholes is the ordered set applied by the optional optimization pass.
func defaultPeepholes() []peephole {
	return []peephole{
		monetaryFold{},
		deadCode{},
	}
}

// optimize runs the given peepholes to a fixpoint: it sweeps them in order,
// repeating until a full sweep changes nothing. Optimization is OPTIONAL —
// callers may assemble the raw instructions directly and skip this entirely.
func optimize(instrs []vInstr, opts []peephole) []vInstr {
	for {
		changed := false
		for _, opt := range opts {
			next, c := opt.run(instrs)
			if c {
				instrs = next
				changed = true
			}
		}
		if !changed {
			return instrs
		}
	}
}

// --- shared analysis helpers (used by multiple passes) ---

// defCount returns, for each register, how many instructions write it. A
// register with count 1 is single-assignment, which lets a pass substitute its
// uses safely (it holds one value throughout its live range).
func defCount(instrs []vInstr) map[reg]int {
	counts := map[reg]int{}
	for _, in := range instrs {
		for _, d := range in.dests() {
			counts[d]++
		}
	}
	return counts
}

// usedRegs returns the set of registers read by some instruction.
func usedRegs(instrs []vInstr) map[reg]bool {
	used := map[reg]bool{}
	for _, in := range instrs {
		for _, r := range in.sources() {
			used[r] = true
		}
	}
	return used
}
