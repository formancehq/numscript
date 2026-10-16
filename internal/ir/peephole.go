package ir

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
// pruning, and a dropped check frees the values only it read for deadCode.
func DefaultPeepholes() []Peephole {
	return []Peephole{
		constProp,
		dropProvenAsserts,
		pruneDeadBranches,
		postFromUnbounded,
		deadCode,
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
