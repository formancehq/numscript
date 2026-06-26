package runtime

import "math/big"

// MakeAllotment splits amount across portions, returning one integer amount per
// portion such that the parts sum exactly to amount.
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
// Inputs are not mutated; every returned *big.Int is freshly allocated.
func MakeAllotment(amount *big.Int, portions []*big.Rat) []*big.Int {
	parts := make([]*big.Int, len(portions))
	totalAllocated := new(big.Int)
	amountRat := new(big.Rat).SetInt(amount)

	for i, portion := range portions {
		product := new(big.Rat).Mul(portion, amountRat)
		// floor: Denom() is always positive, so this matches big.Int.Div
		floored := new(big.Int).Div(product.Num(), product.Denom())
		parts[i] = floored
		totalAllocated.Add(totalAllocated, floored)
	}

	one := big.NewInt(1)
	for i := range parts {
		if totalAllocated.Cmp(amount) >= 0 {
			break
		}
		parts[i].Add(parts[i], one)
		totalAllocated.Add(totalAllocated, one)
	}

	return parts
}
