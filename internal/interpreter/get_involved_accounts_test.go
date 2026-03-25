package interpreter_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func getInvolvedAccounts(t *testing.T, vars interpreter.VariablesMap, src string) []interpreter.InvolvedAccount {
	out := parser.Parse(src)
	require.Empty(t, out.Errors)
	return interpreter.GetInvolvedAccounts(vars, out.Value)
}

func TestGetInvolvedAccount(t *testing.T) {
	t.Run("simple (no vars)", func(t *testing.T) {
		accs := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		send [USD/2 *] (
			source = @src
			destination = @dest
		)
	`)
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"src"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`),
		)
	})

	t.Run("simple (get var)", func(t *testing.T) {
		accs := getInvolvedAccounts(t, interpreter.VariablesMap{
			"acc": "acc",
		}, `
		vars { account $acc }
		send [USD/2 *] (
			source = $acc
			destination = @dest
		)
	`)

		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"src"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`),
		)
	})

}
