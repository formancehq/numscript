package analysis

import (
	"fmt"
	"slices"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
)

type InlayHint struct {
	Position    parser.Position
	Label       string
	PaddingLeft bool
}

func loadInferenceHints(hints *[]InlayHint, checkResult CheckResult) {
	if checkResult.Program.Vars == nil {
		return
	}

	typePrinter := NewTypePrinter()

	for _, decl := range checkResult.Program.Vars.Declarations {
		shouldShowAsset := slices.Contains(
			[]string{TypeMonetary, TypeAsset},
			decl.Type.Name,
		)

		if !shouldShowAsset {
			continue
		}

		t := checkResult.GetVarDeclType(decl)

		*hints = append(*hints, InlayHint{
			Position: decl.Type.End,
			Label:    fmt.Sprintf("<%s>", typePrinter.Print(t)),
		})
	}
}

func GetInlayHints(
	interpreterHints []interpreter.DbgHint,
	checkResult CheckResult,
	inputs *interpreter.InputsFile,
) []InlayHint {
	var hints []InlayHint

	loadInferenceHints(&hints, checkResult)
	for _, hint := range interpreterHints {
		hints = append(hints, InlayHint{
			Position:    hint.Position,
			Label:       hint.Label,
			PaddingLeft: true,
		})
	}

	return hints
}
