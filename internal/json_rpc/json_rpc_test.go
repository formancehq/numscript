// TODO remove this file in further commits

package json_rpc_test

import (
	"os"

	"github.com/formancehq/numscript/internal/json_rpc"
	"github.com/formancehq/numscript/internal/lsp"
)

// Dead code, just for the sake of example
func Example() {
	objStream := lsp.NewLsObjectStream(os.Stdin, os.Stdout)
	s := json_rpc.NewServer(&objStream)

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_codeAction
	json_rpc.HandleRequest(s, "text/codeAction", func(params lsp.CodeActionParams) any {
		return lsp.CodeAction{}
	})

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#workspace_didChangeConfiguration
	json_rpc.HandleNotification(s, "workspace/didChangeConfiguration", func(params lsp.DidChangeConfigurationParams) {
		// handle
	})

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#workspace_configuration
	response, _ := json_rpc.SendRequest(s, "workspace/configuration", lsp.ConfigurationParams{})
	println(response)

	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_publishDiagnostics
	json_rpc.SendNotification(s, "textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{})

	s.Listen()
}
