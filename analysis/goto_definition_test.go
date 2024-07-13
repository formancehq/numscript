package analysis_test

import (
	"numscript/analysis"
	"numscript/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGotoDefinitionOnSendMonetaryVar(t *testing.T) {
	input := `vars { monetary $amt }

send $amt (
	source = @a
	destination = @b
)`

	rng := RangeOfIndexed(input, "$amt", 1)

	program := parser.Parse(input).Value
	checkResult := analysis.Check(program)

	res := analysis.GotoDefinition(program, rng.Start, checkResult)
	assert.NotNil(t, res)

	assert.Equal(t, &analysis.GotoDefinitionResult{
		Range: RangeOfIndexed(input, "$amt", 0),
	}, res)

}

func TestGotoDefinitionOnNotFound(t *testing.T) {
	input := `send $amt (
	source = @a
	destination = @b
)`

	rng := RangeOfIndexed(input, "$amt", 0)

	program := parser.Parse(input).Value
	checkResult := analysis.Check(program)

	res := analysis.GotoDefinition(program, rng.Start, checkResult)
	assert.Nil(t, res)
}
