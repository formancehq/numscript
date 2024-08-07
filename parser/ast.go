package parser

import "math/big"

type Position struct {
	Character int
	Line      int
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

type Range struct {
	Start Position
	End   Position
}

func (r Range) Contains(position Position) bool {
	//  position >= r.Start && r.End >= position
	return position.GtEq(r.Start) && r.End.GtEq(position)
}

// does it even make sense to have a literal supertype?
type Literal interface {
	literal()
	GetRange() Range
}

func (*AssetLiteral) literal()    {}
func (*MonetaryLiteral) literal() {}
func (*AccountLiteral) literal()  {}
func (*VariableLiteral) literal() {}
func (*RatioLiteral) literal()    {}
func (*NumberLiteral) literal()   {}
func (*StringLiteral) literal()   {}

func (l *AssetLiteral) GetRange() Range    { return l.Range }
func (l *MonetaryLiteral) GetRange() Range { return l.Range }
func (l *AccountLiteral) GetRange() Range  { return l.Range }
func (l *VariableLiteral) GetRange() Range { return l.Range }
func (l *RatioLiteral) GetRange() Range    { return l.Range }
func (l *NumberLiteral) GetRange() Range   { return l.Range }
func (l *StringLiteral) GetRange() Range   { return l.Range }

type (
	AssetLiteral struct {
		Range Range
		Asset string
	}

	NumberLiteral struct {
		Range  Range
		Number int
	}

	StringLiteral struct {
		Range  Range
		String string
	}

	MonetaryLiteral struct {
		Range  Range
		Asset  Literal
		Amount Literal
	}

	AccountLiteral struct {
		Range Range
		Name  string
	}

	RatioLiteral struct {
		Range       Range
		Numerator   uint64
		Denominator uint64
	}

	VariableLiteral struct {
		Range Range
		Name  string
	}
)

func (r RatioLiteral) ToRatio() *big.Rat {
	return big.NewRat(int64(r.Numerator), int64(r.Denominator))
}

type RemainingAllotment struct {
	Range Range
}

func (a *AccountLiteral) IsWorld() bool {
	return a.Name == "world"
}

// Source exprs

type Source interface {
	source()
	GetRange() Range
}

func (*SourceInorder) source()   {}
func (*SourceAllotment) source() {}
func (*AccountLiteral) source()  {}
func (*VariableLiteral) source() {}
func (*SourceCapped) source()    {}
func (*SourceOverdraft) source() {}

func (s *SourceInorder) GetRange() Range   { return s.Range }
func (s *SourceAllotment) GetRange() Range { return s.Range }
func (s *SourceCapped) GetRange() Range    { return s.Range }
func (s *SourceOverdraft) GetRange() Range { return s.Range }

type (
	SourceInorder struct {
		Range   Range
		Sources []Source
	}
	SourceAllotment struct {
		Range Range
		Items []SourceAllotmentItem
	}

	SourceCapped struct {
		Range Range
		From  Source
		Cap   Literal
	}

	SourceOverdraft struct {
		Range   Range
		Address Literal
		Bounded *Literal
	}
)

type AllotmentValue interface{ allotmentValue() }

func (*RemainingAllotment) allotmentValue() {}
func (*RatioLiteral) allotmentValue()       {}
func (*VariableLiteral) allotmentValue()    {}

type SourceAllotmentItem struct {
	Range     Range
	Allotment AllotmentValue
	From      Source
}

// Destination exprs
type Destination interface {
	destination()
	GetRange() Range
}

func (*DestinationInorder) destination()   {}
func (*AccountLiteral) destination()       {}
func (*VariableLiteral) destination()      {}
func (*DestinationAllotment) destination() {}

func (d *DestinationInorder) GetRange() Range   { return d.Range }
func (d *DestinationAllotment) GetRange() Range { return d.Range }

type (
	DestinationInorder struct {
		Range     Range
		Clauses   []DestinationInorderClause
		Remaining KeptOrDestination
	}

	DestinationInorderClause struct {
		Range Range
		Cap   Literal
		To    KeptOrDestination
	}
)

type KeptOrDestination interface {
	keptOrDestination()
}
type DestinationKept struct {
	Range Range
}
type DestinationTo struct {
	Destination Destination
}

func (*DestinationKept) keptOrDestination() {}
func (*DestinationTo) keptOrDestination()   {}

type DestinationAllotment struct {
	Range Range
	Items []DestinationAllotmentItem
}

type DestinationAllotmentItem struct {
	Range     Range
	Allotment AllotmentValue
	To        KeptOrDestination
}

// Statements

type Statement interface {
	statement()
	GetRange() Range
}

func (*FnCall) statement()        {}
func (*SendStatement) statement() {}

func (s *FnCall) GetRange() Range        { return s.Range }
func (s *SendStatement) GetRange() Range { return s.Range }

type FnCallIdentifier struct {
	Range Range
	Name  string
}

type FnCall struct {
	Range  Range
	Caller *FnCallIdentifier
	Args   []Literal
}

type SentValue interface{ sentValue() }
type SentValueLiteral struct {
	Monetary Literal
}
type SentValueAll struct {
	Range Range
	Asset Literal
}

func (*SentValueLiteral) sentValue() {}
func (*SentValueAll) sentValue()     {}

type SendStatement struct {
	Range       Range
	SentValue   SentValue
	Source      Source
	Destination Destination
}

type TypeDecl struct {
	Range Range
	Name  string
}

type VarDeclaration struct {
	Range  Range
	Name   *VariableLiteral
	Type   *TypeDecl
	Origin *FnCall
}

type Program struct {
	Vars       []VarDeclaration
	Statements []Statement
}
