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

func TestMultipleSends(t *testing.T) {
	p := parser.Parse(`
	send [COIN 10] ( source = @src destination = @dest )
	send [COIN 20] ( source = @src destination = @dest )
	`)
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

func TestAllotmentDest(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = @s
  destination = { 1/2 to @d }
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestCapped(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = max [EUR/2 10] from @src
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestCappedVariable(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = max $my_var from @src
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestNested(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = {
    max [COIN 42] from @src
	@a
	@b
  }
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceSend(t *testing.T) {
	p := parser.Parse(`send `)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceMonetary(t *testing.T) {
	p := parser.Parse(`send [COIN]`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceNoAddr(t *testing.T) {
	p := parser.Parse(`send  (
  source = @
  destination = {
	@
  }
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceInvalidDest(t *testing.T) {
	p := parser.Parse(`send [COIN 10] (
    source = @a
    destination =
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceInvalidSrcTk(t *testing.T) {
	p := parser.Parse(`send [COIN 10] (
    source = max
    destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}
