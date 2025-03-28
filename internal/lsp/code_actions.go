package lsp

import (
	"fmt"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/lsp/lsp_types"
	"github.com/formancehq/numscript/internal/parser"
)

func CreateVar(diagnostic analysis.UnboundVariable, program parser.Program) lsp_types.TextEdit {
	declarationLine := fmt.Sprintf("\n  %s $%s\n", diagnostic.Type, diagnostic.Name)

	if program.Vars == nil || len(program.Vars.Declarations) == 0 {
		var rng lsp_types.Range
		text := fmt.Sprintf("vars {%s}", declarationLine)

		if program.Vars != nil {
			rng = toLspRange(program.Vars.Range)
		} else {
			text += "\n\n"
		}

		return lsp_types.TextEdit{
			NewText: text,
			Range:   rng,
		}
	}

	lastVarEnd := program.Vars.Declarations[len(program.Vars.Declarations)-1].End

	varsEndPosition := program.Vars.Range.End
	varsEndPosition.Character--

	return lsp_types.TextEdit{
		NewText: declarationLine,
		Range: lsp_types.Range{
			Start: ParserToLspPosition(lastVarEnd),
			End:   ParserToLspPosition(varsEndPosition),
		},
	}

}
