package analysis_test

import (
	"numscript/analysis"
	"numscript/parser"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentSymbolsWhenEmpty(t *testing.T) {
	input := `set_tx_meta(@a, "k", 42)`

	program := parser.Parse(input).Value
	checkResult := analysis.CheckProgram(program)

	symbols := checkResult.GetSymbols()
	assert.Empty(t, symbols)
}

func TestDocumentSymbols(t *testing.T) {
	input := `vars {
		monetary $mon
		account $acc
	}`

	program := parser.Parse(input).Value
	checkResult := analysis.CheckProgram(program)

	symbols := checkResult.GetSymbols()

	assert.Len(t, symbols, 2)

	indexMon := slices.IndexFunc(symbols, func(s analysis.DocumentSymbol) bool {
		return s.Name == "mon"
	})
	assert.Equal(t, symbols[indexMon], analysis.DocumentSymbol{
		Name:           "mon",
		Detail:         "monetary",
		Range:          RangeOfIndexed(input, "$mon", 0),
		SelectionRange: RangeOfIndexed(input, "$mon", 0),
		Kind:           analysis.DocumentSymbolVariable,
	})

	indexSrc := slices.IndexFunc(symbols, func(s analysis.DocumentSymbol) bool {
		return s.Name == "acc"
	})
	assert.Equal(t, symbols[indexSrc], analysis.DocumentSymbol{
		Name:           "acc",
		Detail:         "account",
		Range:          RangeOfIndexed(input, "$acc", 0),
		SelectionRange: RangeOfIndexed(input, "$acc", 0),
		Kind:           analysis.DocumentSymbolVariable,
	})

}
