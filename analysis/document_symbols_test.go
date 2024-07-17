package analysis_test

import (
	"numscript/analysis"
	"numscript/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentSymbolsWhenEmpty(t *testing.T) {
	input := `set_tx_meta(@a, "k", 42)`

	program := parser.Parse(input).Value
	checkResult := analysis.Check(program)

	symbols := checkResult.GetSymbols()
	assert.Empty(t, symbols)
}

func TestDocumentSymbols(t *testing.T) {
	input := `vars {
		monetary $mon
		account $src
	}`

	program := parser.Parse(input).Value
	checkResult := analysis.Check(program)

	symbols := checkResult.GetSymbols()

	assert.Len(t, symbols, 2)

	assert.Equal(t, symbols[0], analysis.DocumentSymbol{
		Name:   "mon",
		Detail: "monetary",
		Range:  RangeOfIndexed(input, "$mon", 0),
		Kind:   analysis.DocumentSymbolVariable,
	})

}
