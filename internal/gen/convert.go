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

// ToBuilderScript converts a full generated Script (vars declarations,
// seed-funding statements, the core send-only program, and extra non-send
// statements) into a flat list of builder statements, ready for
// builder.BuildProgram.
func ToBuilderScript(s Script) []builder.Statement {
	varExprs := make([]builder.Expression[builder.ExprTypeMonetary], len(s.Vars))
	for i, v := range s.Vars {
		account := builder.UnsafeAccount(v.Account)
		varExprs[i] = builder.NewMonetaryVarFromBalance(account, builder.ExprAsset(v.Asset))
	}

	out := make([]builder.Statement, 0, len(s.Seeds)+len(s.Program)+len(s.Extra))
	out = append(out, ToBuilder(s.Seeds)...)
	out = append(out, ToBuilder(s.Program)...)
	for _, e := range s.Extra {
		out = append(out, toBuilderExtra(e, varExprs))
	}
	return out
}

func toBuilderExtra(e ExtraStatement, varExprs []builder.Expression[builder.ExprTypeMonetary]) builder.Statement {
	switch e.Kind {
	case ExtraSave:
		var mon builder.Expression[builder.ExprTypeMonetary]
		if e.VarIdx != nil {
			mon = varExprs[*e.VarIdx]
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
			varExprs[*e.VarIdx],
			builder.SrcAccountOverdraft(builder.UnsafeAccount(e.Account), builder.UnboundedOverdraft()),
			builder.DestAccount(builder.UnsafeAccount(e.Destination)),
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
