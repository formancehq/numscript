// Package smt is the SMT-LIB2 layer of the prover: a small term AST over the
// Int and Bool sorts whose constructors fold constants, a printer, a script
// builder, a reader for solver output, a subprocess driver and a concrete
// evaluator.
//
// Only linear integer arithmetic is constructible: multiplication, division and
// modulo take a literal operand, so every script stays in QF_LIA.
package smt

import (
	"math/big"
)

type Sort uint8

const (
	SortInt Sort = iota
	SortBool
)

func (s Sort) String() string {
	if s == SortBool {
		return "Bool"
	}
	return "Int"
}

type Term interface {
	Sort() Sort
	// nonNeg reports whether an Int term is known to be >= 0 without solving.
	nonNeg() bool
}

type intLit struct{ v *big.Int }
type boolLit struct{ v bool }
type symbol struct {
	name string
	sort Sort
}
type app struct {
	op   string
	args []Term
	sort Sort
	nn   bool
}

func (intLit) Sort() Sort     { return SortInt }
func (boolLit) Sort() Sort    { return SortBool }
func (s symbol) Sort() Sort   { return s.sort }
func (a app) Sort() Sort      { return a.sort }
func (l intLit) nonNeg() bool { return l.v.Sign() >= 0 }
func (boolLit) nonNeg() bool  { return false }
func (symbol) nonNeg() bool   { return false }
func (a app) nonNeg() bool    { return a.nn }

var (
	zero = intLit{big.NewInt(0)}
	one  = intLit{big.NewInt(1)}
)

func Int(v *big.Int) Term { return intLit{new(big.Int).Set(v)} }
func Int64(v int64) Term  { return intLit{big.NewInt(v)} }
func Bool(b bool) Term    { return boolLit{b} }
func Symbol(name string, sort Sort) Term {
	return symbol{name, sort}
}

// SymbolName returns the name of a symbol term.
func SymbolName(t Term) (string, bool) {
	s, ok := t.(symbol)
	return s.name, ok
}

func AsInt(t Term) (*big.Int, bool) {
	l, ok := t.(intLit)
	if !ok {
		return nil, false
	}
	return new(big.Int).Set(l.v), true
}

func AsBool(t Term) (bool, bool) {
	l, ok := t.(boolLit)
	return l.v, ok
}

func IsTrue(t Term) bool  { b, ok := AsBool(t); return ok && b }
func IsFalse(t Term) bool { b, ok := AsBool(t); return ok && !b }

// IsLeaf reports whether printing t costs nothing to repeat: a literal or a
// symbol. Script.Define aliases those instead of naming them.
func IsLeaf(t Term) bool {
	switch t.(type) {
	case intLit, boolLit, symbol:
		return true
	}
	return false
}

// same is a cheap structural equality: literals by value, symbols by name,
// anything else by identity. It is what the folds below need and nothing more.
func same(a, b Term) bool {
	switch a := a.(type) {
	case intLit:
		b, ok := b.(intLit)
		return ok && a.v.Cmp(b.v) == 0
	case boolLit:
		b, ok := b.(boolLit)
		return ok && a.v == b.v
	case symbol:
		b, ok := b.(symbol)
		return ok && a == b
	}
	return false
}

func lit(a, b Term) (*big.Int, *big.Int, bool) {
	x, okx := a.(intLit)
	y, oky := b.(intLit)
	if !okx || !oky {
		return nil, nil, false
	}
	return x.v, y.v, true
}

func Add(a, b Term) Term {
	if x, y, ok := lit(a, b); ok {
		return intLit{new(big.Int).Add(x, y)}
	}
	if same(a, zero) {
		return b
	}
	if same(b, zero) {
		return a
	}
	return app{"+", []Term{a, b}, SortInt, a.nonNeg() && b.nonNeg()}
}

func Sub(a, b Term) Term {
	if x, y, ok := lit(a, b); ok {
		return intLit{new(big.Int).Sub(x, y)}
	}
	if same(b, zero) {
		return a
	}
	if same(a, b) {
		return zero
	}
	return app{"-", []Term{a, b}, SortInt, false}
}

func Neg(a Term) Term {
	if x, ok := a.(intLit); ok {
		return intLit{new(big.Int).Neg(x.v)}
	}
	return app{"-", []Term{a}, SortInt, false}
}

// MulConst is c * a, with c a literal so the term stays linear.
func MulConst(c *big.Int, a Term) Term {
	if x, ok := a.(intLit); ok {
		return intLit{new(big.Int).Mul(c, x.v)}
	}
	switch {
	case c.Sign() == 0:
		return zero
	case c.Cmp(one.v) == 0:
		return a
	}
	return app{"*", []Term{intLit{new(big.Int).Set(c)}, a}, SortInt, c.Sign() > 0 && a.nonNeg()}
}

// DivConst is floor(a / d) for a literal d > 0 (SMT-LIB div is Euclidean,
// which is the floor for a positive divisor, like big.Int.Div).
func DivConst(a Term, d *big.Int) Term {
	if d.Sign() <= 0 {
		panic("smt.DivConst: divisor must be positive")
	}
	if x, ok := a.(intLit); ok {
		return intLit{new(big.Int).Div(x.v, d)}
	}
	if d.Cmp(one.v) == 0 {
		return a
	}
	return app{"div", []Term{a, intLit{new(big.Int).Set(d)}}, SortInt, a.nonNeg()}
}

// ModConst is a mod d for a literal d > 0 (Euclidean, always in [0, d)).
func ModConst(a Term, d *big.Int) Term {
	if d.Sign() <= 0 {
		panic("smt.ModConst: divisor must be positive")
	}
	if x, ok := a.(intLit); ok {
		return intLit{new(big.Int).Mod(x.v, d)}
	}
	if d.Cmp(one.v) == 0 {
		return zero
	}
	return app{"mod", []Term{a, intLit{new(big.Int).Set(d)}}, SortInt, true}
}

func Lt(a, b Term) Term {
	if x, y, ok := lit(a, b); ok {
		return boolLit{x.Cmp(y) < 0}
	}
	if same(a, b) {
		return boolLit{false}
	}
	return app{"<", []Term{a, b}, SortBool, false}
}

func Le(a, b Term) Term {
	if x, y, ok := lit(a, b); ok {
		return boolLit{x.Cmp(y) <= 0}
	}
	if same(a, b) {
		return boolLit{true}
	}
	return app{"<=", []Term{a, b}, SortBool, false}
}

func Gt(a, b Term) Term { return Lt(b, a) }
func Ge(a, b Term) Term { return Le(b, a) }

func Eq(a, b Term) Term {
	if a.Sort() == SortBool {
		return Iff(a, b)
	}
	if x, y, ok := lit(a, b); ok {
		return boolLit{x.Cmp(y) == 0}
	}
	if same(a, b) {
		return boolLit{true}
	}
	return app{"=", []Term{a, b}, SortBool, false}
}

func Ne(a, b Term) Term { return Not(Eq(a, b)) }

func Not(a Term) Term {
	if b, ok := a.(boolLit); ok {
		return boolLit{!b.v}
	}
	if n, ok := a.(app); ok && n.op == "not" {
		return n.args[0]
	}
	return app{"not", []Term{a}, SortBool, false}
}

func And(ts ...Term) Term {
	var args []Term
	for _, t := range ts {
		if b, ok := t.(boolLit); ok {
			if !b.v {
				return boolLit{false}
			}
			continue
		}
		if n, ok := t.(app); ok && n.op == "and" {
			args = append(args, n.args...)
			continue
		}
		args = append(args, t)
	}
	switch len(args) {
	case 0:
		return boolLit{true}
	case 1:
		return args[0]
	}
	return app{"and", args, SortBool, false}
}

func Or(ts ...Term) Term {
	var args []Term
	for _, t := range ts {
		if b, ok := t.(boolLit); ok {
			if b.v {
				return boolLit{true}
			}
			continue
		}
		if n, ok := t.(app); ok && n.op == "or" {
			args = append(args, n.args...)
			continue
		}
		args = append(args, t)
	}
	switch len(args) {
	case 0:
		return boolLit{false}
	case 1:
		return args[0]
	}
	return app{"or", args, SortBool, false}
}

func Implies(a, b Term) Term {
	if IsFalse(a) || IsTrue(b) {
		return boolLit{true}
	}
	if IsTrue(a) {
		return b
	}
	if IsFalse(b) {
		return Not(a)
	}
	return app{"=>", []Term{a, b}, SortBool, false}
}

func Iff(a, b Term) Term {
	if x, ok := a.(boolLit); ok {
		if x.v {
			return b
		}
		return Not(b)
	}
	if y, ok := b.(boolLit); ok {
		if y.v {
			return a
		}
		return Not(a)
	}
	if same(a, b) {
		return boolLit{true}
	}
	return app{"=", []Term{a, b}, SortBool, false}
}

func Ite(c, t, e Term) Term {
	if IsTrue(c) {
		return t
	}
	if IsFalse(c) {
		return e
	}
	if same(t, e) {
		return t
	}
	if t.Sort() == SortBool {
		if IsTrue(t) && IsFalse(e) {
			return c
		}
		if IsFalse(t) && IsTrue(e) {
			return Not(c)
		}
	}
	return app{"ite", []Term{c, t, e}, t.Sort(), t.nonNeg() && e.nonNeg()}
}

// Max0 is max(0, a).
func Max0(a Term) Term {
	if a.nonNeg() {
		return a
	}
	return app{"ite", []Term{Lt(a, zero), zero, a}, SortInt, true}
}

func Min(a, b Term) Term {
	if x, y, ok := lit(a, b); ok {
		if x.Cmp(y) <= 0 {
			return a
		}
		return b
	}
	if same(a, b) {
		return a
	}
	return app{"ite", []Term{Lt(a, b), a, b}, SortInt, a.nonNeg() && b.nonNeg()}
}

func Max(a, b Term) Term {
	if x, y, ok := lit(a, b); ok {
		if x.Cmp(y) >= 0 {
			return a
		}
		return b
	}
	if same(a, b) {
		return a
	}
	return app{"ite", []Term{Lt(a, b), b, a}, SortInt, a.nonNeg() || b.nonNeg()}
}
