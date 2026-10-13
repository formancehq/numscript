package prop

import (
	"fmt"

	"github.com/formancehq/numscript/internal/prove/smt"
)

// Lower builds the Bool term of a typechecked formula. resolve gives the Int
// term of an atom; success is the Bool term of the run outcome.
func Lower(e Expr, resolve func(Atom) smt.Term, success smt.Term) smt.Term {
	l := lowerer{resolve: resolve, success: success}
	return l.lower(e)
}

type lowerer struct {
	resolve func(Atom) smt.Term
	success smt.Term
}

func (l lowerer) lower(e Expr) smt.Term {
	switch e := e.(type) {
	case *IntLit:
		return smt.Int(e.Value)
	case *BoolLit:
		return smt.Bool(e.Value)
	case *Success:
		return l.success
	case *Fails:
		return smt.Not(l.success)
	case *AtomExpr:
		return l.resolve(e.Atom)
	case *Delta:
		after := l.resolve(Atom{Kind: After, Account: e.Account, Asset: e.Asset})
		before := l.resolve(Atom{Kind: Before, Account: e.Account, Asset: e.Asset})
		return smt.Sub(after, before)
	case *Unary:
		if e.Op == Neg {
			return smt.Neg(l.lower(e.X))
		}
		return smt.Not(l.lower(e.X))
	case *MinMax:
		t := l.lower(e.Args[0])
		for _, a := range e.Args[1:] {
			if e.IsMax {
				t = smt.Max(t, l.lower(a))
			} else {
				t = smt.Min(t, l.lower(a))
			}
		}
		return t
	case *Binary:
		a, b := l.lower(e.L), l.lower(e.R)
		switch e.Op {
		case Add:
			return smt.Add(a, b)
		case Sub:
			return smt.Sub(a, b)
		case Mul:
			if c, ok := smt.AsInt(a); ok {
				return smt.MulConst(c, b)
			}
			if c, ok := smt.AsInt(b); ok {
				return smt.MulConst(c, a)
			}
			panic("prop.Lower: multiplication without a literal operand")
		case Div:
			d, ok := smt.AsInt(b)
			if !ok {
				panic("prop.Lower: division by a non-literal")
			}
			return smt.DivConst(a, d)
		case Mod:
			d, ok := smt.AsInt(b)
			if !ok {
				panic("prop.Lower: modulo by a non-literal")
			}
			return smt.ModConst(a, d)
		case Eq:
			return smt.Eq(a, b)
		case Ne:
			return smt.Ne(a, b)
		case Lt:
			return smt.Lt(a, b)
		case Le:
			return smt.Le(a, b)
		case Gt:
			return smt.Gt(a, b)
		case Ge:
			return smt.Ge(a, b)
		case And:
			return smt.And(a, b)
		case Or:
			return smt.Or(a, b)
		case Implies:
			return smt.Implies(a, b)
		case Iff:
			return smt.Iff(a, b)
		}
	}
	panic(fmt.Sprintf("prop.Lower: unknown node %T", e))
}
