package prop

import (
	"fmt"
	"math/big"
)

// Eval evaluates a typechecked formula. lookup gives the value of an atom.
// Division and modulo are Euclidean, like big.Int.Div and Mod.
func Eval(e Expr, lookup func(Atom) *big.Int, success bool) bool {
	ev := evaluator{lookup: lookup, success: success}
	return ev.boolean(e)
}

type evaluator struct {
	lookup  func(Atom) *big.Int
	success bool
}

func (ev evaluator) boolean(e Expr) bool {
	switch e := e.(type) {
	case *BoolLit:
		return e.Value
	case *Success:
		return ev.success
	case *Fails:
		return !ev.success
	case *Unary:
		return !ev.boolean(e.X)
	case *Binary:
		switch e.Op {
		case And:
			return ev.boolean(e.L) && ev.boolean(e.R)
		case Or:
			return ev.boolean(e.L) || ev.boolean(e.R)
		case Implies:
			return !ev.boolean(e.L) || ev.boolean(e.R)
		case Iff:
			return ev.boolean(e.L) == ev.boolean(e.R)
		}
		c := ev.integer(e.L).Cmp(ev.integer(e.R))
		switch e.Op {
		case Eq:
			return c == 0
		case Ne:
			return c != 0
		case Lt:
			return c < 0
		case Le:
			return c <= 0
		case Gt:
			return c > 0
		case Ge:
			return c >= 0
		}
	}
	panic(fmt.Sprintf("prop.Eval: %T is not a boolean", e))
}

func (ev evaluator) integer(e Expr) *big.Int {
	switch e := e.(type) {
	case *IntLit:
		return e.Value
	case *AtomExpr:
		return ev.lookup(e.Atom)
	case *Delta:
		after := ev.lookup(Atom{Kind: After, Account: e.Account, Asset: e.Asset})
		before := ev.lookup(Atom{Kind: Before, Account: e.Account, Asset: e.Asset})
		return new(big.Int).Sub(after, before)
	case *Unary:
		return new(big.Int).Neg(ev.integer(e.X))
	case *MinMax:
		acc := ev.integer(e.Args[0])
		for _, a := range e.Args[1:] {
			v := ev.integer(a)
			if (e.IsMax && v.Cmp(acc) > 0) || (!e.IsMax && v.Cmp(acc) < 0) {
				acc = v
			}
		}
		return acc
	case *Binary:
		l, r := ev.integer(e.L), ev.integer(e.R)
		switch e.Op {
		case Add:
			return new(big.Int).Add(l, r)
		case Sub:
			return new(big.Int).Sub(l, r)
		case Mul:
			return new(big.Int).Mul(l, r)
		case Div:
			return new(big.Int).Div(l, r)
		case Mod:
			return new(big.Int).Mod(l, r)
		}
	}
	panic(fmt.Sprintf("prop.Eval: %T is not an integer", e))
}
