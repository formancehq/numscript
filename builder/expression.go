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

// UnsafeAsset emits literal directly as an asset reference, bypassing the
// vars pool — the inline counterpart to ExprAsset, and the asset-typed
// analogue of UnsafeAccount (same caveat: the caller owns validity).
func UnsafeAsset(literal string) Expression[ExprTypeAsset] {
	return func(env *env, w int) {
		env.builder.WriteString(literal)
	}
}

func ExprString(name string) Expression[ExprTypeString] {
	return func(env *env, w int) {
		id := env.stringsPool.getItemId(name)
		env.builder.WriteByte('$')
		env.builder.WriteString(stringToName(id))
	}
}

// ExprMonetaryVar routes a monetary through the vars pool, rendering
// `$monetary_N` bound to binding (e.g. "COIN 100"). The inline counterpart is
// ExprMonetary, which writes a `[<asset> <amount>]` literal.
//
// Note the amount slot of an inline monetary literal cannot itself be a
// variable in either grammar (`[COIN $n]` is rejected), so a var-form monetary
// has to replace the whole literal rather than just its amount.
func ExprMonetaryVar(binding string) Expression[ExprTypeMonetary] {
	return func(env *env, w int) {
		id := env.monetariesPool.getItemId(binding)
		env.builder.WriteByte('$')
		env.builder.WriteString(monetaryToName(id))
	}
}

// ExprPortionVar routes a portion through the vars pool, rendering
// `$portion_N` bound to binding (e.g. "1/2"). ExprPortion is the inline form.
func ExprPortionVar(binding string) Expression[ExprTypePortion] {
	return func(env *env, w int) {
		id := env.portionsPool.getItemId(binding)
		env.builder.WriteByte('$')
		env.builder.WriteString(portionToName(id))
	}
}

// ExprPortion renders p inline, as a `num/denom` literal.
func ExprPortion(p Portion) Expression[ExprTypePortion] {
	return func(env *env, w int) {
		p.render(env)
	}
}

func ExprNumberBigInt(amount *big.Int) Expression[ExprTypeNumber] {
	// we don't risk injection with numbers so we can just pprint them right away
	return func(env *env, w int) {
		env.builder.WriteString(amount.String())
	}
}

// ExprNumberVar routes a number through the vars pool, rendering `$number_N`.
// ExprNumberBigInt is the inline form.
//
// Note a number var is not legal everywhere a number literal is: the amount
// slot of a monetary literal (`[COIN $n]`) is rejected by both grammars, so
// this is for value positions such as a meta value.
func ExprNumberVar(amount *big.Int) Expression[ExprTypeNumber] {
	return func(env *env, w int) {
		id := env.numbersPool.getItemId(amount)
		env.builder.WriteByte('$')
		env.builder.WriteString(numberToName(id))
	}
}

// ExprStringLit renders s inline, as a quoted string literal. ExprString is
// the var form.
func ExprStringLit(s string) Expression[ExprTypeString] {
	return func(env *env, w int) {
		writeStringLiteral(env, s)
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
