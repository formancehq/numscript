package lsp

import (
	"fmt"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
	"go.lsp.dev/protocol"
)

func CreateVar(diagnostic analysis.UnboundVariable, program parser.Program) protocol.TextEdit {
	declarationLine := fmt.Sprintf("\n  %s $%s\n", diagnostic.Type, diagnostic.Name)

	if program.Vars == nil || len(program.Vars.Declarations) == 0 {
		var rng protocol.Range
		text := fmt.Sprintf("vars {%s}", declarationLine)

		if program.Vars != nil {
			rng = *toLspRange(program.Vars.Range)
		} else {
			text += "\n\n"
			// The grammar puts the vars block after an optional feature
			// declaration and before the first statement
			if len(program.Statements) != 0 {
				start := ParserToLspPosition(program.Statements[0].GetRange().Start)
				rng = protocol.Range{Start: start, End: start}
			}
		}

		return protocol.TextEdit{
			NewText: text,
			Range:   rng,
		}
	}

	lastVarEnd := program.Vars.Declarations[len(program.Vars.Declarations)-1].End

	varsEndPosition := program.Vars.End
	varsEndPosition.Character--

	return protocol.TextEdit{
		NewText: declarationLine,
		Range: protocol.Range{
			Start: ParserToLspPosition(lastVarEnd),
			End:   ParserToLspPosition(varsEndPosition),
		},
	}

}
