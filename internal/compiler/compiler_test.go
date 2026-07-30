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
  set_current_asset($r1)
  $r3 = "src"
  $r4 = 0
  $r5 = str_eq($r3, $r0)
  jmp_if_false($r5, #not_world_0)
  $r6 = pull_account(account: $r3, cap: $r2)
  jmp(#pull_end_1)
#not_world_0
  $r6 = pull_account(account: $r3, cap: $r2, overdraft: $r4)
#pull_end_1
  check_enough_funds($r6, $r2)
  $r7 = "dest"
  send_to_account(account: $r7)
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
  set_current_asset($r1)
  $r5 = "src"
  $r6 = 0
  $r7 = str_eq($r5, $r0)
  jmp_if_false($r7, #not_world_0)
  $r8 = pull_account(account: $r5, cap: $r4)
  jmp(#pull_end_1)
#not_world_0
  $r8 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
#pull_end_1
  check_enough_funds($r8, $r4)
  $r9 = "dest"
  send_to_account(account: $r9)
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
  set_current_asset($r1)
  $r5 = "src"
  $r6 = 0
  $r7 = str_eq($r5, $r0)
  jmp_if_false($r7, #not_world_0)
  $r8 = pull_account(account: $r5, cap: $r4)
  jmp(#pull_end_1)
#not_world_0
  $r8 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
#pull_end_1
  check_enough_funds($r8, $r4)
  $r9 = "dest"
  send_to_account(account: $r9)
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
  set_current_asset($r1)
  $r6 = "src"
  $r7 = 0
  $r8 = str_eq($r6, $r0)
  jmp_if_false($r8, #not_world_0)
  $r9 = pull_account(account: $r6, cap: $r5)
  jmp(#pull_end_1)
#not_world_0
  $r9 = pull_account(account: $r6, cap: $r5, overdraft: $r7)
#pull_end_1
  check_enough_funds($r9, $r5)
  $r10 = "dest"
  send_to_account(account: $r10)
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
  set_current_asset($r1)
  $r6 = "src"
  $r7 = 0
  $r8 = str_eq($r6, $r0)
  jmp_if_false($r8, #not_world_0)
  $r9 = pull_account(account: $r6, cap: $r5)
  jmp(#pull_end_1)
#not_world_0
  $r9 = pull_account(account: $r6, cap: $r5, overdraft: $r7)
#pull_end_1
  check_enough_funds($r9, $r5)
  $r10 = "dest"
  send_to_account(account: $r10)
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
  set_current_asset($r3)
  $r4 = "src"
  $r5 = 0
  $r6 = str_eq($r4, $r0)
  jmp_if_false($r6, #not_world_0)
  $r7 = pull_account(account: $r4, cap: $r2)
  jmp(#pull_end_1)
#not_world_0
  $r7 = pull_account(account: $r4, cap: $r2, overdraft: $r5)
#pull_end_1
  check_enough_funds($r7, $r2)
  $r8 = "dest"
  send_to_account(account: $r8)
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
  set_current_asset($r1)
  $r4 = "src"
  $r5 = 0
  $r6 = str_eq($r4, $r0)
  jmp_if_false($r6, #not_world_0)
  $r7 = pull_account(account: $r4, cap: $r3)
  jmp(#pull_end_1)
#not_world_0
  $r7 = pull_account(account: $r4, cap: $r3, overdraft: $r5)
#pull_end_1
  check_enough_funds($r7, $r3)
  $r8 = "dest"
  send_to_account(account: $r8)
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
  set_current_asset($r1)
  $r5 = "src"
  $r6 = 0
  $r7 = str_eq($r5, $r0)
  jmp_if_false($r7, #not_world_0)
  $r8 = pull_account(account: $r5, cap: $r4)
  jmp(#pull_end_1)
#not_world_0
  $r8 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
#pull_end_1
  check_enough_funds($r8, $r4)
  $r9 = "dest"
  send_to_account(account: $r9)
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
  set_current_asset($r2)
  $r4 = "src"
  $r5 = 0
  $r6 = str_eq($r4, $r0)
  jmp_if_false($r6, #not_world_0)
  $r7 = pull_account(account: $r4, cap: $r3)
  jmp(#pull_end_1)
#not_world_0
  $r7 = pull_account(account: $r4, cap: $r3, overdraft: $r5)
#pull_end_1
  check_enough_funds($r7, $r3)
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
  $r0 = "world"
  $r1 = "alice"
  $r2 = "USD/2"
  $r3 = 10
  set_current_asset($r2)
  $r4 = "world"
  $r5 = 0
  $r6 = str_eq($r4, $r0)
  jmp_if_false($r6, #not_world_0)
  $r7 = pull_account(account: $r4, cap: $r3)
  jmp(#pull_end_1)
#not_world_0
  $r7 = pull_account(account: $r4, cap: $r3, overdraft: $r5)
#pull_end_1
  check_enough_funds($r7, $r3)
  $r8 = "users"
  $r9 = ":"
  $r10 = ":"
  $r11 = "wallet"
  $r12 = add_string($r8, $r9)
  $r13 = add_string($r12, $r1)
  $r14 = add_string($r13, $r10)
  $r15 = add_string($r14, $r11)
  assert_valid_account($r15)
  send_to_account(account: $r15)
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
  set_current_asset($r2)
  $r4 = "world"
  $r5 = 0
  $r6 = str_eq($r4, $r0)
  jmp_if_false($r6, #not_world_0)
  $r7 = pull_account(account: $r4, cap: $r3)
  jmp(#pull_end_1)
#not_world_0
  $r7 = pull_account(account: $r4, cap: $r3, overdraft: $r5)
#pull_end_1
  check_enough_funds($r7, $r3)
  $r8 = "account"
  $r9 = ":"
  $r10 = int_to_string($r1)
  $r11 = add_string($r8, $r9)
  $r12 = add_string($r11, $r10)
  assert_valid_account($r12)
  send_to_account(account: $r12)
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
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = "a"
  $r6 = 0
  $r7 = str_eq($r5, $r0)
  jmp_if_false($r7, #not_world_1)
  $r8 = pull_account(account: $r5, cap: $r4)
  jmp(#pull_end_2)
#not_world_1
  $r8 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
#pull_end_2
  $r3 += $r8
  $r4 -= $r8
  $r9 = is_zero($r4)
  jmp_if_true($r9, #inorder_end_0)
  $r10 = "b"
  $r11 = 0
  $r12 = str_eq($r10, $r0)
  jmp_if_false($r12, #not_world_3)
  $r13 = pull_account(account: $r10, cap: $r4)
  jmp(#pull_end_4)
#not_world_3
  $r13 = pull_account(account: $r10, cap: $r4, overdraft: $r11)
#pull_end_4
  $r3 += $r13
  $r4 -= $r13
  $r14 = is_zero($r4)
  jmp_if_true($r14, #inorder_end_0)
  $r15 = "c"
  $r16 = 0
  $r17 = str_eq($r15, $r0)
  jmp_if_false($r17, #not_world_5)
  $r18 = pull_account(account: $r15, cap: $r4)
  jmp(#pull_end_6)
#not_world_5
  $r18 = pull_account(account: $r15, cap: $r4, overdraft: $r16)
#pull_end_6
  $r3 += $r18
#inorder_end_0
  check_enough_funds($r3, $r2)
  $r19 = "dest"
  send_to_account(account: $r19)
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
  set_current_asset($r1)
  $r3 = 0
  $r4 = int_copy($r2)
  $r5 = "a"
  $r6 = 0
  $r7 = str_eq($r5, $r0)
  jmp_if_false($r7, #not_world_1)
  $r8 = pull_account(account: $r5, cap: $r4)
  jmp(#pull_end_2)
#not_world_1
  $r8 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
#pull_end_2
  $r3 += $r8
  $r4 -= $r8
  $r9 = is_zero($r4)
  jmp_if_true($r9, #inorder_end_0)
  $r10 = "USD/2"
  $r11 = 5
  assert_same_asset($r10, $r1)
  $r12 = int_copy($r11)
  $r13 = lt_int($r11, $r4)
  jmp_if_true($r13, #min_end_3)
  $r12 = int_copy($r4)
#min_end_3
  $r14 = "b"
  $r15 = 0
  $r16 = str_eq($r14, $r0)
  jmp_if_false($r16, #not_world_4)
  $r17 = pull_account(account: $r14, cap: $r12)
  jmp(#pull_end_5)
#not_world_4
  $r17 = pull_account(account: $r14, cap: $r12, overdraft: $r15)
#pull_end_5
  $r3 += $r17
  $r4 -= $r17
  $r18 = is_zero($r4)
  jmp_if_true($r18, #inorder_end_0)
  $r19 = "c"
  $r20 = 0
  $r21 = str_eq($r19, $r0)
  jmp_if_false($r21, #not_world_6)
  $r22 = pull_account(account: $r19, cap: $r4)
  jmp(#pull_end_7)
#not_world_6
  $r22 = pull_account(account: $r19, cap: $r4, overdraft: $r20)
#pull_end_7
  $r3 += $r22
#inorder_end_0
  check_enough_funds($r3, $r2)
  $r23 = "dest"
  send_to_account(account: $r23)
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
  set_current_asset($r1)
  $r3 = "world"
  $r4 = 0
  $r5 = str_eq($r3, $r0)
  jmp_if_false($r5, #not_world_0)
  $r6 = pull_account(account: $r3, cap: $r2)
  jmp(#pull_end_1)
#not_world_0
  $r6 = pull_account(account: $r3, cap: $r2, overdraft: $r4)
#pull_end_1
  check_enough_funds($r6, $r2)
  $r7 = int_copy($r6)
  $r8 = "USD/2"
  $r9 = 4
  assert_same_asset($r8, $r1)
  $r10 = int_copy($r7)
  $r11 = lt_int($r7, $r9)
  jmp_if_true($r11, #min_end_2)
  $r10 = int_copy($r9)
#min_end_2
  $r12 = "d1"
  send_to_account(account: $r12, cap: $r10)
  $r7 -= $r10
  $r13 = "d2"
  send_to_account(account: $r13, cap: $r7)
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
  set_current_asset($r1)
  $r3 = snapshot()
  $r4 = "a"
  $r5 = 0
  $r6 = str_eq($r4, $r0)
  jmp_if_false($r6, #not_world_1)
  $r7 = pull_account(account: $r4, cap: $r2)
  jmp(#pull_end_2)
#not_world_1
  $r7 = pull_account(account: $r4, cap: $r2, overdraft: $r5)
#pull_end_2
  $r8 = int_copy($r7)
  $r9 = $r2 - $r7
  $r10 = is_zero($r9)
  jmp_if_true($r10, #oneof_end_0)
  restore($r3)
  $r11 = "b"
  $r12 = 0
  $r13 = str_eq($r11, $r0)
  jmp_if_false($r13, #not_world_3)
  $r14 = pull_account(account: $r11, cap: $r2)
  jmp(#pull_end_4)
#not_world_3
  $r14 = pull_account(account: $r11, cap: $r2, overdraft: $r12)
#pull_end_4
  $r8 = int_copy($r14)
  $r15 = $r2 - $r14
  $r16 = is_zero($r15)
  jmp_if_true($r16, #oneof_end_0)
  restore($r3)
  $r17 = "c"
  $r18 = 0
  $r19 = str_eq($r17, $r0)
  jmp_if_false($r19, #not_world_5)
  $r20 = pull_account(account: $r17, cap: $r2)
  jmp(#pull_end_6)
#not_world_5
  $r20 = pull_account(account: $r17, cap: $r2, overdraft: $r18)
#pull_end_6
  $r8 = int_copy($r20)
#oneof_end_0
  check_enough_funds($r8, $r2)
  $r21 = "dest"
  send_to_account(account: $r21)
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
  set_current_asset($r1)
  $r3 = snapshot()
  $r4 = "a"
  $r5 = 0
  $r6 = str_eq($r4, $r0)
  jmp_if_false($r6, #not_world_1)
  $r7 = pull_account(account: $r4, cap: $r2)
  jmp(#pull_end_2)
#not_world_1
  $r7 = pull_account(account: $r4, cap: $r2, overdraft: $r5)
#pull_end_2
  $r8 = int_copy($r7)
  $r9 = $r2 - $r7
  $r10 = is_zero($r9)
  jmp_if_true($r10, #oneof_end_0)
  restore($r3)
  $r11 = "b"
  $r12 = 0
  $r13 = str_eq($r11, $r0)
  jmp_if_false($r13, #not_world_3)
  $r14 = pull_account(account: $r11, cap: $r2)
  jmp(#pull_end_4)
#not_world_3
  $r14 = pull_account(account: $r11, cap: $r2, overdraft: $r12)
#pull_end_4
  $r8 = int_copy($r14)
#oneof_end_0
  check_enough_funds($r8, $r2)
  $r15 = "dest"
  send_to_account(account: $r15)
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
  set_current_asset($r1)
  $r3 = "world"
  $r4 = 0
  $r5 = str_eq($r3, $r0)
  jmp_if_false($r5, #not_world_0)
  $r6 = pull_account(account: $r3, cap: $r2)
  jmp(#pull_end_1)
#not_world_0
  $r6 = pull_account(account: $r3, cap: $r2, overdraft: $r4)
#pull_end_1
  check_enough_funds($r6, $r2)
  $r7 = "USD/2"
  $r8 = 4
  assert_same_asset($r7, $r1)
  $r9 = int_copy($r6)
  $r10 = lt_int($r6, $r8)
  jmp_if_true($r10, #min_end_4)
  $r9 = int_copy($r8)
#min_end_4
  $r11 = $r6 - $r9
  $r12 = is_zero($r11)
  jmp_if_true($r12, #oneof_dest_clause_3)
  $r13 = "b"
  send_to_account(account: $r13)
  jmp(#oneof_dest_end_2)
#oneof_dest_clause_3
  $r14 = "a"
  send_to_account(account: $r14)
  jmp(#oneof_dest_end_2)
#oneof_dest_end_2
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
  set_current_asset($r1)
  $r3 = "src"
  $r4 = "RED"
  assert_valid_color($r4)
  $r5 = 0
  $r6 = str_eq($r3, $r0)
  jmp_if_false($r6, #not_world_0)
  $r7 = pull_account(account: $r3, cap: $r2, color: $r4)
  jmp(#pull_end_1)
#not_world_0
  $r7 = pull_account(account: $r3, cap: $r2, overdraft: $r5, color: $r4)
#pull_end_1
  check_enough_funds($r7, $r2)
  $r8 = "dest"
  send_to_account(account: $r8)
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
  set_current_asset($r1)
  $r3 = "src"
  $r4 = "RED"
  assert_valid_color($r4)
  $r5 = pull_account(account: $r3, cap: $r2, color: $r4)
  check_enough_funds($r5, $r2)
  $r6 = "dest"
  send_to_account(account: $r6)
`))
}
