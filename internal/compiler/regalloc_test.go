package compiler

import "testing"

// touch is (reg, bank) referenced at a given instruction index. buildInput
// replays a sequence of touches into an allocInput exactly as the discovery pass
// would, so the allocator can be tested as a pure function — no assembler, no
// vm.Program — which is the point: the packing logic is independently unit
// testable even though linear scan is the only allocator wired into assemble.
type touch struct {
	r    reg
	bank bankId
	idx  int
}

func buildInput(touches []touch, bundles ...bundleReq) *allocInput {
	in := newAllocInput()
	for _, t := range touches {
		in.touch(t.bank, t.r, t.idx)
	}
	for _, b := range bundles {
		if b.regs != nil {
			// mark bound-bundle regs (their slots come from the reserved run)
			for _, r := range b.regs {
				in.bundleReg[r] = true
			}
		}
		in.bundles = append(in.bundles, b)
	}
	return in
}

func TestLinearScan_OverlappingRegsGetDistinctSlots(t *testing.T) {
	// three regs all live at instruction 0 -> all interfere -> 3 slots
	in := buildInput([]touch{
		{1, bankInt, 0}, {2, bankInt, 0}, {3, bankInt, 0},
	})
	plan, err := linearScanPlan(in)
	if err != nil {
		t.Fatal(err)
	}
	if plan.widths[bankInt] != 3 {
		t.Errorf("width = %d, want 3", plan.widths[bankInt])
	}
	seen := map[byte]bool{}
	for _, r := range []reg{1, 2, 3} {
		seen[plan.idx[r]] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 distinct slots, got %v", plan.idx)
	}
}

func TestLinearScan_DeadRegSlotIsReused(t *testing.T) {
	// reg1: [0,1], reg2: [1,2], reg3: [2,2]. reg1 dies at 1, so reg3 (born at 2)
	// reuses its slot -> peak pressure 2, width 2.
	in := buildInput([]touch{
		{1, bankInt, 0},
		{1, bankInt, 1}, {2, bankInt, 1},
		{2, bankInt, 2}, {3, bankInt, 2},
	})
	plan, err := linearScanPlan(in)
	if err != nil {
		t.Fatal(err)
	}
	if plan.widths[bankInt] != 2 {
		t.Fatalf("width = %d, want 2", plan.widths[bankInt])
	}
	if plan.idx[1] != plan.idx[3] {
		t.Errorf("reg3 should reuse reg1's slot: reg1=%d reg3=%d", plan.idx[1], plan.idx[3])
	}
	if plan.idx[2] == plan.idx[1] {
		t.Errorf("reg2 overlaps reg1 and must not share its slot")
	}
}

func TestLinearScan_WidthIsPeakPressure(t *testing.T) {
	// A long-lived reg spanning the whole function, plus a rolling pair of
	// short-lived temps. Peak simultaneous liveness is 3 (base + two temps at
	// one point), regardless of how many temps are minted overall.
	touches := []touch{{100, bankInt, 0}} // base reg, live throughout
	// temps a,b co-live at idx 1; c,d co-live at idx 3; e,f at idx 5
	pairs := [][2]reg{{1, 2}, {3, 4}, {5, 6}}
	at := 1
	for _, p := range pairs {
		touches = append(touches,
			touch{p[0], bankInt, at}, touch{p[1], bankInt, at},
			touch{100, bankInt, at}, // keep base alive across the whole span
		)
		at += 2
	}
	in := buildInput(touches)
	plan, err := linearScanPlan(in)
	if err != nil {
		t.Fatal(err)
	}
	if plan.widths[bankInt] != 3 {
		t.Errorf("width = %d, want 3 (base + peak temp pair)", plan.widths[bankInt])
	}
}

func TestLinearScan_BanksAreIndependent(t *testing.T) {
	in := buildInput([]touch{
		{1, bankInt, 0}, {2, bankInt, 0},
		{3, bankStr, 0},
		{4, bankPortion, 0}, {5, bankPortion, 0}, {6, bankPortion, 0},
	})
	plan, err := linearScanPlan(in)
	if err != nil {
		t.Fatal(err)
	}
	if plan.widths[bankInt] != 2 || plan.widths[bankStr] != 1 ||
		plan.widths[bankPortion] != 3 || plan.widths[bankMonetary] != 0 {
		t.Errorf("widths = %v, want [2 1 3 0]", plan.widths)
	}
}

func TestLinearScan_ContiguousBundleReservedAtBankBottom(t *testing.T) {
	// A bound run of 3 int regs (allotment outputs) plus a couple of ordinary
	// int temps. The run occupies indices 0..2 as a block; ordinary temps pack
	// above it.
	bundle := bundleReq{bank: bankInt, regs: []reg{10, 11, 12}, n: 3}
	in := buildInput([]touch{
		{10, bankInt, 0}, {11, bankInt, 0}, {12, bankInt, 0}, // the run
		{20, bankInt, 1}, // ordinary temp
	}, bundle)
	plan, err := linearScanPlan(in)
	if err != nil {
		t.Fatal(err)
	}
	// bundle is contiguous starting at 0
	if plan.idx[10] != 0 || plan.idx[11] != 1 || plan.idx[12] != 2 {
		t.Errorf("bundle not contiguous at 0: %d %d %d", plan.idx[10], plan.idx[11], plan.idx[12])
	}
	if len(plan.contig) != 1 || plan.contig[0] != 0 {
		t.Errorf("contig starts = %v, want [0]", plan.contig)
	}
	// ordinary temp packs above the reserved region
	if plan.idx[20] < 3 {
		t.Errorf("ordinary temp = %d, want >= 3 (above the reserved run)", plan.idx[20])
	}
}
