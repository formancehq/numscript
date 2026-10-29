//go:build js && wasm

// Command wasm is the browser-facing entry point of the Numscript WebAssembly
// binding. It is a thin syscall/js shim over internal/wasmbinding: it registers
// a single namespace object on globalThis.__numscript and then blocks forever so
// the Go runtime (and the registered callbacks) stay alive for the lifetime of
// the page.
//
// All the real work — parsing, static analysis, execution, JSON shaping, panic
// recovery — lives in internal/wasmbinding, which is pure Go and unit-tested on
// the host toolchain. Each registered function takes a single string argument
// and returns a JSON string the TS binding parses.
package main

import (
	"syscall/js"

	"github.com/formancehq/numscript/internal/wasmbinding"
)

func stringArg(args []js.Value) (string, bool) {
	if len(args) < 1 || args[0].Type() != js.TypeString {
		return "", false
	}
	return args[0].String(), true
}

func main() {
	ns := js.Global().Get("Object").New()

	// check(source) -> JSON { diagnostics: [...] }
	ns.Set("check", js.FuncOf(func(_ js.Value, args []js.Value) any {
		source, ok := stringArg(args)
		if !ok {
			return `{"diagnostics":[]}`
		}
		return string(wasmbinding.Check(source))
	}))

	// run(inputJSON) -> JSON { ok, value } | { ok:false, error }
	ns.Set("run", js.FuncOf(func(_ js.Value, args []js.Value) any {
		input, ok := stringArg(args)
		if !ok {
			return `{"ok":false,"error":"run expects a JSON string argument"}`
		}
		return string(wasmbinding.Run([]byte(input)))
	}))

	// getNeededVariables(source) -> JSON { ok, variables } | { ok:false, error }
	ns.Set("getNeededVariables", js.FuncOf(func(_ js.Value, args []js.Value) any {
		source, ok := stringArg(args)
		if !ok {
			return `{"ok":false,"error":"getNeededVariables expects a string argument"}`
		}
		return string(wasmbinding.NeededVariables(source))
	}))

	// version() -> the binding's ABI version string
	ns.Set("version", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return wasmbinding.Version
	}))

	js.Global().Set("__numscript", ns)

	// Block forever: if main returned, the Go runtime would exit and every
	// registered js.Func would become invalid.
	select {}
}
