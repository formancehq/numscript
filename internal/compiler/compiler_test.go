package compiler

import (
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func getCompiledOutput(t *testing.T, source string) string {
	program := parser.Parse(source)
	require.Empty(t, program.Errors)
	compiled, err := compileProgramToVirtual(program.Value)
	require.Nil(t, err)

	out := dump(compiled.instructions)
	return "\n" + out
}

// getOptimizedOutput compiles source and runs the full peephole pipeline
// (defaultPeepholes, to a fixpoint) before dumping — the same optimization
// CompileWithOptimizations applies before assembling.
func getOptimizedOutput(t *testing.T, source string) string {
	program := parser.Parse(source)
	require.Empty(t, program.Errors)
	compiled, err := compileProgramToVirtual(program.Value)
	require.Nil(t, err)

	out := dump(optimize(compiled.instructions, defaultPeepholes()))
	return "\n" + out
}

func TestSimpleProgram(t *testing.T) {
	out := getCompiledOutput(t, `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 <- load_const("USD/2")
  $r1 <- load_const(10)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- get_asset($r2)
  set_current_asset($r3)
  $r4 <- get_amount($r2)
  $r5 <- load_const("src")
  $r6 <- pull_account(account: $r5, cap: $r4, overdraft: 0)
  check_enough_funds($r6, $r4)
  $r7 <- load_const("dest")
  send_to_account(account: $r7)
`),
	)
}

// TestOptimizedSimpleProgram is the same script as TestSimpleProgram, but with
// all peepholes applied. Contrast the two snapshots:
//   - monetaryFold + deadCode strip the mk_monetary / get_asset / get_amount
//     round-trip (the asset/amount registers are read directly);
//   - fundsBypass fuses the single-source/single-destination pull_account +
//     send_to_account into take_account (debit, at the source site) +
//     post_account (posting, at the destination site), skipping the funds queue.
func TestOptimizedSimpleProgram(t *testing.T) {
	out := getOptimizedOutput(t, `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 <- load_const("USD/2")
  $r1 <- load_const(10)
  set_current_asset($r0)
  $r5 <- load_const("src")
  $r6 <- take_account(account: $r5, cap: $r1, overdraft: 0)
  check_enough_funds($r6, $r1)
  $r7 <- load_const("dest")
  post_account(src: $r5, dst: $r7, amount: $r6)
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(4)
  $r2 <- load_const(6)
  $r3 <- add_int($r1, $r2)
  $r4 <- mk_monetary($r0, $r3)
  $r5 <- get_asset($r4)
  set_current_asset($r5)
  $r6 <- get_amount($r4)
  $r7 <- load_const("src")
  $r8 <- pull_account(account: $r7, cap: $r6, overdraft: 0)
  check_enough_funds($r8, $r6)
  $r9 <- load_const("dest")
  send_to_account(account: $r9)
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(16)
  $r2 <- load_const(6)
  $r3 <- sub_int($r1, $r2)
  $r4 <- mk_monetary($r0, $r3)
  $r5 <- get_asset($r4)
  set_current_asset($r5)
  $r6 <- get_amount($r4)
  $r7 <- load_const("src")
  $r8 <- pull_account(account: $r7, cap: $r6, overdraft: 0)
  check_enough_funds($r8, $r6)
  $r9 <- load_const("dest")
  send_to_account(account: $r9)
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(3)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- load_const("USD/2")
  $r4 <- load_const(7)
  $r5 <- mk_monetary($r3, $r4)
  $r6 <- get_asset($r2)
  $r7 <- get_asset($r5)
  assert_same_asset($r6, $r7)
  $r8 <- get_amount($r2)
  $r9 <- get_amount($r5)
  $r10 <- add_int($r8, $r9)
  $r11 <- mk_monetary($r6, $r10)
  $r12 <- get_asset($r11)
  set_current_asset($r12)
  $r13 <- get_amount($r11)
  $r14 <- load_const("src")
  $r15 <- pull_account(account: $r14, cap: $r13, overdraft: 0)
  check_enough_funds($r15, $r13)
  $r16 <- load_const("dest")
  send_to_account(account: $r16)
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(30)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- load_const("USD/2")
  $r4 <- load_const(20)
  $r5 <- mk_monetary($r3, $r4)
  $r6 <- get_asset($r2)
  $r7 <- get_asset($r5)
  assert_same_asset($r6, $r7)
  $r8 <- get_amount($r2)
  $r9 <- get_amount($r5)
  $r10 <- sub_int($r8, $r9)
  $r11 <- mk_monetary($r6, $r10)
  $r12 <- get_asset($r11)
  set_current_asset($r12)
  $r13 <- get_amount($r11)
  $r14 <- load_const("src")
  $r15 <- pull_account(account: $r14, cap: $r13, overdraft: 0)
  check_enough_funds($r15, $r13)
  $r16 <- load_const("dest")
  send_to_account(account: $r16)
`),
	)
}

func TestGetAmount(t *testing.T) {
	out := getCompiledOutput(t, `
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(42)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- get_amount($r2)
  $r4 <- load_const("USD/2")
  $r5 <- mk_monetary($r4, $r3)
  $r6 <- get_asset($r5)
  set_current_asset($r6)
  $r7 <- get_amount($r5)
  $r8 <- load_const("src")
  $r9 <- pull_account(account: $r8, cap: $r7, overdraft: 0)
  check_enough_funds($r9, $r7)
  $r10 <- load_const("dest")
  send_to_account(account: $r10)
`),
	)
}

func TestGetAsset(t *testing.T) {
	out := getCompiledOutput(t, `
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(42)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- get_asset($r2)
  $r4 <- load_const(10)
  $r5 <- mk_monetary($r3, $r4)
  $r6 <- get_asset($r5)
  set_current_asset($r6)
  $r7 <- get_amount($r5)
  $r8 <- load_const("src")
  $r9 <- pull_account(account: $r8, cap: $r7, overdraft: 0)
  check_enough_funds($r9, $r7)
  $r10 <- load_const("dest")
  send_to_account(account: $r10)
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(10)
  $r2 <- neg_int($r1)
  $r3 <- mk_monetary($r0, $r2)
  $r4 <- get_amount($r3)
  $r5 <- neg_int($r4)
  $r6 <- get_asset($r3)
  $r7 <- mk_monetary($r6, $r5)
  $r8 <- get_asset($r7)
  set_current_asset($r8)
  $r9 <- get_amount($r7)
  $r10 <- load_const("src")
  $r11 <- pull_account(account: $r10, cap: $r9, overdraft: 0)
  check_enough_funds($r11, $r9)
  $r12 <- load_const("dest")
  send_to_account(account: $r12)
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
  $r0 <- load_const("src")
  $r1 <- load_const("USD/2")
  $r2 <- balance($r0, $r1)
  assert_non_negative_balance($r2, $r0)
  $r3 <- get_asset($r2)
  set_current_asset($r3)
  $r4 <- get_amount($r2)
  $r5 <- load_const("src")
  $r6 <- pull_account(account: $r5, cap: $r4, overdraft: 0)
  check_enough_funds($r6, $r4)
  $r7 <- load_const("dest")
  send_to_account(account: $r7)
`),
	)
}

func TestAccountInterpolation(t *testing.T) {
	out := getCompiledOutput(t, `
		vars {
			string $id = "alice"
		}
		send [USD/2 10] (
			source = @world
			destination = @users:$id:wallet
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 <- load_const("alice")
  $r1 <- load_const("USD/2")
  $r2 <- load_const(10)
  $r3 <- mk_monetary($r1, $r2)
  $r4 <- get_asset($r3)
  set_current_asset($r4)
  $r5 <- get_amount($r3)
  $r6 <- load_const("world")
  $r7 <- pull_account(account: $r6, cap: $r5, overdraft: 0)
  check_enough_funds($r7, $r5)
  $r8 <- load_const("users")
  $r9 <- load_const(":")
  $r10 <- load_const(":")
  $r11 <- load_const("wallet")
  $r12 <- add_string($r8, $r9)
  $r13 <- add_string($r12, $r0)
  $r14 <- add_string($r13, $r10)
  $r15 <- add_string($r14, $r11)
  assert_valid_account($r15)
  send_to_account(account: $r15)
`),
	)
}

func TestAccountInterpolationInt(t *testing.T) {
	out := getCompiledOutput(t, `
		vars {
			number $n = 42
		}
		send [USD/2 10] (
			source = @world
			destination = @account:$n
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 <- load_const(42)
  $r1 <- load_const("USD/2")
  $r2 <- load_const(10)
  $r3 <- mk_monetary($r1, $r2)
  $r4 <- get_asset($r3)
  set_current_asset($r4)
  $r5 <- get_amount($r3)
  $r6 <- load_const("world")
  $r7 <- pull_account(account: $r6, cap: $r5, overdraft: 0)
  check_enough_funds($r7, $r5)
  $r8 <- load_const("account")
  $r9 <- load_const(":")
  $r10 <- int_to_string($r0)
  $r11 <- add_string($r8, $r9)
  $r12 <- add_string($r11, $r10)
  assert_valid_account($r12)
  send_to_account(account: $r12)
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(10)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- get_asset($r2)
  set_current_asset($r3)
  $r4 <- get_amount($r2)
  $r5 <- load_const(0)
  $r6 <- int_copy($r4)
  $r7 <- load_const("a")
  $r8 <- pull_account(account: $r7, cap: $r6, overdraft: 0)
  $r5 <- add_int($r5, $r8)
  $r6 <- sub_int($r6, $r8)
  jmp_if_zero($r6, #inorder_end_0)
  $r9 <- load_const("b")
  $r10 <- pull_account(account: $r9, cap: $r6, overdraft: 0)
  $r5 <- add_int($r5, $r10)
  $r6 <- sub_int($r6, $r10)
  jmp_if_zero($r6, #inorder_end_0)
  $r11 <- load_const("c")
  $r12 <- pull_account(account: $r11, cap: $r6, overdraft: 0)
  $r5 <- add_int($r5, $r12)
#inorder_end_0
  check_enough_funds($r5, $r4)
  $r13 <- load_const("dest")
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(10)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- get_asset($r2)
  set_current_asset($r3)
  $r4 <- get_amount($r2)
  $r5 <- load_const(0)
  $r6 <- int_copy($r4)
  $r7 <- load_const("a")
  $r8 <- pull_account(account: $r7, cap: $r6, overdraft: 0)
  $r5 <- add_int($r5, $r8)
  $r6 <- sub_int($r6, $r8)
  jmp_if_zero($r6, #inorder_end_0)
  $r9 <- load_const("USD/2")
  $r10 <- load_const(5)
  $r11 <- mk_monetary($r9, $r10)
  $r12 <- get_asset($r11)
  assert_same_asset($r12, $r3)
  $r13 <- get_amount($r11)
  $r14 <- min_int($r13, $r6)
  $r15 <- load_const("b")
  $r16 <- pull_account(account: $r15, cap: $r14, overdraft: 0)
  $r5 <- add_int($r5, $r16)
  $r6 <- sub_int($r6, $r16)
  jmp_if_zero($r6, #inorder_end_0)
  $r17 <- load_const("c")
  $r18 <- pull_account(account: $r17, cap: $r6, overdraft: 0)
  $r5 <- add_int($r5, $r18)
#inorder_end_0
  check_enough_funds($r5, $r4)
  $r19 <- load_const("dest")
  send_to_account(account: $r19)
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(10)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- get_asset($r2)
  set_current_asset($r3)
  $r4 <- get_amount($r2)
  $r5 <- load_const("world")
  $r6 <- pull_account(account: $r5, cap: $r4, overdraft: 0)
  check_enough_funds($r6, $r4)
  $r7 <- int_copy($r6)
  $r8 <- load_const("USD/2")
  $r9 <- load_const(4)
  $r10 <- mk_monetary($r8, $r9)
  $r11 <- get_asset($r10)
  assert_same_asset($r11, $r3)
  $r12 <- get_amount($r10)
  $r13 <- min_int($r7, $r12)
  $r14 <- load_const("d1")
  send_to_account(account: $r14, cap: $r13)
  $r7 <- sub_int($r7, $r13)
  $r15 <- load_const("d2")
  send_to_account(account: $r15, cap: $r7)
`),
	)
}
