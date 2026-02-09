package analysis

import (
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

func DetectAccounts(
	program parser.Program,
	vars map[string]string,
) {
	for _, statement := range program.Statements {
		switch statement := statement.(type) {
		case *parser.FnCall:
			return nil

		case *parser.SaveStatement:
			asset, _, err := st.evaluateSentAmt(statement.SentValue)
			if err != nil {
				return err
			}

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

}
