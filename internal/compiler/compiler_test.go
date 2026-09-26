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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 10
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = lt_int($r3, $r2)
  jmp_if_true($r5, #max_end_0)
  $r4 = int_copy($r3)
#max_end_0
  $r6 = "src"
  $r7 = 0
  $r8 = str_eq($r6, $r0)
  jmp_if_false($r8, #not_world_1)
  $r9 = pull_account(account: $r6, cap: $r4)
  jmp(#pull_end_2)
#not_world_1
  $r9 = pull_account(account: $r6, cap: $r4, overdraft: $r7)
#pull_end_2
  check_enough_funds($r9, $r2)
  $r10 = "dest"
  send_to_account(account: $r10)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 4
  $r3 = 6
  $r4 = $r2 + $r3
  assert_non_negative_amount($r4)
  set_current_asset($r1)
  $r5 = 0
  $r6 = int_copy($r4)
  $r7 = lt_int($r5, $r4)
  jmp_if_true($r7, #max_end_0)
  $r6 = int_copy($r5)
#max_end_0
  $r8 = "src"
  $r9 = 0
  $r10 = str_eq($r8, $r0)
  jmp_if_false($r10, #not_world_1)
  $r11 = pull_account(account: $r8, cap: $r6)
  jmp(#pull_end_2)
#not_world_1
  $r11 = pull_account(account: $r8, cap: $r6, overdraft: $r9)
#pull_end_2
  check_enough_funds($r11, $r4)
  $r12 = "dest"
  send_to_account(account: $r12)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 16
  $r3 = 6
  $r4 = $r2 - $r3
  assert_non_negative_amount($r4)
  set_current_asset($r1)
  $r5 = 0
  $r6 = int_copy($r4)
  $r7 = lt_int($r5, $r4)
  jmp_if_true($r7, #max_end_0)
  $r6 = int_copy($r5)
#max_end_0
  $r8 = "src"
  $r9 = 0
  $r10 = str_eq($r8, $r0)
  jmp_if_false($r10, #not_world_1)
  $r11 = pull_account(account: $r8, cap: $r6)
  jmp(#pull_end_2)
#not_world_1
  $r11 = pull_account(account: $r8, cap: $r6, overdraft: $r9)
#pull_end_2
  check_enough_funds($r11, $r4)
  $r12 = "dest"
  send_to_account(account: $r12)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 3
  $r3 = "USD/2"
  $r4 = 7
  assert_same_asset($r1, $r3)
  $r5 = $r2 + $r4
  assert_non_negative_amount($r5)
  set_current_asset($r1)
  $r6 = 0
  $r7 = int_copy($r5)
  $r8 = lt_int($r6, $r5)
  jmp_if_true($r8, #max_end_0)
  $r7 = int_copy($r6)
#max_end_0
  $r9 = "src"
  $r10 = 0
  $r11 = str_eq($r9, $r0)
  jmp_if_false($r11, #not_world_1)
  $r12 = pull_account(account: $r9, cap: $r7)
  jmp(#pull_end_2)
#not_world_1
  $r12 = pull_account(account: $r9, cap: $r7, overdraft: $r10)
#pull_end_2
  check_enough_funds($r12, $r5)
  $r13 = "dest"
  send_to_account(account: $r13)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 30
  $r3 = "USD/2"
  $r4 = 20
  assert_same_asset($r1, $r3)
  $r5 = $r2 - $r4
  assert_non_negative_amount($r5)
  set_current_asset($r1)
  $r6 = 0
  $r7 = int_copy($r5)
  $r8 = lt_int($r6, $r5)
  jmp_if_true($r8, #max_end_0)
  $r7 = int_copy($r6)
#max_end_0
  $r9 = "src"
  $r10 = 0
  $r11 = str_eq($r9, $r0)
  jmp_if_false($r11, #not_world_1)
  $r12 = pull_account(account: $r9, cap: $r7)
  jmp(#pull_end_2)
#not_world_1
  $r12 = pull_account(account: $r9, cap: $r7, overdraft: $r10)
#pull_end_2
  check_enough_funds($r12, $r5)
  $r13 = "dest"
  send_to_account(account: $r13)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 42
  $r3 = "USD/2"
  assert_non_negative_amount($r2)
  set_current_asset($r3)
  $r4 = 0
  $r5 = int_copy($r2)
  $r6 = lt_int($r4, $r2)
  jmp_if_true($r6, #max_end_0)
  $r5 = int_copy($r4)
#max_end_0
  $r7 = "src"
  $r8 = 0
  $r9 = str_eq($r7, $r0)
  jmp_if_false($r9, #not_world_1)
  $r10 = pull_account(account: $r7, cap: $r5)
  jmp(#pull_end_2)
#not_world_1
  $r10 = pull_account(account: $r7, cap: $r5, overdraft: $r8)
#pull_end_2
  check_enough_funds($r10, $r2)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 42
  $r3 = 10
  assert_non_negative_amount($r3)
  set_current_asset($r1)
  $r4 = 0
  $r5 = int_copy($r3)
  $r6 = lt_int($r4, $r3)
  jmp_if_true($r6, #max_end_0)
  $r5 = int_copy($r4)
#max_end_0
  $r7 = "src"
  $r8 = 0
  $r9 = str_eq($r7, $r0)
  jmp_if_false($r9, #not_world_1)
  $r10 = pull_account(account: $r7, cap: $r5)
  jmp(#pull_end_2)
#not_world_1
  $r10 = pull_account(account: $r7, cap: $r5, overdraft: $r8)
#pull_end_2
  check_enough_funds($r10, $r3)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 10
  $r3 = neg_int($r2)
  $r4 = neg_int($r3)
  assert_non_negative_amount($r4)
  set_current_asset($r1)
  $r5 = 0
  $r6 = int_copy($r4)
  $r7 = lt_int($r5, $r4)
  jmp_if_true($r7, #max_end_0)
  $r6 = int_copy($r5)
#max_end_0
  $r8 = "src"
  $r9 = 0
  $r10 = str_eq($r8, $r0)
  jmp_if_false($r10, #not_world_1)
  $r11 = pull_account(account: $r8, cap: $r6)
  jmp(#pull_end_2)
#not_world_1
  $r11 = pull_account(account: $r8, cap: $r6, overdraft: $r9)
#pull_end_2
  check_enough_funds($r11, $r4)
  $r12 = "dest"
  send_to_account(account: $r12)
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
  $r0 = "world"
  $r1 = "src"
  $r2 = "USD/2"
  $r3 = balance($r1, $r2)
  assert_non_negative_balance($r3, $r1)
  assert_non_negative_amount($r3)
  set_current_asset($r2)
  $r4 = 0
  $r5 = int_copy($r3)
  $r6 = lt_int($r4, $r3)
  jmp_if_true($r6, #max_end_0)
  $r5 = int_copy($r4)
#max_end_0
  $r7 = "src"
  $r8 = 0
  $r9 = str_eq($r7, $r0)
  jmp_if_false($r9, #not_world_1)
  $r10 = pull_account(account: $r7, cap: $r5)
  jmp(#pull_end_2)
#not_world_1
  $r10 = pull_account(account: $r7, cap: $r5, overdraft: $r8)
#pull_end_2
  check_enough_funds($r10, $r3)
  $r11 = "dest"
  send_to_account(account: $r11)
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
  $r0 = "world"
  $r1 = "alice"
  $r2 = "USD/2"
  $r3 = 10
  assert_non_negative_amount($r3)
  set_current_asset($r2)
  $r4 = 0
  $r5 = int_copy($r3)
  $r6 = lt_int($r4, $r3)
  jmp_if_true($r6, #max_end_0)
  $r5 = int_copy($r4)
#max_end_0
  $r7 = "world"
  $r8 = 0
  $r9 = str_eq($r7, $r0)
  jmp_if_false($r9, #not_world_1)
  $r10 = pull_account(account: $r7, cap: $r5)
  jmp(#pull_end_2)
#not_world_1
  $r10 = pull_account(account: $r7, cap: $r5, overdraft: $r8)
#pull_end_2
  check_enough_funds($r10, $r3)
  $r11 = "users"
  $r12 = ":"
  $r13 = ":"
  $r14 = "wallet"
  $r15 = add_string($r11, $r12)
  $r16 = add_string($r15, $r1)
  $r17 = add_string($r16, $r13)
  $r18 = add_string($r17, $r14)
  assert_valid_account($r18)
  send_to_account(account: $r18)
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
  $r0 = "world"
  $r1 = 42
  $r2 = "USD/2"
  $r3 = 10
  assert_non_negative_amount($r3)
  set_current_asset($r2)
  $r4 = 0
  $r5 = int_copy($r3)
  $r6 = lt_int($r4, $r3)
  jmp_if_true($r6, #max_end_0)
  $r5 = int_copy($r4)
#max_end_0
  $r7 = "world"
  $r8 = 0
  $r9 = str_eq($r7, $r0)
  jmp_if_false($r9, #not_world_1)
  $r10 = pull_account(account: $r7, cap: $r5)
  jmp(#pull_end_2)
#not_world_1
  $r10 = pull_account(account: $r7, cap: $r5, overdraft: $r8)
#pull_end_2
  check_enough_funds($r10, $r3)
  $r11 = "account"
  $r12 = ":"
  $r13 = int_to_string($r1)
  $r14 = add_string($r11, $r12)
  $r15 = add_string($r14, $r13)
  assert_valid_account($r15)
  send_to_account(account: $r15)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 10
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = lt_int($r3, $r2)
  jmp_if_true($r5, #max_end_0)
  $r4 = int_copy($r3)
#max_end_0
  $r6 = 0
  $r7 = int_copy($r4)
  $r8 = "a"
  $r9 = 0
  $r10 = str_eq($r8, $r0)
  jmp_if_false($r10, #not_world_1)
  $r11 = pull_account(account: $r8, cap: $r7)
  jmp(#pull_end_2)
#not_world_1
  $r11 = pull_account(account: $r8, cap: $r7, overdraft: $r9)
#pull_end_2
  $r6 += $r11
  $r7 -= $r11
  $r12 = "b"
  $r13 = 0
  $r14 = str_eq($r12, $r0)
  jmp_if_false($r14, #not_world_3)
  $r15 = pull_account(account: $r12, cap: $r7)
  jmp(#pull_end_4)
#not_world_3
  $r15 = pull_account(account: $r12, cap: $r7, overdraft: $r13)
#pull_end_4
  $r6 += $r15
  $r7 -= $r15
  $r16 = "c"
  $r17 = 0
  $r18 = str_eq($r16, $r0)
  jmp_if_false($r18, #not_world_5)
  $r19 = pull_account(account: $r16, cap: $r7)
  jmp(#pull_end_6)
#not_world_5
  $r19 = pull_account(account: $r16, cap: $r7, overdraft: $r17)
#pull_end_6
  $r6 += $r19
  check_enough_funds($r6, $r2)
  $r20 = "dest"
  send_to_account(account: $r20)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 10
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = lt_int($r3, $r2)
  jmp_if_true($r5, #max_end_0)
  $r4 = int_copy($r3)
#max_end_0
  $r6 = 0
  $r7 = int_copy($r4)
  $r8 = "a"
  $r9 = 0
  $r10 = str_eq($r8, $r0)
  jmp_if_false($r10, #not_world_1)
  $r11 = pull_account(account: $r8, cap: $r7)
  jmp(#pull_end_2)
#not_world_1
  $r11 = pull_account(account: $r8, cap: $r7, overdraft: $r9)
#pull_end_2
  $r6 += $r11
  $r7 -= $r11
  $r12 = "USD/2"
  $r13 = 5
  assert_same_asset($r12, $r1)
  $r14 = int_copy($r13)
  $r15 = lt_int($r13, $r7)
  jmp_if_true($r15, #min_end_3)
  $r14 = int_copy($r7)
#min_end_3
  $r16 = 0
  $r17 = int_copy($r14)
  $r18 = lt_int($r16, $r14)
  jmp_if_true($r18, #max_end_4)
  $r17 = int_copy($r16)
#max_end_4
  $r19 = "b"
  $r20 = 0
  $r21 = str_eq($r19, $r0)
  jmp_if_false($r21, #not_world_5)
  $r22 = pull_account(account: $r19, cap: $r17)
  jmp(#pull_end_6)
#not_world_5
  $r22 = pull_account(account: $r19, cap: $r17, overdraft: $r20)
#pull_end_6
  $r6 += $r22
  $r7 -= $r22
  $r23 = "c"
  $r24 = 0
  $r25 = str_eq($r23, $r0)
  jmp_if_false($r25, #not_world_7)
  $r26 = pull_account(account: $r23, cap: $r7)
  jmp(#pull_end_8)
#not_world_7
  $r26 = pull_account(account: $r23, cap: $r7, overdraft: $r24)
#pull_end_8
  $r6 += $r26
  check_enough_funds($r6, $r2)
  $r27 = "dest"
  send_to_account(account: $r27)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 10
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = lt_int($r3, $r2)
  jmp_if_true($r5, #max_end_0)
  $r4 = int_copy($r3)
#max_end_0
  $r6 = "world"
  $r7 = 0
  $r8 = str_eq($r6, $r0)
  jmp_if_false($r8, #not_world_1)
  $r9 = pull_account(account: $r6, cap: $r4)
  jmp(#pull_end_2)
#not_world_1
  $r9 = pull_account(account: $r6, cap: $r4, overdraft: $r7)
#pull_end_2
  check_enough_funds($r9, $r2)
  $r10 = int_copy($r9)
  $r11 = is_zero($r10)
  jmp_if_true($r11, #dest_inorder_end_3)
  $r12 = "USD/2"
  $r13 = 4
  assert_same_asset($r12, $r1)
  $r14 = 0
  $r15 = int_copy($r10)
  $r16 = lt_int($r10, $r13)
  jmp_if_true($r16, #min_end_4)
  $r15 = int_copy($r13)
#min_end_4
  $r17 = int_copy($r15)
  $r18 = lt_int($r14, $r15)
  jmp_if_true($r18, #max_end_5)
  $r17 = int_copy($r14)
#max_end_5
  $r19 = is_zero($r17)
  jmp_if_true($r19, #dest_inorder_skip_6)
  $r20 = "d1"
  send_to_account(account: $r20, cap: $r17)
  $r10 -= $r17
#dest_inorder_skip_6
#dest_inorder_end_3
  $r21 = is_zero($r10)
  jmp_if_true($r21, #dest_inorder_rem_skip_7)
  $r22 = "d2"
  send_to_account(account: $r22, cap: $r10)
#dest_inorder_rem_skip_7
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 10
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = lt_int($r3, $r2)
  jmp_if_true($r5, #max_end_0)
  $r4 = int_copy($r3)
#max_end_0
  mark_push()
  $r6 = "a"
  $r7 = 0
  $r8 = str_eq($r6, $r0)
  jmp_if_false($r8, #not_world_2)
  $r9 = pull_account(account: $r6, cap: $r4)
  jmp(#pull_end_3)
#not_world_2
  $r9 = pull_account(account: $r6, cap: $r4, overdraft: $r7)
#pull_end_3
  $r10 = int_copy($r9)
  $r11 = $r4 - $r9
  $r12 = is_zero($r11)
  jmp_if_true($r12, #oneof_end_1)
  mark_rewind()
  mark_push()
  $r13 = "b"
  $r14 = 0
  $r15 = str_eq($r13, $r0)
  jmp_if_false($r15, #not_world_4)
  $r16 = pull_account(account: $r13, cap: $r4)
  jmp(#pull_end_5)
#not_world_4
  $r16 = pull_account(account: $r13, cap: $r4, overdraft: $r14)
#pull_end_5
  $r10 = int_copy($r16)
  $r17 = $r4 - $r16
  $r18 = is_zero($r17)
  jmp_if_true($r18, #oneof_end_1)
  mark_rewind()
  mark_push()
  $r19 = "c"
  $r20 = 0
  $r21 = str_eq($r19, $r0)
  jmp_if_false($r21, #not_world_6)
  $r22 = pull_account(account: $r19, cap: $r4)
  jmp(#pull_end_7)
#not_world_6
  $r22 = pull_account(account: $r19, cap: $r4, overdraft: $r20)
#pull_end_7
  $r10 = int_copy($r22)
#oneof_end_1
  mark_commit()
  check_enough_funds($r10, $r2)
  $r23 = "dest"
  send_to_account(account: $r23)
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 10
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = lt_int($r3, $r2)
  jmp_if_true($r5, #max_end_0)
  $r4 = int_copy($r3)
#max_end_0
  mark_push()
  $r6 = "a"
  $r7 = 0
  $r8 = str_eq($r6, $r0)
  jmp_if_false($r8, #not_world_2)
  $r9 = pull_account(account: $r6, cap: $r4)
  jmp(#pull_end_3)
#not_world_2
  $r9 = pull_account(account: $r6, cap: $r4, overdraft: $r7)
#pull_end_3
  $r10 = int_copy($r9)
  $r11 = $r4 - $r9
  $r12 = is_zero($r11)
  jmp_if_true($r12, #oneof_end_1)
  mark_rewind()
  mark_push()
  $r13 = "b"
  $r14 = 0
  $r15 = str_eq($r13, $r0)
  jmp_if_false($r15, #not_world_4)
  $r16 = pull_account(account: $r13, cap: $r4)
  jmp(#pull_end_5)
#not_world_4
  $r16 = pull_account(account: $r13, cap: $r4, overdraft: $r14)
#pull_end_5
  $r10 = int_copy($r16)
#oneof_end_1
  mark_commit()
  check_enough_funds($r10, $r2)
  $r17 = "dest"
  send_to_account(account: $r17)
`))
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
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 10
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = lt_int($r3, $r2)
  jmp_if_true($r5, #max_end_0)
  $r4 = int_copy($r3)
#max_end_0
  $r6 = "world"
  $r7 = 0
  $r8 = str_eq($r6, $r0)
  jmp_if_false($r8, #not_world_1)
  $r9 = pull_account(account: $r6, cap: $r4)
  jmp(#pull_end_2)
#not_world_1
  $r9 = pull_account(account: $r6, cap: $r4, overdraft: $r7)
#pull_end_2
  check_enough_funds($r9, $r2)
  $r10 = "USD/2"
  $r11 = 4
  assert_same_asset($r10, $r1)
  $r12 = int_copy($r9)
  $r13 = lt_int($r9, $r11)
  jmp_if_true($r13, #min_end_5)
  $r12 = int_copy($r11)
#min_end_5
  $r14 = $r9 - $r12
  $r15 = is_zero($r14)
  jmp_if_true($r15, #oneof_dest_clause_4)
  $r16 = "b"
  send_to_account(account: $r16)
  jmp(#oneof_dest_end_3)
#oneof_dest_clause_4
  $r17 = "a"
  send_to_account(account: $r17)
  jmp(#oneof_dest_end_3)
#oneof_dest_end_3
`))
}

func TestColoredSource(t *testing.T) {
	out := getCompiledOutput(t, `
		#![feature("experimental-asset-colors")]
		send [COIN 10] (
			source = @src \ "RED"
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "world"
  $r1 = "COIN"
  $r2 = 10
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = lt_int($r3, $r2)
  jmp_if_true($r5, #max_end_0)
  $r4 = int_copy($r3)
#max_end_0
  $r6 = "src"
  $r7 = "RED"
  assert_valid_color($r7)
  $r8 = 0
  $r9 = str_eq($r6, $r0)
  jmp_if_false($r9, #not_world_1)
  $r10 = pull_account(account: $r6, cap: $r4, color: $r7)
  jmp(#pull_end_2)
#not_world_1
  $r10 = pull_account(account: $r6, cap: $r4, overdraft: $r8, color: $r7)
#pull_end_2
  check_enough_funds($r10, $r2)
  $r11 = "dest"
  send_to_account(account: $r11)
`))
}

func TestColoredOverdraftSource(t *testing.T) {
	out := getCompiledOutput(t, `
		#![feature("experimental-asset-colors")]
		send [COIN 10] (
			source = @src \ "RED" allowing unbounded overdraft
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "world"
  $r1 = "COIN"
  $r2 = 10
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = lt_int($r3, $r2)
  jmp_if_true($r5, #max_end_0)
  $r4 = int_copy($r3)
#max_end_0
  $r6 = "src"
  $r7 = "RED"
  assert_valid_color($r7)
  $r8 = pull_account(account: $r6, cap: $r4, color: $r7)
  check_enough_funds($r8, $r2)
  $r9 = "dest"
  send_to_account(account: $r9)
`))
}
