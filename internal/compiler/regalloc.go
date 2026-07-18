package compiler

import (
	"fmt"
	"sort"
)

// Register allocation
// ===================
//
// The VM has four typed register banks (int, string, portion, monetary). Each
// bank is a flat slice the VM allocates once per run (see vm.NewMachine), sized
// by the per-bank register count the assembler reports. So the number of
// registers a program declares is (a) a hard ceiling on script complexity — the
// bank index is a single byte, and 0xFF is reserved as the "no register"
// sentinel, leaving 0..254 usable — and (b) a per-run memory/zeroing cost.
//
// A virtual register (`reg`) is minted per definition by the compiler; the IR is
// effectively SSA (almost every reg is written exactly once). Which bank a reg
// lives in is NOT stored on the reg: it is implied by which resolver an op's
// assemble() calls (intReg/strReg/...). Allocation therefore happens in two
// passes over the virtual-instruction stream:
//
//  1. discovery: assemble once in discovery mode, recording each reg's bank and
//     its [firstTouch, lastTouch] index interval (see allocInput). The emitted
//     instructions are thrown away; only the liveness is kept. Under SSA the
//     touch-interval IS the live range, so we never need a def/use split.
//  2. emit: build a plan from that liveness and assemble again for real,
//     resolving each reg to its planned physical index.
//
// The plan packs a bank so that a physical slot is reused as soon as its previous
// occupant is dead. A bank's width therefore collapses from "distinct regs ever
// minted" (what a naive bump allocator would give) down to "peak number of
// simultaneously-live regs".
//
// Why linear scan is as good as graph coloring here
// -------------------------------------------------
// Register allocation is classically phrased as coloring the interference graph
// (two regs interfere iff their live ranges overlap) with as few colors as
// possible — NP-hard in general, which is why production compilers reach for
// Chaitin/Briggs graph coloring with iterated spilling.
//
// None of that cost is warranted for numscript, because of the shape of the code
// we generate: it is straight-line with FORWARD-ONLY conditional skips. The only
// branch the compiler emits is `jmpIfZero -> endLabel`, and every label is placed
// AFTER the code that jumps to it (the inorder-source lowering). There are no
// loops and no back-edges. Consequences:
//
//   - Every reg's live range is a single contiguous interval [first, last] in
//     instruction-index order. A forward jump can only skip over uses (shrinking
//     the dynamic range), never wrap around to re-enter a range, so the static
//     interval is always a safe superset of the true liveness.
//   - The interference graph of a set of intervals is an INTERVAL GRAPH, and
//     interval graphs are PERFECT: their chromatic number equals the size of the
//     largest clique. Here the largest clique = the most intervals overlapping at
//     any single point = the peak register pressure.
//   - The expiring linear-scan sweep colors an interval graph with exactly that
//     many colors, in O(n log n). It only mints a new physical slot when every
//     existing slot is occupied — i.e. only when pressure forces it.
//
// So linear scan already achieves the information-theoretic minimum number of
// registers for this CFG shape. Graph coloring cannot use fewer; it would just
// cost more to arrive at the same width. That is the whole argument for stopping
// here and not implementing coloring.

// bankId identifies one of the VM's four typed register banks. The bank a reg
// belongs to is fixed by its type (which the typechecker has already validated
// to be consistent across all of a reg's uses).
type bankId int

const (
	bankInt bankId = iota
	bankStr
	bankPortion
	bankMonetary
	numBanks
)

// --- discovery: liveness collected during the first pass ---

// bundleReq records a request for a contiguous run of slots in the emit stream.
// regs is nil for a pure scratch run (reserveContig); otherwise it is the bound
// regs (bindContig). Bundles are replayed in the same stream order in both
// passes, so their planned start indices line up positionally.
type bundleReq struct {
	bank bankId
	regs []reg // nil => scratch-only run
	n    int
}

// allocInput accumulates, during the discovery pass, everything the plan builder
// needs: each reg's bank and live interval, first-appearance order (for stable
// tie-breaking), and the contiguous-run requests in stream order.
type allocInput struct {
	bankOf    map[reg]bankId
	first     map[reg]int
	last      map[reg]int
	order     []reg // regs in first-appearance order over the whole stream
	bundles   []bundleReq
	bundleReg map[reg]bool // regs bound inside a contiguous run
}

func newAllocInput() *allocInput {
	return &allocInput{
		bankOf:    map[reg]bankId{},
		first:     map[reg]int{},
		last:      map[reg]int{},
		bundleReg: map[reg]bool{},
	}
}

// touch records that r (in bank) is referenced at instruction index idx.
func (in *allocInput) touch(bank bankId, r reg, idx int) {
	in.bankOf[r] = bank
	if _, ok := in.first[r]; !ok {
		in.first[r] = idx
		in.order = append(in.order, r)
	}
	in.last[r] = idx
}

func (in *allocInput) addScratch(bank bankId, n int) {
	in.bundles = append(in.bundles, bundleReq{bank: bank, n: n})
}

func (in *allocInput) addBound(bank bankId, regs []reg, idx int) {
	for _, r := range regs {
		in.touch(bank, r, idx)
		in.bundleReg[r] = true
	}
	in.bundles = append(in.bundles, bundleReq{bank: bank, regs: regs, n: len(regs)})
}

// --- plan: the linear-scan result the emit pass replays ---

// regPlan is a fully-precomputed allocation: a physical index for every reg
// (including bound-bundle regs), the start index of each contiguous run in
// stream order, and each bank's width.
type regPlan struct {
	idx    map[reg]byte
	contig []byte // one start per bundleReq, in stream order
	widths [numBanks]int
	cursor int // emit-pass state: next contiguous run to hand out
}

func (p *regPlan) resolve(r reg) (byte, error) {
	// Every reg touched in discovery is in the plan, so a miss is an internal
	// invariant break rather than a user-facing error.
	v, ok := p.idx[r]
	if !ok {
		return 0, fmt.Errorf("internal error: register $r%d has no allocation", int(r))
	}
	return v, nil
}

func (p *regPlan) nextContig() byte {
	s := p.contig[p.cursor]
	p.cursor++
	return s
}

// linearScanPlan turns the discovered liveness into a packed allocation.
//
// Contiguous runs (allotment portions and outputs) are placed first, at the
// bottom of their bank, and reserved for the whole function. This is a
// deliberate simplification: allotment is rare and the runs are small, and a
// bound output run's liveness usually extends past the allotment anyway (the
// following sends read it), so reserving it wholesale is both simple and, in the
// common case, not wasteful. Every other reg is then packed above the reserved
// region by an expiring linear scan that reuses a slot as soon as its occupant's
// interval has ended.
func linearScanPlan(in *allocInput) (*regPlan, error) {
	plan := &regPlan{idx: map[reg]byte{}}

	// reserved region per bank, filled by the contiguous runs
	var reserved [numBanks]int
	for _, b := range in.bundles {
		start := reserved[b.bank]
		if start+b.n > maxReg {
			return nil, errBankOverflow
		}
		reserved[b.bank] += b.n
		plan.contig = append(plan.contig, byte(start))
		for i, r := range b.regs {
			plan.idx[r] = byte(start + i)
		}
	}

	for bank := bankId(0); bank < numBanks; bank++ {
		width, err := scanBank(in, bank, reserved[bank], plan)
		if err != nil {
			return nil, err
		}
		plan.widths[bank] = width
	}
	return plan, nil
}

type liveInterval struct {
	r           reg
	first, last int
}

// scanBank packs one bank's non-bundle regs above its reserved region and
// returns the bank width. It writes chosen indices into plan.idx.
func scanBank(in *allocInput, bank bankId, reserved int, plan *regPlan) (int, error) {
	var ivals []liveInterval
	for _, r := range in.order {
		if in.bankOf[r] != bank || in.bundleReg[r] {
			continue
		}
		ivals = append(ivals, liveInterval{r, in.first[r], in.last[r]})
	}
	// Sort by start, then end, then reg id — a total order, so the scan is
	// deterministic across runs.
	sort.Slice(ivals, func(i, j int) bool {
		a, b := ivals[i], ivals[j]
		if a.first != b.first {
			return a.first < b.first
		}
		if a.last != b.last {
			return a.last < b.last
		}
		return a.r < b.r
	})

	type slot struct {
		last int
		idx  byte
	}
	var active []slot
	var free []byte // freed physical indices, reused before minting new ones
	next := reserved
	width := reserved

	for _, iv := range ivals {
		// Expire every interval that ended strictly before this one starts.
		// Strict `<` is intentional: an interval ending exactly where another
		// begins shares an instruction with it, so they must not share a slot.
		kept := active[:0]
		for _, s := range active {
			if s.last < iv.first {
				free = append(free, s.idx)
			} else {
				kept = append(kept, s)
			}
		}
		active = kept

		var idx byte
		if n := len(free); n > 0 {
			idx = free[n-1]
			free = free[:n-1]
		} else {
			if next >= maxReg {
				return 0, errBankOverflow
			}
			idx = byte(next)
			next++
		}
		plan.idx[iv.r] = idx
		active = append(active, slot{iv.last, idx})
		if int(idx)+1 > width {
			width = int(idx) + 1
		}
	}
	return width, nil
}

var errBankOverflow = fmt.Errorf("register bank overflow: more than %d registers live at once in one bank", maxReg)
