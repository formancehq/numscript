package interpreter

import (
	"math/big"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestFilterQuery(t *testing.T) {
	fullBalance := Balances{
		"alice": AccountBalance{
			"EUR/2": Uncolored(big.NewInt(1)),
			"USD/2": Uncolored(big.NewInt(2)),
		},
		"bob": AccountBalance{
			"BTC": Uncolored(big.NewInt(3)),
		},
	}

	filteredQuery := fullBalance.filterQuery(BalanceQuery{
		"alice":   []AssetColor{{Asset: "GBP/2"}, {Asset: "YEN"}, {Asset: "EUR/2"}},
		"bob":     []AssetColor{{Asset: "BTC"}},
		"charlie": []AssetColor{{Asset: "ETH"}},
	})

	require.Equal(t, BalanceQuery{
		"alice":   []AssetColor{{Asset: "GBP/2"}, {Asset: "YEN"}},
		"charlie": []AssetColor{{Asset: "ETH"}},
	}, filteredQuery)
}

func TestFilterQueryDistinguishesColors(t *testing.T) {
	t.Parallel()

	fullBalance := Balances{
		"alice": AccountBalance{
			"USD/2": ColorBalance{
				"":      big.NewInt(1),
				"GRANTS": big.NewInt(2),
			},
		},
	}

	filteredQuery := fullBalance.filterQuery(BalanceQuery{
		"alice": []AssetColor{
			{Asset: "USD/2"},
			{Asset: "USD/2", Color: "GRANTS"},
			{Asset: "USD/2", Color: "OPS"},
		},
	})

	require.Equal(t, BalanceQuery{
		"alice": []AssetColor{{Asset: "USD/2", Color: "OPS"}},
	}, filteredQuery)
}

func TestCloneBalances(t *testing.T) {
	fullBalance := Balances{
		"alice": AccountBalance{
			"EUR/2": Uncolored(big.NewInt(1)),
			"USD/2": Uncolored(big.NewInt(2)),
		},
		"bob": AccountBalance{
			"BTC": Uncolored(big.NewInt(3)),
		},
	}

	cloned := fullBalance.DeepClone()

	fullBalance["alice"]["USD/2"][""].Set(big.NewInt(42))

	require.Equal(t, big.NewInt(2), cloned["alice"]["USD/2"][""])
}

func TestCloneBalancesPreservesColors(t *testing.T) {
	t.Parallel()

	fullBalance := Balances{
		"alice": AccountBalance{
			"USD/2": ColorBalance{
				"RED":  big.NewInt(10),
				"BLUE": big.NewInt(20),
			},
		},
	}

	cloned := fullBalance.DeepClone()
	fullBalance["alice"]["USD/2"]["RED"].Set(big.NewInt(999))

	require.Equal(t, big.NewInt(10), cloned["alice"]["USD/2"]["RED"])
	require.Equal(t, big.NewInt(20), cloned["alice"]["USD/2"]["BLUE"])
}

func TestPrettyPrintBalance(t *testing.T) {
	fullBalance := Balances{
		"alice": AccountBalance{
			"EUR/2":    Uncolored(big.NewInt(1)),
			"USD/1234": Uncolored(big.NewInt(999999)),
		},
		"bob": AccountBalance{
			"BTC": Uncolored(big.NewInt(3)),
		},
	}

	snaps.MatchSnapshot(t, fullBalance.PrettyPrint())
}

func TestCmpMaps(t *testing.T) {

	b1 := Balances{
		"alice": AccountBalance{
			"EUR": Uncolored(big.NewInt(100)),
		},
	}

	b2 := Balances{
		"alice": AccountBalance{
			"EUR": Uncolored(big.NewInt(42)),
		},
	}

	require.Equal(t, false, CompareBalances(b1, b2))
}

func TestCmpMapsDistinguishesColors(t *testing.T) {
	t.Parallel()

	b1 := Balances{
		"alice": AccountBalance{
			"USD/2": ColorBalance{"RED": big.NewInt(100)},
		},
	}

	b2 := Balances{
		"alice": AccountBalance{
			"USD/2": ColorBalance{"BLUE": big.NewInt(100)},
		},
	}

	require.False(t, CompareBalances(b1, b2),
		"same asset but different colors must not compare equal")
}

func TestCmpMapsIncluding(t *testing.T) {

	t.Run("including (subset)", func(t *testing.T) {
		b2 := Balances{
			"alice": AccountBalance{
				"EUR": Uncolored(big.NewInt(100)),
			},
			"bob": AccountBalance{
				"EUR": Uncolored(big.NewInt(100)),
			},
		}

		b1 := Balances{
			"alice": AccountBalance{
				"EUR": Uncolored(big.NewInt(100)),
			},
		}

		require.Equal(t, true, CompareBalancesIncluding(b1, b2))
	})

	t.Run("different value", func(t *testing.T) {
		b2 := Balances{
			"alice": AccountBalance{
				"EUR": Uncolored(big.NewInt(100)),
			},
			"bob": AccountBalance{
				"EUR": Uncolored(big.NewInt(100)),
			},
		}

		b1 := Balances{
			"alice": AccountBalance{
				"EUR": Uncolored(big.NewInt(0)),
			},
		}

		require.Equal(t, false, CompareBalancesIncluding(b1, b2))
	})

	t.Run("extra value", func(t *testing.T) {
		b2 := Balances{
			"alice": AccountBalance{
				"EUR": Uncolored(big.NewInt(100)),
			},
			"bob": AccountBalance{
				"EUR": Uncolored(big.NewInt(100)),
			},
		}

		b1 := Balances{
			"alice": AccountBalance{
				"EUR": Uncolored(big.NewInt(100)),
			},

			"extra-value": AccountBalance{
				"EUR": Uncolored(big.NewInt(100)),
			},
		}

		require.Equal(t, false, CompareBalancesIncluding(b1, b2))
	})

	t.Run("color-aware subset", func(t *testing.T) {
		b2 := Balances{
			"alice": AccountBalance{
				"USD/2": ColorBalance{
					"":     big.NewInt(100),
					"RED":  big.NewInt(50),
					"BLUE": big.NewInt(25),
				},
			},
		}

		b1 := Balances{
			"alice": AccountBalance{
				"USD/2": ColorBalance{"RED": big.NewInt(50)},
			},
		}

		require.True(t, CompareBalancesIncluding(b1, b2),
			"colored subset of an asset must be considered included")
	})

	t.Run("missing color", func(t *testing.T) {
		b2 := Balances{
			"alice": AccountBalance{
				"USD/2": ColorBalance{"": big.NewInt(100)},
			},
		}

		b1 := Balances{
			"alice": AccountBalance{
				"USD/2": ColorBalance{"RED": big.NewInt(50)},
			},
		}

		require.False(t, CompareBalancesIncluding(b1, b2),
			"color present in subset but missing in superset must not be considered included")
	})
}

func TestFetchBalanceCreatesEntriesLazily(t *testing.T) {
	t.Parallel()

	b := Balances{}
	got := b.fetchBalance("alice", "USD/2", "RED")
	require.NotNil(t, got)
	require.Equal(t, 0, got.Sign(), "freshly created entry should be zero")

	// mutating the returned big.Int should be reflected in subsequent fetches
	got.SetInt64(42)
	require.Equal(t, big.NewInt(42), b.fetchBalance("alice", "USD/2", "RED"))

	// distinct colors must be independent
	require.Equal(t, big.NewInt(0), b.fetchBalance("alice", "USD/2", "BLUE"))
}

func TestMergeBalancesIsColorAware(t *testing.T) {
	t.Parallel()

	b := Balances{
		"alice": AccountBalance{
			"USD/2": ColorBalance{"RED": big.NewInt(1)},
		},
	}

	b.Merge(Balances{
		"alice": AccountBalance{
			"USD/2": ColorBalance{
				"RED":  big.NewInt(99), // overwrites
				"BLUE": big.NewInt(50), // adds new color under existing asset
			},
			"EUR/2": Uncolored(big.NewInt(7)), // adds new asset
		},
		"bob": AccountBalance{
			"BTC": Uncolored(big.NewInt(3)), // adds new account
		},
	})

	require.Equal(t, big.NewInt(99), b["alice"]["USD/2"]["RED"])
	require.Equal(t, big.NewInt(50), b["alice"]["USD/2"]["BLUE"])
	require.Equal(t, big.NewInt(7), b["alice"]["EUR/2"][""])
	require.Equal(t, big.NewInt(3), b["bob"]["BTC"][""])
}

func TestUncoloredHelper(t *testing.T) {
	t.Parallel()

	got := Uncolored(big.NewInt(42))
	require.Equal(t, ColorBalance{"": big.NewInt(42)}, got)
}
