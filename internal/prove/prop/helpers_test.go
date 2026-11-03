package prop

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// tree is the fully parenthesized form of a formula, which exposes its
// structure regardless of Dump's parenthesization.
func tree(e Expr) string {
	switch e := e.(type) {
	case *IntLit:
		return e.Value.String()
	case *BoolLit:
		return fmt.Sprint(e.Value)
	case *Success:
		return "success"
	case *Fails:
		return "fails"
	case *AtomExpr:
		return e.Atom.String()
	case *Delta:
		return fmt.Sprintf("delta(@%s, %s)", e.Account, e.Asset)
	case *Unary:
		if e.Op == Neg {
			return "(- " + tree(e.X) + ")"
		}
		return "(not " + tree(e.X) + ")"
	case *Binary:
		return "(" + e.Op.String() + " " + tree(e.L) + " " + tree(e.R) + ")"
	case *MinMax:
		parts := make([]string, len(e.Args))
		for i, a := range e.Args {
			parts[i] = tree(a)
		}
		name := "min"
		if e.IsMax {
			name = "max"
		}
		return "(" + name + " " + strings.Join(parts, " ") + ")"
	}
	panic(fmt.Sprintf("tree: unknown node %T", e))
}

// sameExpr is structural equality ignoring ranges.
func sameExpr(a, b Expr) bool {
	switch a := a.(type) {
	case *IntLit:
		b, ok := b.(*IntLit)
		return ok && a.Value.Cmp(b.Value) == 0
	case *BoolLit:
		b, ok := b.(*BoolLit)
		return ok && a.Value == b.Value
	case *Success:
		_, ok := b.(*Success)
		return ok
	case *Fails:
		_, ok := b.(*Fails)
		return ok
	case *AtomExpr:
		b, ok := b.(*AtomExpr)
		return ok && a.Atom == b.Atom
	case *Delta:
		b, ok := b.(*Delta)
		return ok && a.Account == b.Account && a.Asset == b.Asset
	case *Unary:
		b, ok := b.(*Unary)
		return ok && a.Op == b.Op && sameExpr(a.X, b.X)
	case *Binary:
		b, ok := b.(*Binary)
		return ok && a.Op == b.Op && sameExpr(a.L, b.L) && sameExpr(a.R, b.R)
	case *MinMax:
		b, ok := b.(*MinMax)
		if !ok || a.IsMax != b.IsMax || len(a.Args) != len(b.Args) {
			return false
		}
		for i := range a.Args {
			if !sameExpr(a.Args[i], b.Args[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func mustParse(t *testing.T, src string) Expr {
	t.Helper()
	p, errs := ParseInline(src)
	require.Empty(t, errs, "parsing %q", src)
	return p.Formula
}

func errStrings(errs []Error) []string {
	if len(errs) == 0 {
		return nil
	}
	out := make([]string, len(errs))
	for i, e := range errs {
		out[i] = e.Error()
	}
	return out
}
