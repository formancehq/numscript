package interpreter

import (
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/formancehq/numscript/internal/utils"
)

func assetToScaledAsset(asset string) string {
	parts := strings.Split(asset, "/")
	if len(parts) == 1 {
		return asset + "/*"
	}
	return parts[0] + "/*"
}

func buildScaledAsset(baseAsset string, scale int64) string {
	if scale == 0 {
		return baseAsset
	}
	return fmt.Sprintf("%s/%d", baseAsset, scale)
}

func getAssetScale(asset string) (string, int64) {
	parts := strings.Split(asset, "/")
	if len(parts) == 2 {
		scale, err := strconv.ParseInt(parts[1], 10, 64)
		if err == nil {
			return parts[0], scale
		}
		// fallback if parsing fails
		return parts[0], 0
	}
	return asset, 0
}

func getAssets(balance AccountBalance, baseAsset string) map[int64]*big.Int {
	result := make(map[int64]*big.Int)
	for asset, amount := range balance {
		if strings.HasPrefix(asset, baseAsset) {
			_, scale := getAssetScale(asset)
			result[scale] = amount
		}
	}
	return result
}

type scalePair struct {
	scale  int64
	amount *big.Int
}

// Find a set of conversions from the available "scales", to
// [ASSET/$neededAmtScale $neededAmt], so that there's no rounding error
// and no spare amount
//
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
//
// need=[EUR/2 1], got={EUR: 99}
// => no solution! (if we changed 1 EUR with 100 EUR/2 we'd have 99 spare cents)
func findScalingSolution(
	neededAmt *big.Int, // <- can be nil
	neededAmtScale int64,
	scales map[int64]*big.Int,
) ([]scalePair, *big.Int) {
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

	var out []scalePair

	totalSent := big.NewInt(0)

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

		var leftAmt *big.Int
		var taken *big.Int
		if neededAmt == nil {
			taken = new(big.Int).Set(allowed)
		} else {
			leftAmt = new(big.Int).Sub(neededAmt, totalSent)
			taken = utils.MinBigInt(allowed, leftAmt)
		}

		totalSent.Add(totalSent, taken)

		intPart := new(big.Int).Mul(taken, scalingFactor.Denom())
		intPart.Div(intPart, scalingFactor.Num())

		if intPart.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		out = append(out, scalePair{
			scale:  p.scale,
			amount: intPart,
		})

		if leftAmt != nil && leftAmt.Cmp(big.NewInt(0)) != 1 {
			break
		}
	}

	return out, totalSent
}
