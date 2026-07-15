package interpreter

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestCompareSetAccountsMetadata(t *testing.T) {
	x := SetAccountMetadataRow{Account: "a", Key: "k", Value: NewMonetaryInt(1)}
	y := SetAccountMetadataRow{Account: "a", Key: "k", Value: NewMonetaryInt(2)}

	t.Run("equal regardless of order", func(t *testing.T) {
		require.True(t, CompareSetAccountsMetadata(
			SetAccountsMetadata{x, y},
			SetAccountsMetadata{y, x},
		))
	})

	t.Run("different value is not equal", func(t *testing.T) {
		require.False(t, CompareSetAccountsMetadata(
			SetAccountsMetadata{x},
			SetAccountsMetadata{y},
		))
	})

	t.Run("respects multiplicity: [x, x] != [x, y]", func(t *testing.T) {
		require.False(t, CompareSetAccountsMetadata(
			SetAccountsMetadata{x, x},
			SetAccountsMetadata{x, y},
		))
		// and the symmetric direction
		require.False(t, CompareSetAccountsMetadata(
			SetAccountsMetadata{x, y},
			SetAccountsMetadata{x, x},
		))
	})

	t.Run("identical multisets are equal", func(t *testing.T) {
		require.True(t, CompareSetAccountsMetadata(
			SetAccountsMetadata{x, x},
			SetAccountsMetadata{x, x},
		))
	})
}

func TestScopeValidation(t *testing.T) {
	t.Run("valid scopes", func(t *testing.T) {
		require.True(t, checkScopeName(""))
		require.True(t, checkScopeName("myscope"))
		require.True(t, checkScopeName("x"))
		require.True(t, checkScopeName("x1"))
		require.True(t, checkScopeName("my_scope_with_underscores"))
	})

	t.Run("invalid scopes", func(t *testing.T) {
		require.False(t, checkScopeName("!"))
		require.False(t, checkScopeName("$"))
		require.False(t, checkScopeName("UPPERCASE"))
		require.False(t, checkScopeName("dash-case"))
		require.False(t, checkScopeName("colons:within"))
	})
}

func TestPrettyPrintAccountsMetadata(t *testing.T) {
	t.Run("without scope (no Scope column)", func(t *testing.T) {
		meta := AccountsMetadata{
			{Account: "alice", Key: "kyc", Value: "verified"},
			{Account: "bob", Key: "tier", Value: "gold"},
		}

		snaps.MatchSnapshot(t, meta.PrettyPrint())
	})

	t.Run("with scope (Scope column shown)", func(t *testing.T) {
		meta := AccountsMetadata{
			{Account: "alice", Key: "kyc", Value: "verified"},
			{Account: "alice", Scope: "eu", Key: "kyc", Value: "pending"},
			{Account: "bob", Key: "tier", Value: "gold"},
		}

		snaps.MatchSnapshot(t, meta.PrettyPrint())
	})
}
