package parser_test

import (
	"numscript/parser"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestPlainAddress(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = @src
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestVariable(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = $example_var_src
  destination = $example_var_dest
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestSeq(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = { @s1 @s2 }
  destination = { @d1 @d2 }
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotment(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = { 1/3 from @s1 }
  destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotmentPerc(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = {
    42% from @s1
	1/2 from @s2
  }
  destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotmentPercFloating(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = { 2.42% from @s }
  destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}
