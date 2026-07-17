package runtime_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
)

func portion(num, den int64) runtime.Portion {
	return runtime.Portion{Num: *big.NewInt(num), Den: *big.NewInt(den)}
}

// allot fills a fresh buffer via MakeAllotment and returns it, for ergonomics.
func allot(amount int64, portions []runtime.Portion) []big.Int {
	out := make([]big.Int, len(portions))
	runtime.New(nil).MakeAllotment(out, big.NewInt(amount), portions)
	return out
}

func wantParts(t *testing.T, got []big.Int, want []int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range got {
		if got[i].Cmp(big.NewInt(want[i])) != 0 {
			t.Errorf("part[%d] = %s, want %d", i, got[i].String(), want[i])
		}
	}
}

func TestMakeAllotment_EvenSplit(t *testing.T) {
	wantParts(t, allot(100, []runtime.Portion{portion(1, 2), portion(1, 2)}), []int64{50, 50})
}

func TestMakeAllotment_UnevenSplit(t *testing.T) {
	wantParts(t, allot(100, []runtime.Portion{portion(1, 4), portion(3, 4)}), []int64{25, 75})
}

func TestMakeAllotment_RemainderGoesToEarliest_Thirds(t *testing.T) {
	// 1/3 of 100 floors to 33 each (sum 99); the leftover 1 goes to the first.
	wantParts(t, allot(100, []runtime.Portion{portion(1, 3), portion(1, 3), portion(1, 3)}), []int64{34, 33, 33})
}

func TestMakeAllotment_RemainderTwoUnits(t *testing.T) {
	// 1/6,1/6,4/6 of 100 -> 16,16,66 (sum 98); leftover 2 -> first two get +1.
	wantParts(t, allot(100, []runtime.Portion{portion(1, 6), portion(1, 6), portion(4, 6)}), []int64{17, 17, 66})
}

func TestMakeAllotment_HalvesOfOddAmount(t *testing.T) {
	// 7 split in half -> 3,3 (sum 6); leftover 1 -> first.
	wantParts(t, allot(7, []runtime.Portion{portion(1, 2), portion(1, 2)}), []int64{4, 3})
}

func TestMakeAllotment_SinglePortionWhole(t *testing.T) {
	wantParts(t, allot(100, []runtime.Portion{portion(1, 1)}), []int64{100})
}

func TestMakeAllotment_ZeroAmount(t *testing.T) {
	wantParts(t, allot(0, []runtime.Portion{portion(1, 3), portion(2, 3)}), []int64{0, 0})
}

func TestMakeAllotment_EmptyPortions(t *testing.T) {
	out := []big.Int{}
	runtime.New(nil).MakeAllotment(out, big.NewInt(100), []runtime.Portion{})
	if len(out) != 0 {
		t.Errorf("len = %d, want 0", len(out))
	}
}

func TestMakeAllotment_PercentageLikePortions(t *testing.T) {
	// 19% / 81% of 10_000 -> 1900 / 8100 exactly.
	wantParts(t, allot(10_000, []runtime.Portion{portion(19, 100), portion(81, 100)}), []int64{1900, 8100})
}

func TestMakeAllotment_UnreducedPortionsSameAsReduced(t *testing.T) {
	// 2/4 == 1/2: the impl keeps portions unreduced, but floor(amount*num/den) is
	// invariant under reduction, so the split is identical.
	wantParts(t, allot(100, []runtime.Portion{portion(2, 4), portion(2, 4)}), []int64{50, 50})
}

func TestMakeAllotment_PartsAlwaysSumToAmount(t *testing.T) {
	// A spread that floors awkwardly must still sum exactly to the amount.
	amount := big.NewInt(1001)
	out := make([]big.Int, 3)
	runtime.New(nil).MakeAllotment(out, amount, []runtime.Portion{portion(1, 7), portion(2, 7), portion(4, 7)})
	sum := new(big.Int)
	for i := range out {
		sum.Add(sum, &out[i])
	}
	if sum.Cmp(amount) != 0 {
		t.Errorf("parts sum to %s, want %s (parts=%v)", sum, amount, out)
	}
}

func TestMakeAllotment_BeyondInt64(t *testing.T) {
	amount, _ := new(big.Int).SetString("1000000000000000000000000001", 10) // ~1e27 + 1, odd
	out := make([]big.Int, 2)
	runtime.New(nil).MakeAllotment(out, amount, []runtime.Portion{portion(1, 2), portion(1, 2)})
	// floor halves are equal; the odd unit goes to the first
	half := new(big.Int).Div(amount, big.NewInt(2)) // floor(amount/2)
	first := new(big.Int).Add(half, big.NewInt(1))
	if out[0].Cmp(first) != 0 || out[1].Cmp(half) != 0 {
		t.Errorf("got [%s %s], want [%s %s]", out[0].String(), out[1].String(), first, half)
	}
	sum := new(big.Int).Add(&out[0], &out[1])
	if sum.Cmp(amount) != 0 {
		t.Errorf("sum = %s, want %s", sum, amount)
	}
}

func TestMakeAllotment_ModifiesCallerSliceAndOverwritesStale(t *testing.T) {
	// Pre-fill the buffer with garbage to prove MakeAllotment overwrites it
	// (Div fully replaces each element) and writes through to the caller's slice.
	out := make([]big.Int, 2)
	out[0].SetInt64(999)
	out[1].SetInt64(-7)
	runtime.New(nil).MakeAllotment(out, big.NewInt(100), []runtime.Portion{portion(1, 4), portion(3, 4)})
	wantParts(t, out, []int64{25, 75})
}

func TestMakeAllotment_DoesNotMutateInputs(t *testing.T) {
	portions := []runtime.Portion{portion(1, 3), portion(2, 3)}
	amount := big.NewInt(100)
	out := make([]big.Int, 2)
	runtime.New(nil).MakeAllotment(out, amount, portions)
	if portions[0].Num.Int64() != 1 || portions[0].Den.Int64() != 3 ||
		portions[1].Num.Int64() != 2 || portions[1].Den.Int64() != 3 {
		t.Errorf("portions mutated: %+v", portions)
	}
	if amount.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("amount mutated: %s", amount)
	}
}
