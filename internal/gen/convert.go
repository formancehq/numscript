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
	return builder.ExprMonetary(builder.ExprAsset(m.Asset), builder.ExprNumberBigInt(m.Amount))
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
