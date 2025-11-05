package parser_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/parser"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowErrorLines(t *testing.T) {
	script := `send [EUR/2 100] (
  source = err
  destination = ee
)`
	p := parser.Parse(script)
	snaps.MatchSnapshot(t, parser.ParseErrorsToString(p.Errors, script))
}

func TestPlainAddress(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = @src
  destination = @dest
)`)
	require.Empty(t, p.Errors)
	snaps.MatchSnapshot(t, p.Value)
}

func TestMultipleSends(t *testing.T) {
	p := parser.Parse(`
	send [COIN 10] ( source = @src destination = @dest )
	send [COIN 20] ( source = @src destination = @dest )
	`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestVariable(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = $example_var_src
  destination = $example_var_dest
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestNumberSyntaxUnderscore(t *testing.T) {
	p := parser.Parse(`
set_tx_meta("k1", 1_2_34_567)
`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestForbidInvalidNumberUnderscores(t *testing.T) {
	require.Len(t, parser.Parse(`
	set_tx_meta("k1", _123)
	`).Errors, 1)

	require.Len(t, parser.Parse(`
	set_tx_meta("k1", 123_)
	`).Errors, 1)

	require.Len(t, parser.Parse(`
	set_tx_meta("k1", 1__23)
	`).Errors, 3)
}

func TestVariableMonetary(t *testing.T) {
	p := parser.Parse(`send $example (
  source = @a
  destination = @b
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestInorderSource(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = { @s1 @s2 }
  destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestOneofSource(t *testing.T) {
	p := parser.Parse(`send $amt (
  source = oneof { @s1 @s2 }
  destination = @d
)`)
	require.Empty(t, p.Errors)
	snaps.MatchSnapshot(t, p.Value)
}

func TestNegativeNumberLit(t *testing.T) {
	p := parser.Parse(`send [EUR/2 -100] (
  source = @src
  destination = @dest
)`)
	require.Nil(t, p.Errors)
	snaps.MatchSnapshot(t, p.Value)
}

func TestInorderDestination(t *testing.T) {
	p := parser.Parse(`send $amt (
  source = @s
  destination = {
	max $m1 to @d1
	max [C 42] kept
	remaining to @d3
  }
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestOneofDestination(t *testing.T) {
	p := parser.Parse(`send $amt (
	source = @s
	destination = oneof {
		max $m1 to @d1
		max $m2 kept
		remaining to @d3
  }
	)`)
	require.Empty(t, p.Errors)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotment(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = { 1/3 from @s1 }
  destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotmentPerc(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = {
    42% from @s1
	1/2 from @s2
	remaining from @s3
  }
  destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotmentPercFloating(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = { 2.42% from @s }
  destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotmentVariableSource(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = {
	$x from @a
  }
  destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotmentDest(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = @s
  destination = { 1/2 to @d }
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotmentDestRemaining(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = @s
  destination = {
  	1/2 to @d
	remaining to @d2
  }
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestAllotmentDestKept(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = @s
  destination = { 1/2 kept }
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestCapped(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = max [EUR/2 10] from @src
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestCappedVariable(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = max $my_var from @src
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestNested(t *testing.T) {
	p := parser.Parse(`send [EUR/2 100] (
  source = {
    max [COIN 42] from @src
	@a
	@b
  }
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestEmptyVars(t *testing.T) {
	p := parser.Parse(`vars { }`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestSingleVar(t *testing.T) {
	p := parser.Parse(`vars { monetary $my_var }`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestOverdraftUnbounded(t *testing.T) {
	p := parser.Parse(`send $x (
	source = @src allowing unbounded overdraft
	destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestOverdraftUnboundedVariable(t *testing.T) {
	p := parser.Parse(`send $x (
	source = $my_src_var allowing unbounded overdraft
	destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestBoundedOverdraft(t *testing.T) {
	p := parser.Parse(`send $x (
	source = $my_src_var allowing overdraft up to [COIN 100]
	destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestFunctionCallNoArgs(t *testing.T) {
	p := parser.Parse(`example_fn()`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestFunctionCallOneArg(t *testing.T) {
	p := parser.Parse(`example_fn(@example)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestFunctionCallManyArgs(t *testing.T) {
	p := parser.Parse(`
example_fn(
	[COIN 42],
	100,
	1/2,
	$my_var,
	"example_str"
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestVarOrigin(t *testing.T) {
	p := parser.Parse(`
vars {
	monetary $my_var = origin_fn(@my_account, "str")
}
`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestSendAll(t *testing.T) {
	p := parser.Parse(`send [ASSET *] (
	source = @a
	destination = @b
)
`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestParseFloatingPerc(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		num, fl, err := parser.ParsePercentageRatio("1.23%")
		require.NoError(t, err)
		// (123/100)%
		require.Equal(t, big.NewInt(123), num)
		require.Equal(t, uint16(2), fl)
	})
	t.Run("leading zero", func(t *testing.T) {
		num, fl, err := parser.ParsePercentageRatio("0.23%")
		require.NoError(t, err)
		require.Equal(t, big.NewInt(23), num)
		require.Equal(t, uint16(2), fl)
		// (23/100)%
	})

	t.Run("trim leading zeros (1)", func(t *testing.T) {
		num, fl, err := parser.ParsePercentageRatio("0.01%")
		require.NoError(t, err)
		require.Equal(t, big.NewInt(1), num)
		require.Equal(t, uint16(2), fl)
		// (1/100)%
	})

	t.Run("trim leading zeros (2)", func(t *testing.T) {
		num, fl, err := parser.ParsePercentageRatio("0.019%")
		require.NoError(t, err)
		require.Equal(t, big.NewInt(19), num)
		require.Equal(t, uint16(3), fl)
		// (19/1000)%
	})

	t.Run("trim zero suffix", func(t *testing.T) {
		num, fl, err := parser.ParsePercentageRatio("1.20%")
		require.NoError(t, err)
		require.Equal(t, big.NewInt(120), num)
		require.Equal(t, uint16(2), fl)
		// (120/100)%
	})

	t.Run("zero", func(t *testing.T) {
		num, fl, err := parser.ParsePercentageRatio("0%")
		require.NoError(t, err)
		require.Equal(t, big.NewInt(0), num)
		require.Equal(t, uint16(0), fl)
	})

	t.Run("zero point zero", func(t *testing.T) {
		num, fl, err := parser.ParsePercentageRatio("0.0%")
		require.NoError(t, err)
		require.Equal(t, big.NewInt(0), num)
		require.Equal(t, uint16(1), fl) // 0 / 10^1 is still 0
	})

	t.Run("leading zeros in integer part", func(t *testing.T) {
		num, fl, err := parser.ParsePercentageRatio("007.5%")
		require.NoError(t, err)
		require.Equal(t, big.NewInt(75), num)
		require.Equal(t, uint16(1), fl)
	})

	t.Run("purely fractional small number", func(t *testing.T) {
		num, fl, err := parser.ParsePercentageRatio("0.00009%")
		require.NoError(t, err)
		require.Equal(t, big.NewInt(9), num)
		require.Equal(t, uint16(5), fl)
	})

}

func TestWhitespaceInRatio(t *testing.T) {
	p := parser.Parse(`
send $var (
  source = @world
  destination = {
    1 / 6 to @player:1
  }
)
	`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestSaveStatementSimple(t *testing.T) {
	p := parser.Parse(`
save [EUR/2 100] from @alice
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestSaveAllStatement(t *testing.T) {
	p := parser.Parse(`
save [EUR/2 *] from @alice
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestSaveStatementVar(t *testing.T) {
	p := parser.Parse(`
save $amt from $acc
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestInfix(t *testing.T) {
	p := parser.Parse(`
set_tx_meta("k1", 1 + "invalid arg")
set_tx_meta("k2", 1/2 - [COIN 10])
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestInfixPrec(t *testing.T) {
	// 1 + 2 - 3
	// should be the same as
	// (1 + 2) - 3
	p := parser.Parse(`
set_tx_meta("k1", 1 + 2 - 3)
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestParensInfixPrec(t *testing.T) {
	p := parser.Parse(`
set_tx_meta("k1", 1 + (2 - 3))
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestInfixSumDiv(t *testing.T) {
	// 1 + 2 / 3
	// should be the same as
	// 1 + (2/3)
	p := parser.Parse(`
set_tx_meta("k1", 1 + 2 / 3)
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestForbidInvalidAssetNames(t *testing.T) {
	for _, name := range []string{"", "/", "EUR/", "/42"} {
		t.Run(name, func(t *testing.T) {
			p := parser.Parse(fmt.Sprintf(
				`set_tx_meta("k1", %s)`,
				name,
			))
			require.NotEmpty(t, p.Errors)
		})
	}
}

func TestAllowValidAssetNames(t *testing.T) {
	for _, name := range []string{"EUR", "EUR/2", "TK123", "TK123/2"} {
		t.Run(name, func(t *testing.T) {
			p := parser.Parse(fmt.Sprintf(
				`set_tx_meta("k1", %s)`,
				name,
			))
			require.Empty(t, p.Errors)
		})
	}
}

func TestDivInfix(t *testing.T) {
	p := parser.Parse(`
set_tx_meta("k1", $x / $y)
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestDivVariableDenominator(t *testing.T) {
	p := parser.Parse(`
set_tx_meta("k1",  10/$y)
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestStringTemplate(t *testing.T) {
	p := parser.Parse(
		"set_tx_meta(0, 42)",
	)

	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestInterpAccount(t *testing.T) {
	p := parser.Parse(
		"set_tx_meta(@abc:cde:$id, 42)",
	)

	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestIncludeComments(t *testing.T) {
	p := parser.Parse(
		`
		// first comment
		f()

		// second comment
		g()
		`,
	)

	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestExprInVarOrigin(t *testing.T) {
	p := parser.Parse(`
vars {
	number $id = 42
}
	`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestMidScriptFnCall(t *testing.T) {
	p := parser.Parse(`
set_tx_meta(
	"k",
	"v",
	balance(@user, EUR/2)
)
`)
	require.Len(t, p.Errors, 0)
	snaps.MatchSnapshot(t, p.Value)
}

func TestColorRestriction(t *testing.T) {
	p := parser.Parse(`send $sent (
  source = @s \ "red"
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestColorRestrictionBoundedOverdraft(t *testing.T) {
	p := parser.Parse(`send $sent (
  source = @s \ "cl" allowing overdraft up to $ov
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestColorRestrictionUnboundedOverdraft(t *testing.T) {
	p := parser.Parse(`send $sent (
  source = $acc \ $col allowing unbounded overdraft
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}

func TestScalingSyntax(t *testing.T) {
	p := parser.Parse(`send $sent (
  source = @src with scaling
  destination = @dest
)`)
	snaps.MatchSnapshot(t, p.Value)
	assert.Empty(t, p.Errors)
}
