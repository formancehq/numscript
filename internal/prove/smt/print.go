package smt

import (
	"strings"
)

// Print renders a term in SMT-LIB2 syntax. Negative literals print as (- n)
// because a bare "-5" is not a numeral in the standard.
func Print(t Term) string {
	var sb strings.Builder
	printTo(&sb, t)
	return sb.String()
}

func printTo(sb *strings.Builder, t Term) {
	switch t := t.(type) {
	case intLit:
		if t.v.Sign() < 0 {
			sb.WriteString("(- ")
			sb.WriteString(t.v.String()[1:])
			sb.WriteString(")")
		} else {
			sb.WriteString(t.v.String())
		}
	case boolLit:
		if t.v {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case symbol:
		sb.WriteString(Quote(t.name))
	case app:
		sb.WriteString("(")
		sb.WriteString(t.op)
		for _, a := range t.args {
			sb.WriteString(" ")
			printTo(sb, a)
		}
		sb.WriteString(")")
	default:
		panic("smt.Print: unknown term")
	}
}

// Quote wraps name in |..| unless it is a simple SMT-LIB symbol.
func Quote(name string) string {
	if isSimpleSymbol(name) {
		return name
	}
	return "|" + name + "|"
}

func isSimpleSymbol(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
			if i == 0 {
				return false
			}
		case strings.ContainsRune("~!@$%^&*_-+=<>.?/", r):
		default:
			return false
		}
	}
	return true
}
