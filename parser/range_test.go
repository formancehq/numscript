package parser_test

import (
	"testing"

	"github.com/formancehq/numscript/parser"
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
