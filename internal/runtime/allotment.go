package runtime

import "math/big"

// MakeAllotment splits amount across portions, writing one integer amount per
// portion into out (out[i] for portions[i]) such that the written parts sum
// exactly to amount.
//
// out must have the same length as portions; its elements are overwritten. They
// are big.Int values, not pointers: MakeAllotment mutates them in place through
// the slice (out[i] is addressable, so out[i].Div(...) writes the element). This
// lets the caller allocate the whole result as one contiguous []big.Int rather
// than len(portions) separate *big.Int.
//
// Portions are fractions of the whole (big.Rat) and are expected to sum to 1;
// any "remaining" portion must already be resolved by the caller (i.e. computed
// as 1 minus the others). MakeAllotment does not validate the sum.
//
// Algorithm (matching the interpreter's allotment logic):
//  1. each part is floor(portion * amount);
//  2. the leftover from flooring — amount minus the sum of the floored parts —
//     is handed out one unit at a time to the earliest portions, until the parts
//     sum exactly to amount.
//
// Because flooring loses strictly less than 1 unit per portion, the leftover is
// strictly less than len(portions), so a single front-to-back pass distributes
// it fully (given portions that sum to 1).
//
// Inputs amount and portions are not mutated.
func MakeAllotment(out []big.Int, amount *big.Int, portions []big.Rat) {
	totalAllocated := new(big.Int)
	amountRat := new(big.Rat).SetInt(amount)

	for i := range portions {
		product := new(big.Rat).Mul(&portions[i], amountRat)
		// floor into out[i] in place; Denom() is always positive (matches Div)
		out[i].Div(product.Num(), product.Denom())
		totalAllocated.Add(totalAllocated, &out[i])
	}

	one := big.NewInt(1)
	for i := range out {
		if totalAllocated.Cmp(amount) >= 0 {
			break
		}
		out[i].Add(&out[i], one)
		totalAllocated.Add(totalAllocated, one)
	}
}
