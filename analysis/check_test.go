package analysis_test

import (
	"numscript/analysis"
	"numscript/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidType(t *testing.T) {
	input := `vars { invalid $my_var }`
	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Character: 7},
			End:   parser.Position{Character: 7 + len("invalid") - 1},
		},
		d1.Range,
	)

	assert.Equal(t,
		&analysis.InvalidType{Name: "invalid"},
		d1.Kind,
	)
}
