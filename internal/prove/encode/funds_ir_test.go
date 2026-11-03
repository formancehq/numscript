package encode

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

const irPrologue = `
  $asset = "USD/2"
  set_current_asset($asset)
  $zero = 0
  $src = "src"
  $dest = "dest"
  $a = "a"
  $b = "b"
  $c = "c"
`

// Shapes the compiler never emits, as hand-written IR.
var handIR = map[string]string{
	"capped send then drain": irPrologue + `
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  $thirty = 30
  send_to_account(account: $dest, cap: $thirty)
  send_to_account(account: $a)
`,
	"negative cap": irPrologue + `
  $cap = -5
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  send_to_account(account: $dest)
`,
	"uncapped overdraft pull": irPrologue + `
  $od = 20
  $pulled = pull_account(account: $src, overdraft: $od)
  send_to_account(account: $dest)
`,
	"unbounded pull": irPrologue + `
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap)
  send_to_account(account: $dest)
`,
	"asset switch between pull and send": irPrologue + `
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  $eur = "EUR/2"
  set_current_asset($eur)
  send_to_account(account: $dest)
`,
	"undrained queue": irPrologue + `
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
`,
	"kept then send": irPrologue + `
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  check_enough_funds($pulled, $cap)
  $forty = 40
  send_to_account(cap: $forty)
  send_to_account(account: $dest)
`,
	"save then pull": irPrologue + `
  $fifty = 50
  save(account: $src, asset: $asset, amount: $fifty)
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  check_enough_funds($pulled, $cap)
  send_to_account(account: $dest)
`,
	"save all then pull": irPrologue + `
  save(account: $src, asset: $asset)
  $cap = 100
  $od = 30
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $od)
  send_to_account(account: $dest)
`,
	"balance read mid script": irPrologue + `
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  $bal = balance($src, $asset)
  assert_non_negative_balance($bal, $src)
  send_to_account(account: $dest, cap: $bal)
  send_to_account(account: $a)
`,
	"portion split": irPrologue + `
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  $one = 1
  $three = 3
  $third = mk_portion($one, $three)
  $whole = mk_portion($one, $one)
  $rest = sub_portion($whole, $third)
  assert_leftover($rest)
  $pp = int_to_portion($pulled)
  $prod1 = mul_portion($third, $pp)
  $share1 = portion_to_int($prod1)
  $prod2 = mul_portion($rest, $pp)
  $share2 = portion_to_int($prod2)
  $total = $share1 + $share2
  $short = lt_int($total, $pulled)
  jmp_if_false($short, #allot_end)
  $share1 += $one
#allot_end
  send_to_account(account: $a, cap: $share1)
  send_to_account(account: $b, cap: $share2)
`,
	"division by zero": irPrologue + `
  $one = 1
  $p = mk_portion($one, $zero)
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  send_to_account(account: $dest)
`,
	"asset mismatch": irPrologue + `
  $eur = "EUR/2"
  assert_same_asset($asset, $eur)
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  send_to_account(account: $dest)
`,
	"oneof": irPrologue + `
  $cap = 10
  mark_push()
  $pa = pull_account(account: $a, cap: $cap, overdraft: $zero)
  $res = int_copy($pa)
  $missing = $cap - $pa
  $done = is_zero($missing)
  jmp_if_true($done, #end)
  mark_rewind()
  mark_push()
  $pb = pull_account(account: $b, cap: $cap, overdraft: $zero)
  $res = int_copy($pb)
#end
  mark_commit()
  check_enough_funds($res, $cap)
  send_to_account(account: $dest)
`,
	"rewind after a join": irPrologue + `
  $cap = 10
  $five = 5
  mark_push()
  $pa = pull_account(account: $a, cap: $cap, overdraft: $zero)
  $small = lt_int($pa, $five)
  jmp_if_true($small, #join)
  mark_rewind()
  mark_push()
  $pb = pull_account(account: $b, cap: $cap, overdraft: $zero)
#join
  mark_rewind()
  mark_push()
  $pc = pull_account(account: $c, cap: $cap, overdraft: $zero)
  mark_commit()
  send_to_account(account: $dest)
`,
	"two statements": irPrologue + `
  $cap = 100
  $pulled = pull_account(account: $src, cap: $cap, overdraft: $zero)
  check_enough_funds($pulled, $cap)
  send_to_account(account: $dest)
  $cap2 = 30
  $pulled2 = pull_account(account: $dest, cap: $cap2, overdraft: $zero)
  check_enough_funds($pulled2, $cap2)
  send_to_account(account: $a)
`,
}

var assignments = []map[Key]*big.Int{
	{},
	bal("src", "USD/2", 100, "a", "USD/2", 100, "b", "USD/2", 100, "c", "USD/2", 100),
	bal("src", "USD/2", 3, "a", "USD/2", 4, "b", "USD/2", 12, "c", "USD/2", 1),
	bal("src", "USD/2", -3, "a", "USD/2", -4, "b", "USD/2", 10, "dest", "USD/2", -7),
	bal("src", "USD/2", 25, "a", "USD/2", 10, "b", "USD/2", 7, "c", "USD/2", 10, "dest", "USD/2", 5),
	bal("src", "USD/2", 150, "a", "USD/2", 0, "b", "USD/2", 2),
	bal("src", "USD/2", 70, "src", "EUR/2", 9, "dest", "EUR/2", 1),
}

func TestHandIRAgreesWithVM(t *testing.T) {
	for name, src := range handIR {
		t.Run(name, func(t *testing.T) {
			instrs := parseIR(t, src)
			for i, before := range assignments {
				t.Run(fmt.Sprint(i), func(t *testing.T) {
					checkFixed(t, instrs, before)
				})
			}
		})
	}
}

func TestHandIRExpectedOutcomes(t *testing.T) {
	out := checkFixed(t, parseIR(t, handIR["capped send then drain"]), bal("src", "USD/2", 100))
	require.True(t, out.Success)
	require.Equal(t, "30", out.Received[Key{"dest", "USD/2"}].String())
	require.Equal(t, "70", out.Received[Key{"a", "USD/2"}].String())

	out = checkFixed(t, parseIR(t, handIR["kept then send"]), bal("src", "USD/2", 100))
	require.True(t, out.Success)
	require.Equal(t, "60", out.Sent[Key{"src", "USD/2"}].String())
	require.Equal(t, "40", out.After[Key{"src", "USD/2"}].String())

	out = checkFixed(t, parseIR(t, handIR["save then pull"]), bal("src", "USD/2", 120))
	require.False(t, out.Success)
	require.Equal(t, MissingFunds, out.FailKind)

	out = checkFixed(t, parseIR(t, handIR["oneof"]), bal("a", "USD/2", 4, "b", "USD/2", 12))
	require.True(t, out.Success)
	require.Equal(t, "0", out.Sent[Key{"a", "USD/2"}].String())
	require.Equal(t, "10", out.Sent[Key{"b", "USD/2"}].String())
	require.Equal(t, "4", out.After[Key{"a", "USD/2"}].String())

	out = checkFixed(t, parseIR(t, handIR["division by zero"]), bal("src", "USD/2", 100))
	require.False(t, out.Success)
	require.Equal(t, DivideByZero, out.FailKind)

	out = checkFixed(t, parseIR(t, handIR["asset switch between pull and send"]), bal("src", "USD/2", 100))
	require.True(t, out.Success)
	require.Equal(t, "100", out.After[Key{"src", "USD/2"}].String(), "no USD posting exists, so the ledger never debits USD")
	require.Equal(t, "100", out.Received[Key{"dest", "EUR/2"}].String())
}
