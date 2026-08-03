package runtime

import (
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/formancehq/numscript/internal/utils"
)

// Asset-scaling arithmetic: converting between different scales of the same base
// asset (e.g. EUR, EUR/2, EUR/4) with no rounding error and no spare amount.
// Value-free — operates on asset strings, scales, and amounts — so both the
// interpreter and the VM can drive it.

// GetBaseAndScale splits an asset into its base and scale, e.g. "EUR/2" ->
// ("EUR", 2) and "EUR" -> ("EUR", 0).
func GetBaseAndScale(asset string) (string, int64) {
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

// AssetToScaledAsset maps an asset to its wildcard-scale form, e.g. "EUR/2" ->
// "EUR/*" and "EUR" -> "EUR/*".
func AssetToScaledAsset(asset string) string {
	parts := strings.Split(asset, "/")
	if len(parts) == 1 {
		return asset + "/*"
	}
	return parts[0] + "/*"
}

// BuildScaledAsset composes a base asset and a scale into an asset string, e.g.
// ("EUR", 2) -> "EUR/2" and ("EUR", 0) -> "EUR".
func BuildScaledAsset(baseAsset string, scale int64) string {
	if scale == 0 {
		return baseAsset
	}
	return fmt.Sprintf("%s/%d", baseAsset, scale)
}

// GetAssets collects, per scale, the (uncolored) amount an account holds of
// baseAsset. Scaling converts only uncolored balances and emits uncolored
// postings, so colored balances are excluded.
func GetAssets(accountBalances []AccountBalance, baseAsset string) map[int64]*big.Int {
	result := make(map[int64]*big.Int)
	for _, accBalance := range accountBalances {
		if accBalance.Color != "" {
			continue
		}
		accBalanceAsset, scale := GetBaseAndScale(accBalance.Asset)
		if accBalanceAsset == baseAsset {
			result[scale] = new(big.Int).Set(accBalance.Amount)
		}
	}
	return result
}

// ScalePair is an (amount at scale) entry: a conversion output of
// FindScalingSolution.
type ScalePair struct {
	Scale  int64
	Amount *big.Int
}

func getSortedAssets(scales map[int64]*big.Int) []ScalePair {
	var assets []ScalePair
	for k, v := range scales {
		assets = append(assets, ScalePair{
			Scale:  k,
			Amount: v,
		})
	}

	// Sort in DESC order (e.g. EUR/4, .., EUR/1, EUR)
	slices.SortFunc(assets, func(p ScalePair, other ScalePair) int {
		return int(other.Scale - p.Scale)
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

func applyScaling(amt *big.Int, scalingFactor *big.Rat) *big.Int {
	availableCurrencyScaled := new(big.Int)
	availableCurrencyScaled.Mul(amt, scalingFactor.Num())
	availableCurrencyScaled.Div(availableCurrencyScaled, scalingFactor.Denom())

	return availableCurrencyScaled
}

func applyScalingInv(amt *big.Int, scalingFactor *big.Rat) *big.Int {
	rem := new(big.Int)

	availableCurrencyScaled := new(big.Int)
	availableCurrencyScaled.Mul(amt, scalingFactor.Denom())
	availableCurrencyScaled.QuoRem(availableCurrencyScaled, scalingFactor.Num(), rem)

	if rem.Sign() != 0 {
		availableCurrencyScaled.Add(availableCurrencyScaled, big.NewInt(1))
	}

	return availableCurrencyScaled
}

// FindScalingSolution finds a set of conversions from the available "scales" to
// [ASSET/$neededAmtScale $neededAmt], so that there's no rounding error and no
// spare amount. neededAmt may be nil (meaning "convert everything available").
func FindScalingSolution(
	neededAmt *big.Int, // <- can be nil
	neededAmtScale int64,
	scales map[int64]*big.Int,
) ([]ScalePair, *big.Int) {
	if ownedAmt, ok := scales[neededAmtScale]; ok && neededAmt != nil {
		// Note we don't mutate the input value
		neededAmt = new(big.Int).Sub(neededAmt, ownedAmt)
	}

	var out []ScalePair
	totalSent := big.NewInt(0)

	for _, p := range getSortedAssets(scales) {
		if neededAmtScale == p.Scale {
			// We don't convert assets we already have
			continue
		}

		if neededAmt != nil && totalSent.Cmp(neededAmt) != -1 {
			break
		}

		scalingFactor := getScalingFactor(neededAmtScale, p.Scale)

		// scale the original amount to the current currency
		// availableCurrencyScaled := floor(p.amount * scalingFactor)
		availableCurrencyScaled := applyScaling(p.Amount, scalingFactor)

		var taken *big.Int // := min(availableCurrencyScaled, (neededAmt-totalSent) ?? ∞)
		if neededAmt == nil {
			taken = new(big.Int).Set(availableCurrencyScaled)
		} else {
			leftAmt := new(big.Int).Sub(neededAmt, totalSent)
			taken = utils.MinBigInt(availableCurrencyScaled, leftAmt)
		}

		// intPart := floor(p.amount * 1/scalingFactor) == (p.amount * scalingFactor.Denom)/scalingFactor.Num)
		intPart := applyScalingInv(taken, scalingFactor)

		if intPart.Sign() == 0 {
			continue
		}

		actuallyTaken := applyScaling(intPart, scalingFactor)

		totalSent.Add(totalSent, actuallyTaken)

		out = append(out, ScalePair{
			Scale:  p.Scale,
			Amount: intPart,
		})
	}

	return out, totalSent
}
