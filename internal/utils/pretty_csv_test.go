package utils_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/utils"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestPrettyCsv(t *testing.T) {
	out := utils.CsvPretty([]string{
		"Account", "Asset", "Balance",
	}, [][]string{
		{"alice", "EUR/2", "1"},
		{"alice", "USD/1234", "999999"},
		{"bob", "BTC", "3"},
	})

	snaps.MatchSnapshot(t, out)
}
