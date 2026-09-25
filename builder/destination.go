package builder

type Destination render

func DestAccount(expr Expression[ExprTypeAccount]) Destination {
	return Destination(expr)
}

// Either `kept` or `to <destination>`, as used in inorder/allotment clauses.
type KeptOrDest render

func Kept() KeptOrDest {
	return func(env *env, w int) {
		env.builder.WriteString("kept")
	}
}

func To(dest Destination) KeptOrDest {
	return func(env *env, w int) {
		env.builder.WriteString("to ")
		dest(env, w)
	}
}

// One `max <amount> <keptOrDest>` clause of a DestInorder block.
type DestInorderClause struct {
	Max  Expression[ExprTypeMonetary]
	Dest KeptOrDest
}

// A destination inorder block. Per the numscript grammar, the `max` clauses
// are mandatory and a trailing `remaining` clause is required.
func DestInorder(clauses []DestInorderClause, remaining KeptOrDest) Destination {
	return func(env *env, w int) {
		env.builder.WriteString("{\n")
		for _, clause := range clauses {
			writeIndentation(env, w+1)
			env.builder.WriteString("max ")
			clause.Max(env, w+1)
			env.builder.WriteString(" ")
			clause.Dest(env, w+1)
			env.builder.WriteByte('\n')
		}
		writeIndentation(env, w+1)
		env.builder.WriteString("remaining ")
		remaining(env, w+1)
		env.builder.WriteByte('\n')
		writeIndentation(env, w)
		env.builder.WriteByte('}')
	}
}

func DestAllotment(clauses ...AllotmentClause[KeptOrDest]) Destination {
	return destAllotment(clauses, nil)
}

// DestAllotmentWithRemaining is DestAllotment with a trailing
// `remaining <keptOrDest>` clause. The clause portions must then sum to less
// than 100% (both grammars reject an allotment whose known portions already
// reach it), and it is the only allotment form in which a portion var
// compiles.
func DestAllotmentWithRemaining(clauses []AllotmentClause[KeptOrDest], remaining KeptOrDest) Destination {
	return destAllotment(clauses, remaining)
}

func destAllotment(clauses []AllotmentClause[KeptOrDest], remaining KeptOrDest) Destination {
	return func(env *env, w int) {
		env.builder.WriteString("{\n")
		for _, clause := range clauses {
			writeIndentation(env, w+1)
			clause.Portion(env, w)
			env.builder.WriteString(" ")
			clause.Payload(env, w+1)
			env.builder.WriteByte('\n')
		}
		if remaining != nil {
			writeIndentation(env, w+1)
			env.builder.WriteString("remaining ")
			remaining(env, w+1)
			env.builder.WriteByte('\n')
		}
		writeIndentation(env, w)
		env.builder.WriteByte('}')
	}
}
