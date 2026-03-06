package analysis

import (
	"context"
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

func loadInputsHints(hints *[]InlayHint, program parser.Program, inputs *interpreter.InputsFile) {
	dbg := interpreter.NewDbgBuf()

	// TODO use output, err
	interpreter.RunProgramWithDbg(
		context.Background(),
		program,
		inputs.Variables,
		interpreter.StaticStore{
			Balances: inputs.Balances,
			Meta:     inputs.Meta,
		},
		inputs.GetFeatureFlagsMap(),
		&dbg,
	)

	for _, hint := range dbg.Hints {
		*hints = append(*hints, InlayHint{
			Position:    hint.Position,
			Label:       hint.Label,
			PaddingLeft: true,
		})
	}
}

func GetInlayHints(
	checkResult CheckResult,
	inputs *interpreter.InputsFile,
) []InlayHint {

	var hints []InlayHint

	loadInferenceHints(&hints, checkResult)
	if checkResult.GetErrorsCount() == 0 {
		loadInputsHints(&hints, checkResult.Program, inputs)
	}

	return hints
}
