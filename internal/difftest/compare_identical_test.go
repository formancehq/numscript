package difftest

import "testing"

func TestCompareIdenticalToleratesOnlyLiftedCapacityBounds(t *testing.T) {
	overflow := SideResult{CompileErr: "register bank overflow: ...", RegisterOverflow: true}
	tooLarge := SideResult{CompileErr: "program too large: ...", ProgramTooLarge: true}
	compiled := SideResult{}

	for _, plain := range []SideResult{overflow, tooLarge} {
		if v := CompareIdentical(compiled, plain, "optimized vm", "vm"); v != tolerated("capacity bound lifted by optimization") {
			t.Errorf("lifted bound %q: got %+v", plain.CompileErr, v)
		}
	}

	// imposing a bound, or trading one for another, is a mismatch: the compiler
	// falls back to the plain stream when the optimized one hits a bound
	if v := CompareIdentical(overflow, compiled, "optimized vm", "vm"); !v.Mismatch {
		t.Errorf("imposed bound: got %+v", v)
	}
	if v := CompareIdentical(overflow, tooLarge, "optimized vm", "vm"); !v.Mismatch {
		t.Errorf("traded bound: got %+v", v)
	}
}
