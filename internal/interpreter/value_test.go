package interpreter_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"

	"github.com/stretchr/testify/require"
)

func TestMetaString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		value    interpreter.Value
		expected string
	}{
		{"string is not quoted", interpreter.String("abc"), "abc"},
		{"empty string", interpreter.String(""), ""},
		{"a string that looks like a number", interpreter.String("42"), "42"},
		{"asset", interpreter.Asset("EUR/2"), "EUR/2"},
		{"account has no @ prefix", interpreter.AccountAddress{Name: "alice"}, "alice"},
		{"number", interpreter.NewMonetaryInt(42), "42"},
		{"negative number", interpreter.NewMonetaryInt(-7), "-7"},
		{"monetary", interpreter.Monetary{Asset: "USD/2", Amount: interpreter.NewMonetaryInt(100)}, "USD/2 100"},
		{"portion", interpreter.Portion(*big.NewRat(2, 3)), "2/3"},
		{"portion is normalized", interpreter.Portion(*big.NewRat(2, 4)), "1/2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, interpreter.MetaString(tc.value))
		})
	}
}

// The metadata format is untyped on purpose: it stores the rendered value, so
// values of different types that render the same are indistinguishable there.
// This is the trade-off that keeps the format stringly-typed, and it is what the
// pre-tagged format did too.
func TestMetaStringConflatesTypesThatRenderAlike(t *testing.T) {
	t.Parallel()

	require.Equal(t,
		interpreter.MetaString(interpreter.String("42")),
		interpreter.MetaString(interpreter.NewMonetaryInt(42)),
	)

	require.Equal(t,
		interpreter.MetaString(interpreter.String("COIN")),
		interpreter.MetaString(interpreter.Asset("COIN")),
	)

	require.Equal(t,
		interpreter.MetaString(interpreter.String("alice")),
		interpreter.MetaString(interpreter.AccountAddress{Name: "alice"}),
	)
}

// String() is the diagnostic form and stays delimited, so error messages can
// still tell a string from an account. MetaString is the wire form.
func TestStringDiffersFromMetaString(t *testing.T) {
	t.Parallel()

	require.Equal(t, `"abc"`, interpreter.String("abc").String())
	require.Equal(t, "abc", interpreter.MetaString(interpreter.String("abc")))

	require.Equal(t, "@alice", interpreter.AccountAddress{Name: "alice"}.String())
	require.Equal(t, "alice", interpreter.MetaString(interpreter.AccountAddress{Name: "alice"}))
}
