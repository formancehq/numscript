package analysis_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"

	"github.com/stretchr/testify/require"
)

func TestHoverOnSendMonetaryVar(t *testing.T) {
	input := `vars { monetary $amt }

send $amt (
	source = @a
	destination = @b
)`

	rng := parser.RangeOfIndexed(input, "$amt", 1)

	program := parser.Parse(input).Value
	hover := analysis.HoverOn(program, rng.Start)
	require.NotNil(t, hover)

	variableHover, ok := hover.(*analysis.VariableHover)
	require.True(t, ok, "Expected VariableHover")

	require.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	require.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	require.NotNil(t, resolved)
}

func TestHoverOnSaveMonetaryVar(t *testing.T) {
	input := `vars { monetary $amt }

save $amt from @acc
`

	rng := parser.RangeOfIndexed(input, "$amt", 1)

	program := parser.Parse(input).Value
	hover := analysis.HoverOn(program, rng.Start)
	require.NotNil(t, hover)

	variableHover, ok := hover.(*analysis.VariableHover)
	require.True(t, ok, "Expected VariableHover")

	require.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	require.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	require.NotNil(t, resolved)
}

func TestHoverOnAccountVar(t *testing.T) {
	input := `vars { account $acc }

save [COIN 100] from $acc
`

	rng := parser.RangeOfIndexed(input, "$acc", 1)

	program := parser.Parse(input).Value
	hover := analysis.HoverOn(program, rng.Start)
	require.NotNil(t, hover)

	variableHover, ok := hover.(*analysis.VariableHover)
	require.True(t, ok, "Expected VariableHover")

	require.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	require.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	require.NotNil(t, resolved)
}

func TestHoverOnMonetaryVarAmt(t *testing.T) {
	input := `vars { number $amt }

send [COIN 10] + [COIN $amt] (
	source = @a
	destination = @b
)`

	rng := parser.RangeOfIndexed(input, "$amt", 1)

	program := parser.Parse(input).Value
	hover := analysis.HoverOn(program, rng.Start)
	require.NotNil(t, hover)
}

func TestHoverOnMonetaryVarAsset(t *testing.T) {
	input := `vars { asset $asset }

send [$asset 100] (
	source = @a
	destination = @b
)`

	rng := parser.RangeOfIndexed(input, "$asset", 1)

	program := parser.Parse(input).Value
	hover := analysis.HoverOn(program, rng.Start)
	require.NotNil(t, hover)
}

func TestHoverOnSource(t *testing.T) {
	input := `vars { account $src }

send [C 10] (
	source = {
		1/2 from {
			@z
			max [C 1] from $src
		}
		1/2 from @b
	}
	destination = @dest
)`

	rng := parser.RangeOfIndexed(input, "$src", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNil(t, hover, "hover should not be nil")

	variableHover, ok := hover.(*analysis.VariableHover)
	require.True(t, ok, "Expected VariableHover")

	require.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	require.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	require.NotNil(t, resolved)
}

func TestHoverOnDestinationInorder(t *testing.T) {
	input := `vars { account $dest }

send [C 10] (
	source = @src
	destination = {
		1/2 to {
			max [C 10] to $dest
			remaining to @x
		}
		1/2 to @b
	}
)`

	rng := parser.RangeOfIndexed(input, "$dest", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")

	variableHover, ok := hover.(*analysis.VariableHover)
	require.Truef(t, ok, "Expected VariableHover")

	require.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	require.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	require.NotNil(t, resolved)
}

func TestHoverOnStringInterp(t *testing.T) {
	input := `vars { number $id }

send [ASSET *] (
	source = @world
	destination = @user:$id
)
`

	rng := parser.RangeOfIndexed(input, "$id", 1)

	program := parser.Parse(input).Value
	hover := analysis.HoverOn(program, rng.Start)
	require.NotNil(t, hover)

	variableHover, ok := hover.(*analysis.VariableHover)
	require.True(t, ok, "Expected VariableHover")

	require.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	require.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	require.NotNil(t, resolved)
}

func TestHoverOnDestinationInorderRemaining(t *testing.T) {
	input := `vars { account $dest }

send [C 10] (
	source = @src
	destination = {
		1/2 to {
			max [C 10] to @z
			remaining to $dest
		}
		1/2 to @b
	}
)`

	rng := parser.RangeOfIndexed(input, "$dest", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")

	variableHover, ok := hover.(*analysis.VariableHover)
	require.Truef(t, ok, "Expected VariableHover")

	require.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	require.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	require.NotNil(t, resolved)
}

func TestHoverOnSrcVariableAllotment(t *testing.T) {
	input := `vars { portion $portion }

send [C 10] (
	source = { $portion from @a }
	destination = @dest
)`

	rng := parser.RangeOfIndexed(input, "$portion", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnDestVariableAllotment(t *testing.T) {
	input := `vars { portion $portion }

send [C 10] (
	source = @s
	destination = { $portion to @a }
)`

	rng := parser.RangeOfIndexed(input, "$portion", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnUnboundSrc(t *testing.T) {
	input := `vars { account $acc }

send [C 10] (
	source = $acc allowing unbounded overdraft
	destination = @dest
)`

	rng := parser.RangeOfIndexed(input, "$acc", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnBoundedCap(t *testing.T) {
	input := `vars { monetary $mon }

send [C 10] (
	source = @s allowing overdraft up to $mon
	destination = @dest
)`

	rng := parser.RangeOfIndexed(input, "$mon", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnFnCall(t *testing.T) {
	input := `vars { string $str }
set_tx_meta($str, 42)
`

	rng := parser.RangeOfIndexed(input, "$str", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnOriginFnCall(t *testing.T) {
	input := `vars {
		string $arg
		number $meta_variable = meta(@account, $arg)
	}`

	rng := parser.RangeOfIndexed(input, "$arg", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnFnCallDocs(t *testing.T) {
	input := `set_tx_meta()`

	rng := parser.RangeOfIndexed(input, "set_tx_meta", 0)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")

	fnHover, ok := hover.(*analysis.BuiltinFnHover)
	if !ok {
		t.Fatalf("Expected a BuiltinFnHover")
	}

	require.Equal(t, fnHover.Range, rng)
	require.NotNil(t, fnHover.Node)
	require.Equal(t, "set_tx_meta", fnHover.Node.Caller.Name)
}

func TestHoverOnFnOriginDocs(t *testing.T) {
	input := `vars { monetary $m = balance(@a, COIN) }`

	rng := parser.RangeOfIndexed(input, "balance", 0)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	require.NotNilf(t, hover, "hover should not be nil")

	fnHover, ok := hover.(*analysis.BuiltinFnHover)
	if !ok {
		t.Fatalf("Expected a BuiltinFnHover")
	}

	require.Equal(t, fnHover.Range, rng)
	require.NotNil(t, fnHover.Node)
	require.Equal(t, "balance", fnHover.Node.Caller.Name)

	res := analysis.CheckProgram(program)
	resolution := res.ResolveBuiltinFn(fnHover.Node.Caller)
	require.NotNil(t, resolution)

	resolution, ok = resolution.(analysis.VarOriginFnCallResolution)
	if !ok {
		t.Fatalf("Expected a VarOriginFnCallResolution (got %v)", resolution)
	}

}

func TestHoverFaultTolerance(t *testing.T) {
	t.Run("missing lit", func(t *testing.T) {
		input := `
			send [COIN 10] (
				source = max <invalidtk> from @a
				destination = @a
			)
		`

		rng := parser.RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		require.Nil(t, hover)
	})

	t.Run("missing source", func(t *testing.T) {
		input := `
			send [COIN 10] (
				source = <invalidtk>
				destination = @a
			)
		`

		rng := parser.RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		require.Nil(t, hover)
	})

	t.Run("missing source in inorder", func(t *testing.T) {
		input := `
			send [COIN 10] (
				source = {
				 	@a
					<invalidtk>
				}
				destination = @a
			)
		`

		rng := parser.RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		require.Nil(t, hover)
	})

	t.Run("missing dest", func(t *testing.T) {
		input := `
			send [COIN 10] (
				source = @a
				destination = <invalidtk>
			)
		`
		rng := parser.RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		require.Nil(t, hover)
	})

	t.Run("invalid call arg", func(t *testing.T) {
		input := `
			set_tx_meta(<invalidtk>)
		`
		rng := parser.RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		require.Nil(t, hover)
	})

	t.Run("missing sent value", func(t *testing.T) {
		input := `
			send <invalidtk> (
			)
		`
		rng := parser.RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		require.Nil(t, hover)
	})

	t.Run("missing inorder clause", func(t *testing.T) {
		input := `
			send [COIN 10] (
				source = @a
				destination = {
					<invalidtk-0>
					remaining to <invalidtk>
				}
			)
		`
		rng := parser.RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		require.Nil(t, hover)
	})
}
