package parser

import (
	"math/big"
	"strings"
)

type ValueExpr interface {
	Ranged
	valueExpr()
}

func (*Variable) valueExpr()             {}
func (*AssetLiteral) valueExpr()         {}
func (*MonetaryLiteral) valueExpr()      {}
func (*AccountInterpLiteral) valueExpr() {}
func (*RatioLiteral) valueExpr()         {}
func (*NumberLiteral) valueExpr()        {}
func (*StringLiteral) valueExpr()        {}
func (*BinaryInfix) valueExpr()          {}

type InfixOperator string

const (
	InfixOperatorPlus  InfixOperator = "+"
	InfixOperatorMinus InfixOperator = "-"
)

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
		Asset  ValueExpr
		Amount ValueExpr
	}

	AccountInterpLiteral struct {
		Range
		Parts []AccountNamePart
	}

	RatioLiteral struct {
		Range
		Numerator   *big.Int
		Denominator *big.Int
	}

	Variable struct {
		Range
		Name string
	}

	BinaryInfix struct {
		Range
		Operator InfixOperator
		Left     ValueExpr
		Right    ValueExpr
	}
)

type AccountNamePart interface{ accountNamePart() }
type AccountTextPart struct{ Name string }

func (AccountTextPart) accountNamePart() {}
func (*Variable) accountNamePart()       {}

func (r RatioLiteral) ToRatio() *big.Rat {
	return new(big.Rat).SetFrac(r.Numerator, r.Denominator)
}

func (a AccountInterpLiteral) IsWorld() bool {
	if len(a.Parts) != 1 {
		return false
	}
	switch part := a.Parts[0].(type) {
	case AccountTextPart:
		return part.Name == "world"

	default:
		return false
	}
}

func (expr AccountInterpLiteral) String() string {
	// TODO we might want to parse this instead of computing it
	var parts []string
	for _, part := range expr.Parts {
		switch part := part.(type) {
		case AccountTextPart:
			parts = append(parts, part.Name)
		case *Variable:
			parts = append(parts, "$"+part.Name)
		}
	}
	return strings.Join(parts, ":")
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
		ValueExpr
	}

	SourceInorder struct {
		Range
		Sources []Source
	}
	SourceAllotment struct {
		Range
		Items []SourceAllotmentItem
	}

	SourceAllotmentItem struct {
		Range
		Allotment AllotmentValue
		From      Source
	}

	SourceCapped struct {
		Range
		From Source
		Cap  ValueExpr
	}

	SourceOverdraft struct {
		Range
		Address ValueExpr
		Bounded *ValueExpr
	}
)

type AllotmentValue interface{ allotmentValue() }

func (*RemainingAllotment) allotmentValue() {}
func (*RatioLiteral) allotmentValue()       {}
func (*Variable) allotmentValue()           {}

type RemainingAllotment struct {
	Range
}

// Destination exprs
type Destination interface {
	destination()
	Ranged
}

func (*DestinationAccount) destination()   {}
func (*DestinationInorder) destination()   {}
func (*DestinationAllotment) destination() {}

type (
	DestinationAccount struct {
		ValueExpr
	}

	DestinationInorder struct {
		Range
		Clauses   []DestinationInorderClause
		Remaining KeptOrDestination
	}

	DestinationInorderClause struct {
		Range
		Cap ValueExpr
		To  KeptOrDestination
	}

	DestinationAllotment struct {
		Range
		Items []DestinationAllotmentItem
	}

	DestinationAllotmentItem struct {
		Range
		Allotment AllotmentValue
		To        KeptOrDestination
	}
)

type KeptOrDestination interface {
	keptOrDestination()
}

func (*DestinationKept) keptOrDestination() {}
func (*DestinationTo) keptOrDestination()   {}

type (
	DestinationKept struct {
		Range
	}

	DestinationTo struct {
		Destination Destination
	}
)

// Statements

type Statement interface {
	statement()
	Ranged
}

func (*FnCall) statement()        {}
func (*SendStatement) statement() {}
func (*SaveStatement) statement() {}

type FnCallIdentifier struct {
	Range
	Name string
}

type FnCall struct {
	Range
	Caller *FnCallIdentifier
	Args   []ValueExpr
}

type SentValue interface {
	sentValue()
	Ranged
}

type SentValueLiteral struct {
	Range
	Monetary ValueExpr
}
type SentValueAll struct {
	Range
	Asset ValueExpr
}

func (*SentValueLiteral) sentValue() {}
func (*SentValueAll) sentValue()     {}

type SendStatement struct {
	Range
	SentValue   SentValue
	Source      Source
	Destination Destination
}

type SaveStatement struct {
	Range
	SentValue SentValue
	Amount    ValueExpr
}

type TypeDecl struct {
	Range
	Name string
}

type VarDeclaration struct {
	Range
	Name   *Variable
	Type   *TypeDecl
	Origin *FnCall
}

type Program struct {
	Vars       []VarDeclaration
	Statements []Statement
}
