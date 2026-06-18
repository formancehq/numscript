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
