package interpreter

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func resolveTest(t *testing.T, script string, vars map[string]string, store Store) *ResolvedDependencies {
	t.Helper()
	parsed := parser.Parse(script)
	require.Empty(t, parsed.Errors, "script should parse without errors")

	deps, err := ResolveDependencies(context.Background(), parsed.Value, vars, store, ResolveDependenciesOptions{})
	require.NoError(t, err)
	require.NotNil(t, deps)

	return deps
}

func TestResolveDependencies_SimpleTransfer(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		send [USD/2 100] (
			source = @alice
			destination = @bob
		)
	`, nil, StaticStore{
		Balances: Balances{
			"alice": AccountBalance{"USD/2": big.NewInt(500)},
		},
	})

	require.Contains(t, deps.Volumes, "alice")
	require.Contains(t, deps.Volumes["alice"], "USD/2")
	require.Equal(t, big.NewInt(500), deps.Volumes["alice"]["USD/2"])
}

func TestResolveDependencies_WorldSource(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		send [USD/2 100] (
			source = @world
			destination = @bob
		)
	`, nil, StaticStore{})

	require.Empty(t, deps.Volumes)
}

func TestResolveDependencies_MetaCall(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		vars {
			account $dest = meta(@config, "default_dest")
		}
		send [USD/2 100] (
			source = @world
			destination = $dest
		)
	`, nil, StaticStore{
		Meta: AccountsMetadata{
			"config": AccountMetadata{"default_dest": "treasury"},
		},
	})

	require.Contains(t, deps.Metadata, "config")
	require.Equal(t, "treasury", deps.Metadata["config"]["default_dest"])
}

func TestResolveDependencies_MultipleSources(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		send [USD/2 200] (
			source = {
				@checking
				@savings
			}
			destination = @merchant
		)
	`, nil, StaticStore{
		Balances: Balances{
			"checking": AccountBalance{"USD/2": big.NewInt(50)},
			"savings":  AccountBalance{"USD/2": big.NewInt(300)},
		},
	})

	require.Contains(t, deps.Volumes, "checking")
	require.Contains(t, deps.Volumes, "savings")
}

func TestResolveDependencies_Variables(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		vars {
			account $src
			monetary $amount
		}
		send $amount (
			source = $src
			destination = @dest
		)
	`, map[string]string{
		"src":    "users:alice",
		"amount": "EUR/2 1000",
	}, StaticStore{
		Balances: Balances{
			"users:alice": AccountBalance{"EUR/2": big.NewInt(5000)},
		},
	})

	require.Contains(t, deps.Volumes, "users:alice")
	require.Contains(t, deps.Volumes["users:alice"], "EUR/2")
}

func TestResolveDependencies_BalanceFunction(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		vars {
			monetary $bal = balance(@src, USD/2)
		}
		send $bal (
			source = @src
			destination = @dest
		)
	`, nil, StaticStore{
		Balances: Balances{
			"src": AccountBalance{"USD/2": big.NewInt(750)},
		},
	})

	require.Contains(t, deps.Volumes, "src")
	require.Equal(t, big.NewInt(750), deps.Volumes["src"]["USD/2"])
}

func TestResolveDependencies_MultipleSends(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		send [USD/2 50] (
			source = @world
			destination = @a
		)
		send [EUR/2 100] (
			source = @b
			destination = @c
		)
	`, nil, StaticStore{
		Balances: Balances{
			"b": AccountBalance{"EUR/2": big.NewInt(200)},
		},
	})

	require.NotContains(t, deps.Volumes, "world")
	require.Contains(t, deps.Volumes, "b")
}

func TestResolveDependencies_SetAccountMeta(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		set_account_meta(@alice, "status", "active")
		send [USD/2 100] (
			source = @world
			destination = @alice
		)
	`, nil, StaticStore{})

	require.Empty(t, deps.Metadata, "set_account_meta should not produce metadata reads")
}

func TestResolveDependencies_MetaChain(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		vars {
			string $key = meta(@config, "key_name")
			account $dest = meta(@routing, $key)
		}
		send [USD/2 100] (
			source = @world
			destination = $dest
		)
	`, nil, StaticStore{
		Meta: AccountsMetadata{
			"config":  AccountMetadata{"key_name": "destination"},
			"routing": AccountMetadata{"destination": "treasury"},
		},
	})

	require.Contains(t, deps.Metadata, "config")
	require.Equal(t, "destination", deps.Metadata["config"]["key_name"])
	require.Contains(t, deps.Metadata, "routing")
	require.Equal(t, "treasury", deps.Metadata["routing"]["destination"])
}

func TestResolveDependencies_SendAll(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		send [USD/2 *] (
			source = @src
			destination = @dest
		)
	`, nil, StaticStore{
		Balances: Balances{
			"src": AccountBalance{"USD/2": big.NewInt(999)},
		},
	})

	require.Contains(t, deps.Volumes, "src")
	require.Equal(t, big.NewInt(999), deps.Volumes["src"]["USD/2"])
}

func TestResolveDependencies_EmptyReads(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		send [USD/2 100] (
			source = @world
			destination = @dest
		)
	`, nil, StaticStore{})

	require.Empty(t, deps.Volumes)
	require.Empty(t, deps.Metadata)
}

func TestResolveDependencies_ForbiddenFlag(t *testing.T) {
	t.Parallel()

	script := `
#![feature("experimental-mid-script-function-call")]
send [USD/2 100] (
  source = @world
  destination = @acc
)
send balance(@acc, USD/2) (
  source = @acc
  destination = @dest
)
`
	parsed := parser.Parse(script)
	require.Empty(t, parsed.Errors)

	_, err := ResolveDependencies(context.Background(), parsed.Value, nil, StaticStore{}, ResolveDependenciesOptions{
		ForbiddenFlags: map[string]struct{}{
			flags.ExperimentalMidScriptFunctionCall: {},
		},
	})
	require.Error(t, err)

	var forbiddenErr ForbiddenFeature
	require.ErrorAs(t, err, &forbiddenErr)
	require.Equal(t, flags.ExperimentalMidScriptFunctionCall, forbiddenErr.FlagName)
}

func TestResolveDependencies_Nested(t *testing.T) {
	t.Parallel()

	script := `vars {
	account $s1
	account $s2 = meta(@account_that_needs_meta, "k")
	number $b = balance(@account_that_needs_balance, USD/2)
}

send [COIN 100] (
	source = {
		$s1
		$s2
		@source3
		@world
	}
  	destination = @dest
)
`

	parsed := parser.Parse(script)
	require.Empty(t, parsed.Errors)

	deps := resolveTest(t,
		script,
		map[string]string{"s1": "source1"},
		StaticStore{
			Balances: Balances{
				"source1": {
					"COIN": big.NewInt(123),
				},
				"source2": {
					"COIN": big.NewInt(456),
				},
				"source3": {
					"COIN": big.NewInt(55),
				},
				"account_that_needs_balance": {
					"USD/2": big.NewInt(42),
				},
			},
			Meta: AccountsMetadata{"account_that_needs_meta": {"k": "source2"}},
		})

	require.Equal(t, deps.Volumes, map[string]map[string]*big.Int{
		"source1": {
			"COIN": big.NewInt(123),
		},
		"source2": {
			"COIN": big.NewInt(456),
		},
		"source3": {
			"COIN": big.NewInt(55),
		},
		"account_that_needs_balance": {
			"USD/2": big.NewInt(42),
		},
	})

}
