package ir

import "slices"

// foldConstConditions replaces a comparison whose operands are constants with
// the bool it always yields: str_eq, eq_int and lt_int over two loads, is_zero
// over one. These are the conditions of the branches the compiler emits around
// every literal account (is it @world?) and every literal amount (is the cap
// already zero?), and a constant condition is what pruneDeadBranches acts on.
func foldConstConditions(instrs []Instr) []Instr {
	defs := regDefs(instrs)

	var out []Instr
	fold := func(i int, dest Reg, value bool) {
		if out == nil {
			out = slices.Clone(instrs)
		}
		out[i] = ConstBool{Dest: dest, Value: value}
	}

	for i, in := range instrs {
		switch in := in.(type) {
		case BinaryOp:
			if v, ok := foldComparison(instrs, defs, in); ok {
				fold(i, in.Dest, v)
			}
		case UnaryOp:
			if _, isZero := in.Op.(OpIsZero); !isZero {
				continue
			}
			if n, ok := constIntOf(instrs, defs, in.Arg); ok {
				fold(i, in.Dest, n.Sign() == 0)
			}
		}
	}
	return out
}

func foldComparison(instrs []Instr, defs map[Reg][]int, in BinaryOp) (bool, bool) {
	switch in.Op.(type) {
	case OpStrEq:
		l, lok := constStrOf(instrs, defs, in.Left)
		r, rok := constStrOf(instrs, defs, in.Right)
		return l == r, lok && rok

	case OpEqInt, OpLtInt:
		l, lok := constIntOf(instrs, defs, in.Left)
		r, rok := constIntOf(instrs, defs, in.Right)
		if !lok || !rok {
			return false, false
		}
		cmp := l.Cmp(r)
		if _, eq := in.Op.(OpEqInt); eq {
			return cmp == 0, true
		}
		return cmp < 0, true

	default:
		return false, false
	}
}
