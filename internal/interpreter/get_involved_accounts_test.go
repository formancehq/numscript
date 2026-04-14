package interpreter_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func getInvolvedAccounts(t *testing.T, vars interpreter.VariablesMap, src string) ([]interpreter.InvolvedAccount, []interpreter.InvolvedMeta) {
	out := parser.Parse(src)
	require.Empty(t, out.Errors)
	return interpreter.GetInvolvedAccounts(vars, out.Value)
}

func TestGetInvolvedAccount(t *testing.T) {
	t.Run("simple (no vars)", func(t *testing.T) {
		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		send [USD/2 *] (
			source = @src
			destination = @dest
		)
	`)
		require.Nil(t, meta)
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"src"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`),
		)
	})

	t.Run("simple (get var)", func(t *testing.T) {
		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{
			"acc": "acc_value_after_subst",
		}, `
		vars { account $acc }
		send [USD/2 *] (
			source = $acc
			destination = @dest
		)
	`)
		require.Nil(t, meta)
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"acc_value_after_subst"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`),
		)
	})

	t.Run("simple (account interp var)", func(t *testing.T) {
		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{
			"id": "42",
		}, `
		vars { number $id }
		send [USD/2 *] (
			source = @user:$id:pending
			destination = @dest
		)
	`)
		require.Nil(t, meta)
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.Add{
            Left: interpreter.Add{
                Left:  interpreter.AccountLiteral{Account:"user"},
                Right: interpreter.NumberLiteral{
                    Amount: &big.Int{
                        neg: false,
                        abs: {0x2a},
                    },
                },
            },
            Right: interpreter.AccountLiteral{Account:"pending"},
        },
        AssetExpr: interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
	})

	t.Run("eval var expr", func(t *testing.T) {
		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		vars { account $acc = [EUR/2 (100 + 42)] }
		send [USD/2 *] (
			source = $acc
			destination = @dest
		)
	`)
		require.Nil(t, meta)
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.MakeMonetary{
            Asset:  interpreter.AssetLiteral{Asset:"EUR/2"},
            Amount: interpreter.Add{
                Left: interpreter.NumberLiteral{
                    Amount: &big.Int{
                        neg: false,
                        abs: {0x64},
                    },
                },
                Right: interpreter.NumberLiteral{
                    Amount: &big.Int{
                        neg: false,
                        abs: {0x2a},
                    },
                },
            },
        },
        AssetExpr: interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
	})

	t.Run("required meta", func(t *testing.T) {
		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		// this does not involve accounts, but it's required meta
		vars { account $acc = meta(@acc, "k") }
	`)
		snaps.MatchInlineSnapshot(t, meta, snaps.Inline(`[]interpreter.InvolvedMeta{
    {
        Account: interpreter.AccountLiteral{Account:"acc"},
        Key:     interpreter.StringLiteral{String:"k"},
    },
}`))
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline("[]interpreter.InvolvedAccount(nil)"))
	})

	t.Run("meta fn check", func(t *testing.T) {
		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		vars { account $acc = meta(@acc, "k") }
		send [USD/2 *] (
			source = $acc
			destination = @dest
		)
	`)
		snaps.MatchInlineSnapshot(t, meta, snaps.Inline(`[]interpreter.InvolvedMeta{
    {
        Account: interpreter.AccountLiteral{Account:"acc"},
        Key:     interpreter.StringLiteral{String:"k"},
    },
}`))
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.FnMeta{
            ExpectedType: "account",
            Account:      interpreter.AccountLiteral{Account:"acc"},
            Key:          interpreter.StringLiteral{String:"k"},
        },
        AssetExpr: interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
	})

	t.Run("unresolved meta under string addition", func(t *testing.T) {
		accs, _ := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		vars { account $acc = meta(@acc, "k") }
		send [USD/2 *] (
			source = @user:$acc:pending
			destination = @dest
		)
	`)

		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.Add{
            Left: interpreter.Add{
                Left:  interpreter.AccountLiteral{Account:"user"},
                Right: interpreter.FnMeta{
                    ExpectedType: "account",
                    Account:      interpreter.AccountLiteral{Account:"acc"},
                    Key:          interpreter.StringLiteral{String:"k"},
                },
            },
            Right: interpreter.AccountLiteral{Account:"pending"},
        },
        AssetExpr: interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
	})

	t.Run("nested meta fn check", func(t *testing.T) {
		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		vars {
			string $s = meta(@a1, "k")
			account $acc = meta(@acc, $s)
		}

		send [USD/2 *] (
			source = $acc
			destination = @dest
		)
	`)
		snaps.MatchInlineSnapshot(t, meta, snaps.Inline(`[]interpreter.InvolvedMeta{
    {
        Account: interpreter.AccountLiteral{Account:"a1"},
        Key:     interpreter.StringLiteral{String:"k"},
    },
    {
        Account: interpreter.AccountLiteral{Account:"acc"},
        Key:     interpreter.FnMeta{
            ExpectedType: "string",
            Account:      interpreter.AccountLiteral{Account:"a1"},
            Key:          interpreter.StringLiteral{String:"k"},
        },
    },
}`))
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.FnMeta{
            ExpectedType: "account",
            Account:      interpreter.AccountLiteral{Account:"acc"},
            Key:          interpreter.FnMeta{
                ExpectedType: "string",
                Account:      interpreter.AccountLiteral{Account:"a1"},
                Key:          interpreter.StringLiteral{String:"k"},
            },
        },
        AssetExpr: interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
	})

	t.Run("involved account in balance check", func(t *testing.T) {
		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		vars {
			// even if this is dead code, we don't need to care about
			// this kind of dead code elimination. It'd make the inference
			// harder and we'd have no real world perf gain
			// users should just avoid this kind of dead code, and it's marked
			// as warning by the "numscript check" command
			monetary $acc = balance(@acc, USD/2)
		}
	`)
		require.Nil(t, meta)
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"acc"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
	})

	t.Run("forbid invalid meta keys", func(t *testing.T) {

		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		vars {
			monetary $acc = balance(@acc, USD/2)
			number $amt = get_amount($acc)
			account $acc = @user:$amt
			string $m = meta($acc, "k")
		}
	`)
		snaps.MatchInlineSnapshot(t, meta, snaps.Inline(`[]interpreter.InvolvedMeta{
    {
        Account: interpreter.Add{
            Left:  interpreter.AccountLiteral{Account:"user"},
            Right: interpreter.GetAmount{
                Monetary: interpreter.GetBalance{
                    Account: interpreter.AccountLiteral{Account:"acc"},
                    Asset:   interpreter.AssetLiteral{Asset:"USD/2"},
                },
            },
        },
        Key: interpreter.StringLiteral{String:"k"},
    },
}`))

		require.False(t, interpreter.IsValidCall(meta[0].Account))
		require.True(t, interpreter.IsValidCall(meta[0].Key))
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"acc"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
	})

	t.Run("bounded send statements", func(t *testing.T) {

		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		send [USD/2 100] (
			source = @acc
			destination = @dest
		)
	`)
		snaps.MatchInlineSnapshot(t, meta, snaps.Inline("[]interpreter.InvolvedMeta(nil)"))
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"acc"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`),
		)
	})

	t.Run("TODO descr", func(t *testing.T) {

		accs, _ := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		vars {
			asset $a = meta(@acc, "k")
		}
		send [$a 100] (
			source = @acc
			destination = @dest
		)
	`)

		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"acc"},
        AssetExpr:   interpreter.FnMeta{
            ExpectedType: "asset",
            Account:      interpreter.AccountLiteral{Account:"acc"},
            Key:          interpreter.StringLiteral{String:"k"},
        },
    },
}`))

	})

	t.Run("TODO descr 2", func(t *testing.T) {

		accs, _ := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		vars {
			monetary $a = meta(@acc, "k")
		}
		send $a (
			source = @acc
			destination = @dest
		)
	`)

		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"acc"},
        AssetExpr:   interpreter.GetAsset{
            Monetary: interpreter.FnMeta{
                ExpectedType: "monetary",
                Account:      interpreter.AccountLiteral{Account:"acc"},
                Key:          interpreter.StringLiteral{String:"k"},
            },
        },
    },
}`))

	})

}
