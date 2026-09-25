package builder

type Source render

type Overdraft render

func UnboundedOverdraft() Overdraft {
	return func(env *env, w int) {
		env.builder.WriteString(" allowing unbounded overdraft")
	}
}

func BoundedOverdraft(amt Expression[ExprTypeMonetary]) Overdraft {
	return func(env *env, w int) {
		env.builder.WriteString(" allowing overdraft up to ")
		amt(env, w)
	}
}

func SrcAccount(expr Expression[ExprTypeAccount]) Source {
	return Source(expr)
}

func SrcAccountOverdraft(
	expr Expression[ExprTypeAccount],
	overdraft Overdraft,
) Source {
	return func(env *env, w int) {
		SrcAccount(expr)(env, w)
		overdraft(env, w)
	}
}

func SrcColored(
	accountExpr Expression[ExprTypeAccount],
	colorExpr Expression[ExprTypeString],
) Source {
	return func(env *env, w int) {
		accountExpr(env, w)
		env.builder.WriteString(" \\ ")
		colorExpr(env, w)
	}
}

func SrcColoredOverdraft(
	accountExpr Expression[ExprTypeAccount],
	colorExpr Expression[ExprTypeString],
	overdraft Overdraft,
) Source {
	return func(env *env, w int) {
		SrcColored(accountExpr, colorExpr)(env, w)
		overdraft(env, w)
	}
}

func SrcInorder(sources ...Source) Source {
	return func(env *env, w int) {
		env.builder.WriteString("{\n")
		for _, src := range sources {
			writeIndentation(env, w+1)
			src(env, w+1)
			env.builder.WriteByte('\n')
		}
		writeIndentation(env, w)
		env.builder.WriteByte('}')
	}
}

// A source capped by a maximum amount: `max <amount> from <source>`
func SrcCapped(max Expression[ExprTypeMonetary], source Source) Source {
	return func(env *env, w int) {
		env.builder.WriteString("max ")
		max(env, w)
		env.builder.WriteString(" from ")
		source(env, w)
	}
}

func SrcAllotment(clauses ...AllotmentClause[Source]) Source {
	return srcAllotment(clauses, nil)
}

// SrcAllotmentWithRemaining is SrcAllotment with a trailing
// `remaining from <source>` clause.
//
// The engines constrain the clause differently. The legacy machine's compiler
// rejects it when the known portions already sum to 100%, and rejects a
// portion var in an allotment that lacks it; the interpreter runs the first
// (the clause receives zero) and checks the second's sum at run time. Past
// 100%, the machine errors at run time while the interpreter computes a
// negative remaining portion instead of rejecting (oracle/DIVERGENCES.md #6).
// internal/gen therefore keeps generated sums strictly below 100%.
func SrcAllotmentWithRemaining(clauses []AllotmentClause[Source], remaining Source) Source {
	return srcAllotment(clauses, remaining)
}

func srcAllotment(clauses []AllotmentClause[Source], remaining Source) Source {
	return func(env *env, w int) {
		env.builder.WriteString("{\n")
		for _, clause := range clauses {
			writeIndentation(env, w+1)
			clause.Portion(env, w)
			env.builder.WriteString(" from ")
			clause.Payload(env, w+1)
			env.builder.WriteByte('\n')
		}
		if remaining != nil {
			writeIndentation(env, w+1)
			env.builder.WriteString("remaining from ")
			remaining(env, w+1)
			env.builder.WriteByte('\n')
		}
		writeIndentation(env, w)
		env.builder.WriteByte('}')
	}
}
