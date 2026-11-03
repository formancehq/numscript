package prop

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// Known is the set of accounts (bare names) and assets a script references.
// A nil map disables that check.
type Known struct {
	Accounts map[string]struct{}
	Assets   map[string]struct{}
}

type sort_ int

const (
	sortInt sort_ = iota
	sortBool
)

type checker struct {
	known Known
	errs  []Error
}

// Typecheck reports every type error of a formula, which must be boolean.
func Typecheck(e Expr, known Known) []Error {
	c := &checker{known: known}
	c.want(e, sortBool)
	return c.errs
}

func (c *checker) err(e Expr, msg string) {
	c.errs = append(c.errs, Error{Range: e.GetRange(), Msg: msg})
}

func (c *checker) want(e Expr, s sort_) {
	if c.infer(e) == s {
		return
	}
	if s == sortInt {
		c.err(e, "expected an integer expression, found a boolean")
	} else {
		c.err(e, "expected a boolean expression, found an integer")
	}
}

func (c *checker) infer(e Expr) sort_ {
	switch e := e.(type) {
	case *IntLit:
		return sortInt
	case *BoolLit, *Success, *Fails:
		return sortBool
	case *AtomExpr:
		c.checkAccount(e, e.Atom.Account)
		if e.Atom.Kind == Moved {
			c.checkAccount(e, e.Atom.Account2)
		}
		c.checkAsset(e, e.Atom.Asset)
		return sortInt
	case *Delta:
		c.checkAccount(e, e.Account)
		c.checkAsset(e, e.Asset)
		return sortInt
	case *Unary:
		if e.Op == Neg {
			c.want(e.X, sortInt)
			return sortInt
		}
		c.want(e.X, sortBool)
		return sortBool
	case *Binary:
		switch e.Op {
		case Add, Sub:
			c.want(e.L, sortInt)
			c.want(e.R, sortInt)
			return sortInt
		case Mul:
			c.want(e.L, sortInt)
			c.want(e.R, sortInt)
			_, lok := constInt(e.L)
			_, rok := constInt(e.R)
			if !lok && !rok {
				c.err(e, "multiplication must have a literal operand (linear arithmetic only)")
			}
			return sortInt
		case Div, Mod:
			c.want(e.L, sortInt)
			c.want(e.R, sortInt)
			if d, ok := constInt(e.R); !ok || d.Sign() <= 0 {
				if e.Op == Div {
					c.err(e.R, "division must be by a positive literal")
				} else {
					c.err(e.R, "modulo must be by a positive literal")
				}
			}
			return sortInt
		case Eq, Ne, Lt, Le, Gt, Ge:
			c.want(e.L, sortInt)
			c.want(e.R, sortInt)
			return sortBool
		default:
			c.want(e.L, sortBool)
			c.want(e.R, sortBool)
			return sortBool
		}
	case *MinMax:
		for _, a := range e.Args {
			c.want(a, sortInt)
		}
		return sortInt
	}
	panic(fmt.Sprintf("prop.Typecheck: unknown node %T", e))
}

func (c *checker) checkAccount(e Expr, account string) {
	if c.known.Accounts == nil {
		return
	}
	if _, ok := c.known.Accounts[account]; ok {
		return
	}
	names := make([]string, 0, len(c.known.Accounts))
	for a := range c.known.Accounts {
		names = append(names, "@"+a)
	}
	c.err(e, fmt.Sprintf("account @%s is not referenced by the script (known: %s)", account, knownList(names)))
}

func (c *checker) checkAsset(e Expr, asset string) {
	if c.known.Assets == nil {
		return
	}
	if _, ok := c.known.Assets[asset]; ok {
		return
	}
	names := make([]string, 0, len(c.known.Assets))
	for a := range c.known.Assets {
		names = append(names, a)
	}
	c.err(e, fmt.Sprintf("asset %s is not referenced by the script (known: %s)", asset, knownList(names)))
}

func knownList(names []string) string {
	if len(names) == 0 {
		return "none"
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// constInt is the value of an atom-free integer expression. A division by a
// non-positive constant has no value.
func constInt(e Expr) (*big.Int, bool) {
	switch e := e.(type) {
	case *IntLit:
		return e.Value, true
	case *Unary:
		if e.Op != Neg {
			return nil, false
		}
		x, ok := constInt(e.X)
		if !ok {
			return nil, false
		}
		return new(big.Int).Neg(x), true
	case *Binary:
		l, ok := constInt(e.L)
		if !ok {
			return nil, false
		}
		r, ok := constInt(e.R)
		if !ok {
			return nil, false
		}
		switch e.Op {
		case Add:
			return new(big.Int).Add(l, r), true
		case Sub:
			return new(big.Int).Sub(l, r), true
		case Mul:
			return new(big.Int).Mul(l, r), true
		case Div:
			if r.Sign() <= 0 {
				return nil, false
			}
			return new(big.Int).Div(l, r), true
		case Mod:
			if r.Sign() <= 0 {
				return nil, false
			}
			return new(big.Int).Mod(l, r), true
		}
		return nil, false
	case *MinMax:
		var acc *big.Int
		for _, a := range e.Args {
			v, ok := constInt(a)
			if !ok {
				return nil, false
			}
			if acc == nil || (e.IsMax && v.Cmp(acc) > 0) || (!e.IsMax && v.Cmp(acc) < 0) {
				acc = v
			}
		}
		return acc, true
	}
	return nil, false
}

// Atoms lists the distinct atoms of a formula, sorted by their text. A delta
// contributes its Before and After atoms.
func Atoms(e Expr) []Atom {
	set := map[Atom]struct{}{}
	var walk func(Expr)
	walk = func(e Expr) {
		switch e := e.(type) {
		case *AtomExpr:
			set[e.Atom] = struct{}{}
		case *Delta:
			set[Atom{Kind: Before, Account: e.Account, Asset: e.Asset}] = struct{}{}
			set[Atom{Kind: After, Account: e.Account, Asset: e.Asset}] = struct{}{}
		case *Unary:
			walk(e.X)
		case *Binary:
			walk(e.L)
			walk(e.R)
		case *MinMax:
			for _, a := range e.Args {
				walk(a)
			}
		}
	}
	walk(e)
	out := make([]Atom, 0, len(set))
	for a := range set {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}
