package cmd

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestShowDiff(t *testing.T) {
	var buf bytes.Buffer
	showDiff(
		&buf,
		map[string]any{"x": 42},
		map[string]any{"x": 100},
	)
	snaps.MatchSnapshot(t, buf.String())
}
