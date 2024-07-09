package parser_test

import (
	"numscript/parser"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestParser(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = @src
)`)
	snaps.MatchSnapshot(t, p.Value)
}
