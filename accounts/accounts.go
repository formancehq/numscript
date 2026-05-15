package accounts

import "github.com/formancehq/numscript/internal/interpreter"

type (
	InvolvedAccountExpr = interpreter.InvolvedAccountExpr
	InvolvedAccount     = interpreter.InvolvedAccount
	InvolvedMeta        = interpreter.InvolvedMeta

	AssetLiteral   = interpreter.AssetLiteral
	AccountLiteral = interpreter.AccountLiteral
	MakeMonetary   = interpreter.MakeMonetary
	NumberLiteral  = interpreter.NumberLiteral
	StringLiteral  = interpreter.StringLiteral
	Add            = interpreter.Add
	ConcatAccount  = interpreter.ConcatAccount
	Sub            = interpreter.Sub
	Div            = interpreter.Div
	SubPrefix      = interpreter.SubPrefix
	FnMeta         = interpreter.FnMeta
	GetAmount      = interpreter.GetAmount
	GetAsset       = interpreter.GetAsset
	GetBalance     = interpreter.GetBalance
	GetOverdraft   = interpreter.GetOverdraft
)

var IsValidCall = interpreter.IsValidCall
