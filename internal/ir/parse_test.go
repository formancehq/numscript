package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// parseAndDump parses an IR text and re-dumps it.
func parseAndDump(t *testing.T, source string) ([]Instr, string) {
	t.Helper()
	instrs, errs := Parse(source)
	require.Empty(t, errs, "IR errors: %v", errs)
	return instrs, "\n" + Dump(instrs)
}

// TestParseErrors checks what Parse rejects.
func TestParseErrors(t *testing.T) {
	t.Run("unknown instruction", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = no_such_instr($r1)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "unknown instruction")
	})

	t.Run("invalid arg type", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = neg_int(42)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "expected register")
	})

	t.Run("unbound jmp label", func(t *testing.T) {
		_, errs := Parse(`
  jmp_if_false($r0, #missing_label)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "not defined")
	})

	t.Run("forward jmp", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = 0
  jmp_if_false($r0, #my_label)
#my_label
`)
		require.Empty(t, errs)
	})

	t.Run("backward jmp", func(t *testing.T) {
		_, errs := Parse(`
#my_label
  jmp_if_false($r0, #my_label)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "must go forward")
	})

	t.Run("forward unconditional jmp", func(t *testing.T) {
		_, errs := Parse(`
  jmp(#my_label)
#my_label
`)
		require.Empty(t, errs)
	})

	t.Run("backward unconditional jmp", func(t *testing.T) {
		_, errs := Parse(`
#my_label
  jmp(#my_label)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "must go forward")
	})

	t.Run("unconditional jmp to an undefined label", func(t *testing.T) {
		_, errs := Parse(`
  jmp(#nope)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "not defined")
	})

	t.Run("missing required labeled arg", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = "acc"
  $r1 = 1
  $r2 = pull_account(cap: $r1)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, `missing labeled argument "account"`)
	})

	t.Run("labeled arg no instruction takes", func(t *testing.T) {
		// a stray label on an arg that was consumed positionally: the extra-args
		// loop can't see it, so it's the leftover-label check that reports it
		_, errs := Parse(`
  $r0 = 1
  $r1 = 2
  check_enough_funds(foo: $r0, $r1)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, `unknown labeled argument "foo"`)
	})

	t.Run("labeled arg an instruction can't place", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = "acc"
  $r1 = pull_account(account: $r0, nope: $r0)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "unexpected extra argument")
	})

	t.Run("labels are case sensitive", func(t *testing.T) {
		// IDENTIFIER is lowercase-only, so this doesn't even lex
		_, errs := Parse(`
  $r0 = "acc"
  $r1 = pull_account(Account: $r0)
`)
		require.NotEmpty(t, errs)
	})

	t.Run("duplicate labeled arg", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = "acc"
  $r1 = 1
  $r2 = pull_account(account: $r0, cap: $r1, cap: $r1)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "duplicate labeled argument")
	})

	t.Run("duplicate label", func(t *testing.T) {
		_, errs := Parse(`
  jmp_if_false($r0, #l)
#l
#l
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "duplicate label")
	})

	t.Run("const assigned to a dest list", func(t *testing.T) {
		_, errs := Parse(`
  [$r0, $r1] = 1
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "const assignment requires a single register dest")
	})

	t.Run("infix assigned to a dest list", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = 1
  [$r1, $r2] = $r0 + $r0
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "infix requires a single register dest")
	})

	t.Run("compound assign to a dest list", func(t *testing.T) {
		// the grammar only allows a single register left of `+=`, so this one is
		// rejected before the transform sees it
		_, errs := Parse(`
  $r0 = 1
  [$r1, $r2] += $r0
`)
		require.NotEmpty(t, errs)
	})

	t.Run("labeled arg of the wrong kind", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = pull_account(account: 42)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, `labeled arg "account": expected register, got integer literal`)
	})

	t.Run("register where a label is expected", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = 1
  jmp_if_false($r0, $r0)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "expected label, got register")
	})

	t.Run("register where a register list is expected", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = 1
  [$r1] = mk_allot($r0, $r0)
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "expected register list, got register")
	})

	t.Run("register list where a register is expected", func(t *testing.T) {
		_, errs := Parse(`
  $r0 = "USD/2"
  set_current_asset([$r0, $r0])
`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "expected register, got register list")
	})
}

// The infix forms are sugar: the call form of every binary op parses too, even
// though ir.Dump never prints it for add_int / sub_int.
func TestBinaryOpCallForms(t *testing.T) {
	instrs, errs := Parse(`
  $i = 1
  $s = "x"
  $p = mk_portion($i, $i)
  $sum = add_int($i, $i)
  $diff = sub_int($i, $i)
  $cat = add_string($s, $s)
  $rest = sub_portion($p, $p)
  $label = monetary_to_string($s, $i)
`)
	require.Empty(t, errs)
	require.NoError(t, Typecheck(instrs))

	// the dump switches the two int ops back to their infix spelling
	require.Contains(t, Dump(instrs), "$r3 = $r0 + $r0")
	require.Contains(t, Dump(instrs), "$r4 = $r0 - $r0")
	require.Contains(t, Dump(instrs), "$r5 = add_string($r1, $r1)")
	require.Contains(t, Dump(instrs), "$r6 = sub_portion($r2, $r2)")
}

func TestParseErrorMessageFormat(t *testing.T) {
	_, errs := Parse(`
  $r0 = no_such_instr($r1)
`)
	require.NotEmpty(t, errs)
	// 1-based line:character, then the reason
	require.Regexp(t, `^\d+:\d+: .*unknown instruction`, errs[0].Error())
}

// TestReadBeforeWrite checks that reading a register nothing ever assigned to is
// rejected, and reported under the name the text used.
func TestReadBeforeWrite(t *testing.T) {
	t.Run("never written at all", func(t *testing.T) {
		_, errs := Parse(`
  $a = 42
  $y = lt_int($a, $b)
`)
		require.Len(t, errs, 1)
		require.Contains(t, errs[0].Msg, "$b is read but never written")
	})

	t.Run("written only after the read", func(t *testing.T) {
		_, errs := Parse(`
  $a = 42
  $y = lt_int($a, $b)
  $b = 1
`)
		require.Len(t, errs, 1)
		require.Contains(t, errs[0].Msg, "$b is read but never written")
	})

	t.Run("compound assign reads its own dest", func(t *testing.T) {
		_, errs := Parse(`
  $b = 1
  $acc += $b
`)
		require.Len(t, errs, 1)
		require.Contains(t, errs[0].Msg, "$acc is read but never written")
	})

	t.Run("labeled args are reads too", func(t *testing.T) {
		_, errs := Parse(`
  $acc = "src"
  $pulled = pull_account(account: $acc, cap: $missing)
`)
		require.Len(t, errs, 1)
		require.Contains(t, errs[0].Msg, "$missing is read but never written")
	})

	t.Run("a dest is written for later instructions", func(t *testing.T) {
		_, errs := Parse(`
  $a = 1
  $b = 2
  $sum = $a + $b
  $twice = $sum + $sum
`)
		require.Empty(t, errs)
	})

	t.Run("mk_allot dests count as written", func(t *testing.T) {
		_, errs := Parse(`
  $amount = 100
  $num = 1
  $den = 2
  $half = mk_portion($num, $den)
  [$first, $second] = mk_allot($amount, [$half, $half])
  check_enough_funds($first, $second)
`)
		require.Empty(t, errs)
	})
}

// TestMalformedInputIsRejected checks that text → Instr reports errors on
// invalid input rather than panicking. It doesn't typecheck: only syntax and
// the structural rules (labels resolve, jumps go forward) are checked here.
func TestMalformedInputIsRejected(t *testing.T) {
	sources := []struct {
		name string
		ir   string
	}{
		{"comment", "// not a comment in this format\n  $r0 = 1\n"},
		{"no args at all", "  $r0 = neg_int()\n"},
		{"too few args", "  $r0 = balance($r1)\n"},
		{"too many args", "  $r0 = neg_int($r1, $r2)\n"},
		{"missing required labeled arg", "  $r0 = pull_account(cap: $r1)\n"},
		{"unknown labeled arg", "  $r0 = pull_account(account: $r1, nope: $r2)\n"},
		{"capitalised label", "  $r0 = pull_account(Account: $r1)\n"},
		{"load_var index out of range", "  $r0 = load_var<int>(70000)\n"},
		{"load_var without type param", "  $r0 = load_var(0)\n"},
		{"load_var with a type it doesn't have", "  $r0 = load_var<portion>(0)\n"},
		{"meta without type param", "  $r0 = meta($r1, $r2)\n"},
		{"mk_allot without dest list", "  $r0 = mk_allot($r1, [$r2])\n"},
		{"reg to reg copy", "  $r0 = $r1\n"},
		{"garbage", "$$$ !!!"},
		{"unclosed paren", "  $r0 = neg_int($r1"},
		{"uppercase instr name", "  $r0 = NEG_INT($r1)"},
		{"negative int literal", "  $r0 = -1\n"},
		{"empty dest list", "  [] = mk_allot($r0, [$r1])\n"},
		{"missing dest", "  = neg_int($r0)\n"},
		{"unterminated string", "  $r0 = \"oops\n"},
		{"stray operator", "  $r0 = $r1 * $r2\n"},
		{"type param on plain instr", "  $r0 = neg_int<int>($r1)\n"},
		{"label as instr arg", "  set_current_asset(#lbl)\n"},
		// a bool is a const, never an operand: no instruction takes one inline
		{"bool as instr arg", "  set_current_asset(true)\n"},
		{"bool as labeled arg", "  $r0 = pull_account(account: false)\n"},
		{"bool in a dest list", "  [$r0, $r1] = true\n"},
		{"capitalised bool", "  $r0 = True\n"},
	}

	for _, s := range sources {
		t.Run(s.name, func(t *testing.T) {
			instrs, errs := Parse(s.ir)
			require.NotEmpty(t, errs, "neither the parser nor the transform rejected it")
			// and whatever it did return must be usable: a nil instruction in the
			// stream would blow up in dump or assemble instead
			for _, instr := range instrs {
				require.NotNil(t, instr)
			}
		})
	}
}

// TestRegNamesBindInOrder checks how a name becomes a logical register: the
// first appearance allocates the next one, later appearances reuse it. The name
// itself carries no meaning — `$r<N>` is a convention, not an index.
func TestRegNamesBindInOrder(t *testing.T) {
	_, dumped := parseAndDump(t, `
  $asset = "USD/2"
  $amount = 10
  $label = monetary_to_string($asset, $amount)
  $twice = add_int($amount, $amount)
  $r99 = int_to_string($twice)
`)

	require.Equal(t, `
  $r0 = "USD/2"
  $r1 = 10
  $r2 = monetary_to_string($r0, $r1)
  $r3 = $r1 + $r1
  $r4 = int_to_string($r3)
`, dumped)
}

// TestDiscardDestDesugarsToFreshReg checks that `_` becomes a register no
// statement can name, and that two discards don't alias — otherwise they'd be
// forced to share a type.
func TestDiscardDestDesugarsToFreshReg(t *testing.T) {
	instrs, dumped := parseAndDump(t, `
  $r0 = "acc"
  $r1 = "USD/2"
  _ = pull_account(account: $r0)
  _ = balance($r0, $r1)
`)

	pulled := instrs[2].dests()[0]
	balance := instrs[3].dests()[0]
	require.NotEqual(t, pulled, balance)
	// above every register the text refers to ($r0, $r1)
	require.Greater(t, uint(pulled), uint(1))
	require.Greater(t, uint(balance), uint(1))

	// so the typechecker doesn't see one register written with two types
	require.NoError(t, Typecheck(instrs))

	// the IR has no notion of a discard: it dumps as the register it desugared to
	require.Equal(t, `
  $r0 = "acc"
  $r1 = "USD/2"
  $r2 = pull_account(account: $r0)
  $r3 = balance($r0, $r1)
`, dumped)
}

// TestRoundtripAllInstructions tests every instruction in isolation for roundtrip.
func TestRoundtripAllInstructions(t *testing.T) {
	tests := []struct {
		name string
		ir   string
	}{
		{
			name: "LoadStr",
			ir: `
  $r0 = "hello"
`,
		},
		{
			name: "LoadInt",
			ir: `
  $r0 = 42
`,
		},
		{
			name: "mk_portion",
			ir: `
  $r0 = 1
  $r1 = 2
  $r2 = mk_portion($r0, $r1)
`,
		},
		{
			name: "add_int via infix",
			ir: `
  $r0 = 1
  $r1 = 2
  $r2 = $r0 + $r1
`,
		},
		{
			name: "infix add",
			ir: `
  $r0 = 1
  $r1 = 2
  $r2 = $r0 + $r1
`,
		},
		{
			name: "compound add",
			ir: `
  $r0 = 0
  $r1 = 1
  $r0 += $r1
`,
		},
		{
			name: "infix sub",
			ir: `
  $r0 = 5
  $r1 = 3
  $r2 = $r0 - $r1
`,
		},
		{
			name: "compound sub",
			ir: `
  $r0 = 5
  $r1 = 3
  $r0 -= $r1
`,
		},
		{
			name: "unary ops",
			ir: `
  $r0 = 10
  $r1 = neg_int($r0)
  $r2 = int_copy($r0)
  $r3 = int_to_string($r0)
`,
		},
		{
			name: "pull_account with all labeled args",
			ir: `
  $r0 = "src"
  $r1 = 100
  $r2 = 0
  $r3 = "red"
  $r4 = pull_account(account: $r0, cap: $r1, overdraft: $r2, color: $r3)
`,
		},
		{
			name: "pull_account minimal",
			ir: `
  $r0 = "src"
  $r1 = pull_account(account: $r0)
`,
		},
		{
			name: "send_to_account",
			ir: `
  $r0 = "dest"
  send_to_account(account: $r0)
`,
		},
		{
			name: "send_to_account with cap",
			ir: `
  $r0 = "dest"
  $r1 = 50
  send_to_account(account: $r0, cap: $r1)
`,
		},
		{
			name: "save with amount",
			ir: `
  $r0 = "acct"
  $r1 = "USD/2"
  $r2 = 100
  save(account: $r0, asset: $r1, amount: $r2)
`,
		},
		{
			name: "save all",
			ir: `
  $r0 = "acct"
  $r1 = "USD/2"
  save(account: $r0, asset: $r1)
`,
		},
		{
			name: "mk_allot",
			ir: `
  $r0 = 100
  $r1 = 1
  $r2 = 1
  $r3 = mk_portion($r1, $r2)
  $r4 = 1
  $r5 = 1
  $r6 = mk_portion($r4, $r5)
  [$r7, $r8] = mk_allot($r0, [$r3, $r6])
`,
		},
		{
			name: "check_enough_funds",
			ir: `
  $r0 = 50
  $r1 = 100
  check_enough_funds($r0, $r1)
`,
		},
		{
			name: "assert_leftover",
			ir: `
  $r0 = 1
  $r1 = 1
  $r2 = mk_portion($r0, $r1)
  assert_leftover($r2)
`,
		},
		{
			name: "assert_leftover_exact",
			ir: `
  $r0 = 1
  $r1 = 1
  $r2 = mk_portion($r0, $r1)
  assert_leftover_exact($r2)
`,
		},
		{
			name: "assert_same_asset",
			ir: `
  $r0 = "USD/2"
  $r1 = "EUR/2"
  assert_same_asset($r0, $r1)
`,
		},
		{
			name: "assert_valid_account",
			ir: `
  $r0 = "users:alice"
  assert_valid_account($r0)
`,
		},
		{
			name: "assert_valid_color",
			ir: `
  $r0 = "RED"
  assert_valid_color($r0)
`,
		},
		{
			name: "assert_non_negative_balance",
			ir: `
  $r0 = 100
  $r1 = "src"
  assert_non_negative_balance($r0, $r1)
`,
		},
		{
			name: "set_tx_meta",
			ir: `
  $r0 = "key"
  $r1 = "value"
  set_tx_meta($r0, $r1)
`,
		},
		{
			name: "set_account_meta",
			ir: `
  $r0 = "acct"
  $r1 = "key"
  $r2 = "value"
  set_account_meta($r0, $r1, $r2)
`,
		},
		{
			name: "set_current_asset",
			ir: `
  $r0 = "USD/2"
  set_current_asset($r0)
`,
		},
		{
			name: "balance",
			ir: `
  $r0 = "src"
  $r1 = "USD/2"
  $r2 = balance($r0, $r1)
`,
		},
		{
			name: "meta str",
			ir: `
  $r0 = "acct"
  $r1 = "key"
  $r2 = meta<str>($r0, $r1)
`,
		},
		{
			name: "meta int",
			ir: `
  $r0 = "acct"
  $r1 = "key"
  $r2 = meta<int>($r0, $r1)
`,
		},
		{
			name: "meta portion",
			ir: `
  $r0 = "acct"
  $r1 = "key"
  $r2 = meta<portion>($r0, $r1)
`,
		},
		{
			name: "meta_monetary",
			ir: `
  $r0 = "acct"
  $r1 = "key"
  [$r2, $r3] = meta_monetary($r0, $r1)
`,
		},
		{
			name: "load_var int",
			ir: `
  $r0 = load_var<int>(0)
`,
		},
		{
			name: "load_var str",
			ir: `
  $r0 = load_var<str>(1)
`,
		},
		{
			name: "snapshot and restore",
			ir: `
  $r0 = snapshot()
  restore($r0)
`,
		},
		{
			name: "jmp_if_false and label",
			ir: `
  $r0 = true
  jmp_if_false($r0, #my_label)
#my_label
`,
		},
		{
			name: "jmp_if_true and label",
			ir: `
  $r0 = false
  jmp_if_true($r0, #my_label)
#my_label
`,
		},
		{
			name: "is_zero",
			ir: `
  $r0 = 1
  $r1 = is_zero($r0)
`,
		},
		{
			name: "jmp and label",
			ir: `
  jmp(#my_label)
#my_label
`,
		},
		{
			name: "str_eq",
			ir: `
  $r0 = "a"
  $r1 = "b"
  $r2 = str_eq($r0, $r1)
`,
		},
		{
			name: "sub_int via infix",
			ir: `
  $r0 = 5
  $r1 = 3
  $r2 = $r0 - $r1
`,
		},
		{
			name: "add_string",
			ir: `
  $r0 = "hello"
  $r1 = "world"
  $r2 = add_string($r0, $r1)
`,
		},
		{
			name: "lt_int",
			ir: `
  $r0 = 10
  $r1 = 5
  $r2 = lt_int($r0, $r1)
`,
		},
		{
			name: "eq_int",
			ir: `
  $r0 = 10
  $r1 = 5
  $r2 = eq_int($r0, $r1)
`,
		},
		{
			name: "lt_portion",
			ir: `
  $r0 = 1
  $r1 = 2
  $r2 = mk_portion($r0, $r1)
  $r3 = lt_portion($r2, $r2)
`,
		},
		{
			name: "eq_portion",
			ir: `
  $r0 = 1
  $r1 = 2
  $r2 = mk_portion($r0, $r1)
  $r3 = eq_portion($r2, $r2)
`,
		},
		{
			name: "not",
			ir: `
  $r0 = true
  $r1 = not($r0)
`,
		},
		{
			name: "portion_copy",
			ir: `
  $r0 = 1
  $r1 = 1
  $r2 = mk_portion($r0, $r1)
  $r3 = portion_copy($r2)
`,
		},
		{
			name: "str_copy",
			ir: `
  $r0 = "USD/2"
  $r1 = str_copy($r0)
`,
		},
		{
			name: "bool_copy",
			ir: `
  $r0 = true
  $r1 = bool_copy($r0)
`,
		},
		{
			name: "portion_to_string",
			ir: `
  $r0 = 1
  $r1 = 1
  $r2 = mk_portion($r0, $r1)
  $r3 = portion_to_string($r2)
`,
		},
		{
			name: "monetary_to_string",
			ir: `
  $r0 = "USD/2"
  $r1 = 100
  $r2 = monetary_to_string($r0, $r1)
`,
		},
		{
			name: "bool true",
			ir: `
  $r0 = true
`,
		},
		{
			name: "bool false",
			ir: `
  $r0 = false
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// source already has leading newline and indentation
			source := tt.ir
			_, roundtripped := parseAndDump(t, source)
			require.Equal(t, source, roundtripped)
		})
	}
}
