package compiler_test

import (
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

	enc, program, err := compiler.Compile(parsed.Value)
	require.NoError(t, err)

	vars, err := enc.Encode(map[string]string{
		"dest": "alice",
		"m":    "USD/2 100",
	})
	require.NoError(t, err)

	machine := vm.NewVm(program)
	postings, execErr := vm.Exec(machine, &vars, e2eStore{balances: map[runtime.PairKey]*big.Int{}})
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "world", Destination: "alice", Asset: "USD/2", Amount: big.NewInt(100)},
	}
	requirePostingsEqual(t, want, postings)
}

func compileEncoder(t *testing.T, src string) compiler.VarsEncoder {
	t.Helper()
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	enc, _, err := compiler.Compile(parsed.Value)
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
