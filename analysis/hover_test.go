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

	checkResult := analysis.Check(program)
	assert.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	assert.NotNil(t, resolved)
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

	checkResult := analysis.Check(program)
	assert.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	assert.NotNil(t, resolved)
}

func TestHoverOnDestination(t *testing.T) {
	input := `vars { account $dest }

send [C 10] (
	source = @src
	destination = {
		1/2 to {
			@z
			$dest
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

	checkResult := analysis.Check(program)
	assert.NotNil(t, variableHover.Node)

	resolved := checkResult.ResolveVar(variableHover.Node)
	assert.NotNil(t, resolved)
}
