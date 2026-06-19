package interpreter

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

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
