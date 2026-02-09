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
func (*PercentageLiteral) valueExpr()    {}
func (*NumberLiteral) valueExpr()        {}
func (*StringLiteral) valueExpr()        {}
func (*BinaryInfix) valueExpr()          {}
func (*Prefix) valueExpr()               {}
func (*FnCall) valueExpr()               {}

type PrefixOperator string

const (
	PrefixOperatorMinus PrefixOperator = "-"
)

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
		Number *big.Int
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

	PercentageLiteral struct {
		Range
		Amount         *big.Int
		FloatingDigits uint16
	}

	Variable struct {
		Range
		Name string
	}

	Prefix struct {
		Range
		Operator PrefixOperator
		Expr     ValueExpr
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

func (r PercentageLiteral) ToRatio() *big.Rat {
	denom := new(big.Int).Exp(
		big.NewInt(10),
		big.NewInt(int64(2+r.FloatingDigits)),
		nil,
	)

	return new(big.Rat).SetFrac(r.Amount, denom)
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
func (*SourceOneof) source()     {}
func (*SourceAllotment) source() {}
func (*SourceAccount) source()   {}
func (*SourceCapped) source()    {}
func (*SourceOverdraft) source() {}

type (
	SourceAccount struct {
		Color ValueExpr
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
		Color   ValueExpr
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
	Origin *ValueExpr
}

type VarDeclarations struct {
	Range
	Declarations []VarDeclaration
}

type Comment struct {
	Range
	Content string
}

type Program struct {
	Vars       *VarDeclarations
	Statements []Statement
	Comments   []Comment
}
