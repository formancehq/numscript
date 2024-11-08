package parser

import (
	"math/big"
)

// does it even make sense to have a literal supertype?
type Literal interface {
	Ranged
	literal()
}

func (*AssetLiteral) literal()    {}
func (*MonetaryLiteral) literal() {}
func (*AccountLiteral) literal()  {}
func (*VariableLiteral) literal() {}
func (*RatioLiteral) literal()    {}
func (*NumberLiteral) literal()   {}
func (*StringLiteral) literal()   {}

type (
	AssetLiteral struct {
		Range
		Asset string
	}

	NumberLiteral struct {
		Range
		Number int
	}

	StringLiteral struct {
		Range
		String string
	}

	MonetaryLiteral struct {
		Range
		Asset  Literal
		Amount Literal
	}

	AccountLiteral struct {
		Range
		Name string
	}

	RatioLiteral struct {
		Range
		Numerator   *big.Int
		Denominator *big.Int
	}

	VariableLiteral struct {
		Range
		Name string
	}
)

func (r RatioLiteral) ToRatio() *big.Rat {
	return new(big.Rat).SetFrac(r.Numerator, r.Denominator)
}

type RemainingAllotment struct {
	Range
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
func (*SourceAccount) source()   {}
func (*SourceCapped) source()    {}
func (*SourceOverdraft) source() {}

type (
	SourceAccount struct {
		Literal
	}

	SourceInorder struct {
		Range
		Sources []Source
	}
	SourceAllotment struct {
		Range
		Items []SourceAllotmentItem
	}

	SourceCapped struct {
		Range
		From Source
		Cap  Literal
	}

	SourceOverdraft struct {
		Range
		Address Literal
		Bounded *Literal
	}
)

type AllotmentValue interface{ allotmentValue() }

func (*RemainingAllotment) allotmentValue() {}
func (*RatioLiteral) allotmentValue()       {}
func (*VariableLiteral) allotmentValue()    {}

type SourceAllotmentItem struct {
	Range
	Allotment AllotmentValue
	From      Source
}

// Destination exprs
type Destination interface {
	destination()
	Ranged
}

func (*DestinationInorder) destination()   {}
func (*DestinationAccount) destination()   {}
func (*DestinationAllotment) destination() {}

type (
	DestinationAccount struct {
		Literal
	}

	DestinationInorder struct {
		Range
		Clauses   []DestinationInorderClause
		Remaining KeptOrDestination
	}

	DestinationInorderClause struct {
		Range
		Cap Literal
		To  KeptOrDestination
	}
)

type KeptOrDestination interface {
	keptOrDestination()
}
type DestinationKept struct {
	Range
}
type DestinationTo struct {
	Destination Destination
}

func (*DestinationKept) keptOrDestination() {}
func (*DestinationTo) keptOrDestination()   {}

type DestinationAllotment struct {
	Range
	Items []DestinationAllotmentItem
}

type DestinationAllotmentItem struct {
	Range
	Allotment AllotmentValue
	To        KeptOrDestination
}

// Statements

type Statement interface {
	statement()
	Ranged
}

func (*FnCall) statement()        {}
func (*SendStatement) statement() {}

type FnCallIdentifier struct {
	Range
	Name string
}

type FnCall struct {
	Range
	Caller *FnCallIdentifier
	Args   []Literal
}

type SentValue interface{ sentValue() }
type SentValueLiteral struct {
	Monetary Literal
}
type SentValueAll struct {
	Range
	Asset Literal
}

func (*SentValueLiteral) sentValue() {}
func (*SentValueAll) sentValue()     {}

type SendStatement struct {
	Range
	SentValue   SentValue
	Source      Source
	Destination Destination
}

type TypeDecl struct {
	Range
	Name string
}

type VarDeclaration struct {
	Range
	Name   *VariableLiteral
	Type   *TypeDecl
	Origin *FnCall
}

type Program struct {
	Vars       []VarDeclaration
	Statements []Statement
}
