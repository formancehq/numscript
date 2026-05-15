package interpreter_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func getInvolvedAccounts(t *testing.T, vars interpreter.VariablesMap, src string) ([]interpreter.InvolvedAccount, []interpreter.InvolvedMeta) {
	t.Helper()
	out := parser.Parse(src)
	require.Empty(t, out.Errors)
	accs, meta, err := interpreter.GetInvolvedAccounts(vars, out.Value)
	require.NoError(t, err)
	return accs, meta
}

func getInvolvedAccountsErr(t *testing.T, vars interpreter.VariablesMap, src string) interpreter.InterpreterError {
	t.Helper()
	out := parser.Parse(src)
	_, _, err := interpreter.GetInvolvedAccounts(vars, out.Value)
	require.Error(t, err)
	return err
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
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
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
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
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
        AccountExpr: interpreter.ConcatAccount{
            Left: interpreter.ConcatAccount{
                Left: interpreter.ConcatAccount{
                    Left:  interpreter.AccountLiteral{Account:"user:"},
                    Right: interpreter.NumberLiteral{
                        Amount: &big.Int{
                            neg: false,
                            abs: {0x2a},
                        },
                    },
                },
                Right: interpreter.AccountLiteral{Account:":"},
            },
            Right: interpreter.AccountLiteral{Account:"pending"},
        },
        AssetExpr: interpreter.AssetLiteral{Asset:"USD/2"},
    },
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`),
		)
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
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
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
        Write:   nil,
    },
}`))
		snaps.MatchInlineSnapshot(t, accs, snaps.Inline("[]interpreter.InvolvedAccount(nil)"))
	})

	t.Run("required meta (write)", func(t *testing.T) {
		accs, meta := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		set_account_meta(@acc, "k", 42)
	`)
		snaps.MatchInlineSnapshot(t, meta, snaps.Inline(`[]interpreter.InvolvedMeta{
    {
        Account: interpreter.AccountLiteral{Account:"acc"},
        Key:     interpreter.StringLiteral{String:"k"},
        Write:   interpreter.NumberLiteral{
            Amount: &big.Int{
                neg: false,
                abs: {0x2a},
            },
        },
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
        Write:   nil,
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
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
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
        AccountExpr: interpreter.ConcatAccount{
            Left: interpreter.ConcatAccount{
                Left: interpreter.ConcatAccount{
                    Left:  interpreter.AccountLiteral{Account:"user:"},
                    Right: interpreter.FnMeta{
                        ExpectedType: "account",
                        Account:      interpreter.AccountLiteral{Account:"acc"},
                        Key:          interpreter.StringLiteral{String:"k"},
                    },
                },
                Right: interpreter.AccountLiteral{Account:":"},
            },
            Right: interpreter.AccountLiteral{Account:"pending"},
        },
        AssetExpr: interpreter.AssetLiteral{Asset:"USD/2"},
    },
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`),
		)
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
        Write:   nil,
    },
    {
        Account: interpreter.AccountLiteral{Account:"acc"},
        Key:     interpreter.FnMeta{
            ExpectedType: "string",
            Account:      interpreter.AccountLiteral{Account:"a1"},
            Key:          interpreter.StringLiteral{String:"k"},
        },
        Write: nil,
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
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
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
        Account: interpreter.ConcatAccount{
            Left:  interpreter.AccountLiteral{Account:"user:"},
            Right: interpreter.GetAmount{
                Monetary: interpreter.GetBalance{
                    Account: interpreter.AccountLiteral{Account:"acc"},
                    Asset:   interpreter.AssetLiteral{Asset:"USD/2"},
                },
            },
        },
        Key:   interpreter.StringLiteral{String:"k"},
        Write: nil,
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
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))
	})

	t.Run("eval asset", func(t *testing.T) {

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
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
        AssetExpr:   interpreter.FnMeta{
            ExpectedType: "asset",
            Account:      interpreter.AccountLiteral{Account:"acc"},
            Key:          interpreter.StringLiteral{String:"k"},
        },
    },
}`))

	})

	t.Run("eval asset as meta", func(t *testing.T) {

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
    {
        AccountExpr: interpreter.AccountLiteral{Account:"dest"},
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

	t.Run("allotment in src and dest", func(t *testing.T) {

		accs, _ := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		send [USD/2 42] (
			source = {
				1/2 from @a1
				remaining from @a2
			}
			destination = {
				1/3 to @d1
				1/3 kept
				remaining to @d2
			}
		)
	`)

		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"a1"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
    {
        AccountExpr: interpreter.AccountLiteral{Account:"a2"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
    {
        AccountExpr: interpreter.AccountLiteral{Account:"d1"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
    {
        AccountExpr: interpreter.AccountLiteral{Account:"d2"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
}`))

	})

	t.Run("many assets", func(t *testing.T) {

		accs, _ := getInvolvedAccounts(t, interpreter.VariablesMap{}, `
		send [USD/2 42] (
			source = @a1
			destination = { remaining kept}
		)
		send [EUR/2 42] (
			source = @a2
			destination = { remaining kept}
		)
	`)

		snaps.MatchInlineSnapshot(t, accs, snaps.Inline(`[]interpreter.InvolvedAccount{
    {
        AccountExpr: interpreter.AccountLiteral{Account:"a1"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"USD/2"},
    },
    {
        AccountExpr: interpreter.AccountLiteral{Account:"a2"},
        AssetExpr:   interpreter.AssetLiteral{Asset:"EUR/2"},
    },
}`))

	})

}

func TestGetInvolvedAccountsErrors(t *testing.T) {
	t.Run("missing variable", func(t *testing.T) {
		err := getInvolvedAccountsErr(t, interpreter.VariablesMap{}, `
		vars { account $acc }
		send [USD/2 *] (
			source = $acc
			destination = @dest
		)
	`)
		var target interpreter.MissingVariableErr
		require.ErrorAs(t, err, &target)
		require.Equal(t, "acc", target.Name)
	})

	t.Run("invalid variable value", func(t *testing.T) {
		err := getInvolvedAccountsErr(t, interpreter.VariablesMap{"acc": "not a valid account!!"}, `
		vars { account $acc }
		send [USD/2 *] (
			source = $acc
			destination = @dest
		)
	`)
		var target interpreter.InvalidAccountName
		require.ErrorAs(t, err, &target)
	})

	t.Run("nested meta in var origin args", func(t *testing.T) {
		// meta() as an argument to another meta() hits InvalidNestedMeta in evalExpr
		err := getInvolvedAccountsErr(t, interpreter.VariablesMap{}, `
		vars { string $x = meta(@a, meta(@b, "k")) }
	`)
		var target interpreter.InvalidNestedMeta
		require.ErrorAs(t, err, &target)
	})

	t.Run("unbound variable in account interpolation", func(t *testing.T) {
		// $undeclared is used in source but not declared in vars block;
		// parser reports no error (it's syntactically valid), but
		// GetInvolvedAccounts returns UnboundVariableErr.
		out := parser.Parse(`
		send [USD/2 *] (
			source = @user:$undeclared:end
			destination = @dest
		)
	`)
		_, _, err := interpreter.GetInvolvedAccounts(interpreter.VariablesMap{}, out.Value)
		require.Error(t, err)
		var target interpreter.UnboundVariableErr
		require.ErrorAs(t, err, &target)
		require.Equal(t, "undeclared", target.Name)
	})
}
