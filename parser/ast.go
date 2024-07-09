package parser

type Position struct {
	Character int
	Line      int
}

type Range struct {
	Start Position
	End   Position
}

// Literals

type Literal interface{ literal() }

func (*MonetaryLiteral) literal() {}
func (*AccountLiteral) literal()  {}
func (*VariableLiteral) literal() {}

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

// Source exprs

type Source interface{ source() }

func (*AccountLiteral) source()  {}
func (*VariableLiteral) source() {}

// Statements

type Statement interface{ statement() }

func (*SendStatement) statement() {}

type SendStatement struct {
	Range    Range
	Monetary Literal
	Source   Source
}

type Program struct {
	Statements []Statement
}
