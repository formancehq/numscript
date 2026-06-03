package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScalingAvoidSwappingAlreadyHaveAsset(t *testing.T) {
	// Need [USD/2 200]
	// Got: {USD/2 100, USD 2}
	// we only want [USD 1] to be swapped
	sol, got := findScalingSolution(
		big.NewInt(200),
		2,
		map[int64]*big.Int{
			2: big.NewInt(100),
			0: big.NewInt(2),
		})

	require.Equal(t, []scalePair{
		{0, big.NewInt(1)},
	}, sol)
	require.Equal(t, big.NewInt(100), got)
}

func TestScalingAvoidSpareAmt(t *testing.T) {
	// Need [USD/2 1]
	// Got: {USD 99}
	sol, got := findScalingSolution(
		big.NewInt(1),
		2,
		map[int64]*big.Int{
			0: big.NewInt(99),
		})

	require.Equal(t, []scalePair{
		{0, big.NewInt(1)},
	}, sol)
	require.Equal(t, big.NewInt(100), got)
}

func TestScalingAvoidSpareAmt2(t *testing.T) {
	// Need [USD/2 1]
	// Got: {USD 99}
	sol, got := findScalingSolution(
		big.NewInt(399),
		2,
		map[int64]*big.Int{
			0: big.NewInt(9999999),
		})

	require.Equal(t, []scalePair{
		{0, big.NewInt(4)},
	}, sol)
	require.Equal(t, big.NewInt(400), got)
}

func TestScalingDownAvoidSpareAmt(t *testing.T) {
	sol, got := findScalingSolution(
		big.NewInt(1),
		0,
		map[int64]*big.Int{
			2: big.NewInt(123),
		})

	require.Equal(t, []scalePair{
		{2, big.NewInt(100)},
	}, sol)
	require.Equal(t, big.NewInt(1), got)
}

func TestScalingZeroNeeded(t *testing.T) {
	sol, tot := findScalingSolution(
		big.NewInt(0),
		42,
		map[int64]*big.Int{
			2: big.NewInt(100),
			1: big.NewInt(1),
		})

	require.Equal(t, []scalePair(nil), sol)
	require.Equal(t, big.NewInt(0), tot)
}

func TestDoNotAllowSpare(t *testing.T) {
	sol, tot := findScalingSolution(
		// Need [EUR/2 1]
		big.NewInt(1),
		2,

		// Have: {EUR: 99}
		map[int64]*big.Int{
			0: big.NewInt(99),
		})

	require.Equal(t, []scalePair{
		{0, big.NewInt(1)},
	}, sol)
	require.Equal(t, big.NewInt(100), tot)
}

func TestRepro(t *testing.T) {
	sol, tot := findScalingSolution(
		// Need [EUR/2 400]
		big.NewInt(400),
		2,

		// Have: {EUR: 99, EUR/2: 1}
		map[int64]*big.Int{
			2: big.NewInt(1),
			0: big.NewInt(99),
		})

	require.Equal(t, []scalePair{
		{0, big.NewInt(4)},
	}, sol)
	require.Equal(t, big.NewInt(400), tot)
}

func TestScalingSameAsset(t *testing.T) {
	sol, tot := findScalingSolution(
		// Need [EUR/2 200]
		big.NewInt(200),
		2,

		// Have: {EUR/2: 201}
		map[int64]*big.Int{
			2: big.NewInt(201),
		})

	require.Equal(t, []scalePair(nil), sol)
	require.Equal(t, big.NewInt(0), tot)
}

func TestScalingSolutionLowerScale(t *testing.T) {
	sol, _ := findScalingSolution(
		// Need [COIN 1]
		big.NewInt(1),
		0,
		// Got: {COIN/2 900}
		map[int64]*big.Int{
			2: big.NewInt(900),
		})

	require.Equal(t, []scalePair{
		{2, big.NewInt(100)},
	}, sol)
}

func TestScalingSolutionHigherScale(t *testing.T) {
	sol, _ := findScalingSolution(
		// Need [EUR/2 200]
		big.NewInt(200),
		2,

		// Have: {EUR: 4} (eq to EUR/2 400)
		map[int64]*big.Int{
			0: big.NewInt(4),
		})

	require.Equal(t, []scalePair{
		{0, big.NewInt(2)},
	}, sol)
}

func TestScalingSolutionHigherScaleNoSolution(t *testing.T) {
	// TODO change name
	sol, _ := findScalingSolution(
		// Needed: [COIN/2 1]
		big.NewInt(1),
		2,

		// Got: {COIN: 100, COIN/1: 100}
		map[int64]*big.Int{
			0: big.NewInt(100),
			1: big.NewInt(100),
		})

	require.Equal(t, []scalePair{
		{1, big.NewInt(1)},
	}, sol)
}

func TestNoSolution(t *testing.T) {
	sol, got := findScalingSolution(
		// Need [USD/2 400]
		big.NewInt(400),
		2,
		map[int64]*big.Int{
			0: big.NewInt(1),
			2: big.NewInt(1),
			3: big.NewInt(1),
		})

	require.Equal(t, big.NewInt(100), got)
	require.Equal(t, []scalePair{
		{0, big.NewInt(1)},
	}, sol)
}

func TestNoSolution2(t *testing.T) {
	sol, tot := findScalingSolution(
		// Need [USD/2 400]
		big.NewInt(400),
		2,
		map[int64]*big.Int{
			0: big.NewInt(1),
			2: big.NewInt(200),
			3: big.NewInt(10),
		})

	require.Equal(t, []scalePair{
		{3, big.NewInt(10)},
		{0, big.NewInt(1)},
	}, sol)
	require.Equal(t, big.NewInt(100+1), tot)
}

func TestUnboundedScalingSameAsset(t *testing.T) {
	sol, tot := findScalingSolution(
		// Need [USD/2 *]
		nil,
		2,
		// Have: {EUR/2: 201}
		map[int64]*big.Int{
			2: big.NewInt(123),
		})

	require.Equal(t, []scalePair(nil), sol)
	require.Equal(t, big.NewInt(0), tot)
}

func TestUnboundedScalingLowerAsset(t *testing.T) {
	sol, _ := findScalingSolution(
		nil,
		2,
		map[int64]*big.Int{
			0: big.NewInt(1),
		})

	require.Equal(t, []scalePair{
		{0, big.NewInt(1)},
	}, sol)
}

func TestUnboundedScalinHigherAsset(t *testing.T) {
	sol, _ := findScalingSolution(
		nil,
		2,
		map[int64]*big.Int{
			3: big.NewInt(10),
		})

	require.Equal(t,
		[]scalePair{
			{3, big.NewInt(10)},
		},
		sol)
}

func TestUnboundedScalinHigherAssetTrimRemainder(t *testing.T) {
	sol, _ := findScalingSolution(
		nil,
		2,
		map[int64]*big.Int{
			3: big.NewInt(15),
		})

	require.Equal(t, []scalePair{
		{3, big.NewInt(10)},
	}, sol)
}

// getAssets must match a base asset exactly: assets that merely share the
// base name as a string prefix (e.g. "USDT" or "USD_RED") are NOT scaled
// forms of "USD" and must be ignored. Only the bare base and its
// "base/N" precision variants belong in the scaling pool.
//
// Sub-tests are deliberately split so each is deterministic regardless of
// Go's randomized map iteration order. A combined fixture is flaky:
// `getAssets` keys its result map by precision alone, so when both
// `USD/2 = 2` and `USD_RED/2 = 500` are present the buggy implementation
// writes both to `result[2]`, and the final value depends on iteration
// order — sometimes the correct one wins, sometimes the spurious one.
func TestGetAssetsRejectsSpuriousPrefixMatches(t *testing.T) {
	t.Run("USDT does not pollute USD pool", func(t *testing.T) {
		balance := AccountBalance{
			"USDT":   big.NewInt(100),
			"USDT/2": big.NewInt(200),
		}
		got := getAssets(balance, "USD")
		require.Empty(t, got, "USDT-prefixed entries must not be folded into USD")
	})

	t.Run("color-suffixed variants do not pollute USD pool", func(t *testing.T) {
		balance := AccountBalance{
			"USD_RED":   big.NewInt(50),
			"USD_RED/2": big.NewInt(500),
		}
		got := getAssets(balance, "USD")
		require.Empty(t, got, "USD_RED-prefixed entries must not be folded into USD")
	})

	t.Run("USD precision variants are picked up", func(t *testing.T) {
		balance := AccountBalance{
			"USD":   big.NewInt(1),
			"USD/2": big.NewInt(2),
			"USD/3": big.NewInt(3),
		}
		got := getAssets(balance, "USD")
		require.Equal(t, map[int64]*big.Int{
			0: big.NewInt(1),
			2: big.NewInt(2),
			3: big.NewInt(3),
		}, got)
	})

}
