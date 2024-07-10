package lsp

import (
	"github.com/sourcegraph/jsonrpc2"
)

type numscriptHandlers struct{}

func (*numscriptHandlers) Handle(r jsonrpc2.Request) any {
	switch r.Method {
	case "initialize":
		return InitializeResult{
			Capabilities: ServerCapabilities{},
			// This is ugly. Is there a shortcut?
			ServerInfo: struct {
				Name    string "json:\"name\""
				Version string "json:\"version,omitempty\""
			}{
				Name:    "numscript-ls",
				Version: "0.0.1",
			},
		}

	default:
		// Unhandled method
		// TODO should it panic?
		return nil
	}
}

func NewHandler() Handler {
	return &numscriptHandlers{}
}
