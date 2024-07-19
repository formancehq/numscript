package parser_test

import (
	"numscript/parser"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

func TestPositionGt(t *testing.T) {
	assert.True(t, (&parser.Position{
		Line:      10,
		Character: 20,
	}).GtEq(parser.Position{
		Line:      10,
		Character: 20,
	}), "a position is GtEq to itself")

	assert.True(t, (&parser.Position{
		Line:      100,
		Character: 0,
	}).GtEq(parser.Position{
		Line:      10,
		Character: 20,
	}), "x is GtEq to y when x line is greater")

	assert.False(t, (&parser.Position{
		Line:      10,
		Character: 0,
	}).GtEq(parser.Position{
		Line:      100,
		Character: 20,
	}), "x is not GtEq to y when x line is greater")

	assert.True(t, (&parser.Position{
		Line:      0,
		Character: 100,
	}).GtEq(parser.Position{
		Line:      0,
		Character: 20,
	}), "x is GtEq to y when they are in the same line and the char is higher")

	assert.False(t, (&parser.Position{
		Line:      0,
		Character: 19,
	}).GtEq(parser.Position{
		Line:      0,
		Character: 20,
	}), "x is not GtEq to y when they are in the same line and the char is lower")

}

func TestContainsSameLine(t *testing.T) {
	rng := parser.Range{
		Start: parser.Position{Line: 42, Character: 5},
		End:   parser.Position{Line: 42, Character: 10},
	}

	assert.True(t, rng.Contains(parser.Position{
		Line:      42,
		Character: 8,
	}), "contains position within the same line")

	assert.True(t, rng.Contains(parser.Position{
		Line:      42,
		Character: 5,
	}), "contains position with same start")

	assert.True(t, rng.Contains(parser.Position{
		Line:      42,
		Character: 10,
	}), "contains position with same end")

	assert.False(t, rng.Contains(parser.Position{
		Line:      42,
		Character: 4,
	}), "does not contain position before the start")

	assert.False(t, rng.Contains(parser.Position{
		Line:      42,
		Character: 12,
	}), "does not contain position after the end")

	assert.False(t, rng.Contains(parser.Position{
		Line:      10,
		Character: 8,
	}), "does not contain position before the line")

	assert.False(t, rng.Contains(parser.Position{
		Line:      100,
		Character: 8,
	}), "does not contain position after the line")
}

func TestContainsManyLines(t *testing.T) {
	rng := parser.Range{
		Start: parser.Position{Line: 10, Character: 5},
		End:   parser.Position{Line: 100, Character: 10},
	}

	assert.True(t, rng.Contains(parser.Position{
		Line:      50,
		Character: 0,
	}), "contains position between the lines even if char is before end")

	assert.True(t, rng.Contains(parser.Position{
		Line:      50,
		Character: 99,
	}), "contains position between the lines even if char is after end")
}

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

func TestVariableMonetary(t *testing.T) {
	p := parser.Parse(`send $example (
  source = @a
  destination = @b
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestInorderSource(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = { @s1 @s2 }
  destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestInorderDestination(t *testing.T) {
	p := parser.Parse(`send $amt (
  source = @s
  destination = {
	max $m1 to @d1
	max [C 42] kept
	remaining to @d3
  }
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
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
	remaining from @s3
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

func TestAllotmentVariableSource(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = {
	$x from @a
  }
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

func TestAllotmentDestRemaining(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = @s
  destination = {
  	1/2 to @d
	remaining to @d2
  }
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotmentDestKept(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = @s
  destination = { 1/2 kept }
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
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

func TestEmptyVars(t *testing.T) {
	p := parser.Parse(`vars { }`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestSingleVar(t *testing.T) {
	p := parser.Parse(`vars { monetary $my_var }`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestOverdraftUnbounded(t *testing.T) {
	p := parser.Parse(`send $x (
	source = @src allowing unbounded overdraft
	destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestOverdraftUnboundedVariable(t *testing.T) {
	p := parser.Parse(`send $x (
	source = $my_src_var allowing unbounded overdraft
	destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestBoundedOverdraft(t *testing.T) {
	p := parser.Parse(`send $x (
	source = $my_src_var allowing overdraft up to [COIN 100]
	destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestFunctionCallNoArgs(t *testing.T) {
	p := parser.Parse(`example_fn()`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestFunctionCallOneArg(t *testing.T) {
	p := parser.Parse(`example_fn(@example)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestFunctionCallManyArgs(t *testing.T) {
	p := parser.Parse(`
example_fn(
	[COIN 42],
	100,
	1/2,
	$my_var,
	"example_str"
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestVarOrigin(t *testing.T) {
	p := parser.Parse(`
vars {
	monetary $my_var = origin_fn(@my_account, "str")
}
`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestSendAll(t *testing.T) {
	p := parser.Parse(`send [ASSET *] (
	source = @a
	destination = @b
)
`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}
