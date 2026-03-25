package interpreter

import (
	"fmt"

	"github.com/formancehq/numscript/internal/parser"
)

type InvolvedAccountExpr interface{ involvedAccountExpr() }

type AccountNamePart interface{ accountNamePart() }
type AccountTextPart struct{ Name string }
type AccountVariablePart struct{ Expr InvolvedAccountExpr }

func (AccountTextPart) accountNamePart()     {}
func (AccountVariablePart) accountNamePart() {}

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
	NumberLiteral struct{}
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
		Right InvolvedAccountExpr
	}
	FnMeta struct {
		Account InvolvedAccountExpr
		Key     InvolvedAccountExpr
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

type InvolvedAccount struct {
	AccountExpr InvolvedAccountExpr
	AssetExpr   InvolvedAccountExpr
}

func GetInvolvedAccounts(vars VariablesMap, program parser.Program) []InvolvedAccount {
	st := involvedAccountsAnalysisState{}

	for _, stmt := range program.Statements {
		switch stmt := stmt.(type) {
		case *parser.SendStatement:
			st.evalSendStmt(*stmt)

		case *parser.SaveStatement:
			// TODO
		case *parser.FnCall:
			// TODO
		}

	}

	return st.output
}

type involvedAccountsAnalysisState struct {
	vars         VariablesMap
	currentAsset InvolvedAccountExpr
	output       []InvolvedAccount
}

func (st *involvedAccountsAnalysisState) evalSendStmt(stmt parser.SendStatement) {
	switch sentValue := stmt.SentValue.(type) {
	case *parser.SentValueAll:
		st.currentAsset = st.evalExpr(sentValue.Asset)
		st.evalSrc(stmt.Source)

	case *parser.SentValueLiteral:
		// TODO here we should take the value of the monetary
		// st.currentAsset = st.evalExpr(sentValue.Monetary)
		// st.evalSrc(stmt.Source)

	}
}

func (st involvedAccountsAnalysisState) evalAccountNamePart(part parser.AccountNamePart) InvolvedAccountExpr {
	switch part := part.(type) {
	case parser.AccountTextPart:
		return AccountLiteral{Account: part.Name}
	case *parser.Variable:
		panic("TODO subst part")
		// TODO subst var
	}

	return nil
}

func (st involvedAccountsAnalysisState) evalExpr(expr parser.ValueExpr) InvolvedAccountExpr {
	switch expr := expr.(type) {
	case *parser.AccountInterpLiteral:

		var acc InvolvedAccountExpr
		for _, part := range expr.Parts {
			partExpr := st.evalAccountNamePart(part)
			if acc == nil {
				acc = partExpr
			} else {
				acc = Add{
					Left:  acc,
					Right: partExpr,
				}
			}
		}
		return acc

	case *parser.AssetLiteral:
		return AssetLiteral{Asset: expr.Asset}

	// case *parser.Variable:
	// case *parser.MonetaryLiteral:
	// case *parser.PercentageLiteral:
	// case *parser.NumberLiteral:
	// case *parser.StringLiteral:
	// case *parser.BinaryInfix:
	// case *parser.Prefix:
	// case *parser.FnCall:
	default:
		fmt.Printf("TODO: eval %#v\n", expr)
		panic("TODO impl")

	}

	return nil
}

func (st *involvedAccountsAnalysisState) evalSrc(source parser.Source) InvolvedAccountExpr {
	switch source := source.(type) {
	case *parser.SourceAccount:
		accountExpr := st.evalExpr(source.ValueExpr)
		st.output = append(st.output, InvolvedAccount{
			AccountExpr: accountExpr,
			AssetExpr:   st.currentAsset,
		})

	case *parser.SourceInorder:

	case *parser.SourceOneof:
	case *parser.SourceCapped:
	case *parser.SourceOverdraft:
	case *parser.SourceWithScaling:

	}

	return nil
}
