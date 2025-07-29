//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
)

func getDiagnostics(parsed parser.ParseResult) []any {
	var diagnostics []any

	for _, perr := range parsed.Errors {
		diagnostics = append(diagnostics, map[string]any{
			"range":    castRange(perr.Range),
			"message":  perr.Msg,
			"severity": analysis.ErrorSeverity,
		})
	}
	checkResult := analysis.CheckProgram(parsed.Value)
	for _, rawDiagnostic := range checkResult.Diagnostics {
		diagnostics = append(diagnostics, map[string]any{
			"range":    castRange(rawDiagnostic.Range),
			"message":  rawDiagnostic.Kind.Message(),
			"severity": rawDiagnostic.Kind.Severity(),
		})
	}

	return diagnostics
}

func getVariables(parsed parser.ParseResult) any {
	var decls []any

	if parsed.Value.Vars == nil {
		return []any{}
	}

	for _, decl := range parsed.Value.Vars.Declarations {
		if decl.Name == nil || decl.Type == nil {
			continue
		}

		if decl.Origin == nil {
			decls = append(decls, map[string]any{
				"range": castRange(decl.Range),
				"name":  decl.Name.Name,
				"type":  decl.Type.Name,
			})
		}
	}

	return decls
}

func getAnalysisOutput(parsed parser.ParseResult, inputs inputOpts) any {
	diagnostics := getDiagnostics(parsed)

	if len(diagnostics) != 0 {
		return map[string]any{
			"ok":          false,
			"errorType":   "analysis",
			"diagnostics": diagnostics,
		}
	}

	featureFlags := map[string]struct{}{}
	for _, v := range inputs.FeatureFlags {
		featureFlags[v] = struct{}{}
	}

	result, iErr := interpreter.RunProgram(
		context.Background(),
		parsed.Value,
		inputs.Variables,
		interpreter.StaticStore{
			Balances: inputs.Balances,
			Meta:     inputs.Meta,
		},
		featureFlags,
	)

	if iErr != nil {
		return map[string]any{
			"ok":           false,
			"errorType":    "execution",
			"errorMessage": iErr.Error(),
			"range":        castRange(iErr.GetRange()),
		}
	}

	return map[string]any{
		"ok":              true,
		"executionResult": castAnything(*result),
	}
}

func analyse(this js.Value, args []js.Value) any {
	script := args[0].String()
	parsed := parser.Parse(script)

	// TODO we'll want to pass them as obj
	inputs := args[1].String()
	var opts inputOpts
	err := json.Unmarshal([]byte(inputs), &opts)
	if err != nil {
		panic(err)
	}

	return map[string]any{
		"declaredVars": getVariables(parsed),
		"output":       getAnalysisOutput(parsed, opts),
	}
}

type inputOpts struct {
	Variables    map[string]string            `json:"variables"`
	Meta         interpreter.AccountsMetadata `json:"metadata"`
	Balances     interpreter.Balances         `json:"balances"`
	FeatureFlags []string                     `json:"featureFlags,omitempty"`
}

func main() {
	ns := js.Global().Call("Object")
	js.Global().Set("Numscript", ns)
	ns.Set("analyse", js.FuncOf(analyse))
	select {}
}

// Conversion boilerplate
func castPosition(pos parser.Position) any {
	return map[string]any{
		"character": pos.Character,
		"line":      pos.Line,
	}
}

func castRange(rng parser.Range) any {
	return map[string]any{
		"start": castPosition(rng.Start),
		"end":   castPosition(rng.End),
	}
}

func castAnything(struct_ any) any {
	// TODO manually cast without marshaling
	ps, err := json.Marshal(struct_)
	if err != nil {
		panic(err)
	}
	var out any
	json.Unmarshal(ps, &out)
	return out
}
