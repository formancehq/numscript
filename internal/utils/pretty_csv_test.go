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
	}, true)

	snaps.MatchSnapshot(t, out)
}

func TestPrettyCsvMap(t *testing.T) {
	out := utils.CsvPrettyMap("Name", "Value", map[string]string{
		"a":                       "0",
		"b":                       "12345",
		"very-very-very-long-key": "",
	})

	snaps.MatchSnapshot(t, out)
}
