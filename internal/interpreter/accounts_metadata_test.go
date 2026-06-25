package interpreter

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestCompareAccountsMetadata(t *testing.T) {
	x := AccountMetadataRow{Account: "a", Key: "k", Value: "1"}
	y := AccountMetadataRow{Account: "a", Key: "k", Value: "2"}

	t.Run("equal regardless of order", func(t *testing.T) {
		require.True(t, CompareAccountsMetadata(
			AccountsMetadata{x, y},
			AccountsMetadata{y, x},
		))
	})

	t.Run("different value is not equal", func(t *testing.T) {
		require.False(t, CompareAccountsMetadata(
			AccountsMetadata{x},
			AccountsMetadata{y},
		))
	})

	t.Run("respects multiplicity: [x, x] != [x, y]", func(t *testing.T) {
		require.False(t, CompareAccountsMetadata(
			AccountsMetadata{x, x},
			AccountsMetadata{x, y},
		))
		// and the symmetric direction
		require.False(t, CompareAccountsMetadata(
			AccountsMetadata{x, y},
			AccountsMetadata{x, x},
		))
	})

	t.Run("identical multisets are equal", func(t *testing.T) {
		require.True(t, CompareAccountsMetadata(
			AccountsMetadata{x, x},
			AccountsMetadata{x, x},
		))
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
