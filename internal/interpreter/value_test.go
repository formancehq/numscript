package interpreter_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"

	"github.com/stretchr/testify/require"
)

func TestMarshalMonetaryInt(t *testing.T) {
	t.Parallel()

	x := interpreter.NewMonetaryInt(42)

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.JSONEq(t, `{"type":"number","value":"42"}`, string(j))
}

func TestMarshalString(t *testing.T) {
	t.Parallel()

	x := interpreter.String("abc")

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.JSONEq(t, `{"type":"string","value":"abc"}`, string(j))
}

func TestMarshalAsset(t *testing.T) {
	t.Parallel()

	x := interpreter.Asset("EUR/2")

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.JSONEq(t, `{"type":"asset","name":"EUR/2"}`, string(j))
}

func TestMarshalAddress(t *testing.T) {
	t.Parallel()

	j, err := json.Marshal(interpreter.AccountAddress{Name: "abc"})
	require.Nil(t, err)
	require.JSONEq(t, `{"type":"account","name":"abc"}`, string(j))

	j, err = json.Marshal(interpreter.AccountAddress{Name: "abc", Scope: "s"})
	require.Nil(t, err)
	require.JSONEq(t, `{"type":"account","name":"abc","scope":"s"}`, string(j))
}

func TestMarshalPortion(t *testing.T) {
	t.Parallel()

	x := interpreter.Portion(*big.NewRat(2, 3))

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.JSONEq(t, `{"type":"portion","numerator":"2","denominator":"3"}`, string(j))
}

func TestMarshalMonetary(t *testing.T) {
	t.Parallel()

	x := interpreter.Monetary{
		Asset:  interpreter.Asset("USD/2"),
		Amount: interpreter.NewMonetaryInt(100),
	}

	j, err := json.Marshal(x)
	require.Nil(t, err)
	require.JSONEq(t, `{"type":"monetary","asset":"USD/2","amount":"100"}`, string(j))
}

func TestParseTaggedValueRoundTrip(t *testing.T) {
	t.Parallel()

	values := []interpreter.Value{
		interpreter.String("abc"),
		interpreter.Asset("EUR/2"),
		interpreter.AccountAddress{Name: "alice"},
		interpreter.AccountAddress{Name: "alice", Scope: "reserve"},
		interpreter.NewMonetaryInt(42),
		interpreter.Monetary{Asset: "USD/2", Amount: interpreter.NewMonetaryInt(100)},
		interpreter.Portion(*big.NewRat(2, 3)),
	}

	for _, v := range values {
		j, err := json.Marshal(v)
		require.Nil(t, err)

		parsed, err := interpreter.ParseTaggedValue(j)
		require.Nil(t, err)
		// compare on the canonical source form
		require.Equal(t, v.String(), parsed.String())
	}
}

func TestParseTaggedValueRejectsUnknownType(t *testing.T) {
	t.Parallel()

	_, err := interpreter.ParseTaggedValue([]byte(`{"type":"bogus"}`))
	require.Error(t, err)
}
