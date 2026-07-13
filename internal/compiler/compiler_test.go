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
  $r6 <- load_const(0)
  $r7 <- pull_account(account: $r5, cap: $r4, overdraft: $r6)
  check_enough_funds($r7, $r4)
  $r8 <- load_const("dest")
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(4)
  $r2 <- load_const(6)
  $r3 <- add_int($r1, $r2)
  $r4 <- mk_monetary($r0, $r3)
  $r5 <- get_asset($r4)
  set_current_asset($r5)
  $r6 <- get_amount($r4)
  $r7 <- load_const("src")
  $r8 <- load_const(0)
  $r9 <- pull_account(account: $r7, cap: $r6, overdraft: $r8)
  check_enough_funds($r9, $r6)
  $r10 <- load_const("dest")
  send_to_account(account: $r10)
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(10)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- get_asset($r2)
  set_current_asset($r3)
  $r4 <- get_amount($r2)
  $r5 <- load_const(0)
  $r6 <- int_copy($r4)
  $r7 <- load_const("a")
  $r8 <- load_const(0)
  $r9 <- pull_account(account: $r7, cap: $r6, overdraft: $r8)
  $r5 <- add_int($r5, $r9)
  $r6 <- sub_int($r6, $r9)
  jmp_if_zero($r6, #inorder_end_0)
  $r10 <- load_const("b")
  $r11 <- load_const(0)
  $r12 <- pull_account(account: $r10, cap: $r6, overdraft: $r11)
  $r5 <- add_int($r5, $r12)
  $r6 <- sub_int($r6, $r12)
  jmp_if_zero($r6, #inorder_end_0)
  $r13 <- load_const("c")
  $r14 <- load_const(0)
  $r15 <- pull_account(account: $r13, cap: $r6, overdraft: $r14)
  $r5 <- add_int($r5, $r15)
#inorder_end_0
  check_enough_funds($r5, $r4)
  $r16 <- load_const("dest")
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(10)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- get_asset($r2)
  set_current_asset($r3)
  $r4 <- get_amount($r2)
  $r5 <- load_const(0)
  $r6 <- int_copy($r4)
  $r7 <- load_const("a")
  $r8 <- load_const(0)
  $r9 <- pull_account(account: $r7, cap: $r6, overdraft: $r8)
  $r5 <- add_int($r5, $r9)
  $r6 <- sub_int($r6, $r9)
  jmp_if_zero($r6, #inorder_end_0)
  $r10 <- load_const("USD/2")
  $r11 <- load_const(5)
  $r12 <- mk_monetary($r10, $r11)
  $r13 <- get_asset($r12)
  check_eq_current_asset($r13)
  $r14 <- get_amount($r12)
  $r15 <- min_int($r14, $r6)
  $r16 <- load_const("b")
  $r17 <- load_const(0)
  $r18 <- pull_account(account: $r16, cap: $r15, overdraft: $r17)
  $r5 <- add_int($r5, $r18)
  $r6 <- sub_int($r6, $r18)
  jmp_if_zero($r6, #inorder_end_0)
  $r19 <- load_const("c")
  $r20 <- load_const(0)
  $r21 <- pull_account(account: $r19, cap: $r6, overdraft: $r20)
  $r5 <- add_int($r5, $r21)
#inorder_end_0
  check_enough_funds($r5, $r4)
  $r22 <- load_const("dest")
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
  $r0 <- load_const("USD/2")
  $r1 <- load_const(10)
  $r2 <- mk_monetary($r0, $r1)
  $r3 <- get_asset($r2)
  set_current_asset($r3)
  $r4 <- get_amount($r2)
  $r5 <- load_const("world")
  $r6 <- load_const(0)
  $r7 <- pull_account(account: $r5, cap: $r4, overdraft: $r6)
  check_enough_funds($r7, $r4)
  $r8 <- int_copy($r7)
  $r9 <- load_const("USD/2")
  $r10 <- load_const(4)
  $r11 <- mk_monetary($r9, $r10)
  $r12 <- get_asset($r11)
  check_eq_current_asset($r12)
  $r13 <- get_amount($r11)
  $r14 <- min_int($r8, $r13)
  $r15 <- load_const("d1")
  send_to_account(account: $r15, cap: $r14)
  $r8 <- sub_int($r8, $r14)
  $r16 <- load_const("d2")
  send_to_account(account: $r16, cap: $r8)
`))
}
