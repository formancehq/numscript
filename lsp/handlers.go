package lsp

import (
	"encoding/json"
	"os"

	"github.com/sourcegraph/jsonrpc2"
)

type numscriptHandlers struct{}

func (*numscriptHandlers) Handle(r jsonrpc2.Request) any {
	switch r.Method {
	case "initialize":
		return InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync: TextDocumentSyncOptions{
					OpenClose: true,
					Change:    Full,
				},
			},
			// This is ugly. Is there a shortcut?
			ServerInfo: struct {
				Name    string "json:\"name\""
				Version string "json:\"version,omitempty\""
			}{
				Name:    "numscript-ls",
				Version: "0.0.1",
			},
		}

	case "textDocument/didOpen":
		var p DidOpenTextDocumentParams
		json.Unmarshal([]byte(*r.Params), &p)
		os.Stderr.WriteString("OPEN: " + p.TextDocument.Text)

		return nil

	case "textDocument/didChange":
		var p DidChangeTextDocumentParams
		json.Unmarshal([]byte(*r.Params), &p)
		text := p.ContentChanges[len(p.ContentChanges)-1].Text
		os.Stderr.WriteString("CHANGE: " + text)
		return nil

	default:
		// Unhandled method
		// TODO should it panic?
		return nil
	}
}

func NewHandler() Handler {
	return &numscriptHandlers{}
}
