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

// --- foldConstConditions ---

func TestFoldConstConditions(t *testing.T) {
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
			requirePass(t, foldConstConditions, tc.src, tc.want)
		})
	}
}

func TestFoldConstConditionsLeavesNonConstants(t *testing.T) {
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
  $r0 = true
  $r1 = 0
  jmp_if_true($r0, #end)
  $r1 = 1
#end
  $r2 = is_zero($r1)
`,
		},
		{
			name: "operand computed",
			src: `
  $r0 = 1
  $r1 = 2
  $r2 = $r0 + $r1
  $r3 = is_zero($r2)
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requireUnchanged(t, foldConstConditions, tc.src)
		})
	}
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

// --- Optimize ---

// The compiler's output for `send [USD/2 42] (source = @world destination =
// @dest)`, minus the amount clamp: the two passes have to compose for the
// literal @world to end up as a single unbounded pull.
func TestOptimizeCollapsesWorldDiamond(t *testing.T) {
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
	require.Equal(t, `  $r0 = "world"
  $r1 = "USD/2"
  $r2 = 42
  assert_non_negative_amount($r2)
  set_current_asset($r1)
  $r3 = "world"
  $r4 = 0
  $r5 = true
  $r6 = pull_account(account: $r3, cap: $r2)
  check_enough_funds($r6, $r2)
  $r7 = "dest"
  send_to_account(account: $r7)
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
