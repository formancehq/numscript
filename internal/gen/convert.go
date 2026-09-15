package gen

import (
	"math/big"

	"github.com/formancehq/numscript/builder"
)

// ToBuilder converts a (post-cleanup) generated program into builder
// statements, ready to be passed to builder.BuildProgram. Each account and
// asset occurrence is rendered inline or through a var according to its
// AsVar flag — see toBuilderAccount.
func ToBuilder(p Program) []builder.Statement {
	out := make([]builder.Statement, len(p))
	for i, s := range p {
		out[i] = toBuilderStatement(s)
	}
	return out
}

func toBuilderStatement(s Statement) builder.Statement {
	src := toBuilderSource(s.Source)
	dest := toBuilderDestination(s.Destination)

	if s.IsSendAll {
		return builder.StmtSendAll(toBuilderAsset(s.Asset, s.AssetAsVar), src, dest)
	}
	return builder.StmtSend(toBuilderMonetary(s.Amount), src, dest)
}

// toBuilderAccount renders one account occurrence, either inline (`@addr`) or
// through a pooled var (`$accountN` bound to addr). builder.ExprAccount pools
// by address, so every var-form occurrence of an address shares one var while
// inline occurrences stay literal — letting a single script reference the same
// account both ways.
func toBuilderAccount(address string, asVar bool) builder.Expression[builder.ExprTypeAccount] {
	if asVar {
		return builder.ExprAccount(address)
	}
	return builder.UnsafeAccount(address)
}

// toBuilderString renders one string occurrence, var or inline.
func toBuilderString(s string, asVar bool) builder.Expression[builder.ExprTypeString] {
	if asVar {
		return builder.ExprString(s)
	}
	return builder.ExprStringLit(s)
}

// toBuilderAsset renders one asset occurrence, var or inline. Note the asset
// slot of a monetary literal accepts a var in both grammars (`[$assetN 10]`),
// unlike the amount slot — see toBuilderMonetary.
func toBuilderAsset(asset string, asVar bool) builder.Expression[builder.ExprTypeAsset] {
	if asVar {
		return builder.ExprAsset(asset)
	}
	return builder.UnsafeAsset(asset)
}

func toBuilderMonetary(m Monetary) builder.Expression[builder.ExprTypeMonetary] {
	if m.AsVar && m.Amount.Sign() >= 0 {
		return builder.ExprMonetaryVar(m.Asset + " " + m.Amount.String())
	}
	if m.Amount.Sign() < 0 {
		// A bracketed monetary literal's amount slot (`[ASSET N]`) only ever
		// accepts a bare non-negative number in both grammars — there's no
		// legal way to write `[ASSET -7]` directly. `[ASSET 0] - [ASSET 7]`
		// is the only legal way to reach a negative monetary value
		// (verified directly against the oracle); see
		// builder.ExprMonetarySub.
		abs := new(big.Int).Neg(m.Amount)
		zero := builder.ExprMonetary(toBuilderAsset(m.Asset, m.AssetAsVar), builder.ExprNumberBigInt(big.NewInt(0)))
		magnitude := builder.ExprMonetary(toBuilderAsset(m.Asset, m.AssetAsVar), builder.ExprNumberBigInt(abs))
		return builder.ExprMonetarySub(zero, magnitude)
	}
	return builder.ExprMonetary(toBuilderAsset(m.Asset, m.AssetAsVar), builder.ExprNumberBigInt(m.Amount))
}

func toBuilderNumExpr(e NumExpr) builder.Expression[builder.ExprTypeNumber] {
	switch e.Kind {
	case NumLit:
		if e.LitAsVar {
			return builder.ExprNumberVar(e.Lit)
		}
		return builder.ExprNumberBigInt(e.Lit)
	case NumAdd:
		return builder.ExprAdd(toBuilderNumExpr(*e.Left), toBuilderNumExpr(*e.Right))
	case NumSub:
		return builder.ExprSub(toBuilderNumExpr(*e.Left), toBuilderNumExpr(*e.Right))
	default:
		panic("gen: unknown num expr kind")
	}
}

// varExprs holds the converted builder expression for each declared var,
// split by the builder type its Kind produces: a VarFromBalance decl is
// Expression[ExprTypeMonetary], a VarFromMeta decl is
// Expression[ExprTypeNumber]. Both are indexed by the same position in
// Script.Vars; only the slot matching a given VarDecl's Kind is populated.
type varExprs struct {
	monetary []builder.Expression[builder.ExprTypeMonetary]
	number   []builder.Expression[builder.ExprTypeNumber]
	// setTxMeta[i] emits `set_tx_meta(<key>, $var_i)` for a VarFromMeta decl.
	// A meta var can be declared as any of the six types, and Go cannot hold
	// those differently-typed Expression[T] values in one slice, so the
	// statement is built here — where the type is still known — and the
	// consumer only supplies the key.
	setTxMeta []func(key string) builder.Statement
}

func toBuilderVarExprs(vars []VarDecl) varExprs {
	ve := varExprs{
		monetary:  make([]builder.Expression[builder.ExprTypeMonetary], len(vars)),
		number:    make([]builder.Expression[builder.ExprTypeNumber], len(vars)),
		setTxMeta: make([]func(string) builder.Statement, len(vars)),
	}
	for i, v := range vars {
		account := toBuilderAccount(v.Account, v.AccountAsVar)
		switch v.Kind {
		case VarFromBalance:
			ve.monetary[i] = builder.NewMonetaryVarFromBalance(account, toBuilderAsset(v.Asset, v.AssetAsVar))
		case VarFromMeta:
			switch v.MetaType {
			case MetaString:
				x := builder.NewStringVarFromMeta(account, v.Key)
				ve.setTxMeta[i] = func(k string) builder.Statement { return builder.StmtSetTxMeta(k, x) }
			case MetaMonetary:
				x := builder.NewMonetaryVarFromMeta(account, v.Key)
				ve.setTxMeta[i] = func(k string) builder.Statement { return builder.StmtSetTxMeta(k, x) }
			case MetaAsset:
				x := builder.NewAssetVarFromMeta(account, v.Key)
				ve.setTxMeta[i] = func(k string) builder.Statement { return builder.StmtSetTxMeta(k, x) }
			case MetaAccount:
				x := builder.NewAccountVarFromMeta(account, v.Key)
				ve.setTxMeta[i] = func(k string) builder.Statement { return builder.StmtSetTxMeta(k, x) }
			case MetaPortion:
				x := builder.NewPortionVarFromMeta(account, v.Key)
				ve.setTxMeta[i] = func(k string) builder.Statement { return builder.StmtSetTxMeta(k, x) }
			default:
				x := builder.NewNumberVarFromMeta(account, v.Key)
				ve.number[i] = x
				ve.setTxMeta[i] = func(k string) builder.Statement { return builder.StmtSetTxMeta(k, x) }
			}
		default:
			panic("gen: unknown var decl kind")
		}
	}
	return ve
}

// AccountVarFill pairs a declared account-typed var with the literal value
// it must be bound to at run time — collected during ToBuilderScript and
// consumed by GenerateScript (api.go), which merges it into the vars
// bindings map builder.BuildProgram returns. Unlike a VarFromBalance/
// VarFromMeta var (whose value the compiler computes from an origin call),
// a plain account-typed var has no origin — its value can only come from
// the caller-supplied vars map, exactly as GenerateScript already does for
// the account/asset/etc. pools.
type AccountVarFill struct {
	Var   *builder.Var[builder.ExprTypeAccount]
	Value string
}

// ToBuilderScript converts a full generated Script (vars declarations,
// seed-funding statements, the core send-only program interleaved with
// extra non-send statements per Script.Order) into a flat list of builder
// statements, ready for builder.BuildProgram, plus the account-var runtime
// bindings the caller must additionally fill in (see AccountVarFill).
func ToBuilderScript(s Script) ([]builder.Statement, []AccountVarFill) {
	ve := toBuilderVarExprs(s.Vars)

	accountVars := make([]builder.Var[builder.ExprTypeAccount], len(s.AccountVars))
	accountVarExprs := make([]builder.Expression[builder.ExprTypeAccount], len(s.AccountVars))
	fills := make([]AccountVarFill, len(s.AccountVars))
	for i, v := range s.AccountVars {
		accountVars[i] = builder.NewAccountVar()
		accountVarExprs[i] = builder.ExprVar(&accountVars[i])
		fills[i] = AccountVarFill{Var: &accountVars[i], Value: v.Value}
	}

	out := make([]builder.Statement, 0, len(s.Seeds)+len(s.Program)+len(s.Extra))
	out = append(out, ToBuilder(s.Seeds)...)

	pi, ei := 0, 0
	for _, takeProgram := range s.Order {
		if takeProgram {
			out = append(out, toBuilderStatement(s.Program[pi]))
			pi++
		} else {
			out = append(out, toBuilderExtra(s.Extra[ei], ve, accountVarExprs))
			ei++
		}
	}

	return out, fills
}

func toBuilderExtra(e ExtraStatement, ve varExprs, accountVarExprs []builder.Expression[builder.ExprTypeAccount]) builder.Statement {
	switch e.Kind {
	case ExtraSave:
		var mon builder.Expression[builder.ExprTypeMonetary]
		if e.VarIdx != nil {
			mon = ve.monetary[*e.VarIdx]
		} else {
			mon = toBuilderMonetary(*e.Monetary)
		}
		return builder.StmtSave(mon, toBuilderAccount(e.Account, e.AccountAsVar))

	case ExtraSaveAll:
		return builder.StmtSaveAll(toBuilderAsset(e.Asset, e.AssetAsVar), toBuilderAccount(e.Account, e.AccountAsVar))

	case ExtraSetTxMeta:
		if e.StringValue != nil {
			return builder.StmtSetTxMeta(e.Key, toBuilderString(*e.StringValue, e.StringValueAsVar))
		}
		return builder.StmtSetTxMeta(e.Key, toBuilderNumExpr(e.Value))

	case ExtraSetAccountMeta:
		acct := toBuilderAccount(e.Account, e.AccountAsVar)
		if e.StringValue != nil {
			return builder.StmtSetAccountMeta(acct, e.Key, toBuilderString(*e.StringValue, e.StringValueAsVar))
		}
		return builder.StmtSetAccountMeta(acct, e.Key, toBuilderNumExpr(e.Value))

	case ExtraSendVar:
		return builder.StmtSend(
			ve.monetary[*e.VarIdx],
			builder.SrcAccountOverdraft(toBuilderAccount(e.Account, e.AccountAsVar), builder.UnboundedOverdraft()),
			builder.DestAccount(toBuilderAccount(e.Destination, e.DestinationAsVar)),
		)

	case ExtraSetTxMetaVar:
		return ve.setTxMeta[*e.VarIdx](e.Key)

	case ExtraSendFromAccountVar:
		return builder.StmtSend(
			toBuilderMonetary(*e.Monetary),
			builder.SrcAccount(accountVarExprs[*e.AccountVarIdx]),
			builder.DestAccount(toBuilderAccount(e.Account, e.AccountAsVar)),
		)

	case ExtraSendToAccountVar:
		return builder.StmtSend(
			toBuilderMonetary(*e.Monetary),
			builder.SrcAccount(toBuilderAccount(e.Account, e.AccountAsVar)),
			builder.DestAccount(accountVarExprs[*e.AccountVarIdx]),
		)

	default:
		panic("gen: unknown extra statement kind")
	}
}

// toBuilderPortion renders a portion inline, as an `n/d` literal.
//
// Portions are deliberately NOT routed through vars, even though
// builder.ExprPortionVar exists and both grammars accept a portion variable in
// allotment position. The compiler must prove an allotment sums to 100%, and it
// cannot see through a variable, so it rejects the script ("the sum of portions
// might be less than 100%") unless the block ends in a `remaining` clause —
// which this generator does not emit for allotments. Routing portions through
// vars therefore turned ~63% of scripts into oracle-side compile failures,
// which Compare tolerates, silently skipping the comparison. Supporting it
// needs `remaining` support in allotments first.
func toBuilderPortion(r *big.Rat) builder.Expression[builder.ExprTypePortion] {
	return builder.ExprPortion(builder.NewPortion(new(big.Int).Set(r.Num()), new(big.Int).Set(r.Denom())))
}

func toBuilderSource(s Source) builder.Source {
	switch s.Kind {
	case SrcAccount:
		return builder.SrcAccount(toBuilderAccount(s.Account, s.AccountAsVar))

	case SrcAccountOverdraft:
		if s.Overdraft == nil {
			return builder.SrcAccountOverdraft(toBuilderAccount(s.Account, s.AccountAsVar), builder.UnboundedOverdraft())
		}
		return builder.SrcAccountOverdraft(
			toBuilderAccount(s.Account, s.AccountAsVar),
			builder.BoundedOverdraft(toBuilderMonetary(*s.Overdraft)),
		)

	case SrcCapped:
		return builder.SrcCapped(toBuilderMonetary(*s.Cap), toBuilderSource(*s.Inner))

	case SrcInorder:
		sources := make([]builder.Source, len(s.Sources))
		for i, inner := range s.Sources {
			sources[i] = toBuilderSource(inner)
		}
		return builder.SrcInorder(sources...)

	case SrcAllotment:
		clauses := make([]builder.AllotmentClause[builder.Source], len(s.Clauses))
		for i, c := range s.Clauses {
			clauses[i] = builder.AllotmentClause[builder.Source]{
				Portion: toBuilderPortion(c.Portion),
				Payload: toBuilderSource(c.Source),
			}
		}
		return builder.SrcAllotment(clauses...)

	default:
		panic("gen: unknown source kind")
	}
}

func toBuilderDestination(d Destination) builder.Destination {
	switch d.Kind {
	case DestAccount:
		return builder.DestAccount(toBuilderAccount(d.Account, d.AccountAsVar))

	case DestInorder:
		clauses := make([]builder.DestInorderClause, len(d.InorderClauses))
		for i, c := range d.InorderClauses {
			clauses[i] = builder.DestInorderClause{
				Max:  toBuilderMonetary(c.Max),
				Dest: toBuilderKeptOrDest(c.KeptOrDest),
			}
		}
		return builder.DestInorder(clauses, toBuilderKeptOrDest(*d.Remaining))

	case DestAllotment:
		clauses := make([]builder.AllotmentClause[builder.KeptOrDest], len(d.AllotClauses))
		for i, c := range d.AllotClauses {
			clauses[i] = builder.AllotmentClause[builder.KeptOrDest]{
				Portion: toBuilderPortion(c.Portion),
				Payload: toBuilderKeptOrDest(c.KeptOrDest),
			}
		}
		return builder.DestAllotment(clauses...)

	default:
		panic("gen: unknown destination kind")
	}
}

func toBuilderKeptOrDest(k KeptOrDest) builder.KeptOrDest {
	if k.Kind == Kept {
		return builder.Kept()
	}
	return builder.To(toBuilderDestination(*k.Dest))
}
