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
  set_current_asset($r0)
  $r2 = "src"
  $r3 = 0
  $r4 = pull_account(account: $r2, cap: $r1, overdraft: $r3)
  check_enough_funds($r4, $r1)
  $r5 = "dest"
  send_to_account(account: $r5)
`),
	)
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
  set_current_asset($r0)
  $r4 = "src"
  $r5 = 0
  $r6 = pull_account(account: $r4, cap: $r3, overdraft: $r5)
  check_enough_funds($r6, $r3)
  $r7 = "dest"
  send_to_account(account: $r7)
`),
	)
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
  set_current_asset($r0)
  $r4 = "src"
  $r5 = 0
  $r6 = pull_account(account: $r4, cap: $r3, overdraft: $r5)
  check_enough_funds($r6, $r3)
  $r7 = "dest"
  send_to_account(account: $r7)
`),
	)
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
  $r2 = "USD/2"
  $r3 = 7
  assert_same_asset($r0, $r2)
  $r4 = $r1 + $r3
  set_current_asset($r0)
  $r5 = "src"
  $r6 = 0
  $r7 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
  check_enough_funds($r7, $r4)
  $r8 = "dest"
  send_to_account(account: $r8)
`),
	)
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
  $r2 = "USD/2"
  $r3 = 20
  assert_same_asset($r0, $r2)
  $r4 = $r1 - $r3
  set_current_asset($r0)
  $r5 = "src"
  $r6 = 0
  $r7 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
  check_enough_funds($r7, $r4)
  $r8 = "dest"
  send_to_account(account: $r8)
`),
	)
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
  $r2 = "USD/2"
  set_current_asset($r2)
  $r3 = "src"
  $r4 = 0
  $r5 = pull_account(account: $r3, cap: $r1, overdraft: $r4)
  check_enough_funds($r5, $r1)
  $r6 = "dest"
  send_to_account(account: $r6)
`),
	)
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
  $r2 = 10
  set_current_asset($r0)
  $r3 = "src"
  $r4 = 0
  $r5 = pull_account(account: $r3, cap: $r2, overdraft: $r4)
  check_enough_funds($r5, $r2)
  $r6 = "dest"
  send_to_account(account: $r6)
`),
	)
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
  $r3 = neg_int($r2)
  set_current_asset($r0)
  $r4 = "src"
  $r5 = 0
  $r6 = pull_account(account: $r4, cap: $r3, overdraft: $r5)
  check_enough_funds($r6, $r3)
  $r7 = "dest"
  send_to_account(account: $r7)
`),
	)
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
  set_current_asset($r1)
  $r3 = "src"
  $r4 = 0
  $r5 = pull_account(account: $r3, cap: $r2, overdraft: $r4)
  check_enough_funds($r5, $r2)
  $r6 = "dest"
  send_to_account(account: $r6)
`),
	)
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
  set_current_asset($r1)
  $r3 = "world"
  $r4 = 0
  $r5 = pull_account(account: $r3, cap: $r2, overdraft: $r4)
  check_enough_funds($r5, $r2)
  $r6 = "users"
  $r7 = ":"
  $r8 = ":"
  $r9 = "wallet"
  $r10 = add_string($r6, $r7)
  $r11 = add_string($r10, $r0)
  $r12 = add_string($r11, $r8)
  $r13 = add_string($r12, $r9)
  assert_valid_account($r13)
  send_to_account(account: $r13)
`),
	)
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
  set_current_asset($r1)
  $r3 = "world"
  $r4 = 0
  $r5 = pull_account(account: $r3, cap: $r2, overdraft: $r4)
  check_enough_funds($r5, $r2)
  $r6 = "account"
  $r7 = ":"
  $r8 = int_to_string($r0)
  $r9 = add_string($r6, $r7)
  $r10 = add_string($r9, $r8)
  assert_valid_account($r10)
  send_to_account(account: $r10)
`),
	)
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
  set_current_asset($r0)
  $r2 = 0
  $r3 = int_copy($r1)
  $r4 = "a"
  $r5 = 0
  $r6 = pull_account(account: $r4, cap: $r3, overdraft: $r5)
  $r2 += $r6
  $r3 -= $r6
  jmp_if_zero($r3, #inorder_end_0)
  $r7 = "b"
  $r8 = 0
  $r9 = pull_account(account: $r7, cap: $r3, overdraft: $r8)
  $r2 += $r9
  $r3 -= $r9
  jmp_if_zero($r3, #inorder_end_0)
  $r10 = "c"
  $r11 = 0
  $r12 = pull_account(account: $r10, cap: $r3, overdraft: $r11)
  $r2 += $r12
#inorder_end_0
  check_enough_funds($r2, $r1)
  $r13 = "dest"
  send_to_account(account: $r13)
`),
	)
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
  set_current_asset($r0)
  $r2 = 0
  $r3 = int_copy($r1)
  $r4 = "a"
  $r5 = 0
  $r6 = pull_account(account: $r4, cap: $r3, overdraft: $r5)
  $r2 += $r6
  $r3 -= $r6
  jmp_if_zero($r3, #inorder_end_0)
  $r7 = "USD/2"
  $r8 = 5
  assert_same_asset($r7, $r0)
  $r9 = min_int($r8, $r3)
  $r10 = "b"
  $r11 = 0
  $r12 = pull_account(account: $r10, cap: $r9, overdraft: $r11)
  $r2 += $r12
  $r3 -= $r12
  jmp_if_zero($r3, #inorder_end_0)
  $r13 = "c"
  $r14 = 0
  $r15 = pull_account(account: $r13, cap: $r3, overdraft: $r14)
  $r2 += $r15
#inorder_end_0
  check_enough_funds($r2, $r1)
  $r16 = "dest"
  send_to_account(account: $r16)
`),
	)
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
  set_current_asset($r0)
  $r2 = "world"
  $r3 = 0
  $r4 = pull_account(account: $r2, cap: $r1, overdraft: $r3)
  check_enough_funds($r4, $r1)
  $r5 = int_copy($r4)
  $r6 = "USD/2"
  $r7 = 4
  assert_same_asset($r6, $r0)
  $r8 = min_int($r5, $r7)
  $r9 = "d1"
  send_to_account(account: $r9, cap: $r8)
  $r5 -= $r8
  $r10 = "d2"
  send_to_account(account: $r10, cap: $r5)
`),
	)
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
  set_current_asset($r0)
  $r2 = snapshot()
  $r3 = "a"
  $r4 = 0
  $r5 = pull_account(account: $r3, cap: $r1, overdraft: $r4)
  $r6 = int_copy($r5)
  $r7 = $r1 - $r5
  jmp_if_zero($r7, #oneof_end_0)
  restore($r2)
  $r8 = "b"
  $r9 = 0
  $r10 = pull_account(account: $r8, cap: $r1, overdraft: $r9)
  $r6 = int_copy($r10)
  $r11 = $r1 - $r10
  jmp_if_zero($r11, #oneof_end_0)
  restore($r2)
  $r12 = "c"
  $r13 = 0
  $r14 = pull_account(account: $r12, cap: $r1, overdraft: $r13)
  $r6 = int_copy($r14)
#oneof_end_0
  check_enough_funds($r6, $r1)
  $r15 = "dest"
  send_to_account(account: $r15)
`),
	)
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
  set_current_asset($r0)
  $r2 = snapshot()
  $r3 = "a"
  $r4 = 0
  $r5 = pull_account(account: $r3, cap: $r1, overdraft: $r4)
  $r6 = int_copy($r5)
  $r7 = $r1 - $r5
  jmp_if_zero($r7, #oneof_end_0)
  restore($r2)
  $r8 = "b"
  $r9 = 0
  $r10 = pull_account(account: $r8, cap: $r1, overdraft: $r9)
  $r6 = int_copy($r10)
#oneof_end_0
  check_enough_funds($r6, $r1)
  $r11 = "dest"
  send_to_account(account: $r11)
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
  set_current_asset($r0)
  $r2 = "world"
  $r3 = 0
  $r4 = pull_account(account: $r2, cap: $r1, overdraft: $r3)
  check_enough_funds($r4, $r1)
  $r5 = 0
  $r6 = "USD/2"
  $r7 = 4
  assert_same_asset($r6, $r0)
  $r8 = min_int($r4, $r7)
  $r9 = $r4 - $r8
  jmp_if_zero($r9, #oneof_dest_clause_1)
  $r10 = "b"
  send_to_account(account: $r10)
  jmp_if_zero($r5, #oneof_dest_end_0)
#oneof_dest_clause_1
  $r11 = "a"
  send_to_account(account: $r11)
  jmp_if_zero($r5, #oneof_dest_end_0)
#oneof_dest_end_0
`),
	)
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
  $r0 = "COIN"
  $r1 = 10
  set_current_asset($r0)
  $r2 = "src"
  $r3 = "RED"
  assert_valid_color($r3)
  $r4 = 0
  $r5 = pull_account(account: $r2, cap: $r1, overdraft: $r4, color: $r3)
  check_enough_funds($r5, $r1)
  $r6 = "dest"
  send_to_account(account: $r6)
`),
	)
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
  $r0 = "COIN"
  $r1 = 10
  set_current_asset($r0)
  $r2 = "src"
  $r3 = "RED"
  assert_valid_color($r3)
  $r4 = pull_account(account: $r2, cap: $r1, color: $r3)
  check_enough_funds($r4, $r1)
  $r5 = "dest"
  send_to_account(account: $r5)
`),
	)
}
