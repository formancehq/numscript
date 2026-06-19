package interpreter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountAddressString(t *testing.T) {
	t.Run("no scope", func(t *testing.T) {
		require.Equal(t, AccountAddress{Name: "acc"}.String(), "acc")
	})

	t.Run("with scope", func(t *testing.T) {
		require.Equal(t, AccountAddress{Name: "acc", Scope: "xyz"}.String(), "acc/xyz")
	})
}

func TestScopeValidation(t *testing.T) {
	t.Run("valid scopes", func(t *testing.T) {
		require.True(t, validateScope(""))
		require.True(t, validateScope("myscope"))
		require.True(t, validateScope("x"))
		require.True(t, validateScope("x1"))
		require.True(t, validateScope("my_scope_with_underscores"))
	})

	t.Run("invalid scopes", func(t *testing.T) {
		require.False(t, validateScope("!"))
		require.False(t, validateScope("$"))
		require.False(t, validateScope("UPPERCASE"))
		require.False(t, validateScope("dash-case"))
		require.False(t, validateScope("colons:within"))
	})

}
