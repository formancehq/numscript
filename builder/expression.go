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
