package runtime

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// Scalar value parsing / validation: the single source of truth for the textual
// form of numscript scalar values, shared by both runtimes (the tree-walking
// interpreter and the VM) and the compiler's vars encoder. Returns primitives
// (string / *big.Int / *big.Rat) and plain errors, so each caller adapts them
// into its own value/error types.

const accountSegmentRegex = "[a-zA-Z0-9_-]+"

// https://github.com/formancehq/ledger/blob/main/pkg/accounts/accounts.go
var accountNameRegex = regexp.MustCompile("^" + accountSegmentRegex + "(:" + accountSegmentRegex + ")*$")

// https://github.com/formancehq/ledger/blob/main/pkg/assets/asset.go
var assetNameRegex = regexp.MustCompile(`^[A-Z][A-Z0-9]{0,16}(_[A-Z]{1,16})?(\/\d{1,6})?$`)

var percentRegex = regexp.MustCompile(`^([0-9]+)(?:[.]([0-9]+))?[%]$`)
var fractionRegex = regexp.MustCompile(`^([0-9]+)\s?[/]\s?([0-9]+)$`)

func ValidateAccount(addr string) bool { return accountNameRegex.MatchString(addr) }
func ValidateAsset(v string) bool      { return assetNameRegex.MatchString(v) }

// ParseNumber parses a base-10 integer (arbitrary precision).
func ParseNumber(s string) (*big.Int, bool) {
	return new(big.Int).SetString(s, 10)
}

// ParsePortion parses a portion given as a percentage ("12%") or a fraction
// ("1/3"), returning it as a reduced ratio in [0, 1]. The error message is the
// reason (callers may wrap it into their own error type).
func ParsePortion(input string) (*big.Rat, error) {
	var res *big.Rat
	var ok bool

	percentMatch := percentRegex.FindStringSubmatch(input)
	if len(percentMatch) != 0 {
		integral := percentMatch[1]
		fractional := percentMatch[2]
		res, ok = new(big.Rat).SetString(integral + "." + fractional)
		if !ok {
			return nil, errors.New("invalid percent format")
		}
		res.Mul(res, big.NewRat(1, 100))
	} else {
		fractionMatch := fractionRegex.FindStringSubmatch(input)
		if len(fractionMatch) != 0 {
			numerator := fractionMatch[1]
			denominator := fractionMatch[2]
			res, ok = new(big.Rat).SetString(numerator + "/" + denominator)
			if !ok {
				return nil, errors.New("invalid fractional format")
			}
		}
	}
	if res == nil {
		return nil, errors.New("invalid format")
	}

	if res.Cmp(big.NewRat(0, 1)) == -1 || res.Cmp(big.NewRat(1, 1)) == 1 {
		return nil, errors.New("portion must be between 0% and 100% inclusive")
	}

	return res, nil
}

// ParseMonetary parses "ASSET AMOUNT" (e.g. "USD/2 100") into its asset and
// amount.
func ParseMonetary(source string) (asset string, amount *big.Int, err error) {
	parts := strings.Split(source, " ")
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid monetary: %q", source)
	}
	if !ValidateAsset(parts[0]) {
		return "", nil, fmt.Errorf("invalid asset: %q", parts[0])
	}
	n, ok := ParseNumber(parts[1])
	if !ok {
		return "", nil, fmt.Errorf("invalid monetary amount: %q", parts[1])
	}
	return parts[0], n, nil
}
