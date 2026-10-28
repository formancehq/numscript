package compiler

import (
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

// getOptimizedOutput is getCompiledOutput after ir.Optimize. The dump is
// canonicalized through a parse: a pass that deletes an instruction leaves a
// gap in the register numbering, and the parser renumbers by first appearance.
func getOptimizedOutput(t *testing.T, source string) string {
	t.Helper()
	program := parser.Parse(source)
	require.Empty(t, program.Errors)
	compiled, err := compileProgramToIR(program.Value, nil)
	require.Nil(t, err)
	require.NoError(t, ir.Typecheck(compiled.instructions))

	optimized := ir.Optimize(compiled.instructions)
	require.NoError(t, ir.Typecheck(optimized))

	instrs, errs := ir.Parse(ir.Dump(optimized))
	require.Empty(t, errs, "the optimized dump does not parse back")
	return "\n" + ir.Dump(instrs)
}

// A literal @world keeps only the unbounded arm of its branch.
func TestOptimizedWorldSend(t *testing.T) {
	out := getOptimizedOutput(t, `
		send [USD/2 42] (
			source = @world
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = 42
  $r3 = "world"
  $r4 = pull_account(account: $r3, cap: $r2)
  check_enough_funds($r4, $r1)
  $r5 = "dest"
  send_to_account(account: $r5)
`))
}

// Any other literal account keeps only the bounded one.
func TestOptimizedBoundedSend(t *testing.T) {
	out := getOptimizedOutput(t, `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  $r1 = 10
  set_current_asset($r0)
  $r2 = 10
  $r3 = "src"
  $r4 = 0
  $r5 = pull_account(account: $r3, cap: $r2, overdraft: $r4)
  check_enough_funds($r5, $r1)
  $r6 = "dest"
  send_to_account(account: $r6)
`))
}

// An account from a var is decided at run time: both arms stay.
func TestOptimizedSendFromVarKeepsBothArms(t *testing.T) {
	out := getOptimizedOutput(t, `
		vars { account $src }
		send [USD/2 10] (
			source = $src
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "world"
  $r1 = load_var<str>(0)
  $r2 = "USD/2"
  $r3 = 10
  set_current_asset($r2)
  $r4 = 10
  $r5 = 0
  $r6 = str_eq($r1, $r0)
  jmp_if_false($r6, #not_world_1)
  $r7 = pull_account(account: $r1, cap: $r4)
  jmp(#pull_end_2)
#not_world_1
  $r7 = pull_account(account: $r1, cap: $r4, overdraft: $r5)
#pull_end_2
  check_enough_funds($r7, $r3)
  $r8 = "dest"
  send_to_account(account: $r8)
`))
}

// The unbounded arm the compiler emits for a literal @world is reached by a
// branch on str_eq; folding and pruning it must leave a single pull. Checked on
// the plain dump too, so a change in how the compiler spells the branch shows
// up here rather than as a silently unoptimized program.
func TestOptimizeCollapsesLiteralWorldBranch(t *testing.T) {
	src := `send [USD/2 42] (source = @world destination = @dest)`

	plain := getCompiledOutput(t, src)
	require.Equal(t, 2, strings.Count(plain, "pull_account("), "the plain program has both arms")
	require.Contains(t, plain, "str_eq(")

	out := getOptimizedOutput(t, src)
	require.Equal(t, 1, strings.Count(out, "pull_account("))
	require.NotContains(t, out, "overdraft:", "the bounded arm is the dead one")
	require.NotContains(t, out, "str_eq(")
	require.NotContains(t, out, "jmp")
	require.NotContains(t, out, "#")
}
