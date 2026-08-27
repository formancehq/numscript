package builder

import "math/big"

type Expression[T ExprType] render

func ExprVar[T ExprType](v *Var[T]) Expression[T] {
	return func(env *env, w int) {
		varName, hasPreviousLookup := env.varsEnv.bindings[v]
		if !hasPreviousLookup {
			varName = v.alloc(env)
			env.varsEnv.bindings[v] = varName
		}
		env.builder.WriteByte('$')
		env.builder.WriteString(varName)
	}
}

func ExprAccount(name string) Expression[ExprTypeAccount] {
	return func(env *env, w int) {
		id := env.accountsPool.getItemId(name)
		env.builder.WriteByte('$')
		env.builder.WriteString(accountToName(id))
	}
}

// UnsafeAccount emits literal directly as an account reference (`@<literal>`),
// bypassing the vars pool entirely. Unlike ExprAccount, literal is written
// straight into the script text rather than passed through the vars binding
// map, so the caller is responsible for ensuring it's a syntactically valid
// account address (and, if it's not fully trusted, that it can't be used to
// inject additional script content).
func UnsafeAccount(literal string) Expression[ExprTypeAccount] {
	return func(env *env, w int) {
		env.builder.WriteByte('@')
		env.builder.WriteString(literal)
	}
}

func ExprAsset(name string) Expression[ExprTypeAsset] {
	return func(env *env, w int) {
		id := env.assetsPool.getItemId(name)
		env.builder.WriteByte('$')
		env.builder.WriteString(assetToName(id))
	}
}

func ExprString(name string) Expression[ExprTypeString] {
	return func(env *env, w int) {
		id := env.stringsPool.getItemId(name)
		env.builder.WriteByte('$')
		env.builder.WriteString(stringToName(id))
	}
}

func ExprNumberBigInt(amount *big.Int) Expression[ExprTypeNumber] {
	// we don't risk injection with numbers so we can just pprint them right away
	return func(env *env, w int) {
		env.builder.WriteString(amount.String())
	}
}

// World is a convenience helper for `@world` — equivalent to
// UnsafeAccount("world").
func World() Expression[ExprTypeAccount] {
	return UnsafeAccount("world")
}

// ExprAdd renders `<a> + <b>`. Numscript only allows +/- between numbers at
// runtime (no other arithmetic operators, no parens, no unary minus).
func ExprAdd(a, b Expression[ExprTypeNumber]) Expression[ExprTypeNumber] {
	return func(env *env, w int) {
		a(env, w)
		env.builder.WriteString(" + ")
		b(env, w)
	}
}

// ExprSub renders `<a> - <b>`.
func ExprSub(a, b Expression[ExprTypeNumber]) Expression[ExprTypeNumber] {
	return func(env *env, w int) {
		a(env, w)
		env.builder.WriteString(" - ")
		b(env, w)
	}
}

// ExprMonetarySub renders `<a> - <b>`, where a and b are monetary literals
// or expressions. A bracketed monetary literal's amount slot (`[ASSET N]`)
// only ever accepts a bare number — never a general expression — so this is
// the only legal way to build a negative-amount monetary: e.g.
// `[ASSET 0] - [ASSET 7]` evaluates to `[ASSET -7]` and is accepted by both
// engines' grammars (verified against the oracle directly), whereas
// `[ASSET -7]` or `[ASSET 0-7]` are not.
func ExprMonetarySub(a, b Expression[ExprTypeMonetary]) Expression[ExprTypeMonetary] {
	return func(env *env, w int) {
		a(env, w)
		env.builder.WriteString(" - ")
		b(env, w)
	}
}

func ExprMonetary(
	asset Expression[ExprTypeAsset],
	amount Expression[ExprTypeNumber],
) Expression[ExprTypeMonetary] {
	return func(env *env, w int) {
		env.builder.WriteString("[")
		asset(env, w)
		env.builder.WriteString(" ")
		amount(env, w)
		env.builder.WriteString("]")
	}
}
