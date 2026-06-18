package builder

import "math/big"

type var_[T ExprType] struct {
	name  string
	set   bool
	alloc func(env *env) string
}

type Var[T ExprType] var_[ExprType]

func NewAccountVar() Var[ExprTypeAccount] {
	return Var[ExprTypeAccount]{
		alloc: func(env *env) string {
			id := env.accountsPool.getFreshId()
			return accountToName(id)
		},
	}
}

func NewAssetVar() Var[ExprTypeAsset] {
	return Var[ExprTypeAsset]{
		alloc: func(env *env) string {
			id := env.assetsPool.getFreshId()
			return assetToName(id)
		},
	}
}

func NewStringVar() Var[ExprTypeString] {
	return Var[ExprTypeString]{
		alloc: func(env *env) string {
			id := env.stringsPool.getFreshId()
			return stringToName(id)
		},
	}
}

func NewNumberVar() Var[ExprTypeNumber] {
	return Var[ExprTypeNumber]{
		alloc: func(env *env) string {
			id := env.numbersPool.getFreshId()
			return numberToName(id)
		},
	}
}

func (v Var[ExprTypeAccount]) FillAccount(account string) (string, string) {
	return v.name, account
}

func (v Var[ExprTypeAsset]) FillAsset(asset string) (string, string) {
	return v.name, asset
}

func (v Var[ExprTypeString]) FillString(str string) (string, string) {
	return v.name, str
}

func (v Var[ExprTypeNumber]) FillNumber(bi *big.Int) (string, string) {
	return v.name, bi.String()
}
