package compiler

import "github.com/formancehq/numscript/internal/ir"

// monetaryValue is a monetary-typed expression after codegen. The VM has no
// monetary register, so a monetary travels as two: the asset in a string
// register and the amount in an int register.
//
// Named fields rather than a returned (asset, amount) pair so that transposing
// the two is a build error instead of an ir.Typecheck error.
type monetaryValue struct {
	Asset  ir.Reg // str
	Amount ir.Reg // int
}

// value is a compiled expression of any type. Mon is set exactly for
// monetary-typed expressions, Reg for every other type.
type value struct {
	Reg ir.Reg
	Mon *monetaryValue
}

func scalarValue(r ir.Reg) value { return value{Reg: r} }

func monValue(m monetaryValue) value { return value{Mon: &m} }
