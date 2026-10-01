package compiler_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/funds"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// e2eStore is a minimal vm.Store for the end-to-end test.
type e2eStore struct {
	balances map[funds.PairKey]*big.Int
	metadata map[e2eMetaKey]string
}

// e2eMetaKey identifies one metadata slot: account, scope and key.
type e2eMetaKey struct {
	account string
	scope   string
	key     string
}

func (s e2eStore) GetBalance(ctx context.Context, account, scope, asset, color string) (*big.Int, error) {
	if v, ok := s.balances[funds.PairKey{Account: account, Scope: scope, Asset: asset, Color: color}]; ok {
		return v, nil
	}
	return new(big.Int), nil
}

func (s e2eStore) GetMetadata(ctx context.Context, account, scope, key string) (string, bool, error) {
	v, ok := s.metadata[e2eMetaKey{account: account, scope: scope, key: key}]
	return v, ok, nil
}

// TestE2E_AllotmentOverSum: portions summing to > 1 must error (leftover < 0).
func TestE2E_AllotmentOverSum(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				2/3 to @a
				2/3 to @b
			}
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[funds.PairKey]*big.Int{}})
	require.Equal(t, vm.InvalidAllotmentSum{ActualSum: *big.NewRat(4, 3)}, execErr)
}

func TestE2E_NegativeAllotmentPortion(t *testing.T) {
	for name, src := range map[string]string{
		"source": `
			send [USD/2 90] (
				source = {
					-1/3 from @s1
					remaining from @s2
				}
				destination = @dest
			)
		`,
		"destination": `
			send [USD/2 90] (
				source = @world
				destination = {
					-1/3 to @a
					remaining to @b
				}
			)
		`,
		"portions summing to one": `
			send [USD/2 90] (
				source = @world
				destination = {
					4/3 to @a
					-1/3 to @b
				}
			)
		`,
	} {
		t.Run(name, func(t *testing.T) {
			parsed := parser.Parse(src)
			require.Empty(t, parsed.Errors)
			_, program, cErr := compiler.Compile(parsed.Value, nil)
			require.Nil(t, cErr)
			machine := vm.NewVm(program)
			_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[funds.PairKey]*big.Int{
				{Account: "s1", Asset: "USD/2", Color: ""}: big.NewInt(500),
				{Account: "s2", Asset: "USD/2", Color: ""}: big.NewInt(500),
			}})
			require.Equal(t, vm.NegativePortionError{Portion: *big.NewRat(-1, 3)}, execErr)
		})
	}
}

// TestE2E_AllotmentUnderSum: without a `remaining` clause the portions must sum
// to exactly 1, so 1/3 + 1/3 = 2/3 must error.
func TestE2E_AllotmentUnderSum(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				1/3 to @a
				1/3 to @b
			}
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[funds.PairKey]*big.Int{}})
	require.Equal(t, vm.InvalidAllotmentSum{ActualSum: *big.NewRat(2, 3)}, execErr)
}

func TestE2E_MonetarySubtractionAssetMismatch(t *testing.T) {
	src := `
		vars {
			monetary $a = [USD/2 30]
			monetary $b = [EUR/2 20]
		}
		send $a - $b (
			source = @src
			destination = @dest
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	_, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, e2eStore{balances: map[funds.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
	require.Equal(t, vm.AssetMismatchError{Expected: "USD/2", Got: "EUR/2"}, execErr)
}

func TestE2E_MonetaryAdditionAssetMismatch(t *testing.T) {
	src := `
		vars {
			monetary $a = [USD/2 3]
			monetary $b = [EUR/2 7]
		}
		send $a + $b (
			source = @src
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	store := e2eStore{balances: map[funds.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	_, execErr := vm.Exec(context.Background(), vm.NewVm(program), nil, store)
	require.Equal(t, vm.AssetMismatchError{Expected: "USD/2", Got: "EUR/2"}, execErr)
}

func TestE2E_CapAssetMismatch(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = max [EUR/2 5] from @a
			destination = @dest
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[funds.PairKey]*big.Int{}})
	require.Equal(t, vm.AssetMismatchError{Expected: "USD/2", Got: "EUR/2"}, execErr)
}

func TestE2E_OverdraftAssetMismatch(t *testing.T) {
	src := `
		send [USD/2 42] (
			source = @a allowing overdraft up to [EUR/2 5]
			destination = @dest
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[funds.PairKey]*big.Int{}})
	require.Equal(t, vm.AssetMismatchError{Expected: "USD/2", Got: "EUR/2"}, execErr)
}

func TestE2E_BalanceNegativeErrors(t *testing.T) {
	src := `
		vars { monetary $b = balance(@acc, USD/2) }
		send $b (source = @world destination = @dest)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[funds.PairKey]*big.Int{
		{Account: "acc", Asset: "USD/2", Color: ""}: big.NewInt(-1),
	}})
	require.Equal(t, vm.NegativeBalanceError{Account: "acc", Amount: *big.NewInt(-1)}, execErr)
}

func TestE2E_DivideByZero(t *testing.T) {
	src := `
		send [USD/2 100] (
			source = @world
			destination = {
				1/0 to @a
				remaining kept
			}
		)
	`
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{balances: map[funds.PairKey]*big.Int{}})
	require.Equal(t, vm.DivideByZeroError{Numerator: *big.NewInt(1)}, execErr)
}

func TestE2E_InvalidColor(t *testing.T) {
	src := `
		#![feature("experimental-asset-colors")]
		send [COIN 10] (
			source = @src \ "not a color"
			destination = @dest
		)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)

	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{})
	require.Equal(t, vm.InvalidColor{Color: "not a color"}, execErr)
}

// countingStore is an e2eStore that records how many balances it was asked for.
type countingStore struct {
	e2eStore
	balanceCalls int
}

func (s *countingStore) GetBalance(ctx context.Context, account, scope, asset, color string) (*big.Int, error) {
	s.balanceCalls++
	return s.e2eStore.GetBalance(ctx, account, scope, asset, color)
}

// The compiled world arm has no overdraft operand, which is what makes the pull
// unbounded and therefore free of Store round-trips. numscript_test.go asserts
// the same for the interpreter.
func TestE2E_WorldSourceReadsNoBalance(t *testing.T) {
	src := `send [USD/2 100] (source = @world destination = @dest)`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	store := &countingStore{}
	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, nil, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []funds.Posting{
		{Source: "world", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(100)},
	}, res.Postings)
	require.Zero(t, store.balanceCalls, "a world source must not read any balance")
}

// the run-time branch, not the literal, is what decides it
func TestE2E_DynamicWorldSourceReadsNoBalance(t *testing.T) {
	src := `
		vars { account $src }
		send [USD/2 100] (source = $src destination = @dest)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	enc, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	vars, err := enc.Encode(map[string]string{"src": "world"})
	require.NoError(t, err)
	store := &countingStore{}
	machine := vm.NewVm(program)
	res, execErr := vm.Exec(context.Background(), machine, &vars, store)
	require.Nil(t, execErr)

	requirePostingsEqual(t, []funds.Posting{
		{Source: "world", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(100)},
	}, res.Postings)
	require.Zero(t, store.balanceCalls, "a world source must not read any balance")
}

// A send-all needs a bounded source to know how much "all" is, and @world is
// unbounded. The specs format has no expectation field for this error, so it is
// asserted here; the interpreter's twin is TestInvalidUnboundedWorldInSendAll.
func TestE2E_SendAllFromWorldErrors(t *testing.T) {
	src := `send [USD/2 *] (source = @world destination = @dest)`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	_, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, nil, e2eStore{})

	require.Equal(t, vm.InvalidUncappedSource{Account: "world"}, execErr)
}

// same, but world is only known at run time, so the compiler cannot reject it
func TestE2E_SendAllFromDynamicWorldErrors(t *testing.T) {
	src := `
		vars { account $src }
		send [USD/2 *] (source = $src destination = @dest)
	`

	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	enc, program, cErr := compiler.Compile(parsed.Value, nil)
	require.Nil(t, cErr)
	vars, err := enc.Encode(map[string]string{"src": "world"})
	require.NoError(t, err)
	machine := vm.NewVm(program)
	_, execErr := vm.Exec(context.Background(), machine, &vars, e2eStore{})

	require.Equal(t, vm.InvalidUncappedSource{Account: "world"}, execErr)
}

func requirePostingsEqual(t *testing.T, want, got []funds.Posting) {
	t.Helper()
	require.Len(t, got, len(want))
	for i := range want {
		w, g := want[i], got[i]
		require.Equal(t, w.Source, g.Source, "posting[%d].Source", i)
		require.Equal(t, w.Destination, g.Destination, "posting[%d].Destination", i)
		require.Equal(t, w.Asset, g.Asset, "posting[%d].Asset", i)
		require.Equal(t, w.Color, g.Color, "posting[%d].Color", i)
		require.Zero(t, g.Amount.Cmp(w.Amount), "posting[%d].Amount: got %s want %s", i, g.Amount, w.Amount)
	}
}
