package runtime_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
)

func rat(num, denom int64) *big.Rat { return big.NewRat(num, denom) }

func wantParts(t *testing.T, got []*big.Int, want []int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	var sum int64
	for i, g := range got {
		if g.Cmp(big.NewInt(want[i])) != 0 {
			t.Errorf("part[%d] = %s, want %d", i, g, want[i])
		}
		sum += want[i]
	}
	_ = sum
}

func TestMakeAllotment_EvenSplit(t *testing.T) {
	got := runtime.MakeAllotment(big.NewInt(100), []*big.Rat{rat(1, 2), rat(1, 2)})
	wantParts(t, got, []int64{50, 50})
}

func TestMakeAllotment_UnevenSplit(t *testing.T) {
	got := runtime.MakeAllotment(big.NewInt(100), []*big.Rat{rat(1, 4), rat(3, 4)})
	wantParts(t, got, []int64{25, 75})
}

func TestMakeAllotment_RemainderGoesToEarliest_Thirds(t *testing.T) {
	// 1/3 of 100 floors to 33 each (sum 99); the leftover 1 goes to the first.
	got := runtime.MakeAllotment(big.NewInt(100), []*big.Rat{rat(1, 3), rat(1, 3), rat(1, 3)})
	wantParts(t, got, []int64{34, 33, 33})
}

func TestMakeAllotment_RemainderTwoUnits(t *testing.T) {
	// 1/6,1/6,4/6 of 100 -> 16,16,66 (sum 98); leftover 2 -> first two get +1.
	got := runtime.MakeAllotment(big.NewInt(100), []*big.Rat{rat(1, 6), rat(1, 6), rat(4, 6)})
	wantParts(t, got, []int64{17, 17, 66})
}

func TestMakeAllotment_HalvesOfOddAmount(t *testing.T) {
	// 7 split in half -> 3,3 (sum 6); leftover 1 -> first.
	got := runtime.MakeAllotment(big.NewInt(7), []*big.Rat{rat(1, 2), rat(1, 2)})
	wantParts(t, got, []int64{4, 3})
}

func TestMakeAllotment_SinglePortionWhole(t *testing.T) {
	got := runtime.MakeAllotment(big.NewInt(100), []*big.Rat{rat(1, 1)})
	wantParts(t, got, []int64{100})
}

func TestMakeAllotment_ZeroAmount(t *testing.T) {
	got := runtime.MakeAllotment(big.NewInt(0), []*big.Rat{rat(1, 3), rat(2, 3)})
	wantParts(t, got, []int64{0, 0})
}

func TestMakeAllotment_EmptyPortions(t *testing.T) {
	got := runtime.MakeAllotment(big.NewInt(100), nil)
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestMakeAllotment_PercentageLikePortions(t *testing.T) {
	// 19% / 81% of 10_000 -> 1900 / 8100 exactly.
	got := runtime.MakeAllotment(big.NewInt(10_000), []*big.Rat{rat(19, 100), rat(81, 100)})
	wantParts(t, got, []int64{1900, 8100})
}

func TestMakeAllotment_PartsAlwaysSumToAmount(t *testing.T) {
	// A spread that floors awkwardly must still sum exactly to the amount.
	amount := big.NewInt(1001)
	portions := []*big.Rat{rat(1, 7), rat(2, 7), rat(4, 7)}
	got := runtime.MakeAllotment(amount, portions)
	sum := new(big.Int)
	for _, p := range got {
		sum.Add(sum, p)
	}
	if sum.Cmp(amount) != 0 {
		t.Errorf("parts sum to %s, want %s (parts=%v)", sum, amount, got)
	}
}

func TestMakeAllotment_BeyondInt64(t *testing.T) {
	amount, _ := new(big.Int).SetString("1000000000000000000000000001", 10) // ~1e27 + 1, odd
	got := runtime.MakeAllotment(amount, []*big.Rat{rat(1, 2), rat(1, 2)})
	// floor halves are equal; the odd unit goes to the first
	half := new(big.Int).Div(amount, big.NewInt(2)) // floor(amount/2)
	first := new(big.Int).Add(half, big.NewInt(1))
	if got[0].Cmp(first) != 0 || got[1].Cmp(half) != 0 {
		t.Errorf("got [%s %s], want [%s %s]", got[0], got[1], first, half)
	}
	sum := new(big.Int).Add(got[0], got[1])
	if sum.Cmp(amount) != 0 {
		t.Errorf("sum = %s, want %s", sum, amount)
	}
}

func TestMakeAllotment_DoesNotMutateInputs(t *testing.T) {
	portions := []*big.Rat{rat(1, 3), rat(2, 3)}
	p0, p1 := new(big.Rat).Set(portions[0]), new(big.Rat).Set(portions[1])
	amount := big.NewInt(100)
	_ = runtime.MakeAllotment(amount, portions)
	if portions[0].Cmp(p0) != 0 || portions[1].Cmp(p1) != 0 {
		t.Errorf("portions mutated: %v %v", portions[0], portions[1])
	}
	if amount.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("amount mutated: %s", amount)
	}
}
