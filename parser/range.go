package parser

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
