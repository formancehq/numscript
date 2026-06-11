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
			Balances: interpreter.Balances{},
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
			Meta:     interpreter.AccountsMetadata{"account_that_needs_meta": {"k": "source2"}},
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
				"account_that_needs_meta": {"k"},
			},
		},
		store.GetMetadataCalls)

	require.Equal(t,
		[]numscript.BalanceQuery{
			// TODO maybe those calls can be batched together
			{
				// this is required by the balance() call
				"account_that_needs_balance": {"USD/2"},
			},
			{
				// this is defined in the variables
				"source1": {"COIN"},

				// this is defined in account metadata
				"source2": {"COIN"},

				// this appears as literal
				"source3": {"COIN"},
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
				"a": {"COIN"},
				"b": {"COIN"},
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
				"alice": {"COIN"},
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
				"a": {"COIN": big.NewInt(10000)},
				"b": {"COIN": big.NewInt(10000)},
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
				"a": {"COIN"},
				"b": {"COIN"},
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
				"a": {"COIN"},
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
				"src": {"COIN"},
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
				"src1": {"COIN"},
			},
			{
				"src2": {"COIN"},
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
				"src": {"EUR/2"},
			},
			{
				"src": {"USD/2"},
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
				"alice": {"USD/2"},
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
				"alice": interpreter.AccountBalance{
					"USD/2": big.NewInt(20),
				},
			},
		},
	}
	res, err := parseResult.RunWithFeatureFlags(context.Background(), nil, &store, map[string]struct{}{
		flags.ExperimentalMidScriptFunctionCall: {},
	})
	require.Nil(t, err)

	require.Equal(t, interpreter.Metadata{
		"k": interpreter.NewMonetary("USD/2", 100),
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
				"a": interpreter.AccountMetadata{
					"k": "a2",
				},
			},
			Balances: interpreter.Balances{
				"a": interpreter.AccountBalance{
					"USD/2": big.NewInt(100),
				},
				"a2": interpreter.AccountBalance{
					"USD/2": big.NewInt(1),
				},
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
				"a": {"USD/2"},
			},
			{
				"a2": {"USD/2"},
			},
		},
		store.GetBalancesCalls,
	)

}

func TestMetaFunctionCachesQueries(t *testing.T) {
	parseResult := numscript.Parse(`vars {
	string $a = meta(@acc1, "key1")
	string $b = meta(@acc2, "key2")
	string $a_again = meta(@acc1, "key1")
}
set_tx_meta("a", $a)
set_tx_meta("b", $b)
set_tx_meta("a_again", $a_again)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := scopedMetaStore{
		meta: interpreter.AccountsMetadata{
			"acc1": {"key1": "value1"},
			"acc2": {"key2": "value2"},
		},
	}

	res, err := parseResult.Run(context.Background(), numscript.VariablesMap{}, &store)
	require.Nil(t, err)

	// merging the second query into the cache must not clobber
	// the previously cached acc1 metadata
	require.Equal(t,
		numscript.Metadata{
			"a":       interpreter.String("value1"),
			"b":       interpreter.String("value2"),
			"a_again": interpreter.String("value1"),
		},
		res.Metadata)

	// acc1's "key1" is cached after the first meta() call,
	// so the third meta() call must not query the store again
	require.Equal(t,
		[]numscript.MetadataQuery{
			{"acc1": {"key1"}},
			{"acc2": {"key2"}},
		},
		store.GetMetadataCalls)
}

func TestMetaFunctionMissingKeyStillErrors(t *testing.T) {
	parseResult := numscript.Parse(`vars {
	string $a = meta(@acc1, "key1")
	string $b = meta(@acc1, "missing_key")
}
set_tx_meta("a", $a)
set_tx_meta("b", $b)
`)

	require.Empty(t, parseResult.GetParsingErrors(), "There should not be parsing errors")

	store := scopedMetaStore{
		meta: interpreter.AccountsMetadata{
			"acc1": {"key1": "value1"},
		},
	}

	// even though acc1's metadata is (partially) cached,
	// a key that is absent from the store must still error
	_, err := parseResult.Run(context.Background(), numscript.VariablesMap{}, &store)
	require.NotNil(t, err)

	notFoundErr, ok := err.(interpreter.MetadataNotFound)
	require.True(t, ok, "expected a MetadataNotFound error, got: %v", err)
	require.Equal(t, "acc1", notFoundErr.Account)
	require.Equal(t, "missing_key", notFoundErr.Key)
}

// A store that, unlike StaticStore, only returns the queried
// (account, key) pairs, and records the metadata queries it receives
type scopedMetaStore struct {
	meta             interpreter.AccountsMetadata
	GetMetadataCalls []numscript.MetadataQuery
}

func (s *scopedMetaStore) GetBalances(ctx context.Context, q interpreter.BalanceQuery) (interpreter.Balances, error) {
	return interpreter.StaticStore{}.GetBalances(ctx, q)
}

func (s *scopedMetaStore) GetAccountsMetadata(_ context.Context, q interpreter.MetadataQuery) (interpreter.AccountsMetadata, error) {
	s.GetMetadataCalls = append(s.GetMetadataCalls, q)

	out := interpreter.AccountsMetadata{}
	for account, keys := range q {
		for _, key := range keys {
			value, ok := s.meta[account][key]
			if !ok {
				continue
			}

			accountMeta, ok := out[account]
			if !ok {
				accountMeta = interpreter.AccountMetadata{}
				out[account] = accountMeta
			}
			accountMeta[key] = value
		}
	}
	return out, nil
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
