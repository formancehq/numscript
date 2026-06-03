package interpreter_test

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/flags"
	machine "github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

// runColored is a small helper that parses a script, evaluates it with the
// experimental-asset-colors feature flag enabled, and returns the postings.
// All test cases below exercise the public color-of-money interface — the
// idea is to exercise it from many angles so any future regression in the
// numscript ↔ ledger contract is caught here, not downstream.
func runColored(t *testing.T, src string, store machine.StaticStore) ([]machine.Posting, error) {
	t.Helper()
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors, "unexpected parser errors: %v", parsed.Errors)

	result, err := machine.RunProgram(
		context.Background(),
		parsed.Value,
		machine.VariablesMap{},
		store,
		map[string]struct{}{flags.ExperimentalAssetColors: {}},
	)
	if err != nil {
		return nil, err
	}
	return result.Postings, nil
}

// Single colored send: a posting emitted from a "RED"-restricted source must
// carry Color="RED".
func TestColorSendPropagatesColor(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 100] (
			source = @world \ "RED"
			destination = @alice
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{}}
	postings, err := runColored(t, src, store)
	require.NoError(t, err)

	require.Equal(t, []machine.Posting{
		{Source: "world", Destination: "alice", Amount: big.NewInt(100), Asset: "COIN", Color: "RED"},
	}, postings)
}

// Source-side color constraint: when a source restricts to a color, the
// emitted posting must carry that color and the funds must come from that
// exact bucket.
func TestColorSourceRestrictionEmitsColoredPosting(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 20] (
			source = @acc \ "RED"
			destination = @dest
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{
		"acc": machine.AccountBalance{
			"COIN": machine.ColorBalance{
				"":    big.NewInt(1000),
				"RED": big.NewInt(50),
			},
		},
	}}

	postings, err := runColored(t, src, store)
	require.NoError(t, err)

	require.Equal(t, []machine.Posting{
		{Source: "acc", Destination: "dest", Amount: big.NewInt(20), Asset: "COIN", Color: "RED"},
	}, postings)
}

// Colored funds are strictly segregated: insufficient funds in a color must
// fail even when other colors (and the uncolored bucket) have plenty.
func TestColorIsolationRejectsSpendFromWrongColor(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 100] (
			source = @acc \ "RED"
			destination = @dest
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{
		"acc": machine.AccountBalance{
			"COIN": machine.ColorBalance{
				"":     big.NewInt(10_000),
				"BLUE": big.NewInt(10_000),
				"RED":  big.NewInt(20),
			},
		},
	}}

	_, err := runColored(t, src, store)
	require.Error(t, err)
	var missing machine.MissingFundsErr
	require.ErrorAs(t, err, &missing)
	require.Equal(t, "COIN", missing.Asset)
}

// "Color of money" is immutable: a posting emitted under color X stays
// color X end to end. We verify this by chaining sources and watching the
// emitted postings retain their original colors.
func TestColorImmutabilityThroughInorderSource(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 150] (
			source = {
				@acc \ "RED"
				@acc \ "BLUE"
				@acc
			}
			destination = @dest
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{
		"acc": machine.AccountBalance{
			"COIN": machine.ColorBalance{
				"":     big.NewInt(100),
				"BLUE": big.NewInt(30),
				"RED":  big.NewInt(20),
			},
		},
	}}

	postings, err := runColored(t, src, store)
	require.NoError(t, err)

	require.Equal(t, []machine.Posting{
		{Source: "acc", Destination: "dest", Amount: big.NewInt(20), Asset: "COIN", Color: "RED"},
		{Source: "acc", Destination: "dest", Amount: big.NewInt(30), Asset: "COIN", Color: "BLUE"},
		{Source: "acc", Destination: "dest", Amount: big.NewInt(100), Asset: "COIN", Color: ""},
	}, postings)
}

// The uncolored bucket (Color="") is not pooled with any colored bucket —
// asking for 100 with color="" fails when only colored funds are available.
func TestUncoloredCannotDrawFromColoredFunds(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 100] (
			source = @acc
			destination = @dest
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{
		"acc": machine.AccountBalance{
			"COIN": machine.ColorBalance{
				"RED": big.NewInt(10_000),
			},
		},
	}}

	_, err := runColored(t, src, store)
	require.Error(t, err)
	var missing machine.MissingFundsErr
	require.ErrorAs(t, err, &missing)
}

// Two adjacent colored postings with the same color from the same source
// should compact into a single posting (the funds queue logic).
func TestColoredPostingsCompactByColor(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 30] (
			source = {
				@acc \ "RED"
				@acc \ "RED"
			}
			destination = @dest
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{
		"acc": machine.AccountBalance{
			"COIN": machine.ColorBalance{
				"RED": big.NewInt(100),
			},
		},
	}}

	postings, err := runColored(t, src, store)
	require.NoError(t, err)
	require.Len(t, postings, 1)
	require.Equal(t, "RED", postings[0].Color)
	require.Equal(t, big.NewInt(30), postings[0].Amount)
}

// Colored balance queries must include color in BalanceQuery items.
// We observe the store to verify the contract that gets sent across the
// numscript ↔ ledger boundary.
func TestBalanceQueryIncludesColor(t *testing.T) {
	t.Parallel()

	type observed struct {
		machine.StaticStore
		got []machine.BalanceQuery
	}
	store := &observed{StaticStore: machine.StaticStore{
		Balances: machine.Balances{
			"acc": machine.AccountBalance{
				"COIN": machine.ColorBalance{"RED": big.NewInt(1000)},
			},
		},
	}}

	src := `
		send [COIN 100] (
			source = @acc \ "RED"
			destination = @dest
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	type spyStore struct{ inner machine.Store }
	_ = spyStore{}

	// Wrap the store with an inline implementation that records GetBalances
	// calls so we can assert what crosses the boundary.
	var got []machine.BalanceQuery
	spy := storeFunc{
		getBalances: func(ctx context.Context, q machine.BalanceQuery) (machine.Balances, error) {
			cloned := machine.BalanceQuery{}
			for acc, items := range q {
				cloned[acc] = append([]machine.AssetColor(nil), items...)
			}
			got = append(got, cloned)
			return store.StaticStore.GetBalances(ctx, q)
		},
		getMetadata: store.StaticStore.GetAccountsMetadata,
	}

	_, err := machine.RunProgram(
		context.Background(),
		parsed.Value,
		machine.VariablesMap{},
		spy,
		map[string]struct{}{flags.ExperimentalAssetColors: {}},
	)
	require.NoError(t, err)

	require.Len(t, got, 1, "expected exactly one batched balance query")
	require.Equal(t, machine.BalanceQuery{
		"acc": {{Asset: "COIN", Color: "RED"}},
	}, got[0])
}

// storeFunc is a minimal machine.Store adapter built around function values.
type storeFunc struct {
	getBalances func(context.Context, machine.BalanceQuery) (machine.Balances, error)
	getMetadata func(context.Context, machine.MetadataQuery) (machine.AccountsMetadata, error)
}

func (s storeFunc) GetBalances(ctx context.Context, q machine.BalanceQuery) (machine.Balances, error) {
	return s.getBalances(ctx, q)
}

func (s storeFunc) GetAccountsMetadata(ctx context.Context, q machine.MetadataQuery) (machine.AccountsMetadata, error) {
	return s.getMetadata(ctx, q)
}

// The legal color charset (^[A-Z]*$) is enforced — anything else must be
// rejected as a bad color literal.
func TestColorLiteralCharsetIsEnforced(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		color   string
		wantErr bool
	}{
		{name: "uppercase ok", color: "RED", wantErr: false},
		{name: "empty ok (no color)", color: "", wantErr: false},
		{name: "lowercase rejected", color: "red", wantErr: true},
		{name: "digits rejected", color: "RED1", wantErr: true},
		{name: "punctuation rejected", color: "RED-FOO", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src := `send [COIN 1] (source = @world \ "` + tc.color + `" destination = @dest)`
			parsed := parser.Parse(src)
			require.Empty(t, parsed.Errors)

			_, err := machine.RunProgram(
				context.Background(),
				parsed.Value,
				machine.VariablesMap{},
				machine.StaticStore{Balances: machine.Balances{}},
				map[string]struct{}{flags.ExperimentalAssetColors: {}},
			)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Color survives a JSON marshal/unmarshal roundtrip on the Posting type.
func TestPostingJSONRoundtripPreservesColor(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 100] (
			source = @world \ "GRANTS"
			destination = @alice
		)
	`
	postings, err := runColored(t, src, machine.StaticStore{Balances: machine.Balances{}})
	require.NoError(t, err)
	require.Len(t, postings, 1)
	require.Equal(t, "GRANTS", postings[0].Color)

	encoded, err := json.Marshal(postings[0])
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"color":"GRANTS"`,
		"marshaled posting must carry a non-empty color field, got: %s", string(encoded))

	var decoded machine.Posting
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	require.Equal(t, postings[0].Color, decoded.Color)
}

// Color is always present in JSON output, even when empty, so downstream
// consumers can rely on its presence to distinguish "uncolored" from
// "missing field".
func TestPostingJSONAlwaysIncludesColor(t *testing.T) {
	t.Parallel()

	p := machine.Posting{
		Source:      "world",
		Destination: "dest",
		Asset:       "COIN",
		Amount:      big.NewInt(1),
	}
	encoded, err := json.Marshal(p)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"color":""`,
		"uncolored postings must still expose the color field, got: %s", string(encoded))
}

// Allocation-style send: one source feeding multiple destinations under a
// color constraint. Every emitted posting must carry the source's color.
func TestColoredAllocationPropagatesToEachLeg(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 100] (
			source = @bank \ "GRANTS"
			destination = {
				50% to @alice
				remaining to @bob
			}
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{
		"bank": machine.AccountBalance{
			"COIN": machine.ColorBalance{"GRANTS": big.NewInt(1000)},
		},
	}}

	postings, err := runColored(t, src, store)
	require.NoError(t, err)
	require.Len(t, postings, 2)
	for _, p := range postings {
		require.Equal(t, "GRANTS", p.Color, "every leg of the allocation must keep the color")
		require.Equal(t, "COIN", p.Asset)
	}
}

// send * (send-all) from a colored bucket drains exactly that bucket and
// emits postings carrying the bucket's color.
func TestColoredSendAllDrainsOnlyTheTargetColor(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN *] (
			source = @vault \ "RED"
			destination = @dest
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{
		"vault": machine.AccountBalance{
			"COIN": machine.ColorBalance{
				"":     big.NewInt(1_000_000),
				"RED":  big.NewInt(42),
				"BLUE": big.NewInt(999),
			},
		},
	}}

	postings, err := runColored(t, src, store)
	require.NoError(t, err)
	require.Len(t, postings, 1)
	require.Equal(t, "RED", postings[0].Color)
	require.Equal(t, big.NewInt(42), postings[0].Amount)
}

// Sending an empty-color amount must NOT pull from any colored bucket,
// even when the script forms the source from the union of accounts.
func TestUncoloredSourceIgnoresColoredFunds(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 50] (
			source = @vault
			destination = @dest
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{
		"vault": machine.AccountBalance{
			"COIN": machine.ColorBalance{
				"":    big.NewInt(20), // only 20 here — should fail
				"RED": big.NewInt(1_000_000),
			},
		},
	}}

	_, err := runColored(t, src, store)
	require.Error(t, err, "uncolored source must not be able to dip into colored funds")
	var missing machine.MissingFundsErr
	require.ErrorAs(t, err, &missing)
}

// Two distinct sources with two distinct colors must each contribute a
// posting bearing their own color (no accidental coalescing).
func TestTwoColoredSourcesYieldTwoColoredPostings(t *testing.T) {
	t.Parallel()

	src := `
		send [COIN 60] (
			source = {
				@a \ "RED"
				@b \ "BLUE"
			}
			destination = @dest
		)
	`
	store := machine.StaticStore{Balances: machine.Balances{
		"a": machine.AccountBalance{"COIN": machine.ColorBalance{"RED": big.NewInt(25)}},
		"b": machine.AccountBalance{"COIN": machine.ColorBalance{"BLUE": big.NewInt(100)}},
	}}

	postings, err := runColored(t, src, store)
	require.NoError(t, err)
	require.Len(t, postings, 2)

	require.Equal(t, "RED", postings[0].Color)
	require.Equal(t, big.NewInt(25), postings[0].Amount)

	require.Equal(t, "BLUE", postings[1].Color)
	require.Equal(t, big.NewInt(35), postings[1].Amount)
}

// Colors play orthogonally with the asset precision suffix (e.g. USD/4) —
// the suffix stays on the asset string, the color rides separately.
func TestColorComposesWithAssetPrecision(t *testing.T) {
	t.Parallel()

	src := `
		send [USD/4 10] (
			source = @src \ "COL" allowing unbounded overdraft
			destination = @dest
		)
	`
	postings, err := runColored(t, src, machine.StaticStore{Balances: machine.Balances{}})
	require.NoError(t, err)
	require.Len(t, postings, 1)
	require.Equal(t, "USD/4", postings[0].Asset)
	require.Equal(t, "COL", postings[0].Color)
}

// JSON unmarshal accepts both shapes (flat + colored), as documented in
// Balances.UnmarshalJSON.
func TestBalancesUnmarshalAcceptsBothShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		src  string
		want machine.Balances
	}{
		{
			name: "flat shorthand",
			src:  `{"alice": {"USD/2": 100, "EUR/2": -42}}`,
			want: machine.Balances{
				"alice": machine.AccountBalance{
					"USD/2": machine.Uncolored(big.NewInt(100)),
					"EUR/2": machine.Uncolored(big.NewInt(-42)),
				},
			},
		},
		{
			name: "colored",
			src:  `{"alice": {"USD/2": {"": 100, "RED": 50}}}`,
			want: machine.Balances{
				"alice": machine.AccountBalance{
					"USD/2": machine.ColorBalance{"": big.NewInt(100), "RED": big.NewInt(50)},
				},
			},
		},
		{
			name: "mixed across assets",
			src:  `{"alice": {"USD/2": 100, "EUR/2": {"RED": 5}}}`,
			want: machine.Balances{
				"alice": machine.AccountBalance{
					"USD/2": machine.Uncolored(big.NewInt(100)),
					"EUR/2": machine.ColorBalance{"RED": big.NewInt(5)},
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got machine.Balances
			require.NoError(t, got.UnmarshalJSON([]byte(tc.src)))
			require.True(t, machine.CompareBalances(tc.want, got),
				"unexpected balances: want %v, got %v", tc.want, got)
		})
	}
}
