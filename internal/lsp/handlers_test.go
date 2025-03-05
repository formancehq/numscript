package lsp_test

import (
	"encoding/json"
	"testing"

	"github.com/formancehq/numscript/internal/jsonrpc2"
	"github.com/formancehq/numscript/internal/lsp"
	"github.com/formancehq/numscript/internal/lsp/lsp_types"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestDiagnostics(t *testing.T) {
	client := newTestClient()

	_, _, diagnostics := client.OpenFile("example.num", `
		send [COIN 100] (
			source = @world
			destination = $unbound_var
		)
	`)

	snaps.MatchJSON(t, diagnostics)
}

func TestHoverVariable(t *testing.T) {
	client := newTestClient()

	doc, input, _ := client.OpenFile("example.num", `
		vars {
			account $acc
		}

		send [COIN 100] (
			source = $acc
			destination = @dest
		)
	`)

	hover, err := client.Hover(doc, *parser.PositionOfIndexed(input, "$acc", 1))
	require.Nil(t, err)
	snaps.MatchJSON(t, hover)
}

func TestHoverFnOrigin(t *testing.T) {
	client := newTestClient()

	doc, input, _ := client.OpenFile("example.num", `
		vars {
			monetary $acc = balance(@acc, USD/2)
		}
	`)

	hover, err := client.Hover(doc, *parser.PositionOfIndexed(input, "balance", 0))
	require.Nil(t, err)
	snaps.MatchJSON(t, hover)
}

func TestHoverFnStatement(t *testing.T) {
	client := newTestClient()

	doc, input, _ := client.OpenFile("example.num", `
		set_tx_meta(@acc, "k", 1 + 2)
	`)

	hover, err := client.Hover(doc, *parser.PositionOfIndexed(input, "set_tx_meta", 0))
	require.Nil(t, err)
	snaps.MatchJSON(t, hover)
}

func TestGetSymbols(t *testing.T) {
	client := newTestClient()

	doc, _, _ := client.OpenFile("example.num", `
		vars {
			account $acc
			monetary $mon
		}
	`)

	raw, err := client.GetSymbols(doc)
	require.Nil(t, err)
	snaps.MatchJSON(t, raw)
}

func TestGotoDef(t *testing.T) {
	client := newTestClient()

	doc, input, _ := client.OpenFile("example.num", `
		vars {
			account $acc
		}

		send [USD/2 100] (
			source = $acc
			destination = @dest
		)
 	`)

	raw, err := client.GotoDefinition(doc, *parser.PositionOfIndexed(input, "$acc", 1))
	require.Nil(t, err)
	snaps.MatchJSON(t, raw)
}

// Testing utilities
type TestClient struct {
	conn        *jsonrpc2.Conn
	diagnostics chan json.RawMessage
}

func (c *TestClient) OpenFile(uri string, text string) (lsp_types.TextDocumentIdentifier, string, json.RawMessage) {
	c.conn.SendNotification("textDocument/didOpen", lsp_types.DidOpenTextDocumentParams{
		TextDocument: lsp_types.TextDocumentItem{
			URI:        lsp_types.DocumentURI(uri),
			LanguageID: "numscript",
			Text:       text,
		},
	})

	docIdent := lsp_types.TextDocumentIdentifier{
		URI: lsp_types.DocumentURI(uri),
	}

	return docIdent, text, <-c.diagnostics
}

func (c *TestClient) Hover(doc lsp_types.TextDocumentIdentifier, position parser.Position) (json.RawMessage, *jsonrpc2.ResponseError) {
	return c.conn.SendRequest("textDocument/hover", lsp_types.HoverParams{
		TextDocumentPositionParams: lsp_types.TextDocumentPositionParams{
			TextDocument: doc,
			Position:     lsp.ParserToLspPosition(position),
		},
	})
}

func (c *TestClient) GetSymbols(doc lsp_types.TextDocumentIdentifier) (json.RawMessage, *jsonrpc2.ResponseError) {
	return c.conn.SendRequest("textDocument/documentSymbol", lsp_types.DocumentSymbolParams{
		TextDocument: doc,
	})
}

func (c *TestClient) GotoDefinition(doc lsp_types.TextDocumentIdentifier, position parser.Position) (json.RawMessage, *jsonrpc2.ResponseError) {
	return c.conn.SendRequest("textDocument/definition", lsp_types.DefinitionParams{
		TextDocumentPositionParams: lsp_types.TextDocumentPositionParams{
			TextDocument: doc,
			Position:     lsp.ParserToLspPosition(position),
		},
	})
}

func newTestClient() TestClient {
	in := make(chan jsonrpc2.Message)
	out := make(chan jsonrpc2.Message)

	lsp.NewConn(
		jsonrpc2.NewChanObjStream(in, out),
	)

	diagnostics := make(chan json.RawMessage)
	conn := jsonrpc2.NewConn(
		// note 'out' and 'in' are swapped for the client
		jsonrpc2.NewChanObjStream(out, in),
		jsonrpc2.NewNotificationHandler("textDocument/publishDiagnostics", func(p json.RawMessage, conn *jsonrpc2.Conn) {
			diagnostics <- p
		}),
	)

	return TestClient{
		diagnostics: diagnostics,
		conn:        conn,
	}
}
