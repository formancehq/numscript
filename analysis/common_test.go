package analysis_test

import (
	"numscript/parser"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func indexOfOccurrence(src string, substr string, occurence int) int {
	// TODO this function can probably be simplified
	offset := strings.Index(src, substr)
	if offset == -1 {
		return -1
	}

	for ; occurence > 0; occurence-- {
		shift := offset + len(substr)
		shiftedOffset := strings.Index(src[shift:], substr)
		if shiftedOffset == -1 {
			return -1
		}

		offset += shiftedOffset + len(substr)
	}

	return offset
}
func TestIndexOfOccurence(t *testing.T) {
	assert.Equal(t,
		-1,
		indexOfOccurrence("abc", "Z", 0),
	)

	assert.Equal(t,
		2,
		indexOfOccurrence("0123", "2", 0),
	)

	assert.Equal(t,
		1,
		indexOfOccurrence("xZZbcZZx", "ZZ", 0),
	)

	assert.Equal(t,
		5,
		indexOfOccurrence("xZZbcZZx", "ZZ", 1),
	)

	assert.Equal(t,
		-1,
		indexOfOccurrence("abca", "a", 10),
	)
}

func PositionOf(src string, substr string) *parser.Position {
	return PositionOfIndexed(src, substr, 0)
}

func PositionOfIndexed(src string, substr string, occurrence int) *parser.Position {
	// TODO make offset to position utility
	offset := indexOfOccurrence(src, substr, occurrence)
	if offset == -1 {
		return nil
	}

	pos := parser.Position{}
	for thisOffset, rune := range src {
		if thisOffset == offset {
			break
		}

		if rune == '\n' {
			pos.Line++
			pos.Character = 0
		} else {
			pos.Character++
		}
	}

	return &pos
}

func RangeOfIndexed(src string, substr string, occurrence int) parser.Range {
	start := *PositionOfIndexed(src, substr, occurrence)
	end := start
	end.Character += len(substr)

	return parser.Range{
		Start: start,
		End:   end,
	}
}

func TestPositionOf(t *testing.T) {
	assert.Equal(t,
		&parser.Position{Character: 0},
		PositionOf("abc", "a"),
	)

	assert.Equal(t,
		&parser.Position{Character: 1},
		PositionOf("abc", "b"),
	)

	assert.Equal(t,
		&parser.Position{Line: 1, Character: 0},
		PositionOf("a\nbc", "b"),
	)

	assert.Equal(t,
		&parser.Position{Line: 2, Character: 1},
		PositionOf("a\nb\ncd", "d"),
	)
}

func TestPositionOfIndexed(t *testing.T) {
	assert.Equal(t,
		&parser.Position{Character: 2},
		PositionOfIndexed("aba", "a", 1),
	)

	assert.Equal(t,
		&parser.Position{Line: 2, Character: 1},
		PositionOfIndexed("a\nd\ncd", "d", 1),
	)

}
