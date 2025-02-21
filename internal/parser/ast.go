package parser

import (
	"math/big"
)

type ValueExpr interface {
	Ranged
	valueExpr()
}

func (*Variable) valueExpr()          {}
func (*AssetLiteral) valueExpr()      {}
func (*MonetaryLiteral) valueExpr()   {}
func (*AccountLiteral) valueExpr()    {}
func (*PercentageLiteral) valueExpr() {}
func (*NumberLiteral) valueExpr()     {}
func (*StringLiteral) valueExpr()     {}
func (*BinaryInfix) valueExpr()       {}

type InfixOperator string

const (
	InfixOperatorPlus  InfixOperator = "+"
	InfixOperatorMinus InfixOperator = "-"
	InfixOperatorDiv   InfixOperator = "/"
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

	AccountLiteral struct {
		Range
		Name string
	}

	PercentageLiteral struct {
		Range
		Amount         *big.Int
		FloatingDigits uint16
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

func (r PercentageLiteral) ToRatio() *big.Rat {
	denom := new(big.Int).Exp(
		big.NewInt(10),
		big.NewInt(int64(2+r.FloatingDigits)),
		nil,
	)

	return new(big.Rat).SetFrac(r.Amount, denom)
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
func (*SourceOneof) source()     {}
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

	SourceOneof struct {
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
func (*ValueExprAllotment) allotmentValue() {}

type ValueExprAllotment struct {
	Value ValueExpr
}

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
func (*DestinationOneof) destination()     {}
func (*DestinationAllotment) destination() {}

type (
	DestinationAccount struct {
		ValueExpr
	}

	DestinationInorder struct {
		Range
		Clauses   []CappedKeptOrDestination
		Remaining KeptOrDestination
	}

	DestinationOneof struct {
		Range
		Clauses   []CappedKeptOrDestination
		Remaining KeptOrDestination
	}

	CappedKeptOrDestination struct {
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
