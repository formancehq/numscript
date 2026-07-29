package compiler_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

func TestE2E_ExternalVars(t *testing.T) {
	src := `
		vars {
			account $dest
			monetary $m
		}
		send $m (
			source = @world
			destination = $dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	enc, program, err := compiler.Compile(parsed.Value, nil)
	require.NoError(t, err)

	vars, err := enc.Encode(map[string]string{
		"dest": "alice",
		"m":    "USD/2 100",
	})
	require.NoError(t, err)

	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, &vars, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "world", Destination: "alice", Asset: "USD/2", Amount: big.NewInt(100)},
	}
	requirePostingsEqual(t, want, res.Postings)
}

func TestE2E_InvalidInterpolatedAccount(t *testing.T) {
	src := `
		#![feature("experimental-account-interpolation")]
		vars { string $status }
		set_tx_meta("k", @user:$status)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	enc, program, err := compiler.Compile(parsed.Value, nil)
	require.NoError(t, err)

	vars, err := enc.Encode(map[string]string{"status": "!invalid acc.."})
	require.NoError(t, err)

	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, &vars, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	require.Equal(t, vm.InvalidAccountName{Name: "user:!invalid acc.."}, execErr)
}

func compileEncoder(t *testing.T, src string) compiler.VarsEncoder {
	t.Helper()
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	enc, _, err := compiler.Compile(parsed.Value, nil)
	require.NoError(t, err)
	return enc
}

// A var of each type decomposes into its int/string slots, in declaration order.
func TestVarsEncoder_AllTypes(t *testing.T) {
	enc := compileEncoder(t, `
		vars {
			number $n
			account $acc
			portion $p
			monetary $m
			asset $a
			string $s
		}
		send [COIN 0] (source = @world destination = @world)
	`)

	vars, err := enc.Encode(map[string]string{
		"n":   "42",
		"acc": "alice",
		"p":   "1/4",
		"m":   "USD/2 100",
		"a":   "EUR",
		"s":   "hello",
	})
	require.NoError(t, err)

	// str slots: acc, m.asset, a, s   int slots: n, p.num, p.den, m.amount
	require.Equal(t, []string{"alice", "USD/2", "EUR", "hello"}, vars.StringsPool)
	require.Equal(t, []big.Int{
		*big.NewInt(42), *big.NewInt(1), *big.NewInt(4), *big.NewInt(100),
	}, vars.IntsPool)
}

func TestVarsEncoder_Errors(t *testing.T) {
	enc := compileEncoder(t, `
		vars { number $n account $acc }
		send [COIN 0] (source = @world destination = @world)
	`)

	_, err := enc.Encode(map[string]string{"n": "1"})
	require.ErrorContains(t, err, "missing variable: $acc")

	_, err = enc.Encode(map[string]string{"n": "not-a-number", "acc": "alice"})
	require.ErrorContains(t, err, "variable $n")
}

// Every var type validates its raw value, and the error names the variable.
func TestVarsEncoder_ErrorsPerType(t *testing.T) {
	testCases := []struct {
		typ string
		raw string
		msg string
	}{
		{"number", "4.2", `invalid number: "4.2"`},
		{"account", "not an account", `invalid account: "not an account"`},
		{"asset", "usd", `invalid asset: "usd"`},
		{"portion", "nope", "invalid format"},
		{"portion", "200%", "between 0% and 100%"},
		{"monetary", "USD/2", `invalid monetary: "USD/2"`},
		{"monetary", "usd 1", `invalid asset: "usd"`},
		{"string", "anything goes", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.typ+" "+tc.raw, func(t *testing.T) {
			enc := compileEncoder(t, `
				vars { `+tc.typ+` $v }
				send [COIN 0] (source = @world destination = @world)
			`)
			_, err := enc.Encode(map[string]string{"v": tc.raw})
			if tc.msg == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, "variable $v")
			require.ErrorContains(t, err, tc.msg)
		})
	}
}
