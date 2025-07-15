package interpreter

import (
	"slices"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

// traverse the script to batch in advance required balance queries

func (st *programState) findBalancesQueriesInStatement(statement parser.Statement) InterpreterError {
	switch statement := statement.(type) {
	case *parser.FnCall:
		return nil

	case *parser.SaveStatement:
		asset, _, err := st.evaluateSentAmt(statement.SentValue)
		if err != nil {
			return err
		}

		// Although we don't technically need this account's balance rn,
		// having access to the balance simplifies the "save" statement implementation
		// this means that we would have a needless query in the case in which the account
		// which is selected in the "save" statement never actually appears as source
		//
		// this would mean that the "save" statement was not needed in the first place,
		// so preventing this query would hardly be an useful optimization
		account, err := evaluateExprAs(st, statement.Amount, expectAccount)
		if err != nil {
			return err
		}
		st.batchQuery(*account, *asset, nil)
		return nil

	case *parser.SendStatement:
		asset, _, err := st.evaluateSentAmt(statement.SentValue)
		if err != nil {
			return err
		}
		st.CurrentAsset = *asset

		// traverse source
		return st.findBalancesQueries(statement.Source)

	default:
		utils.NonExhaustiveMatchPanic[any](statement)
		return nil
	}
}

func (st *programState) batchQuery(account string, asset string, color *string) {
	if account == "world" {
		return
	}
	asset = coloredAsset(asset, color)

	previousValues := st.CurrentBalanceQuery[account]
	if !slices.Contains(previousValues, asset) {
		st.CurrentBalanceQuery[account] = append(previousValues, asset)
	}
}

func (st *programState) runBalancesQuery() error {
	filteredQuery := st.CachedBalances.filterQuery(st.CurrentBalanceQuery)

	// avoid updating balances if we don't need to fetch new data
	if len(filteredQuery) == 0 {
		return nil
	}

	queriedBalances, err := st.Store.GetBalances(st.ctx, filteredQuery)
	if err != nil {
		return err
	}
	// reset batch query
	st.CurrentBalanceQuery = BalanceQuery{}

	st.CachedBalances.mergeBalance(queriedBalances)

	return nil
}

func (st *programState) findBalancesQueries(source parser.Source) InterpreterError {
	switch source := source.(type) {
	case *parser.SourceAccount:
		account, err := evaluateExprAs(st, source.ValueExpr, expectAccount)
		if err != nil {
			return err
		}

		color, err := evaluateOptExprAs(st, source.Color, expectString)
		if err != nil {
			return err
		}

		st.batchQuery(*account, st.CurrentAsset, color)
		return nil

	case *parser.SourceOverdraft:
		// Skip balance tracking when balance is overdraft
		if source.Bounded == nil {
			return nil
		}

		account, err := evaluateExprAs(st, source.Address, expectAccount)
		if err != nil {
			return err
		}
		color, err := evaluateOptExprAs(st, source.Color, expectString)
		if err != nil {
			return err
		}

		st.batchQuery(*account, st.CurrentAsset, color)
		return nil

	case *parser.SourceInorder:
		for _, subSource := range source.Sources {
			err := st.findBalancesQueries(subSource)
			if err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceOneof:
		for _, subSource := range source.Sources {
			err := st.findBalancesQueries(subSource)
			if err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceCapped:
		// TODO can this be optimized in some cases?
		return st.findBalancesQueries(source.From)

	case *parser.SourceAllotment:
		for _, item := range source.Items {
			err := st.findBalancesQueries(item.From)
			if err != nil {
				return err
			}
		}
		return nil

	default:
		_ = utils.NonExhaustiveMatchPanic[error](source)
		return nil
	}
}
