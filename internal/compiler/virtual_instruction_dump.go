package compiler

import (
	"fmt"
	"strings"
)

func (r reg) String() string   { return fmt.Sprintf("$r%d", int(r)) }
func (l label) String() string { return fmt.Sprintf("#%s", string(l)) }

func (k binKind) String() string {
	switch k {
	case opMinInt:
		return "min_int"
	case opAddInt:
		return "add_int"
	case opSubInt:
		return "sub_int"
	case opSubPortion:
		return "sub_portion"
	case opMakePortion:
		return "mk_portion"
	case opMakeMonetary:
		return "mk_monetary"
	default:
		return "bin?"
	}
}

func (k unKind) String() string {
	switch k {
	case opIntCopy:
		return "int_copy"
	case opPortionCopy:
		return "portion_copy"
	case opGetAsset:
		return "get_asset"
	case opGetAmount:
		return "get_amount"
	default:
		return "un?"
	}
}

func (i pullAccount) String() string {
	opts := joinOpts(
		optLabel("cap", i.cap),
		optLabel("overdraft", i.overdraft),
		optLabel("color", i.color),
	)
	s := fmt.Sprintf("%s <- pull_account(account: %s", i.dest, i.account)
	if opts != "" {
		s += ", " + opts
	}
	return s + ")"
}

func (i sendToAccount) String() string {
	opts := joinOpts(optLabel("cap", i.cap), optLabel("color", i.color))
	if i.dest == nil {
		return fmt.Sprintf("kept(%s)", opts)
	}
	s := fmt.Sprintf("send_to_account(%s", *i.dest)
	if opts != "" {
		s += ", " + opts
	}
	return s + ")"
}

func (i makeAllotment) String() string {
	return fmt.Sprintf("[%s] <- mk_allot(%s, [%s])", regList(i.dest), i.amount, regList(i.portions))
}

func (i checkEnoughFunds) String() string {
	return fmt.Sprintf("check_enough_funds(%s, %s)", i.got, i.needed)
}

func (i setCurrentAsset) String() string {
	return fmt.Sprintf("set_current_asset(%s)", i.asset)
}

func (i checkEqCurrentAsset) String() string {
	return fmt.Sprintf("check_eq_current_asset(%s)", i.got)
}

func (i fetchVariable) String() string {
	return fmt.Sprintf("%s <- fetch_var(%d)", i.dest, i.index)
}

func (i jmpIfZero) String() string {
	return fmt.Sprintf("jmp_if_zero(%s, %s)", i.cond, i.target)
}

func (i loadInt) String() string {
	return fmt.Sprintf("%s <- load_const(%s)", i.dest, i.value)
}

func (i loadStr) String() string {
	return fmt.Sprintf("%s <- load_const(%q)", i.dest, i.value)
}

func (i binaryOp) String() string {
	return fmt.Sprintf("%s <- %s(%s, %s)", i.dest, i.op, i.left, i.right)
}

func (i unaryOp) String() string {
	return fmt.Sprintf("%s <- %s(%s)", i.dest, i.op, i.src)
}

func (i labelMarker) String() string { return i.label.String() }

// dump renders a program: labels flush-left, instructions indented.
func dump(code []vInstr) string {
	var b strings.Builder
	for _, in := range code {
		if _, ok := in.(labelMarker); ok {
			fmt.Fprintf(&b, "%s\n", in)
		} else {
			fmt.Fprintf(&b, "  %s\n", in)
		}
	}
	return b.String()
}

func optLabel(name string, r *reg) string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", name, *r)
}

func joinOpts(parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ", ")
}

func regList(regs []reg) string {
	parts := make([]string, len(regs))
	for i, r := range regs {
		parts[i] = r.String()
	}
	return strings.Join(parts, ", ")
}
