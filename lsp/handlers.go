package lsp

import (
	"encoding/json"
	"numscript/parser"

	"github.com/sourcegraph/jsonrpc2"
)

type InMemoryDocument struct {
	Text        string
	ParseResult parser.ParseResult[parser.Program]
}

type State struct {
	documents map[DocumentURI]InMemoryDocument
}

func (state *State) updateDocument(uri DocumentURI, text string) {
	parseResult := parser.Parse(text)

	state.documents[uri] = InMemoryDocument{
		Text:        text,
		ParseResult: parseResult,
	}
	var diagnostics []Diagnostic = make([]Diagnostic, 0)
	for _, parseErr := range parseResult.Errors {
		diagnostics = append(diagnostics, Diagnostic{
			Message: parseErr.Msg,
			Range:   ConvertRange(parseErr.Range),
		})
	}
	SendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func InitialState() State {
	return State{
		documents: make(map[DocumentURI]InMemoryDocument),
	}
}

func Handle(r jsonrpc2.Request, state *State) any {
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
		state.updateDocument(p.TextDocument.URI, p.TextDocument.Text)
		return nil

	case "textDocument/didChange":
		var p DidChangeTextDocumentParams
		json.Unmarshal([]byte(*r.Params), &p)
		text := p.ContentChanges[len(p.ContentChanges)-1].Text
		state.updateDocument(p.TextDocument.URI, text)
		return nil

	default:
		// Unhandled method
		// TODO should it panic?
		return nil
	}
}

func ConvertPosition(p parser.Position) Position {
	return Position{
		Line:      p.Line,
		Character: p.Character,
	}
}

func ConvertRange(p parser.Range) Range {
	return Range{
		Start: ConvertPosition(p.Start),
		End:   ConvertPosition(p.End),
	}
}
