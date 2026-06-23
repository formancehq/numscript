package builder

type Source render

func SrcAccount(expr Expression[ExprTypeAccount]) Source {
	return Source(expr)
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

// SrcAllowingUnboundedOverdraft wraps a source (typically SrcAccount or
// SrcColored) and appends the `allowing unbounded overdraft` clause, per the
// Numscript grammar rule `srcAccountUnboundedOverdraft`
// (address colorConstraint? ALLOWING UNBOUNDED OVERDRAFT). It lets a non-world
// source go negative — required, for instance, when minting a colour the
// account does not yet hold.
func SrcAllowingUnboundedOverdraft(source Source) Source {
	return func(env *env, w int) {
		source(env, w)
		env.builder.WriteString(" allowing unbounded overdraft")
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
