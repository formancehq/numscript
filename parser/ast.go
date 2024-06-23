package parser

type Position struct {
	Character int
	Line      int
}

type Range struct {
	Start Position
	End   Position
}

type Literal interface{ literal() }

func (*MonetaryLiteral) literal() {}

type MonetaryLiteral struct {
	Range  Range
	Asset  string
	Amount int
}

type Statement interface{ statement() }

func (*SendStatement) statement() {}

type SendStatement struct {
	Range    Range
	Monetary Literal
}

type Program struct {
	Statements []Statement
}
