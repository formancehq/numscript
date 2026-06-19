package interpreter

import (
	"context"
	"math/big"
	"testing"

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

func readVolume(b Balances, account, asset string) *big.Int {
	for _, row := range b {
		if row.Account == account && row.Asset == asset {
			return row.Amount
		}
	}
	return nil
}

func hasWrite(q BalanceQuery, account, asset string) bool {
	for _, item := range q {
		if item.Account == account && item.Asset == asset {
			return true
		}
	}
	return false
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
			{Account: "alice", Asset: "USD/2", Amount: big.NewInt(500)},
		},
	})

	require.Equal(t, big.NewInt(500), readVolume(deps.Reads.Volumes, "alice", "USD/2"))

	require.True(t, hasWrite(deps.Writes.Volumes, "alice", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "bob", "USD/2"))
}

func TestResolveDependencies_WorldSource(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		send [USD/2 100] (
			source = @world
			destination = @bob
		)
	`, nil, StaticStore{})

	require.Empty(t, deps.Reads.Volumes)
	require.True(t, hasWrite(deps.Writes.Volumes, "world", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "bob", "USD/2"))
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

	require.Equal(t, "treasury", deps.Reads.Metadata["config"]["default_dest"])
	require.True(t, hasWrite(deps.Writes.Volumes, "treasury", "USD/2"))
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
			{Account: "checking", Asset: "USD/2", Amount: big.NewInt(50)},
			{Account: "savings", Asset: "USD/2", Amount: big.NewInt(300)},
		},
	})

	require.NotNil(t, readVolume(deps.Reads.Volumes, "checking", "USD/2"))
	require.NotNil(t, readVolume(deps.Reads.Volumes, "savings", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "checking", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "savings", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "merchant", "USD/2"))
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
			{Account: "users:alice", Asset: "EUR/2", Amount: big.NewInt(5000)},
		},
	})

	require.NotNil(t, readVolume(deps.Reads.Volumes, "users:alice", "EUR/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "users:alice", "EUR/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "dest", "EUR/2"))
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
			{Account: "src", Asset: "USD/2", Amount: big.NewInt(750)},
		},
	})

	require.Equal(t, big.NewInt(750), readVolume(deps.Reads.Volumes, "src", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "src", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "dest", "USD/2"))
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
			{Account: "b", Asset: "EUR/2", Amount: big.NewInt(200)},
		},
	})

	require.Nil(t, readVolume(deps.Reads.Volumes, "world", "USD/2"))
	require.NotNil(t, readVolume(deps.Reads.Volumes, "b", "EUR/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "a", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "c", "EUR/2"))
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

	require.Empty(t, deps.Reads.Metadata, "set_account_meta should not produce metadata reads")
	require.True(t, hasWrite(deps.Writes.Volumes, "alice", "USD/2"))
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

	require.Equal(t, "destination", deps.Reads.Metadata["config"]["key_name"])
	require.Equal(t, "treasury", deps.Reads.Metadata["routing"]["destination"])
	require.True(t, hasWrite(deps.Writes.Volumes, "treasury", "USD/2"))
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
			{Account: "src", Asset: "USD/2", Amount: big.NewInt(999)},
		},
	})

	require.Equal(t, big.NewInt(999), readVolume(deps.Reads.Volumes, "src", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "src", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "dest", "USD/2"))
}

func TestResolveDependencies_EmptyReads(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
		send [USD/2 100] (
			source = @world
			destination = @dest
		)
	`, nil, StaticStore{})

	require.Empty(t, deps.Reads.Volumes)
	require.Empty(t, deps.Reads.Metadata)
	require.True(t, hasWrite(deps.Writes.Volumes, "world", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "dest", "USD/2"))
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

	deps := resolveTest(t,
		script,
		map[string]string{"s1": "source1"},
		StaticStore{
			Balances: Balances{
				{Account: "source1", Asset: "COIN", Amount: big.NewInt(123)},
				{Account: "source2", Asset: "COIN", Amount: big.NewInt(456)},
				{Account: "source3", Asset: "COIN", Amount: big.NewInt(55)},
				{Account: "account_that_needs_balance", Asset: "USD/2", Amount: big.NewInt(42)},
			},
			Meta: AccountsMetadata{
				"account_that_needs_meta": {"k": "source2"},
			},
		})

	require.Equal(t, big.NewInt(123), readVolume(deps.Reads.Volumes, "source1", "COIN"))
	require.Equal(t, big.NewInt(456), readVolume(deps.Reads.Volumes, "source2", "COIN"))
	require.Equal(t, big.NewInt(55), readVolume(deps.Reads.Volumes, "source3", "COIN"))
	require.Equal(t, big.NewInt(42), readVolume(deps.Reads.Volumes, "account_that_needs_balance", "USD/2"))

	require.Equal(t, AccountsMetadata{
		"account_that_needs_meta": {"k": "source2"},
	}, deps.Reads.Metadata)

	// Writes is a conservative over-approximation: every account that appears
	// as a source or destination is listed, even if the actual run would not
	// touch all of them.
	for _, acc := range []string{"source1", "source2", "source3", "world", "dest"} {
		require.True(t, hasWrite(deps.Writes.Volumes, acc, "COIN"), "expected %s in writes", acc)
	}
}

func TestResolveDependencies_MidScriptBalance(t *testing.T) {
	t.Parallel()

	deps := resolveTest(t, `
#![feature("experimental-mid-script-function-call")]
send [USD/2 100] (
  source = @world
  destination = @acc
)
send balance(@acc, USD/2) (
  source = @acc
  destination = @dest
)
`, nil, StaticStore{})

	// The balance call hits the store during preload, recording acc/USD/2 = 0.
	require.Equal(t, big.NewInt(0), readVolume(deps.Reads.Volumes, "acc", "USD/2"))

	require.True(t, hasWrite(deps.Writes.Volumes, "world", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "acc", "USD/2"))
	require.True(t, hasWrite(deps.Writes.Volumes, "dest", "USD/2"))
}
