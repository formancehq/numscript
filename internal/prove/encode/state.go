package encode

import (
	"fmt"
	"sort"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/prove/smt"
)

// slot is a queued funding source: the pull that created it, its account, and
// what is still queued (0 once consumed, rewound, or never pulled on this path).
type slot struct {
	pos     int
	account string
	amount  smt.Term
}

// frame is an open mark region: the guard under which each candidate
// mark_push position is the one this region was opened at. A path through no
// join has a single candidate guarded by true.
type frame map[int]smt.Term

type state struct {
	reach smt.Term
	alive smt.Term
	dead  bool

	regs  map[ir.Reg]regVal
	bal   map[Key]smt.Term
	slots map[int]*slot
	sent  map[Key]smt.Term
	recv  map[Key]smt.Term
	moved map[MoveKey]smt.Term
	asset strVal
	marks []frame
}

func newState() *state {
	return &state{
		reach: smt.Bool(true),
		alive: smt.Bool(true),
		regs:  map[ir.Reg]regVal{},
		bal:   map[Key]smt.Term{},
		slots: map[int]*slot{},
		sent:  map[Key]smt.Term{},
		recv:  map[Key]smt.Term{},
		moved: map[MoveKey]smt.Term{},
	}
}

func deadState() *state {
	s := newState()
	s.reach = smt.Bool(false)
	s.alive = smt.Bool(false)
	s.dead = true
	return s
}

func (s *state) clone() *state {
	out := &state{
		reach: s.reach,
		alive: s.alive,
		dead:  s.dead,
		regs:  make(map[ir.Reg]regVal, len(s.regs)),
		bal:   make(map[Key]smt.Term, len(s.bal)),
		slots: make(map[int]*slot, len(s.slots)),
		sent:  make(map[Key]smt.Term, len(s.sent)),
		recv:  make(map[Key]smt.Term, len(s.recv)),
		moved: make(map[MoveKey]smt.Term, len(s.moved)),
		asset: s.asset,
	}
	for k, v := range s.regs {
		out.regs[k] = v
	}
	for k, v := range s.bal {
		out.bal[k] = v
	}
	for k, v := range s.slots {
		c := *v
		out.slots[k] = &c
	}
	for k, v := range s.sent {
		out.sent[k] = v
	}
	for k, v := range s.recv {
		out.recv[k] = v
	}
	for k, v := range s.moved {
		out.moved[k] = v
	}
	for _, f := range s.marks {
		c := make(frame, len(f))
		for k, v := range f {
			c[k] = v
		}
		out.marks = append(out.marks, c)
	}
	return out
}

func (s *state) sortedSlots() []*slot {
	out := make([]*slot, 0, len(s.slots))
	for _, sl := range s.slots {
		out = append(out, sl)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].pos < out[j].pos })
	return out
}

// merge joins the live edges reaching a label into one state. Every value is
// an ite chain over the edges' reach conditions, named by a define-fun.
func (e *encoder) merge(pos int, edges []*state) (*state, error) {
	live := edges[:0]
	for _, s := range edges {
		if !s.dead {
			live = append(live, s)
		}
	}
	switch len(live) {
	case 0:
		return deadState(), nil
	case 1:
		return live[0], nil
	}

	out := newState()
	reach := make([]smt.Term, len(live))
	alive := make([]smt.Term, len(live))
	for i, s := range live {
		reach[i] = s.reach
		alive[i] = s.alive
	}
	out.reach = e.def(fmt.Sprintf("reach@%d", pos), smt.Or(reach...))
	out.alive = e.def(fmt.Sprintf("alive@%d", pos), smt.Or(alive...))

	// chain builds ite(reach_1, v_1, ite(reach_2, v_2, ... v_n)) over the
	// edges that have a value; the last one is the default
	chain := func(vals []smt.Term, guards []smt.Term) smt.Term {
		t := vals[len(vals)-1]
		for i := len(vals) - 2; i >= 0; i-- {
			t = smt.Ite(guards[i], vals[i], t)
		}
		return t
	}

	regSet := map[ir.Reg]struct{}{}
	for _, s := range live {
		for r := range s.regs {
			regSet[r] = struct{}{}
		}
	}
	regs := make([]ir.Reg, 0, len(regSet))
	for r := range regSet {
		regs = append(regs, r)
	}
	sort.Slice(regs, func(i, j int) bool { return regs[i] < regs[j] })
	for _, r := range regs {
		var vals []regVal
		var guards []smt.Term
		for i, s := range live {
			if v, ok := s.regs[r]; ok {
				vals = append(vals, v)
				guards = append(guards, reach[i])
			}
		}
		first := vals[0]
		switch {
		case first.p != nil:
			for _, v := range vals[1:] {
				if v.p.String() != first.p.String() {
					return nil, e.unsupported(pos, fmt.Sprintf("portion register $r%d takes different values on joining paths", r))
				}
			}
			out.regs[r] = first
		case first.t == nil:
			merged := first.s
			for _, v := range vals[1:] {
				if !v.s.known || !merged.known || v.s.s != merged.s {
					if !v.s.known || !merged.known {
						merged = strVal{}
						continue
					}
					return nil, e.unsupported(pos, fmt.Sprintf("string register $r%d takes different values on joining paths", r))
				}
			}
			out.regs[r] = regVal{s: merged}
		default:
			ts := make([]smt.Term, len(vals))
			for i, v := range vals {
				ts[i] = v.t
			}
			out.regs[r] = regVal{t: e.def(fmt.Sprintf("$r%d@%d", r, pos), chain(ts, guards))}
		}
	}

	keySet := map[Key]struct{}{}
	for _, s := range live {
		for k := range s.bal {
			keySet[k] = struct{}{}
		}
	}
	for _, k := range sortedKeys(keySet) {
		vals := make([]smt.Term, len(live))
		for i, s := range live {
			vals[i] = e.balanceOf(s, k)
		}
		out.bal[k] = e.def(fmt.Sprintf("bal(%s)@%d", k, pos), chain(vals, reach))
	}

	slotSet := map[int]struct{}{}
	for _, s := range live {
		for p := range s.slots {
			slotSet[p] = struct{}{}
		}
	}
	slotPos := make([]int, 0, len(slotSet))
	for p := range slotSet {
		slotPos = append(slotPos, p)
	}
	sort.Ints(slotPos)
	for _, p := range slotPos {
		vals := make([]smt.Term, len(live))
		var account string
		for i, s := range live {
			if sl, ok := s.slots[p]; ok {
				vals[i] = sl.amount
				account = sl.account
			} else {
				vals[i] = smt.Int64(0)
			}
		}
		out.slots[p] = &slot{pos: p, account: account, amount: e.def(fmt.Sprintf("slot#%d@%d", p, pos), chain(vals, reach))}
	}

	mergeKeyed := func(name string, get func(*state) map[Key]smt.Term, dst map[Key]smt.Term) {
		set := map[Key]struct{}{}
		for _, s := range live {
			for k := range get(s) {
				set[k] = struct{}{}
			}
		}
		for _, k := range sortedKeys(set) {
			vals := make([]smt.Term, len(live))
			for i, s := range live {
				vals[i] = orZero(get(s)[k])
			}
			dst[k] = e.def(fmt.Sprintf("%s(%s)@%d", name, k, pos), chain(vals, reach))
		}
	}
	mergeKeyed("sent", func(s *state) map[Key]smt.Term { return s.sent }, out.sent)
	mergeKeyed("received", func(s *state) map[Key]smt.Term { return s.recv }, out.recv)

	moveSet := map[MoveKey]struct{}{}
	for _, s := range live {
		for m := range s.moved {
			moveSet[m] = struct{}{}
		}
	}
	for _, m := range sortedMoveKeys(moveSet) {
		vals := make([]smt.Term, len(live))
		for i, s := range live {
			vals[i] = orZero(s.moved[m])
		}
		out.moved[m] = e.def(fmt.Sprintf("moved(%s)@%d", m, pos), chain(vals, reach))
	}

	out.asset = live[0].asset
	for _, s := range live[1:] {
		if s.asset != out.asset {
			if !s.asset.known || !out.asset.known {
				out.asset = strVal{}
				continue
			}
			return nil, e.unsupported(pos, "the current asset differs on joining paths")
		}
	}

	depth := len(live[0].marks)
	for _, s := range live[1:] {
		if len(s.marks) != depth {
			return nil, fmt.Errorf("encode: instruction %d: paths join with different numbers of open marks", pos)
		}
	}
	for d := 0; d < depth; d++ {
		merged := frame{}
		cands := map[int]struct{}{}
		for _, s := range live {
			for c := range s.marks[d] {
				cands[c] = struct{}{}
			}
		}
		candList := make([]int, 0, len(cands))
		for c := range cands {
			candList = append(candList, c)
		}
		sort.Ints(candList)
		for _, c := range candList {
			var guards []smt.Term
			for i, s := range live {
				if g, ok := s.marks[d][c]; ok {
					guards = append(guards, smt.And(reach[i], g))
				}
			}
			merged[c] = e.def(fmt.Sprintf("mark#%d@%d", c, pos), smt.Or(guards...))
		}
		out.marks = append(out.marks, merged)
	}

	return out, nil
}

func orZero(t smt.Term) smt.Term {
	if t == nil {
		return smt.Int64(0)
	}
	return t
}
