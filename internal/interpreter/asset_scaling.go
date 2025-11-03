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
	neededAmtScale int,
	scales map[int]*big.Int,
) map[int]*big.Int {
	type scalePair struct {
		scale  int
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
		return p.scale - other.scale
	})

	out := map[int]*big.Int{}

	left := new(big.Int).Set(neededAmt)

	for _, p := range assets {
		// "left <= 0"
		if left.Cmp(big.NewInt(0)) != 1 {
			break
		}

		scaleDiff := p.scale - neededAmtScale
		if scaleDiff < 0 {
			scaleDiff = -scaleDiff
		}

		scalingFactor := new(big.Int).Exp(
			big.NewInt(10),
			big.NewInt(int64(scaleDiff)),
			nil,
		)

		if p.scale > neededAmtScale {
			allowed := new(big.Int).Div(p.amount, scalingFactor)
			taken := utils.MinBigInt(left, allowed)
			left.Sub(left, taken)
			out[p.scale] = new(big.Int).Mul(taken, scalingFactor)
		} else if p.scale < neededAmtScale {
			allowed := new(big.Int).Mul(p.amount, scalingFactor)
			taken := utils.MinBigInt(left, allowed)
			left.Sub(left, taken)
			out[p.scale] = new(big.Int).Div(taken, scalingFactor)
		} else {
			allowed := p.amount
			taken := utils.MinBigInt(left, allowed)
			left.Sub(left, taken)
			out[p.scale] = new(big.Int).Set(taken)
		}
	}

	if left.Cmp(big.NewInt(0)) != 0 {
		return nil
	}

	return out
}
