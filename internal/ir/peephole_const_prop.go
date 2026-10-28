package ir

import (
	"maps"
	"math/big"
	"slices"
)

// constProp replaces a pure op whose result is the same int, str or bool on
// every path reaching it with a load of that value. Portion results are not
// materialized, there is no portion load, but they are still tracked for the
// ops that read them.
func constProp(instrs []Instr) []Instr {
	var out []Instr
	walkConsts(instrs, func(i int, env constEnv) {
		if env == nil {
			return
		}
		in := instrs[i]
		switch in.(type) {
		case BinaryOp, UnaryOp:
		default:
			return
		}
		v, ok := evalPure(in, env)
		if !ok {
			return
		}
		load, ok := loadOf(in.dests()[0], v)
		if !ok {
			return
		}
		if out == nil {
			out = slices.Clone(instrs)
		}
		out[i] = load
	})
	return out
}

// A constEnv maps a register to its value: a *big.Int, string, bool or
// *big.Rat, never mutated once stored.
type constEnv map[Reg]any

// walkConsts calls visit with the registers known before each instruction, or
// nil when no path reaches it. env is only valid during the call.
//
// Jumps only go forward, so every edge into a label is known when the walk
// reaches it and one pass is exact.
func walkConsts(instrs []Instr, visit func(i int, env constEnv)) {
	incoming := map[Label]constEnv{}
	addEdge := func(l Label, env constEnv) {
		if prev, ok := incoming[l]; ok {
			incoming[l] = meetEnv(prev, env)
		} else {
			incoming[l] = env
		}
	}

	env := constEnv{}
	for i, in := range instrs {
		if l, ok := in.(LabelMarker); ok {
			if inc, ok := incoming[l.Label]; ok {
				if env == nil {
					env = inc
				} else {
					env = meetEnv(env, inc)
				}
			}
			continue
		}
		visit(i, env)
		if env == nil {
			continue
		}

		switch in := in.(type) {
		case Jmp:
			addEdge(in.Target, env)
			env = nil

		case JmpIfFalse:
			env = branch(env, in.Cond, false, in.Target, addEdge)

		case JmpIfTrue:
			env = branch(env, in.Cond, true, in.Target, addEdge)

		case CheckEnoughFunds:
			// it fails unless got == needed
			if v, ok := env[in.Needed]; ok {
				env[in.Got] = v
			} else if v, ok := env[in.Got]; ok {
				env[in.Needed] = v
			}

		default:
			dests := in.dests()
			if len(dests) == 0 {
				continue
			}
			v, known := evalPure(in, env)
			for _, d := range dests {
				delete(env, d)
			}
			if known {
				env[dests[0]] = v
			}
		}
	}
}

// branch records the taken edge of a jump on cond == jumpsOn and returns the
// fallthrough env, nil when cond makes either edge impossible.
func branch(env constEnv, cond Reg, jumpsOn bool, target Label, addEdge func(Label, constEnv)) constEnv {
	v, known := env[cond].(bool)
	if !known || v == jumpsOn {
		taken := maps.Clone(env)
		taken[cond] = jumpsOn
		addEdge(target, taken)
	}
	if known && v == jumpsOn {
		return nil
	}
	env[cond] = !jumpsOn
	return env
}

func meetEnv(a, b constEnv) constEnv {
	out := constEnv{}
	for r, v := range a {
		if w, ok := b[r]; ok && sameConst(v, w) {
			out[r] = v
		}
	}
	return out
}

func sameConst(a, b any) bool {
	switch a := a.(type) {
	case *big.Int:
		b, ok := b.(*big.Int)
		return ok && a.Cmp(b) == 0
	case *big.Rat:
		b, ok := b.(*big.Rat)
		return ok && a.Cmp(b) == 0
	case string:
		b, ok := b.(string)
		return ok && a == b
	case bool:
		b, ok := b.(bool)
		return ok && a == b
	}
	return false
}

func loadOf(dest Reg, v any) (Instr, bool) {
	switch v := v.(type) {
	case *big.Int:
		l := LoadInt{Dest: dest}
		l.Value.Set(v)
		return l, true
	case string:
		return LoadStr{Dest: dest, Value: v}, true
	case bool:
		return ConstBool{Dest: dest, Value: v}, true
	}
	return nil, false
}

// evalPure is the value a load or pure op writes under env. Each case must
// compute exactly what the VM's opcode does.
func evalPure(in Instr, env constEnv) (any, bool) {
	switch in := in.(type) {
	case LoadInt:
		return &in.Value, true
	case LoadStr:
		return in.Value, true
	case ConstBool:
		return in.Value, true
	case BinaryOp:
		l, lok := env[in.Left]
		r, rok := env[in.Right]
		if lok && rok {
			return evalBinary(in.Op, l, r)
		}
	case UnaryOp:
		if a, ok := env[in.Arg]; ok {
			return evalUnary(in.Op, a)
		}
	}
	return nil, false
}

func evalBinary(op BinKind, l, r any) (any, bool) {
	li, lInt := l.(*big.Int)
	ri, rInt := r.(*big.Int)
	lp, lPort := l.(*big.Rat)
	rp, rPort := r.(*big.Rat)
	ls, lStr := l.(string)
	rs, rStr := r.(string)
	ints, portions, strs := lInt && rInt, lPort && rPort, lStr && rStr

	switch op.(type) {
	case OpAddInt:
		if ints {
			return new(big.Int).Add(li, ri), true
		}
	case OpSubInt:
		if ints {
			return new(big.Int).Sub(li, ri), true
		}
	case OpLtInt:
		if ints {
			return li.Cmp(ri) < 0, true
		}
	case OpEqInt:
		if ints {
			return li.Cmp(ri) == 0, true
		}
	case OpAddString:
		if strs {
			return ls + rs, true
		}
	case OpStrEq:
		if strs {
			return ls == rs, true
		}
	case OpLtPortion:
		if portions {
			return lp.Cmp(rp) < 0, true
		}
	case OpEqPortion:
		if portions {
			return lp.Cmp(rp) == 0, true
		}
	case OpAddPortion:
		if portions {
			return new(big.Rat).Add(lp, rp), true
		}
	case OpSubPortion:
		if portions {
			return new(big.Rat).Sub(lp, rp), true
		}
	case OpMulPortion:
		if portions {
			return new(big.Rat).Mul(lp, rp), true
		}
	case OpMakePortion:
		// a zero denominator is a runtime error, left to the VM
		if ints && ri.Sign() != 0 {
			return new(big.Rat).SetFrac(li, ri), true
		}
	case OpMonetaryToString:
		if lStr && rInt {
			return ls + " " + ri.String(), true
		}
	}
	return nil, false
}

func evalUnary(op UnKind, a any) (any, bool) {
	ai, aInt := a.(*big.Int)
	ap, aPort := a.(*big.Rat)
	ab, aBool := a.(bool)

	switch op.(type) {
	case OpIntCopy, OpPortionCopy, OpStrCopy, OpBoolCopy:
		return a, true
	case OpNegInt:
		if aInt {
			return new(big.Int).Neg(ai), true
		}
	case OpIntToString:
		if aInt {
			return ai.String(), true
		}
	case OpIsZero:
		if aInt {
			return ai.Sign() == 0, true
		}
	case OpNot:
		if aBool {
			return !ab, true
		}
	case OpPortionToString:
		if aPort {
			return ap.String(), true
		}
	case OpIntToPortion:
		if aInt {
			return new(big.Rat).SetInt(ai), true
		}
	case OpPortionToInt:
		if aPort {
			return new(big.Int).Div(ap.Num(), ap.Denom()), true
		}
	}
	return nil, false
}
