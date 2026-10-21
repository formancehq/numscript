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

// accountValue is an account-typed expression after codegen. Scope is a second
// register alongside the name, nil when the expression is provably unscoped (an
// account literal, a plain var, or any account not produced by scoped()) — the
// same nilable-operand idiom PullAccount already uses for Color/Overdraft.
type accountValue struct {
	Name  ir.Reg  // str
	Scope *ir.Reg // str
}

// value is a compiled expression of any type. Mon is set exactly for
// monetary-typed expressions, Acc for account-typed ones, Reg for every other
// type.
type value struct {
	Reg ir.Reg
	Mon *monetaryValue
	Acc *accountValue
}

func scalarValue(r ir.Reg) value { return value{Reg: r} }

func monValue(m monetaryValue) value { return value{Mon: &m} }

func accValue(a accountValue) value { return value{Acc: &a} }
