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
