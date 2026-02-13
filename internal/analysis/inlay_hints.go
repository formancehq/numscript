package analysis

import (
	"fmt"
	"slices"

	"github.com/formancehq/numscript/internal/parser"
)

type InlayHint struct {
	Position parser.Position
	Label    string
}

func GetInlayHints(
	checkResult CheckResult,
) []InlayHint {
	if checkResult.Program.Vars == nil {
		return nil
	}

	typePrinter := NewTypePrinter()

	var hints []InlayHint
	for _, decl := range checkResult.Program.Vars.Declarations {
		shouldShowAsset := slices.Contains(
			[]string{TypeMonetary, TypeAsset},
			decl.Type.Name,
		)

		if !shouldShowAsset {
			continue
		}

		t := checkResult.GetVarDeclType(decl)

		hints = append(hints, InlayHint{
			Position: decl.Type.End,
			Label:    fmt.Sprintf("<%s>", typePrinter.Print(t)),
		})
	}

	return hints
}
