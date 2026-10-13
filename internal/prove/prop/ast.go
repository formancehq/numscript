// Package prop is the property DSL of the prover: a boolean formula over the
// balances and flows of a script run. It parses, dumps, typechecks, lowers to
// smt terms and evaluates formulas.
package prop

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
)

type Error struct {
	Range parser.Range
	Msg   string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Range.Start.Line+1, e.Range.Start.Character+1, e.Msg)
}

type AtomKind int

const (
	Before AtomKind = iota
	After
	Sent
	Received
	Moved
)

var atomKindNames = [...]string{"before", "after", "sent", "received", "moved"}

func (k AtomKind) String() string { return atomKindNames[k] }

// Atom is a quantity of the run. Accounts are bare names, without the '@';
// Account2 is set for Moved only.
type Atom struct {
	Kind     AtomKind
	Account  string
	Account2 string
	Asset    string
}

func (a Atom) String() string {
	if a.Kind == Moved {
		return fmt.Sprintf("moved(@%s, @%s, %s)", a.Account, a.Account2, a.Asset)
	}
	return fmt.Sprintf("%s(@%s, %s)", a.Kind, a.Account, a.Asset)
}

type Expr interface {
	GetRange() parser.Range
	expr()
}

type UnaryOp int

const (
	Neg UnaryOp = iota
	Not
)

type BinaryOp int

const (
	Add BinaryOp = iota
	Sub
	Mul
	Div
	Mod
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
	And
	Or
	Implies
	Iff
)

var binaryOpNames = [...]string{"+", "-", "*", "/", "%", "==", "!=", "<", "<=", ">", ">=", "and", "or", "=>", "<=>"}

func (op BinaryOp) String() string { return binaryOpNames[op] }

type (
	IntLit struct {
		parser.Range
		Value *big.Int
	}
	BoolLit struct {
		parser.Range
		Value bool
	}
	Success  struct{ parser.Range }
	Fails    struct{ parser.Range }
	AtomExpr struct {
		parser.Range
		Atom Atom
	}
	Unary struct {
		parser.Range
		Op UnaryOp
		X  Expr
	}
	Binary struct {
		parser.Range
		Op   BinaryOp
		L, R Expr
	}
	MinMax struct {
		parser.Range
		IsMax bool
		Args  []Expr
	}
	// Delta is after - before of the same account and asset.
	Delta struct {
		parser.Range
		Account string
		Asset   string
	}
)

func (*IntLit) expr()   {}
func (*BoolLit) expr()  {}
func (*Success) expr()  {}
func (*Fails) expr()    {}
func (*AtomExpr) expr() {}
func (*Unary) expr()    {}
func (*Binary) expr()   {}
func (*MinMax) expr()   {}
func (*Delta) expr()    {}

type Property struct {
	Name    string
	Formula Expr
	Range   parser.Range
}
