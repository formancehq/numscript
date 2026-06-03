package mcp_impl

// Failing regression test demonstrating silent truncation in the MCP
// `evaluate` tool's balance parser. JSON numeric values arrive in `any`
// as `float64`, and the current implementation passes them straight
// through `big.NewFloat(amount).Int(new(big.Int))` — which truncates
// fractional parts and silently rounds values outside the exact-integer
// range of float64 (±(2^53 - 1)).
//
// For a financial DSL, silent corruption of balance amounts is not
// acceptable. The MCP handler should reject any value that cannot be
// represented as an exact integer in float64 precision.
//
// CI is intentionally red on this branch. A stacked fix PR validates
// the amount before conversion and turns the test green.

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseBalancesJsonRejectsNonIntegerAmounts(t *testing.T) {
	// 100.9 is unambiguously fractional. The handler must refuse it
	// rather than silently truncate to 100.
	balances := map[string]any{
		"alice": map[string]any{
			"USD/2": float64(100.9),
		},
	}

	_, err := parseBalancesJson(balances)
	require.NotNil(t, err, "fractional amount 100.9 must be rejected, not silently truncated")
}

// Integer amounts that exceed float64's exact-integer range (2^53 - 1 =
// 9_007_199_254_740_991) cannot be recovered losslessly from the JSON
// payload. The handler should refuse them so callers know to switch to a
// string-encoded amount.
func TestParseBalancesJsonRejectsUnsafelyLargeAmounts(t *testing.T) {
	// 1e18 is representable as float64 but the surrounding integer is
	// not unique — many neighbouring int64 values share the same float
	// representation. A future-proof handler refuses anything past 2^53.
	balances := map[string]any{
		"alice": map[string]any{
			"USD/2": float64(1e18),
		},
	}

	_, err := parseBalancesJson(balances)
	require.NotNil(t, err, "amount 1e18 exceeds float64 exact-integer range and must be rejected")
}

// Sanity: a small integer-valued float still parses correctly.
func TestParseBalancesJsonAcceptsExactIntegerAmounts(t *testing.T) {
	balances := map[string]any{
		"alice": map[string]any{
			"USD/2": float64(100),
		},
	}

	got, err := parseBalancesJson(balances)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(100), got["alice"]["USD/2"])
}
