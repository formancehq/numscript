package parser

import "strings"

type Position struct {
	Character int
	Line      int
}

type Range struct {
	Start Position
	End   Position
}

func (p1 *Position) GtEq(p2 Position) bool {
	if p1.Line == p2.Line {
		return p1.Character >= p2.Character
	}

	return p1.Line > p2.Line
}

func (p *Position) AsRange() Range {
	//  position >= r.Start && r.End >= position
	return Range{Start: *p, End: *p}
}

func (r Range) Contains(position Position) bool {
	//  position >= r.Start && r.End >= position
	return position.GtEq(r.Start) && r.End.GtEq(position)
}

// Those functions are mostly used as test utilities
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

func PositionOf(src string, substr string) *Position {
	return PositionOfIndexed(src, substr, 0)
}

func PositionOfIndexed(src string, substr string, occurrence int) *Position {
	// TODO make offset to position utility
	offset := indexOfOccurrence(src, substr, occurrence)
	if offset == -1 {
		return nil
	}

	pos := Position{}
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

func RangeOfIndexed(src string, substr string, occurrence int) Range {
	start := *PositionOfIndexed(src, substr, occurrence)
	end := start
	end.Character += len(substr)

	return Range{
		Start: start,
		End:   end,
	}
}
