package ir

import "math/big"

// A Peephole rewrites an instruction stream and returns the result, or nil when
// it found nothing to change. A pass never mutates its input.
//
// Every pass must preserve the program's observable behavior exactly: the same
// postings, metadata and errors, including which error surfaces when several
// could.
type Peephole func(instrs []Instr) []Instr

// maxOptimizeSweeps bounds the fixed-point iteration. Each pass reaches a fixed
// point on its own; their composition is not proven to.
const maxOptimizeSweeps = 16

// DefaultPeepholes is the ordered set Optimize applies. The order matters for
// how fast a sweep converges, not for the result: the folding feeds the
// pruning.
func DefaultPeepholes() []Peephole {
	return []Peephole{
		foldConstConditions,
		pruneDeadBranches,
	}
}

// Optimize applies DefaultPeepholes in order, sweeping again while a sweep
// changed something, up to maxOptimizeSweeps.
//
// Optimization is opt-in: the compiler assembles the plain stream unless asked
// otherwise, and the optimized stream is behaviorally identical to it.
func Optimize(instrs []Instr) []Instr {
	return optimize(instrs, DefaultPeepholes())
}

func optimize(instrs []Instr, passes []Peephole) []Instr {
	for range maxOptimizeSweeps {
		changed := false
		for _, pass := range passes {
			if next := pass(instrs); next != nil {
				instrs = next
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return instrs
}

// --- analysis helpers shared by the passes ---

// regDefs returns, for each register, the indices of the instructions writing it.
func regDefs(instrs []Instr) map[Reg][]int {
	defs := map[Reg][]int{}
	for i, in := range instrs {
		for _, r := range in.dests() {
			defs[r] = append(defs[r], i)
		}
	}
	return defs
}

// soleDef returns the one instruction writing r, if exactly one does. Jumps only
// go forward and the typechecker rejects a read before any write, so such a
// register holds that instruction's value at every read.
func soleDef(instrs []Instr, defs map[Reg][]int, r Reg) (Instr, bool) {
	d := defs[r]
	if len(d) != 1 {
		return nil, false
	}
	return instrs[d[0]], true
}

func constStrOf(instrs []Instr, defs map[Reg][]int, r Reg) (string, bool) {
	def, ok := soleDef(instrs, defs, r)
	if !ok {
		return "", false
	}
	ls, ok := def.(LoadStr)
	if !ok {
		return "", false
	}
	return ls.Value, true
}

func constIntOf(instrs []Instr, defs map[Reg][]int, r Reg) (*big.Int, bool) {
	def, ok := soleDef(instrs, defs, r)
	if !ok {
		return nil, false
	}
	li, ok := def.(LoadInt)
	if !ok {
		return nil, false
	}
	return &li.Value, true
}

func constBoolOf(instrs []Instr, defs map[Reg][]int, r Reg) (bool, bool) {
	def, ok := soleDef(instrs, defs, r)
	if !ok {
		return false, false
	}
	cb, ok := def.(ConstBool)
	if !ok {
		return false, false
	}
	return cb.Value, true
}
