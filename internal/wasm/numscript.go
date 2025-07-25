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

func serialized(f func(args []js.Value) any) js.Func {
	return js.FuncOf(
		func(this js.Value, args []js.Value) any {
			out := f(args)

			ds, err := json.Marshal(out)
			if err != nil {
				panic(err)
			}

			return string(ds)
		},
	)
}

var currentScript *parser.Program = nil

// The json represantation of both parsing and static analysis diagnostics
type Diagnostic struct {
	Range    parser.Range `json:"range"`
	Message  string       `json:"message"`
	Severity byte         `json:"severity"`
}

func analyse(args []js.Value) any {
	var diagnostics []Diagnostic

	script := args[0].String()
	parsed := parser.Parse(script)

	for _, perr := range parsed.Errors {
		diagnostics = append(diagnostics, Diagnostic{
			Range:    perr.Range,
			Message:  perr.Msg,
			Severity: analysis.ErrorSeverity,
		})
	}

	checkResult := analysis.CheckProgram(parsed.Value)
	for _, rawDiagnostic := range checkResult.Diagnostics {
		diagnostics = append(diagnostics, Diagnostic{
			Range:    rawDiagnostic.Range,
			Message:  rawDiagnostic.Kind.Message(),
			Severity: rawDiagnostic.Kind.Severity(),
		})
	}

	currentScript = &parsed.Value
	return diagnostics
}

type ExecutionResult struct {
	Ok bool `json:"ok"`

	// -- Ok: true

	interpreter.ExecutionResult

	// -- Ok: false

	Error string       `json:"error,omitempty"`
	Range parser.Range `json:"range,omitempty"`
}

type inputOpts struct {
	Script    string                       `json:"script"`
	Variables map[string]string            `json:"variables"`
	Meta      interpreter.AccountsMetadata `json:"metadata"`
	Balances  interpreter.Balances         `json:"balances"`
}

func run(args []js.Value) any {

	inputs := args[0]
	jsonStr := js.Global().Get("JSON").Call("stringify", inputs).String()

	var opts inputOpts

	err := json.Unmarshal([]byte(jsonStr), &opts)
	if err != nil {
		panic(err)
	}

	result, iErr := interpreter.RunProgram(
		context.Background(),
		*currentScript,
		nil,
		nil,
		nil,
	)
	if iErr != nil {
		return ExecutionResult{
			Ok:    false,
			Error: iErr.Error(),
			Range: iErr.GetRange(),
		}
	}

	return ExecutionResult{
		Ok: true,
		ExecutionResult: interpreter.ExecutionResult{
			Postings:         result.Postings,
			Metadata:         result.Metadata,
			AccountsMetadata: result.AccountsMetadata,
		},
	}
}

type VarDeclaration struct {
	Range parser.Range `json:"range"`
	Name  string       `json:"name"`
	Type  string       `json:"type"`
}

func getVariables(this js.Value, args []js.Value) any {
	script := args[0].String()
	parsed := parser.Parse(script)

	var decls []VarDeclaration
	if parsed.Value.Vars == nil {
		return nil
	}

	for _, decl := range parsed.Value.Vars.Declarations {
		if decl.Origin == nil {
			decls = append(decls, VarDeclaration{
				Range: decl.Range,
				Name:  decl.Name.Name,
				Type:  decl.Type.Name,
			})
		}
	}

	ds, err := json.Marshal(decls)
	if err != nil {
		panic(err)
	}

	return string(ds)
}

func main() {
	ns := js.Global().Call("Object")
	js.Global().Set("Numscript", ns)

	ns.Set("analyse", serialized(analyse))
	ns.Set("run", serialized(run))

	ns.Set("getVariables", js.FuncOf(getVariables))
	select {}
}
