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

func TestOptimizedWorldSend(t *testing.T) {
	out := getOptimizedOutput(t, `
		send [USD/2 42] (
			source = @world
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  set_current_asset($r0)
  $r1 = "world"
  $r2 = 42
  $r3 = "dest"
  post_account_no_credit(source: $r1, destination: $r3, amount: $r2)
`))
}

// A bounded source keeps its pull and send: only the branches around the
// literal account and the literal amount are gone.
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
  $r5 = take_account(account: $r3, cap: $r2, overdraft: $r4)
  check_enough_funds($r5, $r1)
  $r6 = "dest"
  post_account_no_credit(source: $r3, destination: $r6, amount: $r5)
`))
}

func TestOptimizedUnboundedOverdraftSend(t *testing.T) {
	out := getOptimizedOutput(t, `
		send [USD/2 42] (
			source = @a allowing unbounded overdraft
			destination = @dest
		)
	`)

	snaps.MatchInlineSnapshot(t, out, snaps.Inline(`
  $r0 = "USD/2"
  set_current_asset($r0)
  $r1 = "a"
  $r2 = 42
  $r3 = "dest"
  post_account_no_credit(source: $r1, destination: $r3, amount: $r2)
`))
}

// Which statements go around the funds queue: post_account against the
// pull_account and send_to_account left in the optimized stream.
func TestPostFromUnboundedShapes(t *testing.T) {
	testCases := []struct {
		name                       string
		src                        string
		posts, takes, pulls, sends int
	}{
		{
			name:  "world to a literal destination",
			src:   `send [USD/2 42] (source = @world destination = @dest)`,
			posts: 1,
		},
		{
			name: "world to a destination from a var",
			src: `vars { account $dest }
send [USD/2 42] (source = @world destination = $dest)`,
			posts: 1,
		},
		{
			name: "world to an interpolated destination",
			src: `#![feature("experimental-account-interpolation")]
vars { string $id }
send [USD/2 42] (source = @world destination = @users:$id)`,
			posts: 1,
		},
		{
			name:  "unbounded overdraft",
			src:   `send [USD/2 42] (source = @a allowing unbounded overdraft destination = @dest)`,
			posts: 1,
		},
		{
			name:  "capped world: the check stays",
			src:   `send [USD/2 42] (source = max [USD/2 5] from @world destination = @dest)`,
			posts: 1,
		},
		{
			name: "amount from a var",
			src: `vars { monetary $amt }
send $amt (source = @world destination = @dest)`,
			posts: 1,
		},
		{
			name: "two world sends",
			src: `send [USD/2 42] (source = @world destination = @a)
send [USD/2 42] (source = @world destination = @b)`,
			posts: 2,
		},
		{
			// an unbounded pull never reads the balance, so the second statement
			// cannot see the first one's dropped debit
			name: "two unbounded-overdraft sends from one account",
			src: `send [USD/2 42] (source = @a allowing unbounded overdraft destination = @x)
send [USD/2 42] (source = @a allowing unbounded overdraft destination = @y)`,
			posts: 2,
		},
		{
			name: "world, with a balance read of another account",
			src: `vars { monetary $amt = balance(@other, USD/2) }
send $amt (source = @world destination = @dest)`,
			posts: 1,
		},
		{
			name: "world, with a bounded send from another literal account",
			src: `send [USD/2 42] (source = @world destination = @a)
send [USD/2 10] (source = @a destination = @b)`,
			posts: 2,
			takes: 1,
			pulls: 0,
			sends: 0,
		},
		{
			name: "colored world source",
			src: `#![feature("experimental-asset-colors")]
send [USD/2 42] (source = @world \ "RED" destination = @dest)`,
			posts: 1,
		},
		{
			name:  "bounded source",
			src:   `send [USD/2 42] (source = @src destination = @dest)`,
			posts: 1,
			takes: 1,
			pulls: 0,
			sends: 0,
		},
		{
			name:  "bounded overdraft",
			src:   `send [USD/2 42] (source = @a allowing overdraft up to [USD/2 100] destination = @dest)`,
			posts: 1,
			takes: 1,
			pulls: 0,
			sends: 0,
		},
		{
			// the @world branch is decided at run time, so both arms stay
			name: "source from a var",
			src: `vars { account $src }
send [USD/2 42] (source = $src destination = @dest)`,
			pulls: 2,
			sends: 1,
		},
		{
			name: "world, with a balance read of world",
			src: `vars { monetary $amt = balance(@world, USD/2) }
send $amt (source = @world destination = @dest)`,
			posts: 1,
			takes: 1,
			pulls: 0,
			sends: 0,
		},
		{
			name: "world, with a balance read of an account from a var",
			src: `vars {
  account $acc
  monetary $amt = balance($acc, USD/2)
}
send $amt (source = @world destination = @dest)`,
			posts: 1,
			takes: 1,
			pulls: 0,
			sends: 0,
		},
		{
			name: "world, with a bounded send from an account from a var",
			src: `vars { account $src }
send [USD/2 42] (source = @world destination = @a)
send [USD/2 10] (source = $src destination = @b)`,
			posts: 1,
			takes: 1,
			pulls: 2,
			sends: 1,
		},
		{
			name: "unbounded overdraft, then a bounded send from the same account",
			src: `send [USD/2 42] (source = @a allowing unbounded overdraft destination = @x)
send [USD/2 10] (source = @a destination = @y)`,
			posts: 2,
			takes: 2,
			pulls: 0,
			sends: 0,
		},
		{
			name: "unbounded overdraft, with a save of the same account",
			src: `save [USD/2 1] from @a
send [USD/2 42] (source = @a allowing unbounded overdraft destination = @x)`,
			posts: 1,
			takes: 1,
			pulls: 0,
			sends: 0,
		},
		{
			name: "fan-out",
			src: `send [USD/2 42] (
  source = @world
  destination = { 1/2 to @a remaining to @b }
)`,
			pulls: 1,
			sends: 2,
		},
		{
			name: "fan-in",
			src: `send [USD/2 42] (
  source = { @world @a }
  destination = @dest
)`,
			pulls: 2,
			sends: 1,
		},
		{
			name: "inorder destination",
			src: `send [USD/2 42] (
  source = @world
  destination = { max [USD/2 10] to @a remaining to @b }
)`,
			pulls: 1,
			sends: 2,
		},
		{
			name: "oneof source",
			src: `#![feature("experimental-oneof")]
send [USD/2 42] (
  source = oneof { @a @world }
  destination = @dest
)`,
			pulls: 2,
			sends: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := getOptimizedOutput(t, tc.src)
			require.Equal(t, tc.posts, strings.Count(out, "post_account(")+strings.Count(out, "post_account_no_credit("), "post_account count in:\n%s", out)
			require.Equal(t, tc.takes, strings.Count(out, "take_account("), "take_account count in:\n%s", out)
			require.Equal(t, tc.pulls, strings.Count(out, "pull_account("), "pull_account count in:\n%s", out)
			require.Equal(t, tc.sends, strings.Count(out, "send_to_account("), "send_to_account count in:\n%s", out)
		})
	}
}

// The unbounded arm the compiler emits for a literal @world is reached by a
// branch on str_eq; folding and pruning it must leave a single pull, which is
// what the fusion needs. Checked on the plain dump too, so a change in how the
// compiler spells the branch shows up here rather than as a silently unfused
// program.
func TestOptimizeCollapsesLiteralWorldBranch(t *testing.T) {
	plain := getCompiledOutput(t, `send [USD/2 42] (source = @world destination = @dest)`)
	require.Equal(t, 2, strings.Count(plain, "pull_account("), "the plain program has both arms")
	require.Contains(t, plain, "str_eq(")

	out := getOptimizedOutput(t, `send [USD/2 42] (source = @world destination = @dest)`)
	require.NotContains(t, out, "str_eq(")
	require.NotContains(t, out, "jmp")
	require.NotContains(t, out, "#")
}
