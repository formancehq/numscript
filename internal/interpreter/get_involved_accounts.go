package interpreter

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/flags"
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

	return nil, InvalidTypeErr{Range: r, Name: fmt.Sprintf("%T", val)}
}

func GetInvolvedAccounts(vars VariablesMap, program parser.Program) ([]InvolvedAccount, []InvolvedMeta, InterpreterError) {
	st := involvedAccountsAnalysisState{
		evaluatedVars: make(map[string]InvolvedAccountExpr),
	}
	if program.Vars != nil {
		if err := st.parseVars(program.Vars.Declarations, vars); err != nil {
			return nil, nil, err
		}
	}

	for _, stmt := range program.Statements {
		switch stmt := stmt.(type) {
		case *parser.SendStatement:
			if err := st.evalSendStmt(*stmt); err != nil {
				return nil, nil, err
			}

		case *parser.SaveStatement:
			if err := st.evalSaveStmt(*stmt); err != nil {
				return nil, nil, err
			}

		case *parser.FnCall:
			switch stmt.Caller.Name {
			case analysis.FnSetTxMeta:
				// we can safely ignore this

			case analysis.FnSetAccountMeta:
				if len(stmt.Args) != 2 {
					return nil, nil, BadArityErr{Range: stmt.Range, ExpectedArity: 2, GivenArguments: len(stmt.Args)}
				}
				acc, err := st.evalExpr(stmt.Args[0])
				if err != nil {
					return nil, nil, err
				}
				key, err := st.evalExpr(stmt.Args[1])
				if err != nil {
					return nil, nil, err
				}
				st.involvedMeta = append(st.involvedMeta, InvolvedMeta{
					Account: acc,
					Key:     key,
				})
			}
		}
	}

	return st.involvedAccounts, st.involvedMeta, nil
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
			value, err := s.evalVar(*varsDecl.Origin, varsDecl.Type.Name)
			if err != nil {
				return err
			}
			s.evaluatedVars[varsDecl.Name.Name] = value
		}
	}
	return nil
}

func (st *involvedAccountsAnalysisState) evalSaveStmt(stmt parser.SaveStatement) InterpreterError {
	account, err := st.evalExpr(stmt.Account)
	if err != nil {
		return err
	}

	switch sentValue := stmt.SentValue.(type) {
	case *parser.SentValueAll:
		asset, err := st.evalExpr(sentValue.Asset)
		if err != nil {
			return err
		}
		st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
			AccountExpr: account,
			AssetExpr:   asset,
		})

	case *parser.SentValueLiteral:
		monetary, err := st.evalExpr(sentValue.Monetary)
		if err != nil {
			return err
		}
		asset := foldedGetAsset(monetary)
		st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
			AccountExpr: account,
			AssetExpr:   asset,
		})
	}
	return nil
}

func (st *involvedAccountsAnalysisState) evalSendStmt(stmt parser.SendStatement) InterpreterError {
	switch sentValue := stmt.SentValue.(type) {
	case *parser.SentValueAll:
		asset, err := st.evalExpr(sentValue.Asset)
		if err != nil {
			return err
		}
		st.currentAsset = asset
		if err := st.evalSrc(stmt.Source); err != nil {
			return err
		}
		return st.evalDest(stmt.Destination)

	case *parser.SentValueLiteral:
		monetary, err := st.evalExpr(sentValue.Monetary)
		if err != nil {
			return err
		}
		st.currentAsset = foldedGetAsset(monetary)
		if err := st.evalSrc(stmt.Source); err != nil {
			return err
		}
		return st.evalDest(stmt.Destination)
	}
	return nil
}

func (st involvedAccountsAnalysisState) evalAccountNamePart(part parser.AccountNamePart) (InvolvedAccountExpr, InterpreterError) {
	switch part := part.(type) {
	case parser.AccountTextPart:
		return AccountLiteral{Account: part.Name}, nil
	case *parser.Variable:
		expr, ok := st.evaluatedVars[part.Name]
		if !ok {
			return nil, UnboundVariableErr{Range: part.Range, Name: part.Name}
		}
		return expr, nil
	}

	return nil, InvalidTypeErr{Range: parser.Range{}, Name: fmt.Sprintf("%T", part)}
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

func (st *involvedAccountsAnalysisState) evalVar(expr parser.ValueExpr, typ string) (InvolvedAccountExpr, InterpreterError) {
	switch expr := expr.(type) {
	case *parser.FnCall:
		switch expr.Caller.Name {
		case analysis.FnVarOriginMeta:
			if len(expr.Args) != 2 {
				return nil, BadArityErr{Range: expr.Range, ExpectedArity: 2, GivenArguments: len(expr.Args)}
			}

			acc, err := st.evalExpr(expr.Args[0])
			if err != nil {
				return nil, err
			}
			key, err := st.evalExpr(expr.Args[1])
			if err != nil {
				return nil, err
			}

			st.involvedMeta = append(st.involvedMeta, InvolvedMeta{
				Account: acc,
				Key:     key,
			})

			return FnMeta{
				ExpectedType: typ,
				Account:      acc,
				Key:          key,
			}, nil
		}
	}

	return st.evalExpr(expr)
}

func (st *involvedAccountsAnalysisState) evalExpr(expr parser.ValueExpr) (InvolvedAccountExpr, InterpreterError) {
	switch expr := expr.(type) {
	case *parser.AccountInterpLiteral:
		var acc InvolvedAccountExpr
		for _, part := range expr.Parts {
			partExpr, err := st.evalAccountNamePart(part)
			if err != nil {
				return nil, err
			}
			if acc == nil {
				acc = partExpr
			} else {
				acc = foldedAdd(acc, partExpr)
			}
		}
		return acc, nil

	case *parser.AssetLiteral:
		return AssetLiteral{Asset: expr.Asset}, nil

	case *parser.Variable:
		varLookup, ok := st.evaluatedVars[expr.Name]
		if !ok {
			return nil, UnboundVariableErr{Range: expr.Range, Name: expr.Name}
		}
		return varLookup, nil

	case *parser.NumberLiteral:
		return NumberLiteral{Amount: expr.Number}, nil

	case *parser.MonetaryLiteral:
		evalAmt, err := st.evalExpr(expr.Amount)
		if err != nil {
			return nil, err
		}
		evalAsset, err := st.evalExpr(expr.Asset)
		if err != nil {
			return nil, err
		}
		return MakeMonetary{Amount: evalAmt, Asset: evalAsset}, nil

	case *parser.StringLiteral:
		return StringLiteral{String: expr.String}, nil

	case *parser.Prefix:
		inner, err := st.evalExpr(expr.Expr)
		if err != nil {
			return nil, err
		}
		switch expr.Operator {
		case parser.PrefixOperatorMinus:
			return SubPrefix{Expr: inner}, nil
		default:
			return nil, InvalidOperatorErr{Range: expr.Range, Operator: string(expr.Operator)}
		}

	case *parser.BinaryInfix:
		evalLeft, err := st.evalExpr(expr.Left)
		if err != nil {
			return nil, err
		}
		evalRight, err := st.evalExpr(expr.Right)
		if err != nil {
			return nil, err
		}
		switch expr.Operator {
		case parser.InfixOperatorMinus:
			return Sub{Left: evalLeft, Right: evalRight}, nil
		case parser.InfixOperatorDiv:
			return Div{Left: evalLeft, Right: evalRight}, nil
		case parser.InfixOperatorPlus:
			return Add{Left: evalLeft, Right: evalRight}, nil
		default:
			return nil, InvalidOperatorErr{Range: expr.Range, Operator: string(expr.Operator)}
		}

	case *parser.PercentageLiteral:
		rat := expr.ToRatio()
		return Div{
			Left:  NumberLiteral{rat.Num()},
			Right: NumberLiteral{rat.Denom()},
		}, nil

	case *parser.FnCall:
		switch expr.Caller.Name {
		case analysis.FnVarOriginMeta:
			return nil, InvalidNestedMeta{Range: expr.Range}

		case analysis.FnVarOriginOverdraft:
			if len(expr.Args) != 2 {
				return nil, BadArityErr{Range: expr.Range, ExpectedArity: 2, GivenArguments: len(expr.Args)}
			}
			acc, err := st.evalExpr(expr.Args[0])
			if err != nil {
				return nil, err
			}
			asset, err := st.evalExpr(expr.Args[1])
			if err != nil {
				return nil, err
			}
			st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
				AccountExpr: acc,
				AssetExpr:   asset,
			})
			return GetOverdraft{
				Account: acc,
				Asset:   asset,
			}, nil

		case analysis.FnVarOriginBalance:
			if len(expr.Args) != 2 {
				return nil, BadArityErr{Range: expr.Range, ExpectedArity: 2, GivenArguments: len(expr.Args)}
			}
			acc, err := st.evalExpr(expr.Args[0])
			if err != nil {
				return nil, err
			}
			asset, err := st.evalExpr(expr.Args[1])
			if err != nil {
				return nil, err
			}
			st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
				AccountExpr: acc,
				AssetExpr:   asset,
			})
			return GetBalance{
				Account: acc,
				Asset:   asset,
			}, nil

		case analysis.FnVarOriginGetAmount:
			if len(expr.Args) != 1 {
				return nil, BadArityErr{Range: expr.Range, ExpectedArity: 1, GivenArguments: len(expr.Args)}
			}
			monetary, err := st.evalExpr(expr.Args[0])
			if err != nil {
				return nil, err
			}
			return foldedGetAmount(monetary), nil

		case analysis.FnVarOriginGetAsset:
			if len(expr.Args) != 1 {
				return nil, BadArityErr{Range: expr.Range, ExpectedArity: 1, GivenArguments: len(expr.Args)}
			}
			monetary, err := st.evalExpr(expr.Args[0])
			if err != nil {
				return nil, err
			}
			return foldedGetAsset(monetary), nil

		default:
			return nil, UnboundFunctionErr{Range: expr.Range, Name: expr.Caller.Name}
		}

	default:
		return nil, InvalidTypeErr{Range: expr.GetRange(), Name: fmt.Sprintf("%T", expr)}
	}
}

func (st *involvedAccountsAnalysisState) evalSrc(source parser.Source) InterpreterError {
	switch source := source.(type) {
	case *parser.SourceWithScaling:
		return ExperimentalFeature{FlagName: flags.AssetScaling}

	case *parser.SourceOverdraft:
		if source.Color != nil {
			return ExperimentalFeature{FlagName: flags.AssetScaling}
		}

		accountExpr, err := st.evalExpr(source.Address)
		if err != nil {
			return err
		}
		st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
			AccountExpr: accountExpr,
			AssetExpr:   st.currentAsset,
		})

	case *parser.SourceAccount:
		if source.Color != nil {
			return ExperimentalFeature{FlagName: flags.AssetScaling}
		}

		accountExpr, err := st.evalExpr(source.ValueExpr)
		if err != nil {
			return err
		}
		st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
			AccountExpr: accountExpr,
			AssetExpr:   st.currentAsset,
		})

	case *parser.SourceInorder:
		for _, acc := range source.Sources {
			if err := st.evalSrc(acc); err != nil {
				return err
			}
		}

	case *parser.SourceOneof:
		for _, acc := range source.Sources {
			if err := st.evalSrc(acc); err != nil {
				return err
			}
		}

	case *parser.SourceCapped:
		return st.evalSrc(source.From)

	case *parser.SourceAllotment:
		for _, acc := range source.Items {
			if err := st.evalSrc(acc.From); err != nil {
				return err
			}
		}
	}
	return nil
}

func (st *involvedAccountsAnalysisState) evalDest(dest parser.Destination) InterpreterError {
	switch dest := dest.(type) {
	case *parser.DestinationAccount:
		accountExpr, err := st.evalExpr(dest.ValueExpr)
		if err != nil {
			return err
		}
		st.involvedAccounts = append(st.involvedAccounts, InvolvedAccount{
			AccountExpr: accountExpr,
			AssetExpr:   st.currentAsset,
		})

	case *parser.DestinationInorder:
		for _, clause := range dest.Clauses {
			if err := st.evalKeptOrDest(clause.To); err != nil {
				return err
			}
		}
		if err := st.evalKeptOrDest(dest.Remaining); err != nil {
			return err
		}

	case *parser.DestinationOneof:
		for _, acc := range dest.Clauses {
			if err := st.evalKeptOrDest(acc.To); err != nil {
				return err
			}
		}
		if err := st.evalKeptOrDest(dest.Remaining); err != nil {
			return err
		}

	case *parser.DestinationAllotment:
		for _, acc := range dest.Items {
			if err := st.evalKeptOrDest(acc.To); err != nil {
				return err
			}
		}
	}
	return nil
}

func (st *involvedAccountsAnalysisState) evalKeptOrDest(keptOrDest parser.KeptOrDestination) InterpreterError {
	switch keptOrDest := keptOrDest.(type) {
	case *parser.DestinationKept:
		// nothing to do here
	case *parser.DestinationTo:
		return st.evalDest(keptOrDest.Destination)
	}
	return nil
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
