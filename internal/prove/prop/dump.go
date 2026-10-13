package prop

import (
	"fmt"
	"strings"
)

// Dump renders a formula in the DSL with the fewest parentheses that keep its
// structure. Parsing the output gives back the same tree.
func Dump(e Expr) string {
	var sb strings.Builder
	dumpTo(&sb, e)
	return sb.String()
}

func prec(e Expr) int {
	switch e := e.(type) {
	case *Unary:
		if e.Op == Neg {
			return precUnary
		}
		return precNot
	case *Binary:
		return binaryOps[e.Op].prec
	}
	return precAtom
}

func dumpChild(sb *strings.Builder, e Expr, parens bool) {
	if parens {
		sb.WriteString("(")
		dumpTo(sb, e)
		sb.WriteString(")")
		return
	}
	dumpTo(sb, e)
}

func dumpTo(sb *strings.Builder, e Expr) {
	switch e := e.(type) {
	case *IntLit:
		sb.WriteString(e.Value.String())
	case *BoolLit:
		if e.Value {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case *Success:
		sb.WriteString("success")
	case *Fails:
		sb.WriteString("fails")
	case *AtomExpr:
		sb.WriteString(e.Atom.String())
	case *Delta:
		fmt.Fprintf(sb, "delta(@%s, %s)", e.Account, e.Asset)
	case *MinMax:
		if e.IsMax {
			sb.WriteString("max(")
		} else {
			sb.WriteString("min(")
		}
		for i, a := range e.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			dumpChild(sb, a, prec(a) < precSum)
		}
		sb.WriteString(")")
	case *Unary:
		if e.Op == Neg {
			sb.WriteString("-")
			dumpChild(sb, e.X, prec(e.X) < precUnary)
		} else {
			sb.WriteString("not ")
			dumpChild(sb, e.X, prec(e.X) < precNot)
		}
	case *Binary:
		info := binaryOps[e.Op]
		lp, rp := prec(e.L), prec(e.R)
		dumpChild(sb, e.L, lp < info.prec || (lp == info.prec && info.assoc != assocLeft))
		sb.WriteString(" " + e.Op.String() + " ")
		dumpChild(sb, e.R, rp < info.prec || (rp == info.prec && info.assoc != assocRight))
	default:
		panic(fmt.Sprintf("prop.Dump: unknown node %T", e))
	}
}
