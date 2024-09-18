package analysis_test

import (
	"testing"

	"github.com/formancehq/numscript/analysis"
	"github.com/formancehq/numscript/parser"

	"github.com/stretchr/testify/assert"
)

func TestGotoDefinitionOnSendMonetaryVar(t *testing.T) {
	input := `vars { monetary $amt }

send $amt (
	source = @a
	destination = @b
)`

	rng := parser.RangeOfIndexed(input, "$amt", 1)

	program := parser.Parse(input).Value
	checkResult := analysis.CheckProgram(program)

	res := analysis.GotoDefinition(program, rng.Start, checkResult)
	assert.NotNil(t, res)

	assert.Equal(t, &analysis.GotoDefinitionResult{
		Range: parser.RangeOfIndexed(input, "$amt", 0),
	}, res)

}

func TestGotoDefinitionOnNotFound(t *testing.T) {
	input := `send $amt (
	source = @a
	destination = @b
)`

	rng := parser.RangeOfIndexed(input, "$amt", 0)

	program := parser.Parse(input).Value
	checkResult := analysis.CheckProgram(program)

	res := analysis.GotoDefinition(program, rng.Start, checkResult)
	assert.Nil(t, res)
}
