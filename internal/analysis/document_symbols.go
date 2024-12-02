package analysis

import (
	"github.com/PagoPlus/numscriptex/internal/parser"
)

type DocumentSymbolKind = float64

// !important! keep in sync with
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_documentSymbol
const (
	DocumentSymbolVariable DocumentSymbolKind = 13
)

type DocumentSymbol struct {
	Name           string
	Detail         string
	Range          parser.Range
	SelectionRange parser.Range
	Kind           DocumentSymbolKind
}

// Note: Results are not sorted
func (r *CheckResult) GetSymbols() []DocumentSymbol {
	var symbols []DocumentSymbol
	for k, v := range r.declaredVars {
		symbols = append(symbols, DocumentSymbol{
			Name:           k,
			Kind:           DocumentSymbolVariable,
			Detail:         v.Type.Name,
			Range:          v.Name.Range,
			SelectionRange: v.Name.Range,
		})

	}

	return symbols
}
