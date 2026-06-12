package interpreter

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestFilterQuery(t *testing.T) {
	fullBalance := Balances{
		"alice": AccountBalance{
			"EUR/2": big.NewInt(1),
			"USD/2": big.NewInt(2),
		},
		"bob": AccountBalance{
			"BTC": big.NewInt(3),
		},
	}

	filteredQuery := fullBalance.filterQuery(BalanceQuery{
		"alice":   []string{"GBP/2", "YEN", "EUR/2"},
		"bob":     []string{"BTC"},
		"charlie": []string{"ETH"},
	})

	require.Equal(t, BalanceQuery{
		"alice":   []string{"GBP/2", "YEN"},
		"charlie": []string{"ETH"},
	}, filteredQuery)
}

func TestCloneBalances(t *testing.T) {
	fullBalance := Balances{
		"alice": AccountBalance{
			"EUR/2": big.NewInt(1),
			"USD/2": big.NewInt(2),
		},
		"bob": AccountBalance{
			"BTC": big.NewInt(3),
		},
	}

	cloned := fullBalance.DeepClone()

	fullBalance["alice"]["USD/2"].Set(big.NewInt(42))

	require.Equal(t, big.NewInt(2), cloned["alice"]["USD/2"])
}

// a Store implementation which returns pointers to its internal state,
// so that we can detect whether the interpreter mutates them
type aliasingStore struct {
	balances Balances
}

func (s *aliasingStore) GetBalances(_ context.Context, q BalanceQuery) (Balances, error) {
	out := Balances{}
	for account, currencies := range q {
		accountBalance := AccountBalance{}
		out[account] = accountBalance
		for _, curr := range currencies {
			// note: this aliases the store's internal *big.Int
			accountBalance[curr] = s.balances[account][curr]
		}
	}
	return out, nil
}

func (s *aliasingStore) GetAccountsMetadata(_ context.Context, _ MetadataQuery) (AccountsMetadata, error) {
	return AccountsMetadata{}, nil
}

func TestRunProgramDoesNotMutateStoreBalances(t *testing.T) {
	t.Parallel()

	aliceBalance := big.NewInt(100)
	store := &aliasingStore{
		balances: Balances{
			"alice": AccountBalance{
				"USD/2": aliceBalance,
			},
		},
	}

	parseResult := parser.Parse(`send [USD/2 100] (
	source = @alice
	destination = @bob
)`)
	require.Empty(t, parseResult.Errors)

	result, err := RunProgram(context.Background(), parseResult.Value, nil, store, nil)
	require.Nil(t, err)
	require.Equal(t, []Posting{
		{
			Source:      "alice",
			Destination: "bob",
			Amount:      big.NewInt(100),
			Asset:       "USD/2",
		},
	}, result.Postings)

	// the store's internal state must not have been mutated by the run
	require.Equal(t, big.NewInt(100), aliceBalance)
}

func TestPrettyPrintBalance(t *testing.T) {
	fullBalance := Balances{
		"alice": AccountBalance{
			"EUR/2":    big.NewInt(1),
			"USD/1234": big.NewInt(999999),
		},
		"bob": AccountBalance{
			"BTC": big.NewInt(3),
		},
	}

	snaps.MatchSnapshot(t, fullBalance.PrettyPrint())
}

func TestCmpMaps(t *testing.T) {

	b1 := Balances{
		"alice": AccountBalance{
			"EUR": big.NewInt(100),
		},
	}

	b2 := Balances{
		"alice": AccountBalance{
			"EUR": big.NewInt(42),
		},
	}

	require.Equal(t, false, CompareBalances(b1, b2))
}

func TestCmpMapsIncluding(t *testing.T) {

	t.Run("including (subset)", func(t *testing.T) {
		b2 := Balances{
			"alice": AccountBalance{
				"EUR": big.NewInt(100),
			},
			"bob": AccountBalance{
				"EUR": big.NewInt(100),
			},
		}

		b1 := Balances{
			"alice": AccountBalance{
				"EUR": big.NewInt(100),
			},
		}

		require.Equal(t, true, CompareBalancesIncluding(b1, b2))
	})

	t.Run("different value", func(t *testing.T) {
		b2 := Balances{
			"alice": AccountBalance{
				"EUR": big.NewInt(100),
			},
			"bob": AccountBalance{
				"EUR": big.NewInt(100),
			},
		}

		b1 := Balances{
			"alice": AccountBalance{
				"EUR": big.NewInt(0),
			},
		}

		require.Equal(t, false, CompareBalancesIncluding(b1, b2))
	})

	t.Run("extra value", func(t *testing.T) {
		b2 := Balances{
			"alice": AccountBalance{
				"EUR": big.NewInt(100),
			},
			"bob": AccountBalance{
				"EUR": big.NewInt(100),
			},
		}

		b1 := Balances{
			"alice": AccountBalance{
				"EUR": big.NewInt(100),
			},

			"extra-value": AccountBalance{
				"EUR": big.NewInt(100),
			},
		}

		require.Equal(t, false, CompareBalancesIncluding(b1, b2))
	})
}
