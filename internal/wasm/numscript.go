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

func main() {
	js.Global().Set("check", js.FuncOf(check))
	select {}
}
