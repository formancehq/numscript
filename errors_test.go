package numscript_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/flags"
	"github.com/stretchr/testify/require"
)

func requireAs[T error](t *testing.T, err error) T {
	t.Helper()
	var target T
	require.True(t, errors.As(err, &target), "expected %T, got %T: %v", target, err, err)
	return target
}

type errorCase struct {
	name   string
	script string
	vars   numscript.VariablesMap
	store  numscript.Store
	flags  []string
	// Whether ResolveDependencies fails with the same error as Run.
	resolve bool
	check   func(t *testing.T, err error)
}

// Hosts classify interpreter failures with errors.As on the public aliases and
// persist their fields, so both must survive refactors of the internal types.
func TestInterpreterErrorsAreClassifiable(t *testing.T) {
	cases := []errorCase{
		{
			name:    "missing variable",
			script:  `vars { account $a } send [USD 1] ( source = $a destination = @b )`,
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "a", requireAs[numscript.MissingVariableErr](t, err).Name)
			},
		},
		{
			name:    "invalid account name",
			script:  `vars { account $a } send [USD 1] ( source = $a destination = @b )`,
			vars:    numscript.VariablesMap{"a": "@world"},
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "@world", requireAs[numscript.InvalidAccountName](t, err).Name)
			},
		},
		{
			name:    "invalid asset",
			script:  `vars { asset $x } send [$x 1] ( source = @world destination = @b )`,
			vars:    numscript.VariablesMap{"x": "Aa"},
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "Aa", requireAs[numscript.InvalidAsset](t, err).Name)
			},
		},
		{
			name:    "invalid monetary literal",
			script:  `vars { monetary $m } send $m ( source = @world destination = @b )`,
			vars:    numscript.VariablesMap{"m": "USD"},
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "USD", requireAs[numscript.InvalidMonetaryLiteral](t, err).Source)
			},
		},
		{
			name:    "invalid number literal",
			script:  `vars { number $n } send [USD $n] ( source = @world destination = @b )`,
			vars:    numscript.VariablesMap{"n": "abc"},
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "abc", requireAs[numscript.InvalidNumberLiteral](t, err).Source)
			},
		},
		{
			name:    "bad portion",
			script:  `vars { portion $p }`,
			vars:    numscript.VariablesMap{"p": "abc"},
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "abc", requireAs[numscript.BadPortionParsingErr](t, err).Source)
			},
		},
		{
			name:    "invalid variable type",
			script:  `vars { invalidt $x }`,
			vars:    numscript.VariablesMap{"x": "42"},
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "invalidt", requireAs[numscript.InvalidTypeErr](t, err).Name)
			},
		},
		{
			name:    "mismatched currency",
			script:  `vars { monetary $m = [USD 1] + [EUR 1] }`,
			resolve: true,
			check: func(t *testing.T, err error) {
				e := requireAs[numscript.MismatchedCurrencyError](t, err)
				require.Equal(t, "USD", e.Expected)
				require.Equal(t, "EUR", e.Got)
			},
		},
		{
			name:    "divide by zero",
			script:  `vars { portion $p = 1 / 0 }`,
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, big.NewInt(1), requireAs[numscript.DivideByZero](t, err).Numerator)
			},
		},
		{
			name:    "metadata not found",
			script:  `vars { number $n = meta(@acc, "k") }`,
			resolve: true,
			check: func(t *testing.T, err error) {
				e := requireAs[numscript.MetadataNotFound](t, err)
				require.Equal(t, "acc", e.Account)
				require.Equal(t, "k", e.Key)
			},
		},
		{
			name:   "negative balance",
			script: `vars { monetary $m = balance(@acc, USD/2) }`,
			store: numscript.StaticStore{Balances: numscript.Balances{
				{Account: "acc", Asset: "USD/2", Amount: big.NewInt(-5)},
			}},
			resolve: true,
			check: func(t *testing.T, err error) {
				e := requireAs[numscript.NegativeBalanceError](t, err)
				require.Equal(t, "acc", e.Account)
				require.Equal(t, "USD/2", e.Asset)
				require.Equal(t, *big.NewInt(-5), e.Amount)
			},
		},
		{
			name:    "unbound function",
			script:  `vars { number $x = unbound_fn(1, 2) }`,
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "unbound_fn", requireAs[numscript.UnboundFunctionErr](t, err).Name)
			},
		},
		{
			name:    "unbound variable",
			script:  `send [USD 1] ( source = $unbound destination = @b )`,
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "unbound", requireAs[numscript.UnboundVariableErr](t, err).Name)
			},
		},
		{
			name:   "bad arity",
			script: `set_tx_meta()`,
			check: func(t *testing.T, err error) {
				e := requireAs[numscript.BadArityErr](t, err)
				require.Equal(t, 2, e.ExpectedArity)
				require.Equal(t, 0, e.GivenArguments)
			},
		},
		{
			name:   "type error",
			script: `set_tx_meta(@acc, "v")`,
			check: func(t *testing.T, err error) {
				require.Equal(t, "string", requireAs[numscript.TypeError](t, err).Expected)
			},
		},
		{
			name:   "missing funds",
			script: `send [USD 10] ( source = @a destination = @b )`,
			check: func(t *testing.T, err error) {
				e := requireAs[numscript.MissingFundsErr](t, err)
				require.Equal(t, "USD", e.Asset)
				require.Equal(t, *big.NewInt(10), e.Needed)
			},
		},
		{
			name:   "negative amount",
			script: `vars { monetary $m } send $m ( source = @world destination = @b )`,
			vars:   numscript.VariablesMap{"m": "USD -1"},
			check: func(t *testing.T, err error) {
				require.Equal(t, "-1", requireAs[numscript.NegativeAmountErr](t, err).Amount.String())
			},
		},
		{
			name: "invalid allotment sum",
			script: `send [USD 10] (
				source = { 1/2 from @world 1/3 from @world }
				destination = @b
			)`,
			check: func(t *testing.T, err error) {
				require.Equal(t, *big.NewRat(5, 6), requireAs[numscript.InvalidAllotmentSum](t, err).ActualSum)
			},
		},
		{
			name: "negative portion",
			script: `vars { number $n } send [USD 9] (
				source = { $n/3 from @world remaining from @world }
				destination = @b
			)`,
			vars: numscript.VariablesMap{"n": "-1"},
			check: func(t *testing.T, err error) {
				require.Equal(t, *big.NewRat(-1, 3), requireAs[numscript.NegativePortion](t, err).Portion)
			},
		},
		{
			name: "duplicate remaining allotment",
			script: `send [USD 10] (
				source = { remaining from @world remaining from @world }
				destination = @b
			)`,
			check: func(t *testing.T, err error) { requireAs[numscript.InvalidRemainingAllotment](t, err) },
		},
		{
			name: "allotment in send all",
			script: `send [USD *] (
				source = { 1/2 from @a remaining from @b }
				destination = @c
			)`,
			check: func(t *testing.T, err error) { requireAs[numscript.InvalidAllotmentInSendAll](t, err) },
		},
		{
			name:   "unbounded source in send all",
			script: `send [USD *] ( source = @world destination = @b )`,
			check: func(t *testing.T, err error) {
				require.Equal(t, "world", requireAs[numscript.InvalidUnboundedInSendAll](t, err).Name)
			},
		},
		{
			name:   "unbounded source in scaling",
			script: `send [EUR/2 *] ( source = @world with scaling through @swap destination = @b )`,
			flags:  []string{flags.AssetScaling},
			check: func(t *testing.T, err error) {
				requireAs[numscript.InvalidUnboundedAddressInScalingAddress](t, err)
			},
		},
		{
			name:    "nested meta",
			script:  `vars { number $x = 1 + meta(@acc, "k") }`,
			flags:   []string{flags.ExperimentalMidScriptFunctionCall},
			resolve: true,
			check:   func(t *testing.T, err error) { requireAs[numscript.InvalidNestedMeta](t, err) },
		},
		{
			name:   "cannot cast to string",
			script: `vars { monetary $m } set_tx_meta("k", @acc:$m)`,
			vars:   numscript.VariablesMap{"m": "USD/2 10"},
			flags:  []string{flags.ExperimentalAccountInterpolationFlag},
			check:  func(t *testing.T, err error) { requireAs[numscript.CannotCastToString](t, err) },
		},
		{
			name:   "cannot cast scoped account to string",
			script: `vars { account $s = scoped(@a, "s") } set_tx_meta("k", @foo:$s)`,
			flags:  []string{flags.ExperimentalScopedFunction, flags.ExperimentalAccountInterpolationFlag},
			check: func(t *testing.T, err error) {
				e := requireAs[numscript.CannotCastScopedAccountToString](t, err)
				require.Equal(t, "a", e.Account)
				require.Equal(t, "s", e.Scope)
			},
		},
		{
			name:   "cannot store scoped account in meta",
			script: `vars { account $s = scoped(@a, "s") } set_tx_meta("k", $s)`,
			flags:  []string{flags.ExperimentalScopedFunction},
			check: func(t *testing.T, err error) {
				e := requireAs[numscript.CannotStoreScopedAccountInMeta](t, err)
				require.Equal(t, "a", e.Account)
				require.Equal(t, "s", e.Scope)
			},
		},
		{
			name:    "invalid scope",
			script:  `vars { account $s = scoped(@b, "not a scope!") }`,
			flags:   []string{flags.ExperimentalScopedFunction},
			resolve: true,
			check: func(t *testing.T, err error) {
				require.Equal(t, "not a scope!", requireAs[numscript.InvalidScope](t, err).Scope)
			},
		},
		{
			name:   "invalid color",
			script: `send [USD 1] ( source = @a \ "red" allowing unbounded overdraft destination = @b )`,
			flags:  []string{flags.ExperimentalAssetColors},
			check: func(t *testing.T, err error) {
				require.Equal(t, "red", requireAs[numscript.InvalidColor](t, err).Color)
			},
		},
		{
			name:   "experimental feature",
			script: `vars { account $s = scoped(@b, "s") }`,
			check: func(t *testing.T, err error) {
				e := requireAs[numscript.ExperimentalFeature](t, err)
				require.Equal(t, flags.ExperimentalScopedFunction, e.FlagName)
			},
		},
		{
			name:   "invalid feature",
			script: `#![feature("nope")]`,
			check: func(t *testing.T, err error) {
				require.Equal(t, "nope", requireAs[numscript.InvalidFeature](t, err).Feature)
			},
		},
		{
			name:   "balance query failure",
			script: `send [USD 1] ( source = @a destination = @b )`,
			store:  &ErrorStore{},
			check: func(t *testing.T, err error) {
				require.EqualError(t, requireAs[numscript.QueryBalanceError](t, err).WrappedError, "Error while fetching balances")
			},
		},
		{
			name:    "metadata query failure",
			script:  `vars { number $n = meta(@acc, "k") }`,
			store:   &ErrorStore{},
			resolve: true,
			check: func(t *testing.T, err error) {
				require.EqualError(t, requireAs[numscript.QueryMetadataError](t, err).WrappedError, "Error while fetching metadata")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed := numscript.Parse(tc.script)
			require.Empty(t, parsed.GetParsingErrors())

			store := tc.store
			if store == nil {
				store = numscript.StaticStore{}
			}
			featureFlags := map[string]struct{}{}
			for _, f := range tc.flags {
				featureFlags[f] = struct{}{}
			}

			_, runErr := parsed.RunWithFeatureFlags(context.Background(), tc.vars, store, featureFlags)
			require.NotNil(t, runErr)
			tc.check(t, runErr)

			if tc.resolve {
				_, resolveErr := parsed.ResolveDependencies(context.Background(), tc.vars, store)
				require.Error(t, resolveErr)
				tc.check(t, resolveErr)
			}
		})
	}
}

func TestResolveDependenciesRejectsScaling(t *testing.T) {
	parsed := numscript.Parse(`send [EUR/2 100] ( source = @a with scaling through @swap destination = @b )`)
	require.Empty(t, parsed.GetParsingErrors())

	_, err := parsed.ResolveDependencies(context.Background(), nil, numscript.StaticStore{})
	require.ErrorIs(t, err, numscript.ErrScalingNotSupported)
}
