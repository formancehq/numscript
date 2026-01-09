package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScalingZeroNeeded(t *testing.T) {
	t.Skip()

	// need [EUR/2 ]
	sol, _ := findSolution(
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

func TestScalingSameAsset(t *testing.T) {
	sol, _ := findSolution(
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
}

func TestScalingSolutionLowerScale(t *testing.T) {
	sol, _ := findSolution(
		big.NewInt(1),
		0,
		map[int64]*big.Int{
			2: big.NewInt(900),
		})

	require.Equal(t, []scalePair{
		{2, big.NewInt(100)},
	}, sol)
}

func TestScalingSolutionHigherScale(t *testing.T) {
	sol, _ := findSolution(
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
	sol, _ := findSolution(
		big.NewInt(1),
		2,
		map[int64]*big.Int{
			0: big.NewInt(100),
			1: big.NewInt(100),
		})

	require.Nil(t, sol)
}

func TestMixedFail(t *testing.T) {
	t.Skip()

	sol, _ := findSolution(
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
	sol, _ := findSolution(
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
	sol, _ := findSolution(
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
	sol, _ := findSolution(
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
	sol, _ := findSolution(
		nil,
		2,
		map[int64]*big.Int{
			3: big.NewInt(15),
		})

	require.Equal(t, []scalePair{
		{3, big.NewInt(10)},
	}, sol)
}
