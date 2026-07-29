package compiler

import (
	"testing"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func getCompiledOutput(t *testing.T, source string) string {
	t.Helper()
	program := parser.Parse(source)
	require.Empty(t, program.Errors)
	compiled, err := compileProgramToIR(program.Value, nil)
	require.Nil(t, err)

	out := "\n" + ir.Dump(compiled.instructions)

	// every snapshot below doubles as a round-trip test of the textual format
	instrs, errs := ir.Parse(out)
	require.Empty(t, errs, "the dump does not parse back")
	require.Equal(t, out, "\n"+ir.Dump(instrs), "the dump does not round-trip")

	return out
}

func TestSimpleProgram(t *testing.T) {
	out := getCompiledOutput(t, `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  set_current_asset($r3)
  $r4 = get_amount($r2)
  $r5 = "src"
  $r6 = 0
  $r7 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
  check_enough_funds($r7, $r4)
  $r8 = "dest"
  send_to_account(account: $r8)
`))
}

func TestIntAddition(t *testing.T) {
	out := getCompiledOutput(t, `
		send [USD/2 4 + 6] (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 4
  $r2 = 6
  $r3 = $r1 + $r2
  $r4 = mk_monetary($r0, $r3)
  $r5 = get_asset($r4)
  set_current_asset($r5)
  $r6 = get_amount($r4)
  $r7 = "src"
  $r8 = 0
  $r9 = pull_account(account: $r7, cap: $r6, overdraft: $r8)
  check_enough_funds($r9, $r6)
  $r10 = "dest"
  send_to_account(account: $r10)
`))
}

func TestIntSubtraction(t *testing.T) {
	out := getCompiledOutput(t, `
		send [USD/2 16 - 6] (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 16
  $r2 = 6
  $r3 = $r1 - $r2
  $r4 = mk_monetary($r0, $r3)
  $r5 = get_asset($r4)
  set_current_asset($r5)
  $r6 = get_amount($r4)
  $r7 = "src"
  $r8 = 0
  $r9 = pull_account(account: $r7, cap: $r6, overdraft: $r8)
  check_enough_funds($r9, $r6)
  $r10 = "dest"
  send_to_account(account: $r10)
`))
}

func TestMonetaryAddition(t *testing.T) {
	out := getCompiledOutput(t, `
		vars {
			monetary $a = [USD/2 3]
			monetary $b = [USD/2 7]
		}
		send $a + $b (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 3
  $r2 = mk_monetary($r0, $r1)
  $r3 = "USD/2"
  $r4 = 7
  $r5 = mk_monetary($r3, $r4)
  $r6 = get_asset($r2)
  $r7 = get_asset($r5)
  assert_same_asset($r6, $r7)
  $r8 = get_amount($r2)
  $r9 = get_amount($r5)
  $r10 = $r8 + $r9
  $r11 = mk_monetary($r6, $r10)
  $r12 = get_asset($r11)
  set_current_asset($r12)
  $r13 = get_amount($r11)
  $r14 = "src"
  $r15 = 0
  $r16 = pull_account(account: $r14, cap: $r13, overdraft: $r15)
  check_enough_funds($r16, $r13)
  $r17 = "dest"
  send_to_account(account: $r17)
`))
}

func TestMonetarySubtraction(t *testing.T) {
	out := getCompiledOutput(t, `
		vars {
			monetary $a = [USD/2 30]
			monetary $b = [USD/2 20]
		}
		send $a - $b (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 30
  $r2 = mk_monetary($r0, $r1)
  $r3 = "USD/2"
  $r4 = 20
  $r5 = mk_monetary($r3, $r4)
  $r6 = get_asset($r2)
  $r7 = get_asset($r5)
  assert_same_asset($r6, $r7)
  $r8 = get_amount($r2)
  $r9 = get_amount($r5)
  $r10 = $r8 - $r9
  $r11 = mk_monetary($r6, $r10)
  $r12 = get_asset($r11)
  set_current_asset($r12)
  $r13 = get_amount($r11)
  $r14 = "src"
  $r15 = 0
  $r16 = pull_account(account: $r14, cap: $r13, overdraft: $r15)
  check_enough_funds($r16, $r13)
  $r17 = "dest"
  send_to_account(account: $r17)
`))
}

func TestGetAmount(t *testing.T) {
	out := getCompiledOutput(t, `
		#![feature("experimental-get-amount-function")]
		vars {
			monetary $m = [USD/2 42]
			number $n = get_amount($m)
		}
		send [USD/2 $n] (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 42
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_amount($r2)
  $r4 = "USD/2"
  $r5 = mk_monetary($r4, $r3)
  $r6 = get_asset($r5)
  set_current_asset($r6)
  $r7 = get_amount($r5)
  $r8 = "src"
  $r9 = 0
  $r10 = pull_account(account: $r8, cap: $r7, overdraft: $r9)
  check_enough_funds($r10, $r7)
  $r11 = "dest"
  send_to_account(account: $r11)
`))
}

func TestGetAsset(t *testing.T) {
	out := getCompiledOutput(t, `
		#![feature("experimental-get-asset-function")]
		vars {
			monetary $m = [USD/2 42]
			asset $a = get_asset($m)
		}
		send [$a 10] (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 42
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  $r4 = 10
  $r5 = mk_monetary($r3, $r4)
  $r6 = get_asset($r5)
  set_current_asset($r6)
  $r7 = get_amount($r5)
  $r8 = "src"
  $r9 = 0
  $r10 = pull_account(account: $r8, cap: $r7, overdraft: $r9)
  check_enough_funds($r10, $r7)
  $r11 = "dest"
  send_to_account(account: $r11)
`))
}

func TestPrefixMinusMonetary(t *testing.T) {
	out := getCompiledOutput(t, `
		vars {
			monetary $neg_mon = [USD/2 -10]
			monetary $pos_mon = -$neg_mon
		}
		send $pos_mon (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 10
  $r2 = neg_int($r1)
  $r3 = mk_monetary($r0, $r2)
  $r4 = get_amount($r3)
  $r5 = neg_int($r4)
  $r6 = get_asset($r3)
  $r7 = mk_monetary($r6, $r5)
  $r8 = get_asset($r7)
  set_current_asset($r8)
  $r9 = get_amount($r7)
  $r10 = "src"
  $r11 = 0
  $r12 = pull_account(account: $r10, cap: $r9, overdraft: $r11)
  check_enough_funds($r12, $r9)
  $r13 = "dest"
  send_to_account(account: $r13)
`))
}

func TestBalance(t *testing.T) {
	out := getCompiledOutput(t, `
		vars {
			monetary $bal = balance(@src, USD/2)
		}
		send $bal (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "src"
  $r1 = "USD/2"
  $r2 = balance($r0, $r1)
  assert_non_negative_balance($r2, $r0)
  $r3 = get_asset($r2)
  set_current_asset($r3)
  $r4 = get_amount($r2)
  $r5 = "src"
  $r6 = 0
  $r7 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
  check_enough_funds($r7, $r4)
  $r8 = "dest"
  send_to_account(account: $r8)
`))
}

func TestAccountInterpolation(t *testing.T) {
	out := getCompiledOutput(t, `
		#![feature("experimental-account-interpolation")]
		vars {
			string $id = "alice"
		}
		send [USD/2 10] (
			source = @world
			destination = @users:$id:wallet
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "alice"
  $r1 = "USD/2"
  $r2 = 10
  $r3 = mk_monetary($r1, $r2)
  $r4 = get_asset($r3)
  set_current_asset($r4)
  $r5 = get_amount($r3)
  $r6 = "world"
  $r7 = 0
  $r8 = pull_account(account: $r6, cap: $r5, overdraft: $r7)
  check_enough_funds($r8, $r5)
  $r9 = "users"
  $r10 = ":"
  $r11 = ":"
  $r12 = "wallet"
  $r13 = add_string($r9, $r10)
  $r14 = add_string($r13, $r0)
  $r15 = add_string($r14, $r11)
  $r16 = add_string($r15, $r12)
  assert_valid_account($r16)
  send_to_account(account: $r16)
`))
}

func TestAccountInterpolationInt(t *testing.T) {
	out := getCompiledOutput(t, `
		#![feature("experimental-account-interpolation")]
		vars {
			number $n = 42
		}
		send [USD/2 10] (
			source = @world
			destination = @account:$n
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = 42
  $r1 = "USD/2"
  $r2 = 10
  $r3 = mk_monetary($r1, $r2)
  $r4 = get_asset($r3)
  set_current_asset($r4)
  $r5 = get_amount($r3)
  $r6 = "world"
  $r7 = 0
  $r8 = pull_account(account: $r6, cap: $r5, overdraft: $r7)
  check_enough_funds($r8, $r5)
  $r9 = "account"
  $r10 = ":"
  $r11 = int_to_string($r0)
  $r12 = add_string($r9, $r10)
  $r13 = add_string($r12, $r11)
  assert_valid_account($r13)
  send_to_account(account: $r13)
`))
}

func TestInorder(t *testing.T) {
	out := getCompiledOutput(t, `
		send [USD/2 10] (
			source = {
				@a
				@b
				@c
			}
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  set_current_asset($r3)
  $r4 = get_amount($r2)
  $r5 = 0
  $r6 = int_copy($r4)
  $r7 = "a"
  $r8 = 0
  $r9 = pull_account(account: $r7, cap: $r6, overdraft: $r8)
  $r5 += $r9
  $r6 -= $r9
  jmp_if_zero($r6, #inorder_end_0)
  $r10 = "b"
  $r11 = 0
  $r12 = pull_account(account: $r10, cap: $r6, overdraft: $r11)
  $r5 += $r12
  $r6 -= $r12
  jmp_if_zero($r6, #inorder_end_0)
  $r13 = "c"
  $r14 = 0
  $r15 = pull_account(account: $r13, cap: $r6, overdraft: $r14)
  $r5 += $r15
#inorder_end_0
  check_enough_funds($r5, $r4)
  $r16 = "dest"
  send_to_account(account: $r16)
`))
}

func TestInorderWithCap(t *testing.T) {
	out := getCompiledOutput(t, `
		send [USD/2 10] (
			source = {
				@a
				max [USD/2 5] from @b
				@c
			}
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  set_current_asset($r3)
  $r4 = get_amount($r2)
  $r5 = 0
  $r6 = int_copy($r4)
  $r7 = "a"
  $r8 = 0
  $r9 = pull_account(account: $r7, cap: $r6, overdraft: $r8)
  $r5 += $r9
  $r6 -= $r9
  jmp_if_zero($r6, #inorder_end_0)
  $r10 = "USD/2"
  $r11 = 5
  $r12 = mk_monetary($r10, $r11)
  $r13 = get_asset($r12)
  assert_same_asset($r13, $r3)
  $r14 = get_amount($r12)
  $r15 = min_int($r14, $r6)
  $r16 = "b"
  $r17 = 0
  $r18 = pull_account(account: $r16, cap: $r15, overdraft: $r17)
  $r5 += $r18
  $r6 -= $r18
  jmp_if_zero($r6, #inorder_end_0)
  $r19 = "c"
  $r20 = 0
  $r21 = pull_account(account: $r19, cap: $r6, overdraft: $r20)
  $r5 += $r21
#inorder_end_0
  check_enough_funds($r5, $r4)
  $r22 = "dest"
  send_to_account(account: $r22)
`))
}

func TestDestInorder(t *testing.T) {
	out := getCompiledOutput(t, `
		send [USD/2 10] (
			source = @world
			destination = {
        max [USD/2 4] to @d1
        remaining to @d2
      }
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  set_current_asset($r3)
  $r4 = get_amount($r2)
  $r5 = "world"
  $r6 = 0
  $r7 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
  check_enough_funds($r7, $r4)
  $r8 = int_copy($r7)
  $r9 = "USD/2"
  $r10 = 4
  $r11 = mk_monetary($r9, $r10)
  $r12 = get_asset($r11)
  assert_same_asset($r12, $r3)
  $r13 = get_amount($r11)
  $r14 = min_int($r8, $r13)
  $r15 = "d1"
  send_to_account(account: $r15, cap: $r14)
  $r8 -= $r14
  $r16 = "d2"
  send_to_account(account: $r16, cap: $r8)
`))
}

func TestSourceOneofSimple(t *testing.T) {
	out := getCompiledOutput(t, `
		#![feature("experimental-oneof")]
		send [USD/2 10] (
			source = oneof {
        @a
        @b
				@c
			}
			destination = @dest
		)
	`)
	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  set_current_asset($r3)
  $r4 = get_amount($r2)
  $r5 = snapshot()
  $r6 = "a"
  $r7 = 0
  $r8 = pull_account(account: $r6, cap: $r4, overdraft: $r7)
  $r9 = int_copy($r8)
  $r10 = $r4 - $r8
  jmp_if_zero($r10, #oneof_end_0)
  restore($r5)
  $r11 = "b"
  $r12 = 0
  $r13 = pull_account(account: $r11, cap: $r4, overdraft: $r12)
  $r9 = int_copy($r13)
  $r14 = $r4 - $r13
  jmp_if_zero($r14, #oneof_end_0)
  restore($r5)
  $r15 = "c"
  $r16 = 0
  $r17 = pull_account(account: $r15, cap: $r4, overdraft: $r16)
  $r9 = int_copy($r17)
#oneof_end_0
  check_enough_funds($r9, $r4)
  $r18 = "dest"
  send_to_account(account: $r18)
`))
}

func TestSourceOneofBounded(t *testing.T) {

	out := getCompiledOutput(t, `
		#![feature("experimental-oneof")]
		send [USD/2 10] (
			source = oneof {
				@a
				@b
			}
			destination = @dest
		)
	`)
	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  set_current_asset($r3)
  $r4 = get_amount($r2)
  $r5 = snapshot()
  $r6 = "a"
  $r7 = 0
  $r8 = pull_account(account: $r6, cap: $r4, overdraft: $r7)
  $r9 = int_copy($r8)
  $r10 = $r4 - $r8
  jmp_if_zero($r10, #oneof_end_0)
  restore($r5)
  $r11 = "b"
  $r12 = 0
  $r13 = pull_account(account: $r11, cap: $r4, overdraft: $r12)
  $r9 = int_copy($r13)
#oneof_end_0
  check_enough_funds($r9, $r4)
  $r14 = "dest"
  send_to_account(account: $r14)
`),
	)
}

func TestDestOneof(t *testing.T) {
	out := getCompiledOutput(t, `
		#![feature("experimental-oneof")]
		send [USD/2 10] (
			source = @world
			destination = oneof {
				max [USD/2 4] to @a
				remaining to @b
			}
		)
	`)
	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  set_current_asset($r3)
  $r4 = get_amount($r2)
  $r5 = "world"
  $r6 = 0
  $r7 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
  check_enough_funds($r7, $r4)
  $r8 = 0
  $r9 = "USD/2"
  $r10 = 4
  $r11 = mk_monetary($r9, $r10)
  $r12 = get_asset($r11)
  assert_same_asset($r12, $r3)
  $r13 = get_amount($r11)
  $r14 = min_int($r7, $r13)
  $r15 = $r7 - $r14
  jmp_if_zero($r15, #oneof_dest_clause_1)
  $r16 = "b"
  send_to_account(account: $r16)
  jmp_if_zero($r8, #oneof_dest_end_0)
#oneof_dest_clause_1
  $r17 = "a"
  send_to_account(account: $r17)
  jmp_if_zero($r8, #oneof_dest_end_0)
#oneof_dest_end_0
`))
}

// pins which meta type the compiler picks per source type — the tag is what lets
// the execution result carry a typed value back out of the VM
func TestMetadataWrites(t *testing.T) {
	out := getCompiledOutput(t, `
		vars {
			monetary $m
			portion $p
		}
		set_tx_meta("str", "abc")
		set_tx_meta("account", @acc)
		set_tx_meta("asset", COIN)
		set_tx_meta("int", 42)
		set_tx_meta("portion", $p)
		set_account_meta(@acc, "monetary", $m)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = load_var<str>(0)
  $r1 = load_var<int>(0)
  $r2 = mk_monetary($r0, $r1)
  $r3 = load_var<int>(1)
  $r4 = load_var<int>(2)
  $r5 = mk_portion($r3, $r4)
  $r6 = "str"
  $r7 = "abc"
  set_tx_meta<str>($r6, $r7)
  $r8 = "account"
  $r9 = "acc"
  set_tx_meta<account>($r8, $r9)
  $r10 = "asset"
  $r11 = "COIN"
  set_tx_meta<asset>($r10, $r11)
  $r12 = "int"
  $r13 = 42
  set_tx_meta<int>($r12, $r13)
  $r14 = "portion"
  set_tx_meta<portion>($r14, $r5)
  $r15 = "acc"
  $r16 = "monetary"
  set_account_meta<monetary>($r15, $r16, $r2)
`),
	)
}
