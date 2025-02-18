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

	client.OpenFile("example.num", `
		send [COIN 100] (
			source = @world
			destination = $unbound_var
		)
	`)

	snaps.MatchJSON(t, <-client.diagnostics)
}

func TestHover(t *testing.T) {
	client := newTestClient()

	uri, input := client.OpenFile("example.num", `
		vars {
			account $acc
		}

		send [COIN 100] (
			source = $acc
			destination = @dest
		)
	`)

	<-client.diagnostics
	hover, err := client.Hover(uri, *parser.PositionOfIndexed(input, "$acc", 1))
	require.Nil(t, err)
	snaps.MatchJSON(t, hover)
}

// Testing utilities
type TestClient struct {
	conn        *jsonrpc2.Conn
	diagnostics chan json.RawMessage
}

func (c *TestClient) OpenFile(uri string, text string) (string, string) {
	c.conn.SendNotification("textDocument/didOpen", lsp_types.DidOpenTextDocumentParams{
		TextDocument: lsp_types.TextDocumentItem{
			URI:        lsp_types.DocumentURI(uri),
			LanguageID: "numscript",
			Text:       text,
		},
	})
	return uri, text
}

func (c *TestClient) Hover(uri string, position parser.Position) (json.RawMessage, error) {
	return c.conn.SendRequest("textDocument/hover", lsp_types.HoverParams{
		TextDocumentPositionParams: lsp_types.TextDocumentPositionParams{
			TextDocument: lsp_types.TextDocumentIdentifier{
				URI: lsp_types.DocumentURI(uri),
			},
			Position: lsp.ParserToLspPosition(position),
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
