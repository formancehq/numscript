package interpreter_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

// queryLog wraps a store and records every balance/metadata actually queried,
// so a test can assert that resolution is a superset of what execution touches.
type queryLog struct {
	inner    interpreter.Store
	balances map[interpreter.AccountDependency]struct{}
	meta     map[interpreter.MetaDependency]struct{}
}

func (q *queryLog) GetBalances(ctx context.Context, query interpreter.BalanceQuery) (interpreter.Balances, error) {
	for _, item := range query {
		q.balances[interpreter.AccountDependency{Account: item.Account, Scope: item.Scope, Color: item.Color, Asset: item.Asset}] = struct{}{}
	}
	return q.inner.GetBalances(ctx, query)
}

func (q *queryLog) GetAccountsMetadata(ctx context.Context, query interpreter.MetadataQuery) (interpreter.AccountsMetadata, error) {
	for _, item := range query {
		for _, key := range item.Keys {
			q.meta[interpreter.MetaDependency{Account: item.Account, Scope: item.Scope, Key: key}] = struct{}{}
		}
	}
	return q.inner.GetAccountsMetadata(ctx, query)
}

// TestResolveDependenciesCoversRuntime checks the soundness property: for a
// script that executes successfully, every balance/metadata the interpreter
// queries is a resolved read, and every account a posting touches is a resolved
// write. Resolution may over-approximate (e.g. both branches of a oneof), so
// these are subset checks.
func TestResolveDependenciesCoversRuntime(t *testing.T) {
	allFlags := map[string]struct{}{}
	for _, f := range flags.AllFlags {
		allFlags[f] = struct{}{}
	}

	testCases := []struct {
		name     string
		src      string
		vars     map[string]string
		balances interpreter.Balances
		meta     interpreter.AccountsMetadata
	}{
		{
			name: "simple send",
			src: `
				send [USD 10] (
					source = @a
					destination = @b
				)
			`,
			balances: interpreter.Balances{{Account: "a", Asset: "USD", Amount: big.NewInt(100)}},
		},
		{
			name: "balance var origin used as sent value",
			src: `
				vars { monetary $x = balance(@t, USD) }
				send $x (
					source = @t
					destination = @o
				)
			`,
			balances: interpreter.Balances{{Account: "t", Asset: "USD", Amount: big.NewInt(100)}},
		},
		{
			name: "bounded overdraft",
			src: `
				send [USD 50] (
					source = @a allowing overdraft up to [USD 100]
					destination = @b
				)
			`,
			balances: interpreter.Balances{{Account: "a", Asset: "USD", Amount: big.NewInt(0)}},
		},
		{
			name: "allotment source",
			src: `
				send [USD 100] (
					source = {
						50% from @a
						50% from @b
					}
					destination = @c
				)
			`,
			balances: interpreter.Balances{
				{Account: "a", Asset: "USD", Amount: big.NewInt(50)},
				{Account: "b", Asset: "USD", Amount: big.NewInt(50)},
			},
		},
		{
			name: "balance inside a cap",
			src: `
				send [USD 5] (
					source = max balance(@r, USD) from @funder
					destination = @out
				)
			`,
			balances: interpreter.Balances{
				{Account: "r", Asset: "USD", Amount: big.NewInt(10)},
				{Account: "funder", Asset: "USD", Amount: big.NewInt(100)},
			},
		},
		{
			name: "meta-derived destination",
			src: `
				vars { account $d = meta(@config, "recipient") }
				send [USD 10] (
					source = @world
					destination = $d
				)
			`,
			meta: interpreter.AccountsMetadata{{Account: "config", Key: "recipient", Value: "out"}},
		},
		{
			name: "destination allotment",
			src: `
				send [USD 100] (
					source = @world
					destination = {
						50% to @a
						remaining to @b
					}
				)
			`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed := parser.Parse(tc.src)
			require.Empty(t, parsed.Errors)

			static := interpreter.StaticStore{Balances: tc.balances, Meta: tc.meta}
			log := &queryLog{
				inner:    static,
				balances: map[interpreter.AccountDependency]struct{}{},
				meta:     map[interpreter.MetaDependency]struct{}{},
			}

			result, runErr := interpreter.RunProgram(context.Background(), parsed.Value, tc.vars, log, allFlags)
			require.Nil(t, runErr)

			deps, err := interpreter.ResolveDependencies(context.Background(), static, tc.vars, parsed.Value)
			require.NoError(t, err)

			for dep := range log.balances {
				require.Contains(t, deps.AccountsReads, dep, "queried balance not resolved as a read")
			}
			for m := range log.meta {
				require.Contains(t, deps.MetaReads, m, "queried metadata not resolved as a meta read")
			}

			// posting accounts must be resolved writes (color isn't statically known for destinations)
			writes := map[[3]string]struct{}{}
			for w := range deps.AccountsWrites {
				writes[[3]string{w.Account, w.Scope, w.Asset}] = struct{}{}
			}
			for _, p := range result.Postings {
				require.Contains(t, writes, [3]string{p.Source, p.SourceScope, p.Asset}, "posting source not resolved as a write")
				require.Contains(t, writes, [3]string{p.Destination, p.DestinationScope, p.Asset}, "posting destination not resolved as a write")
			}
		})
	}
}

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
	// the destination is credited in the source's color
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "src", Asset: "COIN", Color: "RED"}:  {},
		{Account: "dest", Asset: "COIN", Color: "RED"}: {},
	}, deps.AccountsWrites)
}

func TestResolveMixedColorSourceWidensDestination(t *testing.T) {
	// funds from a colored and an uncolored source can both reach the
	// destination, so it's recorded as a write in every source color
	deps := resolve(t, `
		send [COIN 10] (
			source = {
				@s1 \ "C"
				@s2
			}
			destination = @dest
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "s1", Asset: "COIN", Color: "C"}:   {},
		{Account: "s2", Asset: "COIN"}:               {},
		{Account: "dest", Asset: "COIN", Color: "C"}: {},
		{Account: "dest", Asset: "COIN"}:             {},
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

func TestResolveDestinationInorder(t *testing.T) {
	deps := resolve(t, `
		send [USD 100] (
			source = @world
			destination = {
				max [USD 20] to @a
				remaining to @b
			}
		)
	`, nil, interpreter.StaticStore{})

	require.Empty(t, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "world", Asset: "USD"}: {},
		{Account: "a", Asset: "USD"}:     {},
		{Account: "b", Asset: "USD"}:     {},
	}, deps.AccountsWrites)
}

func TestResolveDestinationOneof(t *testing.T) {
	deps := resolve(t, `
		send [USD 100] (
			source = @world
			destination = oneof {
				max [USD 20] to @a
				remaining to @b
			}
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "world", Asset: "USD"}: {},
		{Account: "a", Asset: "USD"}:     {},
		{Account: "b", Asset: "USD"}:     {},
	}, deps.AccountsWrites)
}

func TestResolveDestinationRemainingKept(t *testing.T) {
	deps := resolve(t, `
		send [USD 100] (
			source = @world
			destination = {
				max [USD 20] to @a
				remaining kept
			}
		)
	`, nil, interpreter.StaticStore{})

	// "remaining kept" holds no account
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "world", Asset: "USD"}: {},
		{Account: "a", Asset: "USD"}:     {},
	}, deps.AccountsWrites)
}

func TestResolveBalanceInsideDestinationCap(t *testing.T) {
	deps := resolve(t, `
		send [USD 10] (
			source = @funder
			destination = {
				max balance(@reserve, USD) to @a
				remaining to @b
			}
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "funder", Asset: "USD"}:  {},
		{Account: "reserve", Asset: "USD"}: {},
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "funder", Asset: "USD"}: {},
		{Account: "a", Asset: "USD"}:      {},
		{Account: "b", Asset: "USD"}:      {},
	}, deps.AccountsWrites)
}

func TestResolveNestedDestinations(t *testing.T) {
	deps := resolve(t, `
		send [USD 100] (
			source = @world
			destination = {
				max [USD 50] to {
					max [USD 10] to @a
					remaining to @b
				}
				remaining to @c
			}
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "world", Asset: "USD"}: {},
		{Account: "a", Asset: "USD"}:     {},
		{Account: "b", Asset: "USD"}:     {},
		{Account: "c", Asset: "USD"}:     {},
	}, deps.AccountsWrites)
}

func TestResolveScopedSource(t *testing.T) {
	deps := resolve(t, `
		send [USD 10] (
			source = scoped(@treasury, "reserve")
			destination = @out
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "treasury", Scope: "reserve", Asset: "USD"}: {},
	}, deps.AccountsReads)
	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "treasury", Scope: "reserve", Asset: "USD"}: {},
		{Account: "out", Asset: "USD"}:                        {},
	}, deps.AccountsWrites)
}

func TestResolveScopedDestination(t *testing.T) {
	deps := resolve(t, `
		send [USD 10] (
			source = @world
			destination = scoped(@out, "escrow")
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "world", Asset: "USD"}:                {},
		{Account: "out", Scope: "escrow", Asset: "USD"}: {},
	}, deps.AccountsWrites)
}

func TestResolveScopedBalanceRead(t *testing.T) {
	deps := resolve(t, `
		vars { monetary $b = balance(scoped(@treasury, "reserve"), USD) }
		send $b (
			source = @world
			destination = @out
		)
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[interpreter.AccountDependency]struct{}{
		{Account: "treasury", Scope: "reserve", Asset: "USD"}: {},
	}, deps.AccountsReads)
}

func TestResolveScopedMetaRead(t *testing.T) {
	deps := resolve(t, `
		vars { account $d = meta(scoped(@config, "settings"), "recipient") }
		send [USD 10] (
			source = @world
			destination = @out
		)
	`, nil, interpreter.StaticStore{
		Meta: interpreter.AccountsMetadata{
			{Account: "config", Scope: "settings", Key: "recipient", Value: "out"},
		},
	})

	require.Equal(t, map[interpreter.MetaDependency]struct{}{
		{Account: "config", Scope: "settings", Key: "recipient"}: {},
	}, deps.MetaReads)
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

func TestResolveSetTxMetaIsTxMetaWrite(t *testing.T) {
	deps := resolve(t, `
		set_tx_meta("priority", "high")
	`, nil, interpreter.StaticStore{})

	require.Equal(t, map[string]struct{}{
		"priority": {},
	}, deps.TxMetaWrites)
	require.Empty(t, deps.MetaWrites)
}

func TestResolveUnknownStatementFnErrors(t *testing.T) {
	parsed := parser.Parse(`unknown_fn(@a, "x")`)
	require.Empty(t, parsed.Errors)

	_, err := interpreter.ResolveDependencies(context.Background(), interpreter.StaticStore{}, nil, parsed.Value)
	require.IsType(t, interpreter.UnboundFunctionErr{}, err)
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
