package parser

type Position struct {
	Character int
	Line      int
}

type Range struct {
	Start Position
	End   Position
}

// does it even make sense to have a literal supertype?
type Literal interface{ literal() }

func (*MonetaryLiteral) literal() {}
func (*AccountLiteral) literal()  {}
func (*VariableLiteral) literal() {}
func (*RatioLiteral) literal()    {}

type MonetaryLiteral struct {
	Range  Range
	Asset  string
	Amount int
}

type AccountLiteral struct {
	Range Range
	Name  string
}

type VariableLiteral struct {
	Range Range
	Name  string
}

type RatioLiteral struct {
	Range       Range
	Numerator   uint64
	Denominator uint64
}

type CapLit interface{ capLit() }

func (*MonetaryLiteral) capLit() {}
func (*VariableLiteral) capLit() {}

// Source exprs

type Source interface{ source() }

func (*SourceSeq) source()       {}
func (*SourceAllotment) source() {}
func (*AccountLiteral) source()  {}
func (*VariableLiteral) source() {}
func (*SourceCapped) source()    {}

type SourceSeq struct {
	Range   Range
	Sources []Source
}

type SourceAllotment struct {
	Range Range
	Items []SourceAllotmentItem
}

type SourceCapped struct {
	Range Range
	From  Source
	Cap   CapLit
}

type SourceAllotmentValue interface{ sourceAllotmentValue() }

func (*RatioLiteral) sourceAllotmentValue()    {}
func (*VariableLiteral) sourceAllotmentValue() {}

type SourceAllotmentItem struct {
	Range     Range
	Allotment SourceAllotmentValue
	From      Source
}

// Destination exprs
type Destination interface{ destination() }

func (*DestinationSeq) destination()       {}
func (*AccountLiteral) destination()       {}
func (*VariableLiteral) destination()      {}
func (*DestinationAllotment) destination() {}

type DestinationAllotmentValue interface{ destinationAllotmentValue() }

func (*RatioLiteral) destinationAllotmentValue()    {}
func (*VariableLiteral) destinationAllotmentValue() {}

type DestinationSeq struct {
	Range        Range
	Destinations []Destination
}

type DestinationAllotment struct {
	Range Range
	Items []DestinationAllotmentItem
}

type DestinationAllotmentItem struct {
	Range     Range
	Allotment DestinationAllotmentValue
	To        Destination
}

// Statements

type Statement interface{ statement() }

func (*SendStatement) statement() {}

type SendStatement struct {
	Range       Range
	Monetary    Literal
	Source      Source
	Destination Destination
}

type TypeDecl struct {
	Range Range
	Name  string
}

type VarDeclaration struct {
	Range Range
	Name  *VariableLiteral
	Type  *TypeDecl
}

type Program struct {
	Vars       []VarDeclaration
	Statements []Statement
}
