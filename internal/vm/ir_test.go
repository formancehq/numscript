package vm_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// These tests drive the VM from the IR textual format, without the compiler, so
// they can also cover instruction sequences the compiler doesn't emit.

// irStore is a vm.Store backed by plain maps. A non-nil err fails every lookup.
type irStore struct {
	balances map[runtime.PairKey]*big.Int
	metadata map[string]map[string]string
	err      error
}

func (s irStore) GetBalance(_ context.Context, account, asset, color string) (*big.Int, error) {
	if s.err != nil {
		return nil, s.err
	}
	if v, ok := s.balances[runtime.PairKey{Account: account, Asset: asset, Color: color}]; ok {
		return new(big.Int).Set(v), nil
	}
	return new(big.Int), nil
}

func (s irStore) GetMetadata(_ context.Context, account, key string) (string, bool, error) {
	if s.err != nil {
		return "", false, s.err
	}
	v, ok := s.metadata[account][key]
	return v, ok, nil
}

func meta(rows map[string]map[string]string) irStore {
	return irStore{metadata: rows}
}

func balances(pairs map[string]int64) irStore {
	b := map[runtime.PairKey]*big.Int{}
	for account, amount := range pairs {
		b[runtime.PairKey{Account: account, Asset: "USD/2"}] = big.NewInt(amount)
	}
	return irStore{balances: b}
}

// allot2IR is the sequence the compiler emits to split an amount two ways:
// floor each share, then hand the flooring leftover to the earliest. There is
// no allotment instruction — the split is built out of pure ops — so the three
// tests that need one share this rather than spelling it out each time.
//
// Only one fixup block: flooring loses under a unit per share, so with the two
// portions summing to 1 the shortfall is at most 1 and the second share never
// receives it.
func allot2IR(amount, portion1, portion2, share1, share2 string) string {
	return fmt.Sprintf(`
  $allot_amt = int_to_portion($%[1]s)
  $allot_prod = mul_portion($%[2]s, $allot_amt)
  $%[4]s = portion_to_int($allot_prod)
  $allot_total = int_copy($%[4]s)
  $allot_prod = mul_portion($%[3]s, $allot_amt)
  $%[5]s = portion_to_int($allot_prod)
  $allot_total = add_int($allot_total, $%[5]s)
  $allot_one = 1
  $allot_short = lt_int($allot_total, $%[1]s)
  jmp_if_false($allot_short, #allot_end)
  $%[4]s = add_int($%[4]s, $allot_one)
  $allot_total = add_int($allot_total, $allot_one)
#allot_end
`, amount, portion1, portion2, share1, share2)
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

// The `max [USD/2 20] from @src` shape: the cap is the smaller of the two. There
// is no min opcode — it is lt_int plus a branch, so both arms need covering, and
// the ties too (lt_int is strict).
func TestIRSourceCappedByMin(t *testing.T) {
	// $cap = min($max, $amount), by copying $max and overwriting it unless it
	// already won
	src := `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = load_var<int>(0)
  $max = load_var<int>(1)
  $cap = int_copy($max)
  $lt = lt_int($max, $amount)
  jmp_if_true($lt, #min_end)
  $cap = int_copy($amount)
#min_end
  $src = "src"
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $overdraft)
  $dest = "dest"
  send_to_account(account: $dest)
`

	testCases := []struct {
		name        string
		amount, max int64
		wantSent    int64
	}{
		{"right operand is smaller", 20, 50, 20},
		{"left operand is smaller", 50, 20, 20},
		{"equal operands", 20, 20, 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vars := &vm.Vars{IntsPool: []big.Int{*big.NewInt(tc.amount), *big.NewInt(tc.max)}}
			res := runIR(t, src, balances(map[string]int64{"src": 100}), vars)
			requirePostings(t, []runtime.Posting{posting("src", "dest", tc.wantSent)}, res.Postings)
		})
	}
}

// A comparison drives a real branch end to end, including the `!=` spelling that
// has no opcode of its own (eq_int + not).
func TestIRComparisonBranch(t *testing.T) {
	// send the whole balance only when it differs from the requested amount,
	// otherwise send the amount — a shape numscript can't express yet, which is
	// the point of testing it here
	src := `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "src"
  $amount = 10
  $bal = balance($src, $asset)
  $same = eq_int($bal, $amount)
  $differs = not($same)
  $cap = int_copy($amount)
  jmp_if_false($differs, #end)
  $cap = int_copy($bal)
#end
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $overdraft)
  $dest = "dest"
  send_to_account(account: $dest)
`

	t.Run("balance differs, so it is sent whole", func(t *testing.T) {
		res := runIR(t, src, balances(map[string]int64{"src": 4}), nil)
		requirePostings(t, []runtime.Posting{posting("src", "dest", 4)}, res.Postings)
	})

	t.Run("balance equals the amount, so the amount is sent", func(t *testing.T) {
		res := runIR(t, src, balances(map[string]int64{"src": 10}), nil)
		requirePostings(t, []runtime.Posting{posting("src", "dest", 10)}, res.Postings)
	})
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
  $exhausted = is_zero($remaining)
  jmp_if_true($exhausted, #inorder_end)
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
`+allot2IR("amount", "quarter", "leftover", "small_share", "big_share")+`
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
  set_current_asset($asset)
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $bal, overdraft: $overdraft)
  check_enough_funds($pulled, $bal)
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
  $covered = is_zero($missing)
  jmp_if_true($covered, #oneof_end)
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

func TestIRStrEqAndJmp(t *testing.T) {
	// the if/else shape: str_eq is the only way to branch on a string, and jmp is
	// what skips the else arm. Here the taken arm decides which account is pulled
	// from, which is how @world's unboundedness is expressed in bytecode.
	src := `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 10
  $overdraft = 0
  $expected = "yes"
  $probe = load_var<str>(0)
  $eq = str_eq($probe, $expected)
  jmp_if_false($eq, #else)
  $then_acc = "a"
  $pulled = pull_account(account: $then_acc, cap: $amount, overdraft: $overdraft)
  jmp(#end)
#else
  $else_acc = "b"
  $pulled = pull_account(account: $else_acc, cap: $amount, overdraft: $overdraft)
#end
  $dest = "dest"
  send_to_account(account: $dest)
`

	store := balances(map[string]int64{"a": 10, "b": 10})

	t.Run("equal strings take the then arm", func(t *testing.T) {
		res := runIR(t, src, store, &vm.Vars{StringsPool: []string{"yes"}})
		requirePostings(t, []runtime.Posting{posting("a", "dest", 10)}, res.Postings)
	})

	t.Run("and jmp skips it otherwise", func(t *testing.T) {
		res := runIR(t, src, store, &vm.Vars{StringsPool: []string{"no"}})
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

	t.Run("invalid color", func(t *testing.T) {
		execErr := runIRExpectingError(t, `
  $color = "not a color"
  assert_valid_color($color)
`, balances(nil), nil)
		require.IsType(t, vm.InvalidColor{}, execErr)
	})

	t.Run("empty color is valid", func(t *testing.T) {
		res := runIR(t, `
  $color = ""
  assert_valid_color($color)
`, balances(nil), nil)
		require.Empty(t, res.Postings)
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

func TestIRUncappedPull(t *testing.T) {
	// no cap: the pull is bounded only by the overdraft, i.e. `send *`
	t.Run("with an overdraft", func(t *testing.T) {
		res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "src"
  $overdraft = 0
  $pulled = pull_account(account: $src, overdraft: $overdraft)
  $dest = "dest"
  send_to_account(account: $dest)
`, balances(map[string]int64{"src": 70}), nil)

		requirePostings(t, []runtime.Posting{posting("src", "dest", 70)}, res.Postings)
	})

	t.Run("without one it is unbounded and rejected", func(t *testing.T) {
		execErr := runIRExpectingError(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "src"
  $pulled = pull_account(account: $src)
`, balances(map[string]int64{"src": 70}), nil)

		require.IsType(t, vm.InvalidUncappedSource{}, execErr)
	})
}

func TestIRSaveAll(t *testing.T) {
	// save with no amount withholds the whole balance, so the pull finds nothing
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "src"
  save(account: $src, asset: $asset)
  $amount = 100
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  $dest = "dest"
  send_to_account(account: $dest)
`, balances(map[string]int64{"src": 100}), nil)

	require.Empty(t, res.Postings)
}

func TestIRAssertLeftoverExact(t *testing.T) {
	// the no-`remaining` form: portions must cover exactly 1
	execErr := runIRExpectingError(t, `
  $one = 1
  $two = 2
  $half = mk_portion($one, $two)
  $whole = mk_portion($one, $one)
  $leftover = sub_portion($whole, $half)
  assert_leftover_exact($leftover)
`, balances(nil), nil)

	require.IsType(t, vm.InvalidAllotmentSum{}, execErr)
}

func TestIRMetaTypes(t *testing.T) {
	store := meta(map[string]map[string]string{
		"acc": {
			"portion":  "1/4",
			"monetary": "USD/2 250",
			"oops":     "not a number",
		},
	})

	t.Run("portion", func(t *testing.T) {
		// the portion drives an allotment, so the split proves it parsed
		res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 100
  $world = "world"
  $overdraft = 100
  $pulled = pull_account(account: $world, cap: $amount, overdraft: $overdraft)
  $acc = "acc"
  $key = "portion"
  $quarter = meta<portion>($acc, $key)
  $one = 1
  $whole = mk_portion($one, $one)
  $rest = sub_portion($whole, $quarter)
  assert_leftover($rest)
`+allot2IR("amount", "quarter", "rest", "a_share", "b_share")+`
  $a = "a"
  send_to_account(account: $a, cap: $a_share)
  $b = "b"
  send_to_account(account: $b, cap: $b_share)
`, store, nil)

		requirePostings(t, []runtime.Posting{
			posting("world", "a", 25),
			posting("world", "b", 75),
		}, res.Postings)
	})

	t.Run("monetary", func(t *testing.T) {
		res := runIR(t, `
  $acc = "acc"
  $key = "monetary"
  [$asset, $amount] = meta_monetary($acc, $key)
  set_current_asset($asset)
  $overdraft = 300
  $pulled = pull_account(account: $acc, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $dest = "dest"
  send_to_account(account: $dest)
`, store, nil)

		requirePostings(t, []runtime.Posting{posting("acc", "dest", 250)}, res.Postings)
	})

	t.Run("a value of the wrong shape is an error", func(t *testing.T) {
		for _, read := range []string{
			`  $v = meta<int>($acc, $key)`,
			`  $v = meta<portion>($acc, $key)`,
			`  [$a, $n] = meta_monetary($acc, $key)`,
		} {
			execErr := runIRExpectingError(t, `
  $acc = "acc"
  $key = "oops"
`+read+"\n", store, nil)
			require.IsType(t, vm.BadMetaValueError{}, execErr, "%s", read)
		}
	})
}

func TestIRStoreErrorsPropagate(t *testing.T) {
	failing := irStore{err: errors.New("store is down")}

	t.Run("on a balance read", func(t *testing.T) {
		execErr := runIRExpectingError(t, `
  $src = "src"
  $asset = "USD/2"
  $bal = balance($src, $asset)
`, failing, nil)
		require.IsType(t, vm.StoreError{}, execErr)
		require.ErrorContains(t, execErr, "store is down")
	})

	t.Run("on a pull", func(t *testing.T) {
		execErr := runIRExpectingError(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 10
  $src = "src"
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
`, failing, nil)
		require.IsType(t, vm.StoreError{}, execErr)
	})

	t.Run("on a metadata read", func(t *testing.T) {
		for _, read := range []string{
			`  $v = meta<str>($acc, $key)`,
			`  $v = meta<int>($acc, $key)`,
			`  $v = meta<portion>($acc, $key)`,
			`  [$a, $n] = meta_monetary($acc, $key)`,
		} {
			execErr := runIRExpectingError(t, `
  $acc = "acc"
  $key = "k"
`+read+"\n", failing, nil)
			require.IsType(t, vm.StoreError{}, execErr, "%s", read)
		}
	})

	t.Run("on an uncapped pull", func(t *testing.T) {
		execErr := runIRExpectingError(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "src"
  $overdraft = 0
  $pulled = pull_account(account: $src, overdraft: $overdraft)
`, failing, nil)
		require.IsType(t, vm.StoreError{}, execErr)
	})

	t.Run("on a save", func(t *testing.T) {
		execErr := runIRExpectingError(t, `
  $acc = "acc"
  $asset = "USD/2"
  $amount = 10
  save(account: $acc, asset: $asset, amount: $amount)
`, failing, nil)
		require.IsType(t, vm.StoreError{}, execErr)
	})

	t.Run("but not on a send: crediting a destination reads nothing", func(t *testing.T) {
		// the pull has no overdraft operand, so it is unbounded and reads no balance
		// either — the whole send runs against a store that fails every call. This
		// is the arm the compiler emits for @world.
		res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $world = "world"
  $amount = 10
  $pulled = pull_account(account: $world, cap: $amount)
  $dest = "dest"
  send_to_account(account: $dest)
`, failing, nil)
		requirePostings(t, []runtime.Posting{posting("world", "dest", 10)}, res.Postings)
	})
}

func TestIRRestoreRejectsABogusMark(t *testing.T) {
	// a mark is a queue position, so one that can't fit an int64 is not one
	execErr := runIRExpectingError(t, `
  $mark = 99999999999999999999999999
  restore($mark)
`, balances(nil), nil)

	require.IsType(t, vm.InternalError{}, execErr)
	require.ErrorContains(t, execErr, "invalid snapshot id")
}

func TestIRVmIsReusableAcrossRuns(t *testing.T) {
	// a second run must not see the first one's funds or postings
	program := assembleIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 10
  $src = "src"
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $dest = "dest"
  send_to_account(account: $dest)
`)
	machine := vm.NewVm(program)
	store := balances(map[string]int64{"src": 100})

	want := []runtime.Posting{posting("src", "dest", 10)}
	for run := 1; run <= 3; run++ {
		res, execErr := vm.Exec(context.Background(), machine, nil, store)
		require.Nil(t, execErr, "run %d", run)
		requirePostings(t, want, res.Postings)
	}
}

// TestIRSurvivesTheWireFormat runs one program in memory and again after a trip
// through Encode/DecodeProgram. Nothing else ties the encoder to the VM.
// No instruction reads a bool yet, so this only pins down that the bank is
// allocated separately from the others and that the two ops run.
func TestIRConstBool(t *testing.T) {
	program := assembleIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $t = true
  $f = false
  $amount = 10
  $overdraft = 0
  $src = "src"
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  $dest = "dest"
  send_to_account(account: $dest)
`)

	require.Equal(t, byte(2), program.MaxRegBool)
	// $amount, $overdraft, $pulled — the two bools are not among them
	require.Equal(t, byte(3), program.MaxRegInt, "bools don't consume int registers")

	decoded, err := vm.DecodeProgram(program.Encode())
	require.NoError(t, err)
	require.Equal(t, program, decoded)

	res, execErr := vm.Exec(context.Background(), vm.NewVm(decoded), nil, balances(map[string]int64{"src": 10}))
	require.Nil(t, execErr, "unexpected execution error: %v", execErr)
	requirePostings(t, []runtime.Posting{posting("src", "dest", 10)}, res.Postings)
}

func TestIRSurvivesTheWireFormat(t *testing.T) {
	program := assembleIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 100
  $src = "src"
  $overdraft = 0
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $one = 1
  $two = 2
  $half = mk_portion($one, $two)
  $whole = mk_portion($one, $one)
  $rest = sub_portion($whole, $half)
  assert_leftover($rest)
`+allot2IR("amount", "half", "rest", "a_share", "b_share")+`
  $a = "a"
  send_to_account(account: $a, cap: $a_share)
  $b = "b"
  send_to_account(account: $b, cap: $b_share)
`)

	decoded, err := vm.DecodeProgram(program.Encode())
	require.NoError(t, err)
	require.Equal(t, program, decoded, "the program changed shape on the way through")

	want := []runtime.Posting{posting("src", "a", 50), posting("src", "b", 50)}
	for name, prog := range map[string]vm.Program{"in memory": program, "decoded": decoded} {
		res, execErr := vm.Exec(context.Background(), vm.NewVm(prog), nil, balances(map[string]int64{"src": 100}))
		require.Nil(t, execErr, "%s: %v", name, execErr)
		requirePostings(t, want, res.Postings)
	}
}

// TestIRVarsSurviveTheWireFormat is the same for the vars payload.
func TestIRVarsSurviveTheWireFormat(t *testing.T) {
	vars := vm.Vars{
		StringsPool: []string{"USD/2", "src", "dest"},
		IntsPool:    []big.Int{*big.NewInt(10), *big.NewInt(0)},
	}
	decoded, err := vm.DecodeVars(vars.Encode())
	require.NoError(t, err)

	program := assembleIR(t, `
  $asset = load_var<str>(0)
  set_current_asset($asset)
  $amount = load_var<int>(0)
  $src = load_var<str>(1)
  $overdraft = load_var<int>(1)
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $dest = load_var<str>(2)
  send_to_account(account: $dest)
`)

	res, execErr := vm.Exec(context.Background(), vm.NewVm(program), &decoded, balances(map[string]int64{"src": 100}))
	require.Nil(t, execErr)
	requirePostings(t, []runtime.Posting{posting("src", "dest", 10)}, res.Postings)
}

// --- The int/portion boundary ops -------------------------------------------

func TestIRPortionToIntFloors(t *testing.T) {
	// 7/2 of nothing in particular: the projection floors, it does not round
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $seven = 7
  $two = 2
  $p = mk_portion($seven, $two)
  $amount = portion_to_int($p)
  $world = "world"
  $overdraft = 100
  $pulled = pull_account(account: $world, cap: $amount, overdraft: $overdraft)
  $dest = "dest"
  send_to_account(account: $dest)
`, balances(nil), nil)

	requirePostings(t, []runtime.Posting{posting("world", "dest", 3)}, res.Postings)
}

func TestIRIntToPortionAndMul(t *testing.T) {
	// 1/4 * 100 == 25, computed as mul_portion(int_to_portion(100), 1/4)
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $hundred = 100
  $one = 1
  $four = 4
  $quarter = mk_portion($one, $four)
  $ap = int_to_portion($hundred)
  $prod = mul_portion($quarter, $ap)
  $amount = portion_to_int($prod)
  $world = "world"
  $overdraft = 100
  $pulled = pull_account(account: $world, cap: $amount, overdraft: $overdraft)
  $dest = "dest"
  send_to_account(account: $dest)
`, balances(nil), nil)

	requirePostings(t, []runtime.Posting{posting("world", "dest", 25)}, res.Postings)
}

// A three-way split written out of pure ops: floor each share, then hand the
// flooring leftover to the earliest shares one unit at a time. This is the
// lowering compileAllotmentSplit emits, pinned here on the 34/33/33 case.
//
// Only n-1 fixup blocks: each floor loses < 1, so with portions summing to 1 the
// shortfall is <= n-1 and the last share never receives a unit.
const allotThirdsIR = `
  $asset = "USD/2"
  set_current_asset($asset)
  $amount = 100
  $world = "world"
  $overdraft = 100
  $pulled = pull_account(account: $world, cap: $amount, overdraft: $overdraft)

  $one = 1
  $three = 3
  $third = mk_portion($one, $three)
  $ap = int_to_portion($amount)

  $prod = mul_portion($third, $ap)
  $out0 = portion_to_int($prod)
  $total = int_copy($out0)
  $prod = mul_portion($third, $ap)
  $out1 = portion_to_int($prod)
  $total = add_int($total, $out1)
  $prod = mul_portion($third, $ap)
  $out2 = portion_to_int($prod)
  $total = add_int($total, $out2)

  $lt = lt_int($total, $amount)
  jmp_if_false($lt, #done)
  $out0 = add_int($out0, $one)
  $total = add_int($total, $one)
  $lt = lt_int($total, $amount)
  jmp_if_false($lt, #done)
  $out1 = add_int($out1, $one)
  $total = add_int($total, $one)
#done

  $a = "a"
  send_to_account(account: $a, cap: $out0)
  $b = "b"
  send_to_account(account: $b, cap: $out1)
  $c = "c"
  send_to_account(account: $c, cap: $out2)
`

func TestIRAllotmentFromPureOps(t *testing.T) {
	res := runIR(t, allotThirdsIR, balances(nil), nil)

	requirePostings(t, []runtime.Posting{
		posting("world", "a", 34),
		posting("world", "b", 33),
		posting("world", "c", 33),
	}, res.Postings)
}
