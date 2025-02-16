package lsp

import (
	"fmt"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
)

func CreateVar(diagnostic analysis.UnboundVariable, program parser.Program) TextEdit {
	if len(program.Vars) == 0 {
		return TextEdit{
			NewText: fmt.Sprintf("vars {\n  %s $%s\n}\n\n", diagnostic.Type, diagnostic.Name),
		}
	}

	// firstVar := program.Vars[0]

	return TextEdit{}

}
