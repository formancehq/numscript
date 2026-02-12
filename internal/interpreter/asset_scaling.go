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

func getSortedAssets(scales map[int64]*big.Int) []scalePair {
	var assets []scalePair
	for k, v := range scales {
		assets = append(assets, scalePair{
			scale:  k,
			amount: v,
		})
	}

	// Sort in DESC order (e.g. EUR/4, .., EUR/1, EUR)
	slices.SortFunc(assets, func(p scalePair, other scalePair) int {
		return int(other.scale - p.scale)
	})

	return assets
}

func getScalingFactor(neededAmtScale int64, currentScale int64) *big.Rat {
	scaleDiff := neededAmtScale - currentScale

	exp := big.NewInt(scaleDiff)
	exp.Abs(exp)
	exp.Exp(big.NewInt(10), exp, nil)

	// scalingFactor := 10 ^ (neededAmtScale - p.scale)
	// note that 10^0 == 1 and 10^(-n) == 1/(10^n)
	scalingFactor := new(big.Rat).SetInt(exp)
	if scaleDiff < 0 {
		scalingFactor.Inv(scalingFactor)
	}

	return scalingFactor
}

func applyScaling(amt *big.Int, scalingFactor *big.Rat) (*big.Int, *big.Int) {
	rem := new(big.Int)

	availableCurrencyScaled := new(big.Int)
	availableCurrencyScaled.Mul(amt, scalingFactor.Num())
	availableCurrencyScaled.QuoRem(availableCurrencyScaled, scalingFactor.Denom(), rem)

	return availableCurrencyScaled, rem
}

// Find a set of conversions from the available "scales", to
// [ASSET/$neededAmtScale $neededAmt], so that there's no rounding error
// and no spare amount
func findScalingSolution(
	neededAmt *big.Int, // <- can be nil
	neededAmtScale int64,
	scales map[int64]*big.Int,
) ([]scalePair, *big.Int) {
	if ownedAmt, ok := scales[neededAmtScale]; ok && neededAmt != nil {
		// Note we don't mutate the input value
		neededAmt = new(big.Int).Sub(neededAmt, ownedAmt)
	}

	var out []scalePair
	totalSent := big.NewInt(0)

	for _, p := range getSortedAssets(scales) {
		if neededAmtScale == p.scale {
			// We don't convert assets we already have
			continue
		}

		if neededAmt != nil && totalSent.Cmp(neededAmt) != -1 {
			break
		}

		scalingFactor := getScalingFactor(neededAmtScale, p.scale)

		// scale the original amount to the current currency
		// availableCurrencyScaled := floor(p.amount * scalingFactor)
		availableCurrencyScaled, _ := applyScaling(p.amount, scalingFactor)

		var taken *big.Int // := min(availableCurrencyScaled, (neededAmt-totalSent) ?? ∞)
		if neededAmt == nil {
			taken = new(big.Int).Set(availableCurrencyScaled)
		} else {
			leftAmt := new(big.Int).Sub(neededAmt, totalSent)
			taken = utils.MinBigInt(availableCurrencyScaled, leftAmt)
		}

		// intPart := floor(p.amount * 1/scalingFactor) == (p.amount * scalingFactor.Denom)/scalingFactor.Num)
		intPart, rem := applyScaling(taken, new(big.Rat).Inv(scalingFactor))
		if rem.Sign() == 1 {
			intPart.Add(intPart, big.NewInt(1))
		}

		if intPart.Sign() == 0 {
			continue
		}

		actuallyTaken, remTaken := applyScaling(intPart, scalingFactor)
		if remTaken.Sign() != 0 {
			panic("UNEXPECTED REM")
			actuallyTaken.Add(actuallyTaken, big.NewInt(1))
		}
		totalSent.Add(totalSent, actuallyTaken)

		out = append(out, scalePair{
			scale:  p.scale,
			amount: intPart,
		})
	}

	return out, totalSent
}
