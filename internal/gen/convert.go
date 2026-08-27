package gen

import (
	"math/big"

	"github.com/formancehq/numscript/builder"
)

// ToBuilder converts a (post-cleanup) generated program into builder
// statements, ready to be passed to builder.BuildProgram. Accounts are
// emitted via builder.UnsafeAccount rather than the pooled ExprAccount: the
// account pool here is a small fixed set ("world", "acc0".."accN") reused
// constantly across a program, and pooling all of them into $vars would
// just add noise to every generated script without changing behavior (both
// engines are run with the same vars binding map regardless).
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
		return builder.StmtSendAll(builder.ExprAsset(s.Asset), src, dest)
	}
	return builder.StmtSend(toBuilderMonetary(s.Amount), src, dest)
}

func toBuilderMonetary(m Monetary) builder.Expression[builder.ExprTypeMonetary] {
	if m.Amount.Sign() < 0 {
		// A bracketed monetary literal's amount slot (`[ASSET N]`) only ever
		// accepts a bare non-negative number in both grammars — there's no
		// legal way to write `[ASSET -7]` directly. `[ASSET 0] - [ASSET 7]`
		// is the only legal way to reach a negative monetary value
		// (verified directly against the oracle); see
		// builder.ExprMonetarySub.
		abs := new(big.Int).Neg(m.Amount)
		zero := builder.ExprMonetary(builder.ExprAsset(m.Asset), builder.ExprNumberBigInt(big.NewInt(0)))
		magnitude := builder.ExprMonetary(builder.ExprAsset(m.Asset), builder.ExprNumberBigInt(abs))
		return builder.ExprMonetarySub(zero, magnitude)
	}
	return builder.ExprMonetary(builder.ExprAsset(m.Asset), builder.ExprNumberBigInt(m.Amount))
}

func toBuilderNumExpr(e NumExpr) builder.Expression[builder.ExprTypeNumber] {
	switch e.Kind {
	case NumLit:
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
}

func toBuilderVarExprs(vars []VarDecl) varExprs {
	ve := varExprs{
		monetary: make([]builder.Expression[builder.ExprTypeMonetary], len(vars)),
		number:   make([]builder.Expression[builder.ExprTypeNumber], len(vars)),
	}
	for i, v := range vars {
		account := builder.UnsafeAccount(v.Account)
		switch v.Kind {
		case VarFromBalance:
			ve.monetary[i] = builder.NewMonetaryVarFromBalance(account, builder.ExprAsset(v.Asset))
		case VarFromMeta:
			ve.number[i] = builder.NewNumberVarFromMeta(account, v.Key)
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
		return builder.StmtSave(mon, builder.UnsafeAccount(e.Account))

	case ExtraSaveAll:
		return builder.StmtSaveAll(builder.ExprAsset(e.Asset), builder.UnsafeAccount(e.Account))

	case ExtraSetTxMeta:
		return builder.StmtSetTxMeta(e.Key, toBuilderNumExpr(e.Value))

	case ExtraSetAccountMeta:
		return builder.StmtSetAccountMeta(builder.UnsafeAccount(e.Account), e.Key, toBuilderNumExpr(e.Value))

	case ExtraSendVar:
		return builder.StmtSend(
			ve.monetary[*e.VarIdx],
			builder.SrcAccountOverdraft(builder.UnsafeAccount(e.Account), builder.UnboundedOverdraft()),
			builder.DestAccount(builder.UnsafeAccount(e.Destination)),
		)

	case ExtraSetTxMetaVar:
		return builder.StmtSetTxMeta(e.Key, ve.number[*e.VarIdx])

	case ExtraSendFromAccountVar:
		return builder.StmtSend(
			toBuilderMonetary(*e.Monetary),
			builder.SrcAccount(accountVarExprs[*e.AccountVarIdx]),
			builder.DestAccount(builder.UnsafeAccount(e.Account)),
		)

	case ExtraSendToAccountVar:
		return builder.StmtSend(
			toBuilderMonetary(*e.Monetary),
			builder.SrcAccount(builder.UnsafeAccount(e.Account)),
			builder.DestAccount(accountVarExprs[*e.AccountVarIdx]),
		)

	default:
		panic("gen: unknown extra statement kind")
	}
}

func toBuilderPortion(r *big.Rat) builder.Portion {
	return builder.NewPortion(new(big.Int).Set(r.Num()), new(big.Int).Set(r.Denom()))
}

func toBuilderSource(s Source) builder.Source {
	switch s.Kind {
	case SrcAccount:
		return builder.SrcAccount(builder.UnsafeAccount(s.Account))

	case SrcAccountOverdraft:
		if s.Overdraft == nil {
			return builder.SrcAccountOverdraft(builder.UnsafeAccount(s.Account), builder.UnboundedOverdraft())
		}
		return builder.SrcAccountOverdraft(
			builder.UnsafeAccount(s.Account),
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
		return builder.DestAccount(builder.UnsafeAccount(d.Account))

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
