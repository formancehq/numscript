package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// runPass parses src, applies one pass and returns its dump, or src's own
// canonical dump and false when the pass changed nothing. Registers keep the
// numbers the parse gave them, so an instruction a pass deletes leaves a gap.
func runPass(t *testing.T, pass Peephole, src string) (string, bool) {
	t.Helper()
	instrs, errs := Parse(src)
	require.Empty(t, errs, "IR errors: %v", errs)
	require.NoError(t, Typecheck(instrs))

	out := pass(instrs)
	if out == nil {
		return Dump(instrs), false
	}
	require.NoError(t, Typecheck(out), "the pass broke register typing")
	return Dump(out), true
}

func requirePass(t *testing.T, pass Peephole, src, want string) {
	t.Helper()
	got, changed := runPass(t, pass, src)
	require.True(t, changed, "the pass changed nothing:\n%s", got)
	require.Equal(t, want, got)
}

func requireUnchanged(t *testing.T, pass Peephole, src string) {
	t.Helper()
	got, changed := runPass(t, pass, src)
	require.False(t, changed, "the pass changed the program:\n%s", got)
}

// --- constProp ---

func TestConstProp(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "str_eq of equal literals",
			src: `
  $r0 = "world"
  $r1 = "world"
  $r2 = str_eq($r0, $r1)
`,
			want: `  $r0 = "world"
  $r1 = "world"
  $r2 = true
`,
		},
		{
			name: "str_eq of different literals",
			src: `
  $r0 = "src"
  $r1 = "world"
  $r2 = str_eq($r0, $r1)
`,
			want: `  $r0 = "src"
  $r1 = "world"
  $r2 = false
`,
		},
		{
			name: "eq_int and lt_int",
			src: `
  $r0 = 0
  $r1 = 42
  $r2 = eq_int($r0, $r1)
  $r3 = lt_int($r0, $r1)
  $r4 = lt_int($r1, $r0)
  $r5 = eq_int($r1, $r1)
`,
			want: `  $r0 = 0
  $r1 = 42
  $r2 = false
  $r3 = true
  $r4 = false
  $r5 = true
`,
		},
		{
			name: "is_zero",
			src: `
  $r0 = 0
  $r1 = -3
  $r2 = is_zero($r0)
  $r3 = is_zero($r1)
`,
			want: `  $r0 = 0
  $r1 = -3
  $r2 = true
  $r3 = false
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requirePass(t, constProp, tc.src, tc.want)
		})
	}
}

func TestConstPropLeavesNonConstants(t *testing.T) {
	testCases := []struct {
		name string
		src  string
	}{
		{
			name: "operand from a var",
			src: `
  $r0 = load_var<str>(0)
  $r1 = "world"
  $r2 = str_eq($r0, $r1)
`,
		},
		{
			// written twice, so its value depends on the path taken
			name: "operand written on both arms of a branch",
			src: `
  $r0 = load_var<str>(0)
  $r1 = 0
  $r3 = "world"
  $r4 = str_eq($r0, $r3)
  jmp_if_true($r4, #end)
  $r1 = 1
#end
  $r2 = is_zero($r1)
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requireUnchanged(t, constProp, tc.src)
		})
	}
}

func TestConstPropFoldsThroughOps(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "arithmetic and copies",
			src: `
  $r0 = 1
  $r1 = 2
  $r2 = $r0 + $r1
  $r3 = int_copy($r2)
  $r4 = is_zero($r3)
`,
			want: `  $r0 = 1
  $r1 = 2
  $r2 = 3
  $r3 = 3
  $r4 = false
`,
		},
		{
			// the check fails unless got == needed, so got is needed after it
			name: "check_enough_funds fixes got",
			src: `
  $r0 = "USD/2"
  set_current_asset($r0)
  $r1 = 100
  $r2 = "src"
  $r3 = 0
  $r4 = pull_account(account: $r2, cap: $r1, overdraft: $r3)
  check_enough_funds($r4, $r1)
  $r5 = 1
  $r6 = 2
  $r7 = mk_portion($r5, $r6)
  $r8 = int_to_portion($r4)
  $r9 = mul_portion($r7, $r8)
  $r10 = portion_to_int($r9)
`,
			want: `  $r0 = "USD/2"
  set_current_asset($r0)
  $r1 = 100
  $r2 = "src"
  $r3 = 0
  $r4 = pull_account(account: $r2, cap: $r1, overdraft: $r3)
  check_enough_funds($r4, $r1)
  $r5 = 1
  $r6 = 2
  $r7 = mk_portion($r5, $r6)
  $r8 = int_to_portion($r4)
  $r9 = mul_portion($r7, $r8)
  $r10 = 50
`,
		},
		{
			name: "both arms agree",
			src: `
  $r0 = load_var<str>(0)
  $r1 = 7
  $r2 = "world"
  $r3 = str_eq($r0, $r2)
  jmp_if_true($r3, #end)
  $r1 = 7
#end
  $r4 = is_zero($r1)
`,
			want: `  $r0 = load_var<str>(0)
  $r1 = 7
  $r2 = "world"
  $r3 = str_eq($r0, $r2)
  jmp_if_true($r3, #end)
  $r1 = 7
#end
  $r4 = false
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requirePass(t, constProp, tc.src, tc.want)
		})
	}
}

func TestConstPropLeavesDivisionByZero(t *testing.T) {
	requireUnchanged(t, constProp, `
  $r0 = 1
  $r1 = 0
  $r2 = mk_portion($r0, $r1)
  $r3 = portion_to_int($r2)
`)
}

// --- dropProvenAsserts ---

func TestDropProvenAsserts(t *testing.T) {
	requirePass(t, dropProvenAsserts, `
  $r0 = 42
  assert_non_negative_amount($r0)
  $r1 = "USD/2"
  $r2 = "USD/2"
  assert_same_asset($r1, $r2)
  $r3 = 1
  $r4 = 2
  $r5 = mk_portion($r3, $r4)
  assert_leftover($r5)
`, `  $r0 = 42
  $r1 = "USD/2"
  $r2 = "USD/2"
  $r3 = 1
  $r4 = 2
  $r5 = mk_portion($r3, $r4)
`)
}

func TestDropProvenAssertsKeepsFailingAndUnknown(t *testing.T) {
	testCases := []struct {
		name string
		src  string
	}{
		{
			name: "fails",
			src: `
  $r0 = -1
  assert_non_negative_amount($r0)
`,
		},
		{
			name: "unknown",
			src: `
  $r0 = load_var<int>(0)
  assert_non_negative_amount($r0)
`,
		},
		{
			name: "mismatched assets",
			src: `
  $r0 = "USD/2"
  $r1 = "EUR/2"
  assert_same_asset($r0, $r1)
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requireUnchanged(t, dropProvenAsserts, tc.src)
		})
	}
}

// --- deadCode ---

func TestDeadCode(t *testing.T) {
	requirePass(t, deadCode, `
  $r0 = "USD/2"
  set_current_asset($r0)
  $r1 = "unused"
  $r2 = 1
  $r3 = 2
  $r4 = mk_portion($r2, $r3)
  $r5 = load_var<str>(0)
  $r6 = 10
  $r7 = str_eq($r5, $r0)
  jmp_if_true($r7, #end)
  $r6 = 20
#end
  $r8 = int_copy($r6)
  $r9 = 0
  $r10 = pull_account(account: $r5, cap: $r6, overdraft: $r9)
`, `  $r0 = "USD/2"
  set_current_asset($r0)
  $r5 = load_var<str>(0)
  $r6 = 10
  $r7 = str_eq($r5, $r0)
  jmp_if_true($r7, #end)
  $r6 = 20
#end
  $r9 = 0
  $r10 = pull_account(account: $r5, cap: $r6, overdraft: $r9)
`)
}

func TestDeadCodeKeepsEffects(t *testing.T) {
	requireUnchanged(t, deadCode, `
  $r0 = "USD/2"
  set_current_asset($r0)
  $r1 = load_var<int>(0)
  $r2 = load_var<int>(1)
  $r3 = mk_portion($r1, $r2)
  $r4 = "src"
  $r5 = balance($r4, $r0)
  $r6 = 0
  $r7 = pull_account(account: $r4, cap: $r1, overdraft: $r6)
`)
}

// --- pruneDeadBranches ---

func TestPruneDeadBranches(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "jmp_if_false on true falls through, the label goes",
			src: `
  $r0 = true
  jmp_if_false($r0, #skip)
  $r1 = 1
#skip
  $r2 = 2
`,
			want: `  $r0 = true
  $r1 = 1
  $r2 = 2
`,
		},
		{
			name: "jmp_if_false on false skips the arm",
			src: `
  $r0 = false
  jmp_if_false($r0, #skip)
  $r1 = 1
#skip
  $r2 = 2
`,
			want: `  $r0 = false
  $r2 = 2
`,
		},
		{
			name: "jmp_if_true on true skips the arm",
			src: `
  $r0 = true
  jmp_if_true($r0, #skip)
  $r1 = 1
#skip
  $r2 = 2
`,
			want: `  $r0 = true
  $r2 = 2
`,
		},
		{
			name: "jmp_if_true on false falls through",
			src: `
  $r0 = false
  jmp_if_true($r0, #skip)
  $r1 = 1
#skip
  $r2 = 2
`,
			want: `  $r0 = false
  $r1 = 1
  $r2 = 2
`,
		},
		{
			name: "the world diamond collapses to its unbounded arm",
			src: `
  $r0 = "USD/2"
  set_current_asset($r0)
  $r1 = "world"
  $r2 = 10
  $r3 = 0
  $r4 = true
  jmp_if_false($r4, #not_world)
  $r5 = pull_account(account: $r1, cap: $r2)
  jmp(#pull_end)
#not_world
  $r5 = pull_account(account: $r1, cap: $r2, overdraft: $r3)
#pull_end
  check_enough_funds($r5, $r2)
`,
			want: `  $r0 = "USD/2"
  set_current_asset($r0)
  $r1 = "world"
  $r2 = 10
  $r3 = 0
  $r4 = true
  $r5 = pull_account(account: $r1, cap: $r2)
  check_enough_funds($r5, $r2)
`,
		},
		{
			// the jmp(#b) is itself unreachable, so #b stays unreferenced and goes
			// with it, while #a's code is reachable through the jump to it
			name: "code after an unconditional jmp goes, up to a live label",
			src: `
  $r0 = true
  jmp_if_true($r0, #a)
  $r1 = 1
  jmp(#b)
#a
  $r2 = 2
#b
  $r3 = 3
`,
			want: `  $r0 = true
  $r2 = 2
  $r3 = 3
`,
		},
		{
			name: "a jump to the next instruction goes",
			src: `
  $r0 = load_var<int>(0)
  $r1 = is_zero($r0)
  jmp_if_true($r1, #next)
#next
  $r2 = 1
`,
			want: `  $r0 = load_var<int>(0)
  $r1 = is_zero($r0)
  $r2 = 1
`,
		},
		{
			name: "a label with no jump goes",
			src: `
  $r0 = 1
#orphan
  $r1 = 2
`,
			want: `  $r0 = 1
  $r1 = 2
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requirePass(t, pruneDeadBranches, tc.src, tc.want)
		})
	}
}

func TestPruneDeadBranchesLeavesLiveControlFlow(t *testing.T) {
	testCases := []struct {
		name string
		src  string
	}{
		{
			name: "jump on a run-time condition",
			src: `
  $r0 = load_var<int>(0)
  $r1 = is_zero($r0)
  jmp_if_true($r1, #end)
  $r2 = 1
#end
  $r3 = 2
`,
		},
		{
			name: "both arms reachable",
			src: `
  $r0 = load_var<int>(0)
  $r1 = is_zero($r0)
  jmp_if_false($r1, #else)
  $r2 = 1
  jmp(#end)
#else
  $r2 = 2
#end
  $r3 = int_copy($r2)
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requireUnchanged(t, pruneDeadBranches, tc.src)
		})
	}
}

// --- postFromUnbounded ---

func TestPostFromUnbounded(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "single unbounded source to a single destination",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  $r4 = "dest"
  send_to_account(account: $r4)
`,
			want: `  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = max_zero($r1)
  check_enough_funds($r3, $r1)
  $r4 = "dest"
  post_account(source: $r2, destination: $r4, amount: $r3)
`,
		},
		{
			name: "color and scopes are carried over",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = "RED"
  $r4 = "s1"
  $r5 = pull_account(account: $r2, cap: $r1, color: $r3, scope: $r4)
  check_enough_funds($r5, $r1)
  $r6 = "dest"
  $r7 = "s2"
  send_to_account(account: $r6, scope: $r7)
`,
			want: `  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = "RED"
  $r4 = "s1"
  $r5 = max_zero($r1)
  check_enough_funds($r5, $r1)
  $r6 = "dest"
  $r7 = "s2"
  post_account(source: $r2, destination: $r6, amount: $r5, color: $r3, source_scope: $r4, destination_scope: $r7)
`,
		},
		{
			// an unbounded pull never reads the balance, so two statements from
			// the same unbounded source both go
			name: "two statements from the same unbounded source",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "a"
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  $r4 = "x"
  send_to_account(account: $r4)
  set_current_asset($r0)
  $r5 = "a"
  $r6 = pull_account(account: $r5, cap: $r1)
  check_enough_funds($r6, $r1)
  $r7 = "y"
  send_to_account(account: $r7)
`,
			want: `  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "a"
  $r3 = max_zero($r1)
  check_enough_funds($r3, $r1)
  $r4 = "x"
  post_account(source: $r2, destination: $r4, amount: $r3)
  set_current_asset($r0)
  $r5 = "a"
  $r6 = max_zero($r1)
  check_enough_funds($r6, $r1)
  $r7 = "y"
  post_account(source: $r5, destination: $r7, amount: $r6)
`,
		},
		{
			name: "a balance read of another account does not block",
			src: `
  $r0 = "USD/2"
  $r1 = "other"
  $r2 = balance($r1, $r0)
  set_current_asset($r0)
  $r3 = "world"
  $r4 = pull_account(account: $r3, cap: $r2)
  check_enough_funds($r4, $r2)
  $r5 = "dest"
  send_to_account(account: $r5)
`,
			want: `  $r0 = "USD/2"
  $r1 = "other"
  $r2 = balance($r1, $r0)
  set_current_asset($r0)
  $r3 = "world"
  $r4 = max_zero($r2)
  check_enough_funds($r4, $r2)
  $r5 = "dest"
  post_account(source: $r3, destination: $r5, amount: $r4)
`,
		},
		{
			// the assert stays where it was, after the check
			name: "a check between the pull and the send stays in place",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  $r4 = "users:"
  $r5 = load_var<str>(0)
  $r6 = add_string($r4, $r5)
  assert_valid_account($r6)
  send_to_account(account: $r6)
`,
			want: `  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = max_zero($r1)
  check_enough_funds($r3, $r1)
  $r4 = "users:"
  $r5 = load_var<str>(0)
  $r6 = add_string($r4, $r5)
  assert_valid_account($r6)
  post_account(source: $r2, destination: $r6, amount: $r3)
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requirePass(t, postFromUnbounded, tc.src, tc.want)
		})
	}
}

func TestPostFromUnboundedLeavesOtherShapes(t *testing.T) {
	testCases := []struct {
		name string
		src  string
	}{
		{
			name: "bounded source",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "src"
  $r3 = 0
  $r4 = pull_account(account: $r2, cap: $r1, overdraft: $r3)
  check_enough_funds($r4, $r1)
  $r5 = "dest"
  send_to_account(account: $r5)
`,
		},
		{
			name: "uncapped source",
			src: `
  $r0 = "USD/2"
  set_current_asset($r0)
  $r1 = "src"
  $r2 = 0
  $r3 = pull_account(account: $r1, overdraft: $r2)
  $r4 = "dest"
  send_to_account(account: $r4)
`,
		},
		{
			name: "capped send",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  $r4 = "dest"
  send_to_account(account: $r4, cap: $r1)
`,
		},
		{
			name: "kept",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  send_to_account()
`,
		},
		{
			name: "two sources",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  $r4 = "a"
  $r5 = pull_account(account: $r4, cap: $r1)
  $r6 = "dest"
  send_to_account(account: $r6)
`,
		},
		{
			name: "two destinations",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  $r4 = "a"
  send_to_account(account: $r4, cap: $r1)
  $r5 = "b"
  send_to_account(account: $r5)
`,
		},
		{
			name: "source account from a var",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = load_var<str>(0)
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  $r4 = "dest"
  send_to_account(account: $r4)
`,
		},
		{
			name: "balance read of the source",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  $r4 = "dest"
  send_to_account(account: $r4)
  $r5 = "world"
  $r6 = balance($r5, $r0)
`,
		},
		{
			name: "balance read of an account from a var",
			src: `
  $r0 = "USD/2"
  $r1 = load_var<str>(0)
  $r2 = balance($r1, $r0)
  set_current_asset($r0)
  $r3 = "world"
  $r4 = pull_account(account: $r3, cap: $r2)
  check_enough_funds($r4, $r2)
  $r5 = "dest"
  send_to_account(account: $r5)
`,
		},
		{
			name: "bounded pull of the source elsewhere",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "a"
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  $r4 = "x"
  send_to_account(account: $r4)
  set_current_asset($r0)
  $r5 = "a"
  $r6 = 0
  $r7 = pull_account(account: $r5, cap: $r1, overdraft: $r6)
  check_enough_funds($r7, $r1)
  $r8 = "y"
  send_to_account(account: $r8)
`,
		},
		{
			name: "bounded pull of an account from a var elsewhere",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  $r4 = "x"
  send_to_account(account: $r4)
  set_current_asset($r0)
  $r5 = load_var<str>(0)
  $r6 = 0
  $r7 = pull_account(account: $r5, cap: $r1, overdraft: $r6)
  check_enough_funds($r7, $r1)
  $r8 = "y"
  send_to_account(account: $r8)
`,
		},
		{
			name: "save of the source",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  $r2 = "world"
  save(account: $r2, asset: $r0, amount: $r1)
  set_current_asset($r0)
  $r3 = "world"
  $r4 = pull_account(account: $r3, cap: $r1)
  check_enough_funds($r4, $r1)
  $r5 = "dest"
  send_to_account(account: $r5)
`,
		},
		{
			name: "pulled amount rewritten before the send",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  check_enough_funds($r3, $r1)
  $r3 = 0
  $r4 = "dest"
  send_to_account(account: $r4)
`,
		},
		{
			name: "label between the pull and the send",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  $r4 = is_zero($r3)
  jmp_if_true($r4, #skip)
  check_enough_funds($r3, $r1)
#skip
  $r5 = "dest"
  send_to_account(account: $r5)
`,
		},
		{
			name: "mark in the statement",
			src: `
  $r0 = "USD/2"
  $r1 = 42
  set_current_asset($r0)
  mark_push()
  $r2 = "world"
  $r3 = pull_account(account: $r2, cap: $r1)
  mark_commit()
  check_enough_funds($r3, $r1)
  $r4 = "dest"
  send_to_account(account: $r4)
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requireUnchanged(t, postFromUnbounded, tc.src)
		})
	}
}

// --- Optimize ---

// The compiler's output for `send [USD/2 42] (source = @world destination =
// @dest)`, minus the amount clamp: the three passes have to compose for the
// literal @world to end up as one direct posting.
func TestOptimizeCollapsesWorldDiamondIntoPostAccount(t *testing.T) {
	src := `
  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 42
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = "world"
  $r4 = 0
  $r5 = str_eq($r3, $r0)
  jmp_if_false($r5, #not_world)
  $r6 = pull_account(account: $r3, cap: $r2)
  jmp(#pull_end)
#not_world
  $r6 = pull_account(account: $r3, cap: $r2, overdraft: $r4)
#pull_end
  check_enough_funds($r6, $r2)
  $r7 = "dest"
  send_to_account(account: $r7)
`
	instrs, errs := Parse(src)
	require.Empty(t, errs)
	require.NoError(t, Typecheck(instrs))

	out := Optimize(instrs)
	require.NoError(t, Typecheck(out))
	require.Equal(t, `  $r1 = "USD/2"
  set_current_asset($r1)
  $r3 = "world"
  $r6 = 42
  $r7 = "dest"
  post_account(source: $r3, destination: $r7, amount: $r6)
`, Dump(out))
}

func TestOptimizeReturnsInputWhenNothingApplies(t *testing.T) {
	src := `
  $r0 = "USD/2"
  $r1 = load_var<int>(0)
  set_current_asset($r0)
  $r2 = load_var<str>(0)
  $r3 = 0
  $r4 = pull_account(account: $r2, cap: $r1, overdraft: $r3)
  check_enough_funds($r4, $r1)
  $r5 = "dest"
  send_to_account(account: $r5)
`
	instrs, errs := Parse(src)
	require.Empty(t, errs)
	require.Equal(t, Dump(instrs), Dump(Optimize(instrs)))
}

// A pass whose output is never a fixed point of the next must still terminate.
func TestOptimizeBoundsTheSweeps(t *testing.T) {
	calls := 0
	flip := func(instrs []Instr) []Instr {
		calls++
		return append([]Instr{}, instrs...)
	}
	optimize(nil, []Peephole{flip})
	require.Equal(t, maxOptimizeSweeps, calls)
}
