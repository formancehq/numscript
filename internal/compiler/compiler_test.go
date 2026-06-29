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
  $r6 <- pull_account(account: $r5, cap: $r4, overdraft: 0)
  check_enough_funds($r6, $r4)
  $r7 <- load_const("dest")
  send_to_account($r7)
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
  send_to_account($r13)
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
  $r12 <- get_amount($r11)
  $r13 <- min_int($r12, $r6)
  $r14 <- load_const("b")
  $r15 <- pull_account(account: $r14, cap: $r13, overdraft: 0)
  $r5 <- add_int($r5, $r15)
  $r6 <- sub_int($r6, $r15)
  jmp_if_zero($r6, #inorder_end_0)
  $r16 <- load_const("c")
  $r17 <- pull_account(account: $r16, cap: $r6, overdraft: 0)
  $r5 <- add_int($r5, $r17)
#inorder_end_0
  check_enough_funds($r5, $r4)
  $r18 <- load_const("dest")
  send_to_account($r18)
`),
	)
}
