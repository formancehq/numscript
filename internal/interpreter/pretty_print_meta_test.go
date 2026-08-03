package interpreter

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// Metadata holds already-rendered values (see MetaString), so PrettyPrintMeta
// only has the table layout left to get right.
func TestPrettyPrintMeta(t *testing.T) {
	t.Run("renders plain values", func(t *testing.T) {
		meta := Metadata{
			"greeting": MetaString(String("hello")),
			"count":    MetaString(NewMonetaryInt(42)),
		}

		snaps.MatchSnapshot(t, PrettyPrintMeta(meta))
	})

	t.Run("renders an account value as its bare name", func(t *testing.T) {
		meta := Metadata{
			"greeting": MetaString(String("hello")),
			"owner":    MetaString(AccountAddress{Name: "alice"}),
		}

		snaps.MatchSnapshot(t, PrettyPrintMeta(meta))
	})
}
