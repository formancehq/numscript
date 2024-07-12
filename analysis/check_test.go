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
			End:   parser.Position{Character: 7 + len("invalid")},
		},
		d1.Range,
	)

	assert.Equal(t,
		&analysis.InvalidType{Name: "invalid"},
		d1.Kind,
	)
}

func TestValidType(t *testing.T) {
	input := `vars { monetary $my_var }`
	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 0)
}

func TestDuplicateVariable(t *testing.T) {
	input := `vars {
  asset $x
  account $y
  portion $x
}`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Line: 3, Character: 10},
			End:   parser.Position{Line: 3, Character: 10 + len("$x")},
		},
		d1.Range,
	)

	assert.Equal(t,
		&analysis.DuplicateVariable{Name: "x"},
		d1.Kind,
	)
}

func TestUnboundVarInSource(t *testing.T) {
	input := `send [C 1] (
  source = { max [C 1] from $unbound_var }
  destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Line: 1, Character: 28},
			End:   parser.Position{Line: 1, Character: 28 + len("$unbound_var")},
		},
		d1.Range,
	)

	assert.Equal(t,
		&analysis.UnboundVariable{Name: "unbound_var"},
		d1.Kind,
	)
}

func TestUnboundMany(t *testing.T) {
	input := `send [C 1] (
  	source = {
  		1/3 from $unbound1
		2/3 from $unbound2
	}
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 2)
}

func TestUnboundCurrenciesVars(t *testing.T) {
	input := `send $unbound1 (
  	source = {
		max $unbound2 from @a
	}
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 2)
}

// TODO unbound vars in declr
