package interpreter

import (
	"math/big"
	"slices"

	"github.com/formancehq/numscript/internal/utils"
)

// e.g.
//
// need=[EUR/2 100], got={EUR/2: 100, EUR: 1}
// => {EUR/2: 100, EUR: 1}
//
// need=[EUR 1], got={EUR/2: 100, EUR: 0}
// => {EUR/2: 100, EUR: 0}
//
// need=[EUR/2 199], got={EUR/2: 100, EUR: 2}
// => {EUR/2: 100, EUR: 1}
func findSolution(
	neededAmt *big.Int,
	neededAmtScale int64,
	scales map[int64]*big.Int,
) map[int64]*big.Int {
	// we clone neededAmt so that we can update it
	neededAmt = new(big.Int).Set(neededAmt)

	type scalePair struct {
		scale  int64
		amount *big.Int
	}

	var assets []scalePair
	for k, v := range scales {
		assets = append(assets, scalePair{
			scale:  k,
			amount: v,
		})
	}

	// Sort in ASC order (e.g. EUR, EUR/2, ..)
	slices.SortFunc(assets, func(p scalePair, other scalePair) int {
		return int(p.scale - other.scale)
	})

	out := map[int64]*big.Int{}

	left := new(big.Int).Set(neededAmt)

	for _, p := range assets {
		scaleDiff := neededAmtScale - p.scale

		exp := big.NewInt(scaleDiff)
		exp.Abs(exp)
		exp.Exp(big.NewInt(10), exp, nil)

		// scalingFactor := 10 ^ (neededAmtScale - p.scale)
		// note that 10^0 == 1 and 10^(-n) == 1/(10^n)
		scalingFactor := new(big.Rat).SetInt(exp)
		if scaleDiff < 0 {
			scalingFactor.Inv(scalingFactor)
		}

		allowed := new(big.Int).Mul(p.amount, scalingFactor.Num())
		allowed.Div(allowed, scalingFactor.Denom())

		taken := utils.MinBigInt(allowed, neededAmt)

		intPart := new(big.Int).Mul(taken, scalingFactor.Denom())
		intPart.Div(intPart, scalingFactor.Num())

		if intPart.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		neededAmt.Sub(neededAmt, taken)
		out[p.scale] = intPart

		// if neededAmt <= 0
		if neededAmt.Cmp(big.NewInt(0)) != 1 {
			return out
		}
	}

	if left.Cmp(big.NewInt(0)) != 0 {
		return nil
	}

	return out
}
