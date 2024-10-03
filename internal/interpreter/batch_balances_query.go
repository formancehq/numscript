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

	case *parser.SendStatement:
		// set the current asset
		switch sentValue := statement.SentValue.(type) {
		case *parser.SentValueAll:
			asset, err := evaluateLitExpecting(st, sentValue.Asset, expectAsset)
			if err != nil {
				return err
			}
			st.CurrentAsset = *asset

		case *parser.SentValueLiteral:
			monetary, err := evaluateLitExpecting(st, sentValue.Monetary, expectMonetary)
			if err != nil {
				return err
			}
			st.CurrentAsset = string(monetary.Asset)

		default:
			utils.NonExhaustiveMatchPanic[any](sentValue)
		}

		// traverse source
		return st.findBalancesQueries(statement.Source)

	default:
		utils.NonExhaustiveMatchPanic[any](statement)
		return nil
	}
}

func (st *programState) batchQuery(account string, asset string) {
	if account == "world" {
		return
	}

	previousValues := st.CurrentBalanceQuery[account]
	if !slices.Contains[[]string, string](previousValues, account) {
		st.CurrentBalanceQuery[account] = append(previousValues, asset)
	}
}

func (st *programState) runBalancesQuery() error {
	balances, err := st.Store.GetBalances(st.ctx, st.CurrentBalanceQuery)
	if err != nil {
		return err
	}
	// reset batch query
	st.CurrentBalanceQuery = BalanceQuery{}

	// TODO merge maps
	// TODO handle optimistic balances
	st.CachedBalances = balances
	return nil
}

func (st *programState) findBalancesQueries(source parser.Source) InterpreterError {
	switch source := source.(type) {
	case *parser.SourceAccount:
		account, err := evaluateLitExpecting(st, source.Literal, expectAccount)
		if err != nil {
			return err
		}

		st.batchQuery(*account, st.CurrentAsset)
		return nil

	case *parser.SourceOverdraft:
		// Skip balance tracking when balance is overdraft
		if source.Bounded == nil {
			return nil
		}

		account, err := evaluateLitExpecting(st, source.Address, expectAccount)
		if err != nil {
			return err
		}
		st.batchQuery(*account, st.CurrentAsset)
		return nil

	case *parser.SourceInorder:
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
		utils.NonExhaustiveMatchPanic[error](source)
		return nil
	}
}
