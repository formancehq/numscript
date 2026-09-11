package numscript_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/stretchr/testify/require"
)

func TestGetVars(t *testing.T) {
	parseResult := numscript.Parse(`
	vars {
		monetary $mon
		account $acc
		account $acc2
		
		monetary $do_not_include_in_output = balance(@acc, USD/2)
	}
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	require.Equal(t,
		map[string]string{
			"mon":  "monetary",
			"acc":  "account",
			"acc2": "account",
		},
		parseResult.GetNeededVariables(),
	)

}

func TestGetVarsEmpty(t *testing.T) {
	parseResult := numscript.Parse(`
	vars {}
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")
	require.Equal(t,
		map[string]string{},
		parseResult.GetNeededVariables(),
	)
}

func TestGetVarsNovars(t *testing.T) {
	parseResult := numscript.Parse(``)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")
	require.Equal(t,
		map[string]string{},
		parseResult.GetNeededVariables(),
	)
}

func TestRunRejectsParseErrors(t *testing.T) {
	parseResult := numscript.Parse(`send [COIN 100] (`)
	require.NotEmpty(t, parseResult.GetParsingErrors())

	_, err := parseResult.Run(context.Background(), nil, interpreter.StaticStore{})
	require.Error(t, err)
	require.Equal(t, parseResult.GetParsingErrors()[0].Error(), err.Error())
}

func TestDoNotGetWorldBalance(t *testing.T) {
	parseResult := numscript.Parse(`send [COIN 100] (
	source = @world
  	destination = @dest
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")
	store := ObservableStore{
		StaticStore: interpreter.StaticStore{
			Balances: nil,
			Meta:     interpreter.AccountsMetadata{},
		},
	}
	_, err := parseResult.Run(context.Background(), numscript.VariablesMap{},
		&store,
	)
	require.Nil(t, err)

	require.Equal(t,
		([]numscript.BalanceQuery)(nil),
		store.GetBalancesCalls)
}

func TestGetBalancesInorder(t *testing.T) {
	parseResult := numscript.Parse(`vars {
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
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{
		StaticStore: interpreter.StaticStore{
			Balances: interpreter.Balances{},
			Meta:     interpreter.AccountsMetadata{{Account: "account_that_needs_meta", Key: "k", Value: "source2"}},
		},
	}
	_, err := parseResult.Run(context.Background(), numscript.VariablesMap{
		"s1": "source1",
	},
		&store,
	)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.MetadataQuery{
			{
				{Account: "account_that_needs_meta", Keys: []string{"k"}},
			},
		},
		store.GetMetadataCalls)

	require.Equal(t,
		[]numscript.BalanceQuery{
			// TODO maybe those calls can be batched together
			{
				// this is required by the balance() call
				{Account: "account_that_needs_balance", Asset: "USD/2"},
			},
			{
				// this is defined in the variables
				{Account: "source1", Asset: "COIN"},

				// this is defined in account metadata
				{Account: "source2", Asset: "COIN"},

				// this appears as literal
				{Account: "source3", Asset: "COIN"},
			},
		},
		store.GetBalancesCalls)
}

func TestGetBalancesOneof(t *testing.T) {
	parseResult := numscript.Parse(`
send [COIN 100] (
	source = oneof {
		@a
		@b
		@world
	}
  	destination = @dest
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{
		StaticStore: interpreter.StaticStore{
			Balances: interpreter.Balances{},
		},
	}
	_, err := parseResult.RunWithFeatureFlags(context.Background(), numscript.VariablesMap{
		"s1": "source1",
	},
		&store,
		map[string]struct{}{flags.ExperimentalOneofFeatureFlag: {}},
	)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				{Account: "a", Asset: "COIN"},
				{Account: "b", Asset: "COIN"},
			},
		},
		store.GetBalancesCalls)
}

func TestDoNotGetBalancesTwice(t *testing.T) {
	parseResult := numscript.Parse(`send [COIN 100] (
	source = {
		@alice
		@alice
		@world
	}
  	destination = @dest
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{
		StaticStore: interpreter.StaticStore{
			Balances: interpreter.Balances{},
		},
	}
	_, err := parseResult.Run(context.Background(), numscript.VariablesMap{}, &store)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				{Account: "alice", Asset: "COIN"},
			},
		},
		store.GetBalancesCalls)
}

func TestGetBalancesAllotment(t *testing.T) {
	parseResult := numscript.Parse(`send [COIN 100] (
	source = {
		1/2 from @a
		remaining from @b
	}
  	destination = @dest
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{
		StaticStore: interpreter.StaticStore{
			Balances: interpreter.Balances{
				{Account: "a", Asset: "COIN", Amount: big.NewInt(10000)},
				{Account: "b", Asset: "COIN", Amount: big.NewInt(10000)},
			},
		},
	}

	_, err := parseResult.Run(context.Background(),
		numscript.VariablesMap{},
		&store,
	)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				{Account: "a", Asset: "COIN"},
				{Account: "b", Asset: "COIN"},
			},
		},
		store.GetBalancesCalls)
}

func TestGetBalancesOverdraft(t *testing.T) {
	parseResult := numscript.Parse(`send [COIN 100] (
	source = {
		@a allowing overdraft up to [COIN 10]
		@b allowing unbounded overdraft
	}
  	destination = @dest
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{}

	_, err := parseResult.Run(context.Background(), interpreter.VariablesMap{}, &store)
	require.Nil(t, err)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				{Account: "a", Asset: "COIN"},
			},
		},
		store.GetBalancesCalls)
}

func TestDoNotFetchBalanceTwice(t *testing.T) {
	parseResult := numscript.Parse(`vars { monetary $v = balance(@src, COIN) }

	send $v (
		source = @src
		destination = @dest
	)`)

	store := ObservableStore{}
	parseResult.Run(context.Background(), nil, &store)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				{Account: "src", Asset: "COIN"},
			},
		},
		store.GetBalancesCalls,
	)

}

func TestDoNotFetchBalanceTwice2(t *testing.T) {
	// same test as before, but this time the second batch is not empty
	parseResult := numscript.Parse(`vars { monetary $v = balance(@src1, COIN) }

	send $v (
		source = {
			@src1
			@src2
		}
		destination = @dest
	)`)

	store := ObservableStore{}
	parseResult.Run(context.Background(), nil, &store)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				{Account: "src1", Asset: "COIN"},
			},
			{
				{Account: "src2", Asset: "COIN"},
			},
		},
		store.GetBalancesCalls,
	)

}

func TestDoNotFetchBalanceTwice3(t *testing.T) {
	// same test as before, but this time the second batch requires a _different asset_
	parseResult := numscript.Parse(`vars { monetary $eur_m = balance(@src, EUR/2) }

	
	send [USD/2 100] (
		// note here we are fetching a different currency
		source = @src
		destination = @dest
	)
`)

	store := ObservableStore{}
	parseResult.Run(context.Background(), nil, &store)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				{Account: "src", Asset: "EUR/2"},
			},
			{
				{Account: "src", Asset: "USD/2"},
			},
		},
		store.GetBalancesCalls,
	)

}

func TestQueryBalanceErr(t *testing.T) {
	parseResult := numscript.Parse(`send [COIN 100] (
	source = @src
  	destination = @dest
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	_, err := parseResult.Run(context.Background(), interpreter.VariablesMap{}, &ErrorStore{})
	require.IsType(t, err, interpreter.QueryBalanceError{})
}

func TestMetadataFetchErr(t *testing.T) {
	parseResult := numscript.Parse(`vars {
	number $x = meta(@acc, "k")
}`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	_, err := parseResult.Run(context.Background(), interpreter.VariablesMap{}, &ErrorStore{})
	require.IsType(t, err, interpreter.QueryMetadataError{})
}

func TestBalanceFunctionErr(t *testing.T) {
	parseResult := numscript.Parse(`vars {
	monetary $x = balance(@acc, USD/2)
}`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	_, err := parseResult.Run(context.Background(), interpreter.VariablesMap{}, &ErrorStore{})
	require.IsType(t, err, interpreter.QueryBalanceError{})
}

func TestSaveQuery(t *testing.T) {
	parseResult := numscript.Parse(`
save [USD/2 10] from @alice

send [USD/2 30] (
	source = {
		@alice
		@world
	}
	destination = @bob
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{}
	parseResult.Run(context.Background(), nil, &store)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				{Account: "alice", Asset: "USD/2"},
			},
		},
		store.GetBalancesCalls,
	)

}

func TestMidscriptBalance(t *testing.T) {
	parseResult := numscript.Parse(`
send [USD/2 100] (
	source = @bob allowing unbounded overdraft
	destination = @alice
)

set_tx_meta(
	"k",
	balance(@alice, USD/2)
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{
		StaticStore: interpreter.StaticStore{
			Balances: interpreter.Balances{
				{Account: "alice", Asset: "USD/2", Amount: big.NewInt(20)},
			},
		},
	}
	res, err := parseResult.RunWithFeatureFlags(context.Background(), nil, &store, map[string]struct{}{
		flags.ExperimentalMidScriptFunctionCall: {},
	})
	require.Nil(t, err)

	require.Equal(t, interpreter.Metadata{
		"k": "USD/2 100",
	}, res.Metadata)

	require.Equal(t,
		[]numscript.BalanceQuery(nil),
		store.GetBalancesCalls,
	)

}

func TestInterleavedBalanceBatching(t *testing.T) {
	parseResult := numscript.Parse(`
vars {
	account $a2 = meta(@a, "k") // -> @a2
}

send [USD/2 10] (
  source = {
		// balance(@a2, USD/2) -> [USD/2 1]
		max balance($a2, USD/2) from @a
		@world
	}
  destination = @b
)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := ObservableStore{
		StaticStore: interpreter.StaticStore{
			Meta: interpreter.AccountsMetadata{
				{Account: "a", Key: "k", Value: "a2"},
			},
			Balances: interpreter.Balances{
				{Account: "a", Asset: "USD/2", Amount: big.NewInt(100)},
				{Account: "a2", Asset: "USD/2", Amount: big.NewInt(1)},
			},
		},
	}
	res, err := parseResult.RunWithFeatureFlags(context.Background(), nil, &store, map[string]struct{}{
		flags.ExperimentalMidScriptFunctionCall: {},
	})
	require.Nil(t, err)

	require.Equal(t,
		[]interpreter.Posting{
			{
				Source:      "a",
				Destination: "b",
				Amount:      big.NewInt(1),
				Asset:       "USD/2",
			},
			{
				Source:      "world",
				Destination: "b",
				Amount:      big.NewInt(9),
				Asset:       "USD/2",
			},
		},
		res.Postings,
	)

	require.Equal(t,
		[]numscript.BalanceQuery{
			{
				{Account: "a", Asset: "USD/2"},
			},
			{
				{Account: "a2", Asset: "USD/2"},
			},
		},
		store.GetBalancesCalls,
	)

}

// AliasingStore mimics a real-world Store (e.g. DB-backed with an in-memory
// cache) that returns the *same* *big.Int pointers it keeps in its own internal
// state, instead of defensively cloning them the way StaticStore does.
// If the interpreter mutates the amounts it receives in place, this store's
// internal state gets silently corrupted.
type AliasingStore struct {
	internal map[string]*big.Int // key: account -> amount (single COIN asset, no color)
}

func (s *AliasingStore) GetBalances(_ context.Context, q interpreter.BalanceQuery) (interpreter.Balances, error) {
	var out interpreter.Balances
	for _, item := range q {
		amt, ok := s.internal[item.Account]
		if !ok {
			amt = big.NewInt(0)
			s.internal[item.Account] = amt
		}
		// NOTE: returns the retained pointer directly, no clone.
		out = append(out, interpreter.BalanceRow{
			Account: item.Account,
			Asset:   item.Asset,
			Amount:  amt,
		})
	}
	return out, nil
}

func (s *AliasingStore) GetAccountsMetadata(_ context.Context, _ interpreter.MetadataQuery) (interpreter.AccountsMetadata, error) {
	return nil, nil
}

func TestStoreBalancesNotMutatedInPlace(t *testing.T) {
	parseResult := numscript.Parse(`send [COIN 100] (
	source = @alice
	destination = @bob
)
`)
	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := &AliasingStore{
		internal: map[string]*big.Int{
			"alice": big.NewInt(500),
		},
	}

	_, err := parseResult.Run(context.Background(), numscript.VariablesMap{}, store)
	require.Nil(t, err)

	// The store's own internal balance for @alice must be untouched by the run.
	require.Equal(t, big.NewInt(500), store.internal["alice"],
		"interpreter mutated the Store's balance pointer in place")
}

type ObservableStore struct {
	StaticStore      interpreter.StaticStore
	GetBalancesCalls []numscript.BalanceQuery
	GetMetadataCalls []numscript.MetadataQuery
}

func (os *ObservableStore) GetBalances(ctx context.Context, q interpreter.BalanceQuery) (interpreter.Balances, error) {
	os.GetBalancesCalls = append(os.GetBalancesCalls, q)
	return os.StaticStore.GetBalances(ctx, q)

}

func (os *ObservableStore) GetAccountsMetadata(ctx context.Context, q interpreter.MetadataQuery) (interpreter.AccountsMetadata, error) {
	os.GetMetadataCalls = append(os.GetMetadataCalls, q)
	return os.StaticStore.GetAccountsMetadata(ctx, q)
}

type ErrorStore struct{}

func (*ErrorStore) GetBalances(ctx context.Context, q interpreter.BalanceQuery) (interpreter.Balances, error) {
	return nil, errors.New("Error while fetching balances")
}

func (*ErrorStore) GetAccountsMetadata(ctx context.Context, q interpreter.MetadataQuery) (interpreter.AccountsMetadata, error) {
	return nil, errors.New("Error while fetching metadata")
}

func TestResolveDependenciesPublicAPI(t *testing.T) {
	parsed := numscript.Parse(`
		send [USD 10] (
			source = @alice
			destination = @bob
		)
	`)

	deps, err := parsed.ResolveDependencies(context.Background(), nil, numscript.StaticStore{})
	require.NoError(t, err)

	require.Equal(t, map[numscript.AccountDependency]struct{}{
		{Account: "alice", Asset: "USD"}: {},
	}, deps.AccountsReads)
	require.Equal(t, map[numscript.AccountDependency]struct{}{
		{Account: "alice", Asset: "USD"}: {},
		{Account: "bob", Asset: "USD"}:   {},
	}, deps.AccountsWrites)
}
