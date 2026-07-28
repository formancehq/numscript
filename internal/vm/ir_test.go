package vm_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// These tests drive the VM directly from the IR textual format: no numscript, no
// compiler. That keeps them honest about what the VM itself does — the bytecode
// under test is written out instruction by instruction — and it makes it possible
// to exercise instruction sequences the compiler doesn't currently emit.

// irStore is a vm.Store backed by plain maps.
type irStore struct {
	balances map[runtime.PairKey]*big.Int
	metadata map[string]map[string]string
}

func (s irStore) GetBalance(_ context.Context, account, asset, color string) (*big.Int, error) {
	if v, ok := s.balances[runtime.PairKey{Account: account, Asset: asset, Color: color}]; ok {
		return new(big.Int).Set(v), nil
	}
	return new(big.Int), nil
}

func (s irStore) GetMetadata(_ context.Context, account, key string) (string, bool, error) {
	v, ok := s.metadata[account][key]
	return v, ok, nil
}

func balances(pairs map[string]int64) irStore {
	b := map[runtime.PairKey]*big.Int{}
	for account, amount := range pairs {
		b[runtime.PairKey{Account: account, Asset: "USD/2"}] = big.NewInt(amount)
	}
	return irStore{balances: b}
}

// assembleIR turns an IR text into a runnable program, failing the test on any
// error the format's own layers report.
func assembleIR(t *testing.T, src string) vm.Program {
	t.Helper()

	instrs, errs := ir.Parse(src)
	require.Empty(t, errs, "IR errors: %v", errs)
	require.NoError(t, ir.Typecheck(instrs))

	program, err := ir.Assemble(instrs)
	require.NoError(t, err)
	return program
}

// runIR assembles and runs an IR text, requiring it to succeed.
func runIR(t *testing.T, src string, store irStore, vars *vm.Vars) runtime.ExecutionResult {
	t.Helper()

	res, execErr := vm.Exec(context.Background(), vm.NewVm(assembleIR(t, src)), vars, store)
	require.Nil(t, execErr, "unexpected execution error: %v", execErr)
	return res
}

// runIRExpectingError is runIR for the cases that must fail at run time.
func runIRExpectingError(t *testing.T, src string, store irStore, vars *vm.Vars) vm.ExecutionError {
	t.Helper()

	_, execErr := vm.Exec(context.Background(), vm.NewVm(assembleIR(t, src)), vars, store)
	require.NotNil(t, execErr, "expected an execution error")
	return execErr
}

func requirePostings(t *testing.T, want, got []runtime.Posting) {
	t.Helper()

	require.Len(t, got, len(want))
	for i := range want {
		w, g := want[i], got[i]
		require.Equal(t, w.Source, g.Source, "posting[%d].Source", i)
		require.Equal(t, w.Destination, g.Destination, "posting[%d].Destination", i)
		require.Equal(t, w.Asset, g.Asset, "posting[%d].Asset", i)
		require.Equal(t, w.Color, g.Color, "posting[%d].Color", i)
		require.Zero(t, g.Amount.Cmp(w.Amount), "posting[%d].Amount: got %s, want %s", i, g.Amount, w.Amount)
	}
}

func posting(source, destination string, amount int64) runtime.Posting {
	return runtime.Posting{Source: source, Destination: destination, Asset: "USD/2", Amount: big.NewInt(amount)}
}

func TestIRSend(t *testing.T) {
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 10
  $src = "src"
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $dest = "dest"
  send_to_account(account: $dest)
`, balances(map[string]int64{"src": 100}), nil)

	requirePostings(t, []runtime.Posting{posting("src", "dest", 10)}, res.Postings)
}

func TestIRSourceCappedByMinInt(t *testing.T) {
	// the `max [USD/2 20] from @src` shape: the cap is the smaller of the two
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 50
  $max = 20
  $cap = min_int($max, $amount)
  $src = "src"
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $overdraft)
  $dest = "dest"
  send_to_account(account: $dest)
`, balances(map[string]int64{"src": 100}), nil)

	requirePostings(t, []runtime.Posting{posting("src", "dest", 20)}, res.Postings)
}

func TestIRInorderSourcesStopAtFirstThatCovers(t *testing.T) {
	// @a holds enough, so the forward jump must skip @b entirely
	src := `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 10
  $pulled = 0
  $remaining = int_copy($amount)
  $a = "a"
  $overdraft = 0
  $from_a = pull_account(account: $a, cap: $remaining, overdraft: $overdraft)
  $pulled += $from_a
  $remaining -= $from_a
  jmp_if_zero($remaining, #inorder_end)
  $b = "b"
  $from_b = pull_account(account: $b, cap: $remaining, overdraft: $overdraft)
  $pulled += $from_b
#inorder_end
  check_enough_funds($pulled, $amount)
  $dest = "dest"
  send_to_account(account: $dest)
`

	t.Run("first source covers it", func(t *testing.T) {
		res := runIR(t, src, balances(map[string]int64{"a": 100, "b": 100}), nil)
		requirePostings(t, []runtime.Posting{posting("a", "dest", 10)}, res.Postings)
	})

	t.Run("falls through to the second", func(t *testing.T) {
		res := runIR(t, src, balances(map[string]int64{"a": 4, "b": 100}), nil)
		requirePostings(t, []runtime.Posting{
			posting("a", "dest", 4),
			posting("b", "dest", 6),
		}, res.Postings)
	})

	t.Run("neither covers it", func(t *testing.T) {
		execErr := runIRExpectingError(t, src, balances(map[string]int64{"a": 4, "b": 3}), nil)
		require.IsType(t, vm.MissingFundsError{}, execErr)
	})
}

func TestIRAllotmentDestination(t *testing.T) {
	// 1/4 to @small, the remaining 3/4 to @big
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 100
  $world = "world"
  $overdraft = 100
  $pulled = pull_account(account: $world, cap: $amount, overdraft: $overdraft)
  $one = 1
  $four = 4
  $quarter = mk_portion($one, $four)
  $whole = mk_portion($one, $one)
  $leftover = sub_portion($whole, $quarter)
  assert_leftover($leftover)
  [$small_share, $big_share] = mk_allot($amount, [$quarter, $leftover])
  $small = "small"
  send_to_account(account: $small, cap: $small_share)
  $big = "big"
  send_to_account(account: $big, cap: $big_share)
`, balances(nil), nil)

	requirePostings(t, []runtime.Posting{
		posting("world", "small", 25),
		posting("world", "big", 75),
	}, res.Postings)
}

func TestIRBalanceReadFromStore(t *testing.T) {
	// send exactly what @src holds, read at run time
	res := runIR(t, `
  $src = "src"
  $asset = "USD/2"
  $bal = balance($src, $asset)
  assert_non_negative_balance($bal, $src)
  $cur = get_asset($bal)
  set_current_asset($cur)
  $amount = get_amount($bal)
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $dest = "dest"
  send_to_account(account: $dest)
`, balances(map[string]int64{"src": 42}), nil)

	requirePostings(t, []runtime.Posting{posting("src", "dest", 42)}, res.Postings)
}

func TestIRUnsentFundsAreReturnedToTheSource(t *testing.T) {
	// send_to_account with no account: the `kept` destination. The funds are
	// released back and no posting is emitted for them.
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 100
  $src = "src"
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  $half = 50
  $dest = "dest"
  send_to_account(account: $dest, cap: $half)
  send_to_account()
`, balances(map[string]int64{"src": 100}), nil)

	requirePostings(t, []runtime.Posting{posting("src", "dest", 50)}, res.Postings)
}

func TestIRSnapshotRestoreBacktracks(t *testing.T) {
	// the `oneof` shape: try @a, and if it didn't cover the amount, rewind the
	// funding queue and take from @b instead
	src := `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 10
  $mark = snapshot()
  $a = "a"
  $overdraft = 0
  $from_a = pull_account(account: $a, cap: $amount, overdraft: $overdraft)
  $result = int_copy($from_a)
  $missing = $amount - $from_a
  jmp_if_zero($missing, #oneof_end)
  restore($mark)
  $b = "b"
  $from_b = pull_account(account: $b, cap: $amount, overdraft: $overdraft)
  $result = int_copy($from_b)
#oneof_end
  check_enough_funds($result, $amount)
  $dest = "dest"
  send_to_account(account: $dest)
`

	t.Run("first branch covers it", func(t *testing.T) {
		res := runIR(t, src, balances(map[string]int64{"a": 10, "b": 10}), nil)
		requirePostings(t, []runtime.Posting{posting("a", "dest", 10)}, res.Postings)
	})

	t.Run("rewinds to the second branch", func(t *testing.T) {
		// @a can only cover part of it, so its partial funding must be discarded
		res := runIR(t, src, balances(map[string]int64{"a": 3, "b": 10}), nil)
		requirePostings(t, []runtime.Posting{posting("b", "dest", 10)}, res.Postings)
	})
}

func TestIRSaveWithholdsFunds(t *testing.T) {
	// save reserves part of the balance, so the pull can't reach it
	src := `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "src"
  $reserved = 30
  save(account: $src, asset: $asset, amount: $reserved)
  $amount = 100
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  $dest = "dest"
  send_to_account(account: $dest)
`

	res := runIR(t, src, balances(map[string]int64{"src": 100}), nil)
	requirePostings(t, []runtime.Posting{posting("src", "dest", 70)}, res.Postings)
}

func TestIROverdraftAllowsNegativeBalance(t *testing.T) {
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 40
  $src = "src"
  $overdraft = 25
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $dest = "dest"
  send_to_account(account: $dest)
`, balances(map[string]int64{"src": 15}), nil)

	// 15 on the account plus 25 of allowed overdraft
	requirePostings(t, []runtime.Posting{posting("src", "dest", 40)}, res.Postings)
}

func TestIRMetadata(t *testing.T) {
	res := runIR(t, `
  $account = "acc"
  $key = "k"
  $value = "v"
  set_account_meta($account, $key, $value)
  $tx_key = "tx"
  $tx_value = "yes"
  set_tx_meta($tx_key, $tx_value)
`, balances(nil), nil)

	require.Equal(t, map[string]string{"tx": "yes"}, res.Metadata)
	require.Equal(t, runtime.AccountsMetadata{"acc": {"k": "v"}}, res.AccountsMetadata)
}

func TestIRReadsMetadataFromStore(t *testing.T) {
	// the amount to send is an int read out of @src's metadata
	store := irStore{
		balances: map[runtime.PairKey]*big.Int{
			{Account: "src", Asset: "USD/2"}: big.NewInt(100),
		},
		metadata: map[string]map[string]string{
			"src": {"quota": "7"},
		},
	}

	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "src"
  $key = "quota"
  $amount = meta<int>($src, $key)
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $dest = "dest"
  send_to_account(account: $dest)
`, store, nil)

	requirePostings(t, []runtime.Posting{posting("src", "dest", 7)}, res.Postings)
}

func TestIRMissingMetadataIsAnError(t *testing.T) {
	execErr := runIRExpectingError(t, `
  $src = "src"
  $key = "nope"
  $value = meta<str>($src, $key)
`, balances(nil), nil)

	require.IsType(t, vm.MetadataNotFoundError{}, execErr)
}

func TestIRLoadsVars(t *testing.T) {
	// vars come in as pools, indexed by the load_var instructions
	vars := &vm.Vars{
		StringsPool: []string{"USD/2", "src", "dest"},
		IntsPool:    []big.Int{*big.NewInt(10), *big.NewInt(0)},
	}

	res := runIR(t, `
  $asset = load_var<str>(0)
  set_current_asset($asset)
  $amount = load_var<int>(0)
  $src = load_var<str>(1)
  $overdraft = load_var<int>(1)
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $dest = load_var<str>(2)
  send_to_account(account: $dest)
`, balances(map[string]int64{"src": 100}), vars)

	requirePostings(t, []runtime.Posting{posting("src", "dest", 10)}, res.Postings)
}

func TestIRAssertions(t *testing.T) {
	t.Run("invalid account name", func(t *testing.T) {
		execErr := runIRExpectingError(t, `
  $account = "not a valid account!"
  assert_valid_account($account)
`, balances(nil), nil)
		require.IsType(t, vm.InvalidAccountName{}, execErr)
	})

	t.Run("mismatched assets", func(t *testing.T) {
		execErr := runIRExpectingError(t, `
  $usd = "USD/2"
  $eur = "EUR/2"
  assert_same_asset($usd, $eur)
`, balances(nil), nil)
		require.IsType(t, vm.AssetMismatchError{}, execErr)
	})

	t.Run("negative balance", func(t *testing.T) {
		store := irStore{balances: map[runtime.PairKey]*big.Int{
			{Account: "src", Asset: "USD/2"}: big.NewInt(-1),
		}}
		execErr := runIRExpectingError(t, `
  $src = "src"
  $asset = "USD/2"
  $bal = balance($src, $asset)
  assert_non_negative_balance($bal, $src)
`, store, nil)
		require.IsType(t, vm.NegativeBalanceError{}, execErr)
	})

	t.Run("allotment portions over 100%", func(t *testing.T) {
		execErr := runIRExpectingError(t, `
  $one = 1
  $two = 2
  $half = mk_portion($one, $two)
  $whole = mk_portion($one, $one)
  $leftover = sub_portion($whole, $half)
  $negative = sub_portion($leftover, $whole)
  assert_leftover($negative)
`, balances(nil), nil)
		require.IsType(t, vm.InvalidAllotmentSum{}, execErr)
	})
}
