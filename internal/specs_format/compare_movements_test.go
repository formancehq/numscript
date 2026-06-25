package specs_format

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompareMovementsMultiplicity(t *testing.T) {
	x := Movement{Source: "world", Destination: "a", Asset: "USD", Amount: big.NewInt(1)}
	y := Movement{Source: "world", Destination: "b", Asset: "USD", Amount: big.NewInt(1)}

	// [x, x] must not equal [x, y]
	require.False(t, compareMovements(Movements{x, x}, Movements{x, y}))
	require.False(t, compareMovements(Movements{x, y}, Movements{x, x}))

	// order-independent and multiplicity-exact equality still holds
	require.True(t, compareMovements(Movements{x, y}, Movements{y, x}))
	require.True(t, compareMovements(Movements{x, x}, Movements{x, x}))

	// differing amount on the same key is not equal
	z := Movement{Source: "world", Destination: "a", Asset: "USD", Amount: big.NewInt(2)}
	require.False(t, compareMovements(Movements{x}, Movements{z}))
}
