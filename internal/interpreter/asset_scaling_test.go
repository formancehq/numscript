package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScalingSolution1(t *testing.T) {
	// TODO
	t.Skip()

	sol := findSolution(
		big.NewInt(0),
		2,
		map[int]*big.Int{
			2: big.NewInt(100),
			1: big.NewInt(1),
		})

	require.Equal(t, map[int]*big.Int{
		2: big.NewInt(100),
		1: big.NewInt(1),
	}, sol)
}

func TestScalingSameAsset(t *testing.T) {
	sol := findSolution(
		// Need [EUR/2 200]
		big.NewInt(200),
		2,

		// Have: {EUR/2: 201}
		map[int]*big.Int{
			2: big.NewInt(201),
		})

	require.Equal(t, map[int]*big.Int{
		2: big.NewInt(200),
	}, sol)
}

func TestScalingSolutionLowerScale(t *testing.T) {
	sol := findSolution(
		big.NewInt(1),
		0,
		map[int]*big.Int{
			2: big.NewInt(900),
		})

	require.Equal(t, map[int]*big.Int{
		2: big.NewInt(100),
	}, sol)
}

func TestScalingSolutionHigherScale(t *testing.T) {
	sol := findSolution(
		// Need [EUR/2 200]
		big.NewInt(200),
		2,

		// Have: {EUR: 4} (eq to EUR/2 400)
		map[int]*big.Int{
			0: big.NewInt(4),
		})

	require.Equal(t, map[int]*big.Int{
		0: big.NewInt(2),
	}, sol)
}

func TestScalingSolutionHigherScaleNoSolution(t *testing.T) {
	sol := findSolution(
		big.NewInt(1),
		2,
		map[int]*big.Int{
			0: big.NewInt(100),
			1: big.NewInt(100),
		})

	require.Equal(t, nil, sol)
}
