package runtime

import "math/big"

// bigOne is a shared read-only 1. Safe to share: big.Int.Add reads it, never
// mutates it.
var bigOne = big.NewInt(1)

// Portion is a fraction Num/Den with Den > 0, NOT necessarily in lowest terms.
// It replaces big.Rat for the VM's portion registers so the portion ops
// (mk_portion, sub_portion, copy) and allotment math run integer-only and in
// place, with no per-op allocation. Reduction is never needed for correctness:
// floor(amount * Num / Den) — the only thing allotment does with a portion — is
// invariant under reducing the fraction. (Kept-unreduced denominators can grow
// across sub_portion chains, but stay small for realistic allotments.)
type Portion struct {
	Num big.Int
	Den big.Int
}

// MakeAllotment splits amount across portions, writing one integer amount per
// portion into out (out[i] for portions[i]) such that the written parts sum
// exactly to amount. out must have the same length as portions; its elements are
// overwritten in place.
//
// Portions are fractions of the whole and are expected to sum to 1; any
// "remaining" portion must already be resolved by the caller. MakeAllotment does
// not validate the sum.
//
// Algorithm (matching the interpreter's allotment logic):
//  1. each part is floor(portion * amount) = floor(amount * Num / Den);
//  2. the leftover from flooring — amount minus the sum of the floored parts —
//     is handed out one unit at a time to the earliest portions.
//
// Because flooring loses strictly less than 1 unit per portion, the leftover is
// strictly less than len(portions), so a single front-to-back pass distributes
// it fully (given portions that sum to 1).
//
// It is a method on RunState so it can reuse scratch big.Ints across calls
// (allotTmp/allotTotal), making a whole allotment allocation-free — the previous
// free function allocated a big.Rat per portion. Inputs are not mutated.
func (s *RunState) MakeAllotment(out []big.Int, amount *big.Int, portions []Portion) {
	total := &s.allotTotal
	total.SetInt64(0)
	tmp := &s.allotTmp

	for i := range portions {
		tmp.Mul(amount, &portions[i].Num) // amount * Num
		out[i].Div(tmp, &portions[i].Den) // floor(/ Den); Den > 0 so Div floors
		total.Add(total, &out[i])
	}

	for i := range out {
		if total.Cmp(amount) >= 0 {
			break
		}
		out[i].Add(&out[i], bigOne)
		total.Add(total, bigOne)
	}
}
