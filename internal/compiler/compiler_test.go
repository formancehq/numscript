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
  send_to_account($r8)
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
  send_to_account($r16)
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
  $r13 <- get_amount($r12)
  $r14 <- min_int($r13, $r6)
  $r15 <- load_const("b")
  $r16 <- load_const(0)
  $r17 <- pull_account(account: $r15, cap: $r14, overdraft: $r16)
  $r5 <- add_int($r5, $r17)
  $r6 <- sub_int($r6, $r17)
  jmp_if_zero($r6, #inorder_end_0)
  $r18 <- load_const("c")
  $r19 <- load_const(0)
  $r20 <- pull_account(account: $r18, cap: $r6, overdraft: $r19)
  $r5 <- add_int($r5, $r20)
#inorder_end_0
  check_enough_funds($r5, $r4)
  $r21 <- load_const("dest")
  send_to_account($r21)
`))
}
