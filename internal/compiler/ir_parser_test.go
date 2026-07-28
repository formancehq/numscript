package compiler

import (
	"testing"

	"github.com/formancehq/numscript/internal/irparser"
	"github.com/stretchr/testify/require"
)

// parseAndTransform parses an IR text and transforms it to irInstr.
func parseAndTransform(t *testing.T, source string) ([]irInstr, string) {
	t.Helper()

	result := irparser.Parse(source)
	require.Empty(t, result.Errors, "IR parse errors: %v", result.Errors)

	instrs, errs := Transform(result.Value)
	require.Empty(t, errs, "transform errors: %v", errs)

	dumped := "\n" + dump(instrs)
	return instrs, dumped
}

// TestRoundtripSimpleProgram verifies a simple program roundtrips.
func TestRoundtripSimpleProgram(t *testing.T) {
	// Compile numscript → IR text
	compiledIR := getCompiledOutput(t, `
		send [USD/2 10] (
			source = @src
			destination = @dest
		)
	`)

	// Parse IR text → irInstr → dump again
	_, roundtripped := parseAndTransform(t, compiledIR)

	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripIntAddition verifies int addition roundtrips.
func TestRoundtripIntAddition(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		send [USD/2 4 + 6] (
			source = @src
			destination = @dest
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripMonetaryAddition verifies monetary addition roundtrips.
func TestRoundtripMonetaryAddition(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		vars {
			monetary $a = [USD/2 3]
			monetary $b = [USD/2 7]
		}
		send $a + $b (
			source = @src
			destination = @dest
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripPrefixMinusMonetary verifies monetary negation roundtrips.
func TestRoundtripPrefixMinusMonetary(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		vars {
			monetary $neg_mon = [USD/2 -10]
			monetary $pos_mon = -$neg_mon
		}
		send $pos_mon (
			source = @src
			destination = @dest
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripBalance verifies balance roundtrips.
func TestRoundtripBalance(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		vars {
			monetary $bal = balance(@src, USD/2)
		}
		send $bal (
			source = @src
			destination = @dest
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripAccountInterpolation verifies account interpolation roundtrips.
func TestRoundtripAccountInterpolation(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		vars {
			string $id = "alice"
		}
		send [USD/2 10] (
			source = @world
			destination = @users:$id:wallet
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripInorder verifies inorder source roundtrips.
func TestRoundtripInorder(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		send [USD/2 10] (
			source = {
				@a
				@b
				@c
			}
			destination = @dest
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripInorderWithCap verifies inorder with cap roundtrips.
func TestRoundtripInorderWithCap(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		send [USD/2 10] (
			source = {
				@a
				max [USD/2 5] from @b
				@c
			}
			destination = @dest
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripDestInorder verifies destination inorder roundtrips.
func TestRoundtripDestInorder(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		send [USD/2 10] (
			source = @world
			destination = {
        max [USD/2 4] to @d1
        remaining to @d2
      }
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripGetAmount verifies get_amount roundtrips.
func TestRoundtripGetAmount(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		vars {
			monetary $m = [USD/2 42]
			number $n = get_amount($m)
		}
		send [USD/2 $n] (
			source = @src
			destination = @dest
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripGetAsset verifies get_asset roundtrips.
func TestRoundtripGetAsset(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		vars {
			monetary $m = [USD/2 42]
			asset $a = get_asset($m)
		}
		send [$a 10] (
			source = @src
			destination = @dest
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestRoundtripAccountInterpolationInt verifies account interpolation with int roundtrips.
func TestRoundtripAccountInterpolationInt(t *testing.T) {
	compiledIR := getCompiledOutput(t, `
		vars {
			number $n = 42
		}
		send [USD/2 10] (
			source = @world
			destination = @account:$n
		)
	`)
	_, roundtripped := parseAndTransform(t, compiledIR)
	require.Equal(t, compiledIR, roundtripped)
}

// TestParseAndTransformErrors checks error handling.
func TestParseAndTransformErrors(t *testing.T) {
	t.Run("unknown instruction", func(t *testing.T) {
		result := irparser.Parse(`
  $r0 = no_such_instr($r1)
`)
		require.Empty(t, result.Errors)
		_, errs := Transform(result.Value)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "unknown instruction")
	})

	t.Run("invalid arg type", func(t *testing.T) {
		result := irparser.Parse(`
  $r0 = get_asset(42)
`)
		require.Empty(t, result.Errors)
		_, errs := Transform(result.Value)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "expected register")
	})

	t.Run("unbound jmp label", func(t *testing.T) {
		result := irparser.Parse(`
  jmp_if_zero($r0, #missing_label)
`)
		require.Empty(t, result.Errors)
		_, errs := Transform(result.Value)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Msg, "not defined")
	})

	t.Run("jmp with valid label", func(t *testing.T) {
		result := irparser.Parse(`
#my_label
  jmp_if_zero($r0, #my_label)
`)
		require.Empty(t, result.Errors)
		_, errs := Transform(result.Value)
		require.Empty(t, errs)
	})
}

// TestRegNamesAreBounded checks that no name the grammar accepts resolves past
// maxRegIndex, which is what leaves room for the fresh registers `_` desugars to.
func TestRegNamesAreBounded(t *testing.T) {
	names := []string{
		"$r0", "$r255",
		"$r16777215",             // last index spelled as-is
		"$r16777216",             // first index past the bound: falls back to hashing
		"$r99999999999999999999", // overflows uint32
		"$a", "$my_reg", "$_", "$int",
		"$zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", // hash overflows uint32
	}
	for _, name := range names {
		r := regRefToReg(irparser.RegRef{Name: name})
		require.Less(t, uint(r), uint(maxRegIndex), "%s resolved past the bound", name)
	}
}

// TestDiscardDestDesugarsToFreshReg checks that `_` becomes a register no
// statement can name, and that two discards don't alias — otherwise they'd be
// forced to share a type.
func TestDiscardDestDesugarsToFreshReg(t *testing.T) {
	instrs, dumped := parseAndTransform(t, `
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
	require.NoError(t, typecheckInstructions(instrs))

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
			name: "loadStr",
			ir: `
  $r0 = "hello"
`,
		},
		{
			name: "loadInt",
			ir: `
  $r0 = 42
`,
		},
		{
			name: "mk_monetary",
			ir: `
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
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
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  $r4 = get_amount($r2)
  $r5 = neg_int($r1)
  $r6 = int_copy($r1)
  $r7 = int_to_string($r1)
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
			name: "assert_non_negative_balance",
			ir: `
  $r0 = "USD/2"
  $r1 = 100
  $r2 = mk_monetary($r0, $r1)
  $r3 = "src"
  assert_non_negative_balance($r2, $r3)
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
			name: "meta monetary",
			ir: `
  $r0 = "acct"
  $r1 = "key"
  $r2 = meta<monetary>($r0, $r1)
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
			name: "jmp_if_zero and label",
			ir: `
#my_label
  $r0 = 0
  jmp_if_zero($r0, #my_label)
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
			name: "min_int",
			ir: `
  $r0 = 10
  $r1 = 5
  $r2 = min_int($r0, $r1)
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
  $r2 = mk_monetary($r0, $r1)
  $r3 = monetary_to_string($r2)
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// source already has leading newline and indentation
			source := tt.ir
			_, roundtripped := parseAndTransform(t, source)
			require.Equal(t, source, roundtripped)
		})
	}
}
