package interpreter

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
)

type InvolvedAccountExpr interface{ involvedAccountExpr() }

type (
	AssetLiteral struct {
		Asset string
	}
	AccountLiteral struct {
		Account string
	}
	MakeMonetary struct {
		Asset  InvolvedAccountExpr
		Amount InvolvedAccountExpr
	}
	NumberLiteral struct {
		Amount *big.Int
	}
	StringLiteral struct {
		String string
	}
	Add struct {
		Left  InvolvedAccountExpr
		Right InvolvedAccountExpr
	}
	Sub struct {
		Left  InvolvedAccountExpr
		Right InvolvedAccountExpr
	}
	Div struct {
		Left  InvolvedAccountExpr
		Right InvolvedAccountExpr
	}
	SubPrefix struct {
		Expr InvolvedAccountExpr
	}
	FnMeta struct {
		ExpectedType string
		Account      InvolvedAccountExpr
		Key          InvolvedAccountExpr
	}
	GetAmount struct {
		Monetary InvolvedAccountExpr
	}
	GetAsset struct {
		Monetary InvolvedAccountExpr
	}
	GetBalance struct {
		Account InvolvedAccountExpr
		Asset   InvolvedAccountExpr
	}
	GetOverdraft struct {
		Account InvolvedAccountExpr
		Asset   InvolvedAccountExpr
	}
)

func (AssetLiteral) involvedAccountExpr()   {}
func (MakeMonetary) involvedAccountExpr()   {}
func (AccountLiteral) involvedAccountExpr() {}
func (NumberLiteral) involvedAccountExpr()  {}
func (StringLiteral) involvedAccountExpr()  {}
func (Add) involvedAccountExpr()            {}
func (Sub) involvedAccountExpr()            {}
func (Div) involvedAccountExpr()            {}
func (SubPrefix) involvedAccountExpr()      {}
func (FnMeta) involvedAccountExpr()         {}
func (GetAmount) involvedAccountExpr()      {}
func (GetAsset) involvedAccountExpr()       {}
func (GetBalance) involvedAccountExpr()     {}
func (GetOverdraft) involvedAccountExpr()   {}

type InvolvedAccount struct {
	AccountExpr InvolvedAccountExpr
	AssetExpr   InvolvedAccountExpr
}

type InvolvedMeta struct {
	// TODO add type here?
	Account InvolvedAccountExpr
	Key     InvolvedAccountExpr
}

type involvedAccountsAnalysisState struct {
	vars             VariablesMap
	evaluatedVars    map[string]InvolvedAccountExpr
	currentAsset     InvolvedAccountExpr
	involvedAccounts []InvolvedAccount
	involvedMeta     []InvolvedMeta
}

// A version of parseVar that returns involved account expr instead
func parseVarToInvolvedAccount(type_ string, rawValue string, r parser.Range) (InvolvedAccountExpr, InterpreterError) {
	val, err := parseVar(type_, rawValue, r)
	if err != nil {
		return nil, err
	}

	switch val := val.(type) {
	case String:
		return StringLiteral{String: string(val)}, nil
	case AccountAddress:
		return AccountLiteral{Account: string(val)}, nil
	case Asset:
		return AssetLiteral{Asset: string(val)}, nil

	case MonetaryInt:
		bi := big.Int(val)
		return NumberLiteral{Amount: &bi}, nil

	case Monetary:
		bi := big.Int(val.Amount)
		return MakeMonetary{
			Asset:  AssetLiteral{Asset: string(val.Asset)},
			Amount: NumberLiteral{Amount: &bi},
		}, nil

	case Portion:
		rat := big.Rat(val)
		left := NumberLiteral{Amount: rat.Num()}
		right := NumberLiteral{Amount: rat.Denom()}
		return Div{Left: left, Right: right}, nil
	}

	// TODO(bad_path)
	panic("TODO invalid val")
}

func GetInvolvedAccounts(vars VariablesMap, program parser.Program) ([]InvolvedAccount, []InvolvedMeta) {
	st := involvedAccountsAnalysisState{
		evaluatedVars: make(map[string]InvolvedAccountExpr),
	}
	if program.Vars != nil {
		st.parseVars(program.Vars.Declarations, vars)
	}

	for _, stmt := range program.Statements {
		switch stmt := stmt.(type) {
		case *parser.SendStatement:
			st.evalSendStmt(*stmt)

		case *parser.SaveStatement:
			st.evalSaveStmt(*stmt)

		case *parser.FnCall:
			switch stmt.Caller.Name {
			case analysis.FnSetTxMeta:
				// TODO(check)
				// can we ignore this ?

			case analysis.FnSetAccountMeta:
				acc := st.evalExpr(stmt.Args[0])
				key := st.evalExpr(stmt.Args[1])

				st.involvedMeta = append(st.involvedMeta, InvolvedMeta{
					Account: acc,
					Key:     key,
				})
			}
		}
	}

	return st.involvedAccounts, st.involvedMeta
}

func (s *involvedAccountsAnalysisState) parseVars(varDeclrs []parser.VarDeclaration, rawVars map[string]string) InterpreterError {
	for _, varsDecl := range varDeclrs {
		if varsDecl.Origin == nil {
			raw, ok := rawVars[varsDecl.Name.Name]
			if !ok {
				return MissingVariableErr{Name: varsDecl.Name.Name}
			}

			parsed, err := parseVarToInvolvedAccount(varsDecl.Type.Name, raw, varsDecl.Type.Range)
			if err != nil {
				return err
			}
			s.evaluatedVars[varsDecl.Name.Name] = parsed
		} else {
			value := s.evalVar(*varsDecl.Origin, varsDecl.Type.Name)
			s.evaluatedVars[varsDecl.Name.Name] = value
		}
	}
	return nil
}

func (st *involvedAccountsAnalysisState) evalSaveStmt(stmt parser.SaveStatement) {
	account := st.evalExpr(stmt.Account)

	switch sentValue := stmt.SentValue.(type) {
	case *parser.SentValueAll:
		asset := st.evalExpr(sentValue.Asset)

		st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
			AccountExpr: account,
			AssetExpr:   asset,
		})

	case *parser.SentValueLiteral:
		monetary := st.evalExpr(sentValue.Monetary)
		asset := foldedGetAsset(monetary)

		st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
			AccountExpr: account,
			AssetExpr:   asset,
		})
	}
}

func (st *involvedAccountsAnalysisState) evalSendStmt(stmt parser.SendStatement) {
	switch sentValue := stmt.SentValue.(type) {
	case *parser.SentValueAll:
		st.currentAsset = st.evalExpr(sentValue.Asset)
		st.evalSrc(stmt.Source)

	case *parser.SentValueLiteral:
		monetary := st.evalExpr(sentValue.Monetary)
		st.currentAsset = foldedGetAsset(monetary)
		st.evalSrc(stmt.Source)
	}
}

func (st involvedAccountsAnalysisState) evalAccountNamePart(part parser.AccountNamePart) InvolvedAccountExpr {
	switch part := part.(type) {
	case parser.AccountTextPart:
		return AccountLiteral{Account: part.Name}
	case *parser.Variable:
		expr, ok := st.evaluatedVars[part.Name]
		if !ok {
			// TODO(bad_path)
			panic("TODO unbound var")
		}
		return expr
	}

	return nil
}

// Constant folding for the Add{} node.
func foldedAdd(left InvolvedAccountExpr, right InvolvedAccountExpr) InvolvedAccountExpr {
	switch left.(type) {
	// TODO(impl) bonus: implement folds
	// note that it's correct even without constant folding

	default:
		return Add{Left: left, Right: right}
	}
}

func foldedGetAsset(expr InvolvedAccountExpr) InvolvedAccountExpr {
	switch expr := expr.(type) {
	case MakeMonetary:
		return expr.Asset

	default:
		return GetAsset{Monetary: expr}
	}
}

func foldedGetAmount(expr InvolvedAccountExpr) InvolvedAccountExpr {
	switch expr := expr.(type) {
	case MakeMonetary:
		return expr.Amount

	default:
		return GetAmount{Monetary: expr}
	}
}

func (st *involvedAccountsAnalysisState) evalVar(expr parser.ValueExpr, typ string) InvolvedAccountExpr {
	switch expr := expr.(type) {
	case *parser.FnCall:
		switch expr.Caller.Name {
		case analysis.FnVarOriginMeta:
			if len(expr.Args) != 2 {
				// TODO(bad_path)
				panic("TODO invalid args")
			}

			acc := st.evalExpr(expr.Args[0])
			key := st.evalExpr(expr.Args[1])

			st.involvedMeta = append(st.involvedMeta, InvolvedMeta{
				Account: acc,
				Key:     key,
			})

			return FnMeta{
				ExpectedType: typ,
				Account:      acc,
				Key:          key,
			}
		}
	}

	return st.evalExpr(expr)
}

func (st *involvedAccountsAnalysisState) evalExpr(expr parser.ValueExpr) InvolvedAccountExpr {
	switch expr := expr.(type) {
	case *parser.AccountInterpLiteral:

		var acc InvolvedAccountExpr
		for _, part := range expr.Parts {
			partExpr := st.evalAccountNamePart(part)
			if acc == nil {
				acc = partExpr
			} else {
				acc = foldedAdd(acc, partExpr)
			}
		}
		return acc

	case *parser.AssetLiteral:
		return AssetLiteral{Asset: expr.Asset}

	case *parser.Variable:
		varLookup, ok := st.evaluatedVars[expr.Name]
		if !ok {
			// TODO(bad_path)
			fmt.Printf("Var: %s, all vars: %#v\n\n\n", expr.Name, st.evaluatedVars)
			panic("TODO unbound var")
		}
		return varLookup

	case *parser.NumberLiteral:
		return NumberLiteral{Amount: expr.Number}

	case *parser.MonetaryLiteral:
		evalAmt := st.evalExpr(expr.Amount)
		evalAsset := st.evalExpr(expr.Asset)
		return MakeMonetary{Amount: evalAmt, Asset: evalAsset}

	case *parser.StringLiteral:
		return StringLiteral{String: expr.String}

	case *parser.Prefix:
		evalExpr := st.evalExpr(expr.Expr)
		switch expr.Operator {
		case parser.PrefixOperatorMinus:
			return SubPrefix{Expr: evalExpr}
		default:
			// TODO(bad_path)
			panic("TODO invalid op")
		}

	case *parser.BinaryInfix:
		evalLeft := st.evalExpr(expr.Left)
		evalRight := st.evalExpr(expr.Right)
		switch expr.Operator {
		case parser.InfixOperatorMinus:
			return Sub{Left: evalLeft, Right: evalRight}

		case parser.InfixOperatorDiv:
			return Div{Left: evalLeft, Right: evalRight}

		case parser.InfixOperatorPlus:
			return Add{Left: evalLeft, Right: evalRight}

		default:
			// TODO(bad_path)
			panic("TODO invalid op")
		}

	case *parser.PercentageLiteral:
		rat := expr.ToRatio()
		return Div{
			Left:  NumberLiteral{rat.Num()},
			Right: NumberLiteral{rat.Denom()},
		}

	case *parser.FnCall:
		switch expr.Caller.Name {
		case analysis.FnVarOriginMeta:
			// TODO(bad_path)
			panic("TODO invalid nested fn call")

		case analysis.FnVarOriginOverdraft:
			if len(expr.Args) != 2 {
				// TODO(bad_path)
				panic("TODO invalid args")
			}

			acc := st.evalExpr(expr.Args[0])
			expr := st.evalExpr(expr.Args[1])
			st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
				AccountExpr: acc,
				AssetExpr:   expr,
			})
			return GetOverdraft{
				Account: acc,
				Asset:   expr,
			}

		case analysis.FnVarOriginBalance:
			if len(expr.Args) != 2 {
				// TODO(bad_path)
				panic("TODO invalid args")
			}

			acc := st.evalExpr(expr.Args[0])
			expr := st.evalExpr(expr.Args[1])
			st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
				AccountExpr: acc,
				AssetExpr:   expr,
			})
			return GetBalance{
				Account: acc,
				Asset:   expr,
			}

		case analysis.FnVarOriginGetAmount:
			if len(expr.Args) != 1 {
				// TODO(bad_path)
				panic("TODO invalid args")
			}

			monetary := st.evalExpr(expr.Args[0])
			return foldedGetAmount(monetary)

		case analysis.FnVarOriginGetAsset:
			if len(expr.Args) != 1 {
				// TODO(bad_path)
				panic("TODO invalid args")
			}
			monetary := st.evalExpr(expr.Args[0])
			return foldedGetAsset(monetary)

		default:
			// TODO(bad_path)
			panic("TODO unimplmeented")
		}

	default:
		// TODO(bad_path)
		fmt.Printf("TODO: eval %#v\n", expr)
		panic("TODO impl evalExpr")
	}

}

func (st *involvedAccountsAnalysisState) evalSrc(source parser.Source) {
	switch source := source.(type) {
	case *parser.SourceWithScaling:
		// TODO(impl)
		panic("TODO unimplemented")

	case *parser.SourceOverdraft:
		// TODO(check) do we skip this?
		accountExpr := st.evalExpr(source.Address)
		st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
			AccountExpr: accountExpr,
			AssetExpr:   st.currentAsset,
		})

	case *parser.SourceAccount:
		accountExpr := st.evalExpr(source.ValueExpr)
		st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
			AccountExpr: accountExpr,
			AssetExpr:   st.currentAsset,
		})

	case *parser.SourceInorder:
		for _, acc := range source.Sources {
			st.evalSrc(acc)
		}

	case *parser.SourceOneof:
		for _, acc := range source.Sources {
			st.evalSrc(acc)
		}

	case *parser.SourceCapped:
		st.evalSrc(source.From)
	}
}

type isValidCallState struct {
	isTopLevel bool
}

func (st *isValidCallState) isValidCall(expr InvolvedAccountExpr) bool {
	isTopLevel := st.isTopLevel
	st.isTopLevel = false

	switch expr := expr.(type) {
	case GetBalance:
		return isTopLevel

	case NumberLiteral, StringLiteral, AssetLiteral, AccountLiteral:
		return true

	case Add:
		return st.isValidCall(expr.Left) && st.isValidCall(expr.Right)
	case Sub:
		return st.isValidCall(expr.Left) && st.isValidCall(expr.Right)
	case Div:
		return st.isValidCall(expr.Left) && st.isValidCall(expr.Right)
	case SubPrefix:
		return st.isValidCall(expr.Expr)
	case GetAmount:
		return st.isValidCall(expr.Monetary)
	case GetAsset:
		return st.isValidCall(expr.Monetary)

	case MakeMonetary:
		return st.isValidCall(expr.Amount) && st.isValidCall(expr.Asset)

	case FnMeta:
		return st.isValidCall(expr.Account) && st.isValidCall(expr.Key)
	}

	return false
}

func IsValidCall(expr InvolvedAccountExpr) bool {
	st := isValidCallState{isTopLevel: true}
	return st.isValidCall(expr)
}
