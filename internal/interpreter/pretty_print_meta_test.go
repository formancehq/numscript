package interpreter

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestPrettyPrintMeta(t *testing.T) {
	t.Run("renders plain values", func(t *testing.T) {
		meta := Metadata{
			"greeting": String("hello"),
			"count":    NewMonetaryInt(42),
		}

		snaps.MatchSnapshot(t, PrettyPrintMeta(meta))
	})

	t.Run("renders a scoped account value in its source form", func(t *testing.T) {
		meta := Metadata{
			"greeting": String("hello"),
			"owner":    AccountAddress{Name: "alice", Scope: "reserve"},
		}

		snaps.MatchSnapshot(t, PrettyPrintMeta(meta))
	})
}
