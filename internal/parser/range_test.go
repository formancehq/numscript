package parser_test

import (
	"testing"

	"github.com/PagoPlus/numscript-wasm/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

func TestPositionGt(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

func TestPositionOf(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		&parser.Position{Character: 0},
		parser.PositionOf("abc", "a"),
	)

	assert.Equal(t,
		&parser.Position{Character: 1},
		parser.PositionOf("abc", "b"),
	)

	assert.Equal(t,
		&parser.Position{Line: 1, Character: 0},
		parser.PositionOf("a\nbc", "b"),
	)

	assert.Equal(t,
		&parser.Position{Line: 2, Character: 1},
		parser.PositionOf("a\nb\ncd", "d"),
	)
}

func TestPositionOfIndexed(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		&parser.Position{Character: 2},
		parser.PositionOfIndexed("aba", "a", 1),
	)

	assert.Equal(t,
		&parser.Position{Line: 2, Character: 1},
		parser.PositionOfIndexed("a\nd\ncd", "d", 1),
	)
}

func TestShowRangeOnSourceSameLine(t *testing.T) {
	src := `example error end of line`

	errorRange := parser.RangeOfIndexed(src, "error", 0)

	snaps.MatchSnapshot(t, errorRange.ShowOnSource(src))
}

func TestShowRangeOnMultilineRanges(t *testing.T) {
	src := `example err
or spanning 2 lines`

	pos1 := parser.PositionOfIndexed(src, "err", 0)

	rng := parser.Range{
		Start: *pos1,
		End: parser.Position{
			Line:      1,
			Character: 2,
		},
	}

	snaps.MatchSnapshot(t, rng.ShowOnSource(src))
}

func TestShowRangeComplex(t *testing.T) {
	t.Parallel()

	src := `
example err
or that spans more
lines and then other
words with no error
at all`

	rng := parser.Range{
		Start: parser.Position{
			Line:      1,
			Character: 3,
		},
		End: parser.Position{
			Line:      3,
			Character: 5,
		},
	}

	snaps.MatchSnapshot(t, rng.ShowOnSource(src))
}
