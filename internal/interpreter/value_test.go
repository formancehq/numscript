package interpreter_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/PagoPlus/numscript-wasm/internal/interpreter"

	"github.com/stretchr/testify/require"
)

func TestMarshalMonetaryInt(t *testing.T) {
	t.Parallel()

	x := interpreter.NewMonetaryInt(42)

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.Equal(t, string(j), `"42"`)
}

func TestMarshalString(t *testing.T) {
	t.Parallel()

	x := interpreter.String("abc")

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.Equal(t, string(j), `"abc"`)
}

func TestMarshalAsset(t *testing.T) {
	t.Parallel()

	x := interpreter.Asset("EUR/2")

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.Equal(t, string(j), `"EUR/2"`)
}

func TestMarshalAddress(t *testing.T) {
	t.Parallel()

	x := interpreter.AccountAddress("abc")

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.Equal(t, string(j), `"abc"`)
}

func TestMarshalPortion(t *testing.T) {
	t.Parallel()

	x := interpreter.Portion(*big.NewRat(2, 3))

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.Equal(t, string(j), `"2/3"`)
}

func TestMarshalMonetary(t *testing.T) {
	t.Parallel()

	x := interpreter.Monetary{
		Asset:  interpreter.Asset("USD/2"),
		Amount: interpreter.NewMonetaryInt(100),
	}

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.Equal(t, string(j), `"USD/2 100"`)
}
