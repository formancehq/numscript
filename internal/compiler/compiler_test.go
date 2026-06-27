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
  $r8 <- load_const("dest")
  send_to_account($r8)
`))
}
