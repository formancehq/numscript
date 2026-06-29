package compiler

// monetaryFold removes the round-trip of building a monetary then immediately
// decomposing it:
//
//	D = mk_monetary(A, M)
//	X = get_asset(D)      => uses of X are rewritten to A
//	Y = get_amount(D)     => uses of Y are rewritten to M
//
// The get_* instructions are dropped here; the now-unused mk_monetary is removed
// by the dead-code pass. This is the common shape of `send [ASSET amt] (...)`.
//
// It only fires when the registers involved (D, X/Y, A/M) are single-assignment.
// A single-def register holds one value throughout its live range, and (by the
// def-before-use contract) its definition precedes its uses — so rewriting a use
// of X to A is value-preserving regardless of control flow.
type monetaryFold struct{}

func (monetaryFold) name() string { return "monetary-fold" }

func (monetaryFold) run(instrs []vInstr) ([]vInstr, bool) {
	defs := defCount(instrs)

	type components struct{ asset, amount reg }
	monByDest := map[reg]components{}
	for _, in := range instrs {
		b, ok := in.(binaryOp)
		if !ok {
			continue
		}
		if _, isMk := b.op.(opMakeMonetary); !isMk {
			continue
		}
		if defs[b.dest] == 1 {
			monByDest[b.dest] = components{asset: b.left, amount: b.right}
		}
	}
	if len(monByDest) == 0 {
		return instrs, false
	}

	subst := map[reg]reg{}
	drop := map[int]bool{}
	for idx, in := range instrs {
		u, ok := in.(unaryOp)
		if !ok {
			continue
		}
		m, isMon := monByDest[u.arg]
		if !isMon {
			continue
		}
		switch u.op.(type) {
		case opGetAsset:
			if defs[u.dest] == 1 && defs[m.asset] == 1 {
				subst[u.dest] = m.asset
				drop[idx] = true
			}
		case opGetAmount:
			if defs[u.dest] == 1 && defs[m.amount] == 1 {
				subst[u.dest] = m.amount
				drop[idx] = true
			}
		}
	}
	if len(subst) == 0 {
		return instrs, false
	}

	resolve := func(r reg) reg {
		for {
			n, ok := subst[r]
			if !ok {
				return r
			}
			r = n
		}
	}

	out := make([]vInstr, 0, len(instrs))
	for idx, in := range instrs {
		if drop[idx] {
			continue
		}
		out = append(out, in.mapSources(resolve))
	}
	return out, true
}
