package analysis_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NotNil(t, res)

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
