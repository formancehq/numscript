//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
)

func check(this js.Value, args []js.Value) any {
	script := args[0].String()
	parsed := parser.Parse(script)
	checkResult := analysis.CheckProgram(parsed.Value)
	ds, err := json.Marshal(checkResult.Diagnostics)
	if err != nil {
		panic(err)
	}

	return string(ds)
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

	ns.Set("check", js.FuncOf(check))
	ns.Set("getVariables", js.FuncOf(getVariables))
	select {}
}
