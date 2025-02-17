package lsp

import (
	"fmt"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
)

func CreateVar(diagnostic analysis.UnboundVariable, program parser.Program) TextEdit {
	declarationLine := fmt.Sprintf("  %s $%s\n", diagnostic.Type, diagnostic.Name)

	if program.Vars == nil {
		return TextEdit{
			NewText: fmt.Sprintf("vars {\n%s}\n\n", declarationLine),
		}
	}

	varsEndPosition := program.Vars.Range.End

	// firstVar := program.Vars[0]
	editPosition := Position{
		Line:      uint32(varsEndPosition.Line),
		Character: 0,
	}

	return TextEdit{
		NewText: declarationLine,
		Range:   Range{Start: editPosition, End: editPosition},
	}

}
