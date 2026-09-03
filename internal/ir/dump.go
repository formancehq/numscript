package ir

import (
	"fmt"
	"strings"
)

func (r Reg) String() string   { return fmt.Sprintf("$r%d", uint(r)) }
func (l Label) String() string { return fmt.Sprintf("#%s", string(l)) }

func (OpAddInt) String() string           { return "add_int" }
func (OpSubInt) String() string           { return "sub_int" }
func (OpAddString) String() string        { return "add_string" }
func (OpStrEq) String() string            { return "str_eq" }
func (OpLtInt) String() string            { return "lt_int" }
func (OpEqInt) String() string            { return "eq_int" }
func (OpLtPortion) String() string        { return "lt_portion" }
func (OpEqPortion) String() string        { return "eq_portion" }
func (OpAddPortion) String() string       { return "add_portion" }
func (OpSubPortion) String() string       { return "sub_portion" }
func (OpMulPortion) String() string       { return "mul_portion" }
func (OpMakePortion) String() string      { return "mk_portion" }
func (OpMonetaryToString) String() string { return "monetary_to_string" }

func (OpIntCopy) String() string         { return "int_copy" }
func (OpPortionCopy) String() string     { return "portion_copy" }
func (OpStrCopy) String() string         { return "str_copy" }
func (OpBoolCopy) String() string        { return "bool_copy" }
func (OpNegInt) String() string          { return "neg_int" }
func (OpIntToString) String() string     { return "int_to_string" }
func (OpIsZero) String() string          { return "is_zero" }
func (OpNot) String() string             { return "not" }
func (OpPortionToString) String() string { return "portion_to_string" }
func (OpIntToPortion) String() string    { return "int_to_portion" }
func (OpPortionToInt) String() string    { return "portion_to_int" }

func (i PullAccount) String() string {
	opts := joinOpts(
		optLabel("cap", i.Cap),
		optLabel("overdraft", i.Overdraft),
		optLabel("color", i.Color),
	)
	s := fmt.Sprintf("%s = pull_account(account: %s", i.Dest, i.Account)
	if opts != "" {
		s += ", " + opts
	}
	return s + ")"
}

func (i SendToAccount) String() string {
	opts := joinOpts(optLabel("account", i.Account), optLabel("cap", i.Cap))
	return fmt.Sprintf("send_to_account(%s)", opts)
}

func (i CheckEnoughFunds) String() string {
	return fmt.Sprintf("check_enough_funds(%s, %s)", i.Got, i.Needed)
}

func (i Save) String() string {
	if i.Amount == nil {
		return fmt.Sprintf("save(account: %s, asset: %s)", i.Account, i.Asset)
	}
	return fmt.Sprintf("save(account: %s, asset: %s, amount: %s)", i.Account, i.Asset, *i.Amount)
}

func (i AssertLeftover) String() string {
	if i.Exact {
		return fmt.Sprintf("assert_leftover_exact(%s)", i.Portion)
	}
	return fmt.Sprintf("assert_leftover(%s)", i.Portion)
}

func (i SetCurrentAsset) String() string {
	return fmt.Sprintf("set_current_asset(%s)", i.Asset)
}

func (i AssertSameAsset) String() string {
	return fmt.Sprintf("assert_same_asset(%s, %s)", i.Left, i.Right)
}

func (i AssertValidAccount) String() string {
	return fmt.Sprintf("assert_valid_account(%s)", i.Account)
}

func (i AssertValidColor) String() string {
	return fmt.Sprintf("assert_valid_color(%s)", i.Color)
}

func (i AssertNonNegativeBalance) String() string {
	return fmt.Sprintf("assert_non_negative_balance(%s, %s)", i.Balance, i.Account)
}

func (i AssertNonNegativeAmount) String() string {
	return fmt.Sprintf("assert_non_negative_amount(%s)", i.Amount)
}

func (i SetTxMeta) String() string {
	return fmt.Sprintf("set_tx_meta(%s, %s)", i.Key, i.Value)
}

func (i SetAccountMeta) String() string {
	return fmt.Sprintf("set_account_meta(%s, %s, %s)", i.Account, i.Key, i.Value)
}

func (i MetaVar) String() string {
	return fmt.Sprintf("%s = meta<%s>(%s, %s)", i.Dest, i.Typ, i.Account, i.Key)
}

func (i MetaMonetary) String() string {
	return fmt.Sprintf("[%s, %s] = meta_monetary(%s, %s)", i.DestAsset, i.DestAmount, i.Account, i.Key)
}

func (MetaStr) String() string     { return "str" }
func (MetaInt) String() string     { return "int" }
func (MetaPortion) String() string { return "portion" }

func (i FetchBalance) String() string {
	return fmt.Sprintf("%s = balance(%s, %s)", i.Dest, i.Account, i.Asset)
}

func (i LoadVar) String() string {
	return fmt.Sprintf("%s = load_var<%s>(%d)", i.Dest, i.Typ, i.Index)
}

func (VarInt) String() string { return "int" }
func (VarStr) String() string { return "str" }

func (i JmpIfFalse) String() string {
	return fmt.Sprintf("jmp_if_false(%s, %s)", i.Cond, i.Target)
}

func (i JmpIfTrue) String() string {
	return fmt.Sprintf("jmp_if_true(%s, %s)", i.Cond, i.Target)
}

func (i Jmp) String() string {
	return fmt.Sprintf("jmp(%s)", i.Target)
}

func (i LoadInt) String() string {
	return fmt.Sprintf("%s = %s", i.Dest, &i.Value)
}

func (i LoadStr) String() string {
	return fmt.Sprintf("%s = %q", i.Dest, i.Value)
}

func (i ConstBool) String() string {
	return fmt.Sprintf("%s = %t", i.Dest, i.Value)
}

// infixAlias returns the infix spelling of an op, for the two that have one.
func infixAlias(k BinKind) (string, bool) {
	switch k.(type) {
	case OpAddInt:
		return "+", true
	case OpSubInt:
		return "-", true
	default:
		return "", false
	}
}

func (i BinaryOp) String() string {
	alias, hasAlias := infixAlias(i.Op)
	switch {
	case !hasAlias:
		return fmt.Sprintf("%s = %s(%s, %s)", i.Dest, i.Op, i.Left, i.Right)
	case i.Dest == i.Left:
		// e.g. $acc += $Reg
		return fmt.Sprintf("%s %s= %s", i.Dest, alias, i.Right)
	default:
		// e.g. $tot = $l + $r
		return fmt.Sprintf("%s = %s %s %s", i.Dest, i.Left, alias, i.Right)
	}
}

func (i UnaryOp) String() string {
	return fmt.Sprintf("%s = %s(%s)", i.Dest, i.Op, i.Arg)
}

func (i LabelMarker) String() string { return i.Label.String() }

func (i MarkPush) String() string { return "mark_push()" }

func (i MarkEnd) String() string {
	if i.Rewind {
		return "mark_rewind()"
	}
	return "mark_commit()"
}

// Dump renders a program: labels flush-left, instructions indented.
func Dump(code []Instr) string {
	var b strings.Builder
	for _, in := range code {
		if _, ok := in.(LabelMarker); ok {
			fmt.Fprintf(&b, "%s\n", in)
		} else {
			fmt.Fprintf(&b, "  %s\n", in)
		}
	}
	return b.String()
}

func optLabel(name string, r *Reg) string {
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
