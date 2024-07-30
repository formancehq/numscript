package analysis_test

import (
	"numscript/analysis"
	"numscript/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHoverOnSendMonetaryVar(t *testing.T) {
	input := `vars { monetary $amt }

send $amt (
	source = @a
	destination = @b
)`

	rng := RangeOfIndexed(input, "$amt", 1)

	program := parser.Parse(input).Value
	hover := analysis.HoverOn(program, rng.Start)
	assert.NotNil(t, hover)

	variableHover, ok := hover.(*analysis.VariableHover)
	assert.True(t, ok, "Expected VariableHover")

	assert.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	assert.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	assert.NotNil(t, resolved)
}

func TestHoverOnMonetaryVarAmt(t *testing.T) {
	input := `vars { number $amt }

send [COIN $amt] (
	source = @a
	destination = @b
)`

	rng := RangeOfIndexed(input, "$amt", 1)

	program := parser.Parse(input).Value
	hover := analysis.HoverOn(program, rng.Start)
	assert.NotNil(t, hover)
}

func TestHoverOnMonetaryVarAsset(t *testing.T) {
	input := `vars { asset $asset }

send [$asset 100] (
	source = @a
	destination = @b
)`

	rng := RangeOfIndexed(input, "$asset", 1)

	program := parser.Parse(input).Value
	hover := analysis.HoverOn(program, rng.Start)
	assert.NotNil(t, hover)
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

	rng := RangeOfIndexed(input, "$src", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNil(t, hover, "hover should not be nil")

	variableHover, ok := hover.(*analysis.VariableHover)
	assert.True(t, ok, "Expected VariableHover")

	assert.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	assert.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	assert.NotNil(t, resolved)
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

	rng := RangeOfIndexed(input, "$dest", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")

	variableHover, ok := hover.(*analysis.VariableHover)
	assert.Truef(t, ok, "Expected VariableHover")

	assert.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	assert.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	assert.NotNil(t, resolved)
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

	rng := RangeOfIndexed(input, "$dest", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")

	variableHover, ok := hover.(*analysis.VariableHover)
	assert.Truef(t, ok, "Expected VariableHover")

	assert.Equal(t, rng, variableHover.Range)

	checkResult := analysis.CheckProgram(program)
	assert.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	assert.NotNil(t, resolved)
}

func TestHoverOnSrcVariableAllotment(t *testing.T) {
	input := `vars { portion $portion }

send [C 10] (
	source = { $portion from @a }
	destination = @dest
)`

	rng := RangeOfIndexed(input, "$portion", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnDestVariableAllotment(t *testing.T) {
	input := `vars { portion $portion }

send [C 10] (
	source = @s
	destination = { $portion to @a }
)`

	rng := RangeOfIndexed(input, "$portion", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnUnboundSrc(t *testing.T) {
	input := `vars { account $acc }

send [C 10] (
	source = $acc allowing unbounded overdraft
	destination = @dest
)`

	rng := RangeOfIndexed(input, "$acc", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnBoundedCap(t *testing.T) {
	input := `vars { monetary $mon }

send [C 10] (
	source = @s allowing overdraft up to $mon
	destination = @dest
)`

	rng := RangeOfIndexed(input, "$mon", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnFnCall(t *testing.T) {
	input := `vars { string $str }
set_tx_meta($str, 42)
`

	rng := RangeOfIndexed(input, "$str", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnOriginFnCall(t *testing.T) {
	input := `vars {
		string $arg
		number $meta_variable = meta(@account, $arg)
	}`

	rng := RangeOfIndexed(input, "$arg", 1)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")
}

func TestHoverOnFnCallDocs(t *testing.T) {
	input := `set_tx_meta()`

	rng := RangeOfIndexed(input, "set_tx_meta", 0)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")

	fnHover, ok := hover.(*analysis.BuiltinFnHover)
	if !ok {
		t.Fatalf("Expected a BuiltinFnHover")
	}

	assert.Equal(t, fnHover.Range, rng)
	assert.NotNil(t, fnHover.Node)
	assert.Equal(t, "set_tx_meta", fnHover.Node.Caller.Name)
}

func TestHoverOnFnOriginDocs(t *testing.T) {
	input := `vars { monetary $m = balance(@a, COIN) }`

	rng := RangeOfIndexed(input, "balance", 0)

	program := parser.Parse(input).Value

	hover := analysis.HoverOn(program, rng.Start)

	assert.NotNilf(t, hover, "hover should not be nil")

	fnHover, ok := hover.(*analysis.BuiltinFnHover)
	if !ok {
		t.Fatalf("Expected a BuiltinFnHover")
	}

	assert.Equal(t, fnHover.Range, rng)
	assert.NotNil(t, fnHover.Node)
	assert.Equal(t, "balance", fnHover.Node.Caller.Name)

	res := analysis.CheckProgram(program)
	resolution := res.ResolveBuiltinFn(fnHover.Node.Caller)
	assert.NotNil(t, resolution)

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

		rng := RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		assert.Nil(t, hover)
	})

	t.Run("missing source", func(t *testing.T) {
		input := `
			send [COIN 10] (
				source = <invalidtk>
				destination = @a
			)
		`

		rng := RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		assert.Nil(t, hover)
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

		rng := RangeOfIndexed(input, "<invalidtk>", 0)
		program := parser.Parse(input).Value
		hover := analysis.HoverOn(program, rng.Start)
		assert.Nil(t, hover)
	})
}
