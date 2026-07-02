package interpreter

import (
	"math/big"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestPrettyPrintPostings(t *testing.T) {
	t.Run("no scope, no color (no optional columns)", func(t *testing.T) {
		postings := []Posting{
			{Source: "world", Destination: "alice", Asset: "EUR/2", Amount: big.NewInt(100)},
		}

		snaps.MatchSnapshot(t, PrettyPrintPostings(postings))
	})

	t.Run("only source scope (only Source Scope column shown)", func(t *testing.T) {
		postings := []Posting{
			{Source: "src", SourceScope: "x", Destination: "dest", Asset: "USD", Amount: big.NewInt(10)},
			{Source: "world", Destination: "dest", Asset: "USD", Amount: big.NewInt(5)},
		}

		snaps.MatchSnapshot(t, PrettyPrintPostings(postings))
	})

	t.Run("both scopes (both Scope columns shown)", func(t *testing.T) {
		postings := []Posting{
			{Source: "src", SourceScope: "x", Destination: "dest", DestinationScope: "y", Asset: "USD", Amount: big.NewInt(10)},
		}

		snaps.MatchSnapshot(t, PrettyPrintPostings(postings))
	})
}
