package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScalingZeroNeeded(t *testing.T) {
	t.Skip()

	sol, _ := findScalingSolution(
		big.NewInt(0),
		42,
		map[int64]*big.Int{
			2: big.NewInt(100),
			1: big.NewInt(1),
		})

	require.Equal(t, []scalePair{
		{42, big.NewInt(0)},
	}, sol)
}

func TestAllowSpare(t *testing.T) {
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
	require.Equal(t, big.NewInt(1), tot)
}

func TestRepro(t *testing.T) {
	sol, _ := findScalingSolution(
		// Need [EUR/2 400]
		big.NewInt(400),
		2,

		// Have: {EUR: 99, EUR/2: 1}
		map[int64]*big.Int{
			2: big.NewInt(1),
			0: big.NewInt(99),
		})

	require.Equal(t, []scalePair{
		{2, big.NewInt(1)},
		{0, big.NewInt(4)},
	}, sol)
	// require.Equal(t, big.NewInt(1), tot)
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

	require.Equal(t, []scalePair{
		{2, big.NewInt(200)},
	}, sol)
	require.Equal(t, big.NewInt(200), tot)
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
		big.NewInt(400),
		2,
		map[int64]*big.Int{
			0: big.NewInt(1),
			2: big.NewInt(1),
			3: big.NewInt(1),
		})

	require.Equal(t, big.NewInt(100+1+0), got)
	require.Equal(t, []scalePair{
		{2, big.NewInt(1)},
		{0, big.NewInt(1)},
	}, sol)
}

func TestMixedFail(t *testing.T) {
	t.Skip()

	sol, _ := findScalingSolution(
		big.NewInt(400),
		2,
		map[int64]*big.Int{
			0: big.NewInt(1),
			2: big.NewInt(200),
			3: big.NewInt(10),
		})

	require.Nil(t, sol)
}

func TestUnboundedScalingSameAsset(t *testing.T) {
	sol, _ := findScalingSolution(
		nil,
		// Need USD/2
		2,
		// Have: {EUR/2: 201}
		map[int64]*big.Int{
			2: big.NewInt(123),
		})

	require.Equal(t, []scalePair{
		{2, big.NewInt(123)},
	}, sol)
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
