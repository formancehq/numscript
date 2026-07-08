package interpreter_test

import (
	"context"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func resolve(t *testing.T, src string, vars map[string]string, store interpreter.Store) interpreter.ResolvedDependencies {
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	deps, err := interpreter.ResolveDependencies(context.Background(), store, vars, parsed.Value)
	require.NoError(t, err)
	return deps
}

func TestResolveSendSourceAndDestination(t *testing.T) {
	deps := resolve(t, `
		send [USD 10] (
			source = @alice
			destination = @bob
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "alice", Asset: "USD"}: {},
	}, deps.AccountsReads)

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "alice", Asset: "USD"}: {},
		{Account: "bob", Asset: "USD"}:   {},
	}, deps.AccountsWrites)

	require.Empty(t, deps.MetaReads)
}

func TestResolveWorldIsWrittenButNotRead(t *testing.T) {
	deps := resolve(t, `
		send [USD 10] (
			source = @world
			destination = @bob
		)
	`, nil, interpreter.StaticStore{})

	require.Empty(t, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "world", Asset: "USD"}: {},
		{Account: "bob", Asset: "USD"}:   {},
	}, deps.AccountsWrites)
}

func TestResolveBalanceAndOverdraftAreReads(t *testing.T) {
	deps := resolve(t, `
		vars {
			monetary $b = balance(@alice, USD)
			monetary $o = overdraft(@carol, EUR)
		}
		send $b (
			source = @world
			destination = @bob
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "alice", Asset: "USD"}: {},
		{Account: "carol", Asset: "EUR"}: {},
	}, deps.AccountsReads)

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "world", Asset: "USD"}: {},
		{Account: "bob", Asset: "USD"}:   {},
	}, deps.AccountsWrites)
}

func TestResolveMetaIsMetaRead(t *testing.T) {
	deps := resolve(t, `
		vars {
			account $dst = meta(@config, "recipient")
		}
		send [USD 10] (
			source = @alice
			destination = @bob
		)
	`, nil, interpreter.StaticStore{
		Meta: interpreter.AccountsMetadata{
			{Account: "config", Key: "recipient", Value: "bank"},
		},
	})

	require.Equal(t, map[interpreter.MetaDependency]struct{}{
		{Account: "config", Key: "recipient"}: {},
	}, deps.MetaReads)
}

func TestResolveMetaValueResolvesDependentAccount(t *testing.T) {
	// the meta() value is fetched from the store, so a var bound to it can be
	// used to resolve a downstream account dependency
	deps := resolve(t, `
		vars {
			account $dst = meta(@config, "recipient")
		}
		send [USD 10] (
			source = @world
			destination = $dst
		)
	`, nil, interpreter.StaticStore{
		Meta: interpreter.AccountsMetadata{
			{Account: "config", Key: "recipient", Value: "treasury"},
		},
	})

	require.Equal(t, map[interpreter.MetaDependency]struct{}{
		{Account: "config", Key: "recipient"}: {},
	}, deps.MetaReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "world", Asset: "USD"}:    {},
		{Account: "treasury", Asset: "USD"}: {},
	}, deps.AccountsWrites)
}

func TestResolveInterpolatedAccountUsesVars(t *testing.T) {
	deps := resolve(t, `
		vars {
			string $id
		}
		send [USD 10] (
			source = @users:$id
			destination = @world
		)
	`, map[string]string{"id": "alice"}, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "users:alice", Asset: "USD"}: {},
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "users:alice", Asset: "USD"}: {},
		{Account: "world", Asset: "USD"}:       {},
	}, deps.AccountsWrites)
}

func TestResolveUnboundedOverdraftSourceIsWriteOnly(t *testing.T) {
	deps := resolve(t, `
		send [USD 10] (
			source = @alice allowing unbounded overdraft
			destination = @bob
		)
	`, nil, interpreter.StaticStore{})

	// unbounded overdraft: no balance read, but still a write
	require.Empty(t, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "alice", Asset: "USD"}: {},
		{Account: "bob", Asset: "USD"}:   {},
	}, deps.AccountsWrites)
}

func TestResolveInorderSource(t *testing.T) {
	deps := resolve(t, `
		send [USD 30] (
			source = {
				@a
				@b
			}
			destination = @c
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
		{Account: "b", Asset: "USD"}: {},
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
		{Account: "b", Asset: "USD"}: {},
		{Account: "c", Asset: "USD"}: {},
	}, deps.AccountsWrites)
}

func TestResolveAllotmentSourceAndDestination(t *testing.T) {
	deps := resolve(t, `
		send [USD 100] (
			source = {
				50% from @a
				remaining from @b
			}
			destination = {
				50% to @x
				remaining to @y
			}
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
		{Account: "b", Asset: "USD"}: {},
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
		{Account: "b", Asset: "USD"}: {},
		{Account: "x", Asset: "USD"}: {},
		{Account: "y", Asset: "USD"}: {},
	}, deps.AccountsWrites)
}

func TestResolveBalanceOfWorldIsRead(t *testing.T) {
	deps := resolve(t, `
		vars {
			monetary $w = balance(@world, USD)
		}
		send [USD 10] (
			source = @a
			destination = @b
		)
	`, nil, interpreter.StaticStore{})

	_, ok := deps.AccountsReads[interpreter.AccountDependency{Account: "world", Asset: "USD"}]
	require.True(t, ok, "balance(@world) must be recorded as a read")
}

func TestResolveBalanceInsideCap(t *testing.T) {
	deps := resolve(t, `
		send [USD 10] (
			source = max balance(@reserve, USD) from @a
			destination = @b
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "reserve", Asset: "USD"}: {},
		{Account: "a", Asset: "USD"}:       {},
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
		{Account: "b", Asset: "USD"}: {},
	}, deps.AccountsWrites)
}

func TestResolveOneofSource(t *testing.T) {
	// oneof picks one branch at runtime, but statically every branch is a dependency
	deps := resolve(t, `
		send [USD 10] (
			source = oneof {
				@a
				@b
			}
			destination = @c
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
		{Account: "b", Asset: "USD"}: {},
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
		{Account: "b", Asset: "USD"}: {},
		{Account: "c", Asset: "USD"}: {},
	}, deps.AccountsWrites)
}

func TestResolveBoundedOverdraftIsRead(t *testing.T) {
	deps := resolve(t, `
		send [USD 10] (
			source = @a allowing overdraft up to [USD 100]
			destination = @b
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
		{Account: "b", Asset: "USD"}: {},
	}, deps.AccountsWrites)
}

func TestResolveOverdraftBoundBalanceIsRead(t *testing.T) {
	// a balance() inside an overdraft bound is recorded through the store
	deps := resolve(t, `
		send [USD 10] (
			source = @a allowing overdraft up to balance(@reserve, USD)
			destination = @b
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}:       {},
		{Account: "reserve", Asset: "USD"}: {},
	}, deps.AccountsReads)
}

func TestResolveColoredSource(t *testing.T) {
	deps := resolve(t, `
		send [COIN 100] (
			source = @src \ "RED"
			destination = @dest
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "src", Asset: "COIN", Color: "RED"}: {},
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "src", Asset: "COIN", Color: "RED"}: {},
		{Account: "dest", Asset: "COIN"}:              {},
	}, deps.AccountsWrites)
}

func TestResolveSameAccountDifferentColors(t *testing.T) {
	// the same account with different colors are distinct dependencies
	deps := resolve(t, `
		send [COIN 100] (
			source = {
				@src \ "RED"
				@src \ "BLUE"
				@src
			}
			destination = @dest
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "src", Asset: "COIN", Color: "RED"}:  {},
		{Account: "src", Asset: "COIN", Color: "BLUE"}: {},
		{Account: "src", Asset: "COIN"}:                {},
	}, deps.AccountsReads)
}

func TestResolveDeeplyNestedSources(t *testing.T) {
	deps := resolve(t, `
		send [USD 100] (
			source = oneof {
				max [USD 10] from @a
				{
					@b
					@world
				}
				{
					50% from @c
					remaining from @d allowing unbounded overdraft
				}
			}
			destination = @dest
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}: {},
		{Account: "b", Asset: "USD"}: {},
		{Account: "c", Asset: "USD"}: {},
		// @world (unbounded) and @d (unbounded overdraft) are writes only
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "a", Asset: "USD"}:     {},
		{Account: "b", Asset: "USD"}:     {},
		{Account: "world", Asset: "USD"}: {},
		{Account: "c", Asset: "USD"}:     {},
		{Account: "d", Asset: "USD"}:     {},
		{Account: "dest", Asset: "USD"}:  {},
	}, deps.AccountsWrites)
}

func TestResolveSetAccountMetaIsMetaWrite(t *testing.T) {
	deps := resolve(t, `
		set_account_meta(@alice, "kyc", "verified")
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.MetaDependency]struct{}{
		{Account: "alice", Key: "kyc"}: {},
	}, deps.MetaWrites)
	require.Empty(t, deps.MetaReads)
}

func TestResolveScalingIsNotSupported(t *testing.T) {
	parsed := parser.Parse(`
		send [USD 10] (
			source = @alice with scaling through @pool
			destination = @bob
		)
	`)
	require.Empty(t, parsed.Errors)

	_, err := interpreter.ResolveDependencies(context.Background(), interpreter.StaticStore{}, nil, parsed.Value)
	require.ErrorIs(t, err, interpreter.ErrScalingNotSupported)
}

func TestResolveSaveIsRead(t *testing.T) {
	deps := resolve(t, `
		save [USD 10] from @alice
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "alice", Asset: "USD"}: {},
	}, deps.AccountsReads)
	require.Empty(t, deps.AccountsWrites)
}
