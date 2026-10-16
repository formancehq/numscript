package vm_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/funds"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// post_account credits the destination and leaves the source alone: the
// balances read back after it show the credit and the untouched store value.
func TestIRPostAccount(t *testing.T) {
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "world"
  $dest = "dest"
  $amount = 42
  post_account(source: $src, destination: $dest, amount: $amount)
  $dest_bal = balance($dest, $asset)
  $src_bal = balance($src, $asset)
  $k1 = "dest"
  $v1 = int_to_string($dest_bal)
  set_tx_meta($k1, $v1)
  $k2 = "src"
  $v2 = int_to_string($src_bal)
  set_tx_meta($k2, $v2)
`, balances(map[string]int64{"world": 100, "dest": 1}), nil)

	requirePostings(t, []funds.Posting{posting("world", "dest", 42)}, res.Postings)
	require.Equal(t, map[string]string{"dest": "43", "src": "100"}, res.Metadata)
}

func TestIRPostAccountNonPositiveAmountPostsNothing(t *testing.T) {
	for name, amount := range map[string]string{"zero": "0", "negative": "-5"} {
		t.Run(name, func(t *testing.T) {
			res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "world"
  $dest = "dest"
  $amount = `+amount+`
  post_account(source: $src, destination: $dest, amount: $amount)
`, irStore{}, nil)
			require.Empty(t, res.Postings)
		})
	}
}

func TestIRPostAccountCarriesColorAndScopes(t *testing.T) {
	res := runIR(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "world"
  $dest = "dest"
  $amount = 7
  $color = "RED"
  $src_scope = "s1"
  $dest_scope = "s2"
  post_account(source: $src, destination: $dest, amount: $amount, color: $color, source_scope: $src_scope, destination_scope: $dest_scope)
`, irStore{}, nil)

	require.Equal(t, []funds.Posting{{
		Source:           "world",
		SourceScope:      "s1",
		Destination:      "dest",
		DestinationScope: "s2",
		Asset:            "USD/2",
		Color:            "RED",
		Amount:           big.NewInt(7),
	}}, res.Postings)
}

// The posting of a statement fused by the compiler lands where its send was,
// after the check: a short cap still fails as missing funds, with the clamped
// amount as what was got.
func TestIRPostAccountAfterCheckEnoughFunds(t *testing.T) {
	execErr := runIRExpectingError(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $needed = 42
  $cap = -3
  $src = "world"
  $pulled = max_zero($cap)
  check_enough_funds($pulled, $needed)
  $dest = "dest"
  post_account(source: $src, destination: $dest, amount: $pulled)
`, irStore{}, nil)

	require.Equal(t, vm.MissingFundsError{Asset: "USD/2", Got: big.NewInt(0), Needed: big.NewInt(42)}, execErr)
}

func TestIRPostAccountWhileMarkOpen(t *testing.T) {
	execErr := runUnverifiedIRExpectingError(t, `
  $asset = "USD/2"
  set_current_asset($asset)
  $src = "world"
  $dest = "dest"
  $amount = 42
  mark_push()
  post_account(source: $src, destination: $dest, amount: $amount)
  mark_commit()
`, "while a mark is open", irStore{}, nil)

	var internal vm.InternalError
	require.ErrorAs(t, execErr, &internal)
}

func TestIRMaxZero(t *testing.T) {
	for name, tc := range map[string]struct {
		in, want string
	}{
		"negative": {"-5", "0"},
		"zero":     {"0", "0"},
		"positive": {"7", "7"},
	} {
		t.Run(name, func(t *testing.T) {
			res := runIR(t, `
  $in = `+tc.in+`
  $out = max_zero($in)
  $k = "out"
  $v = int_to_string($out)
  set_tx_meta($k, $v)
`, irStore{}, nil)
			require.Equal(t, map[string]string{"out": tc.want}, res.Metadata)
		})
	}
}
