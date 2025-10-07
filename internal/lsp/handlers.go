package lsp

import (
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/jsonrpc2"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
	"go.lsp.dev/protocol"
)

type InMemoryDocument struct {
	Text        string
	CheckResult analysis.CheckResult
}

type State struct {
	documents documentStore[InMemoryDocument]
}

func (state *State) updateDocument(conn *jsonrpc2.Conn, uri protocol.DocumentURI, text string) {
	checkResult := analysis.CheckSource(text)

	state.documents.Set(uri, InMemoryDocument{
		Text:        text,
		CheckResult: checkResult,
	})

	var diagnostics = make([]protocol.Diagnostic, 0)
	for _, diagnostic := range checkResult.Diagnostics {
		diagnostics = append(diagnostics, toLspDiagnostic(diagnostic))
	}

	if err := conn.SendNotification("textDocument/publishDiagnostics", protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}); err != nil {
		log.Printf("lsp: error publishing diagnostics: %v", err)
	}
}

func (state *State) handleHover(params protocol.HoverParams) *protocol.Hover {
	position := fromLspPosition(params.Position)

	doc, ok := state.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	hoverable := analysis.HoverOn(doc.CheckResult.Program, position)

	switch hoverable := hoverable.(type) {
	case *analysis.VariableHover:

		varLit := hoverable.Node
		resolution := doc.CheckResult.ResolveVar(varLit)

		if resolution == nil {
			return nil
		}

		msg := fmt.Sprintf("```numscript\n$%s: %s\n```", varLit.Name, resolution.Type.Name)

		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Value: msg,
				Kind:  "markdown",
			},
			Range: toLspRange(hoverable.Range),
		}
	case *analysis.BuiltinFnHover:
		resolved := doc.CheckResult.ResolveBuiltinFn(hoverable.Node.Caller)
		if resolved == nil {
			return nil
		}

		var msg string
		switch resolved := resolved.(type) {
		case analysis.StatementFnCallResolution:
			params := "("
			for index, arg := range resolved.Params {
				if index != 0 {
					params += ", "
				}
				params += arg
			}
			params += ")"
			msg = fmt.Sprintf("`%s%s`\n\n%s", hoverable.Node.Caller.Name, params, resolved.Docs)
		case analysis.VarOriginFnCallResolution:
			params := "("
			for index, arg := range resolved.Params {
				if index != 0 {
					params += ", "
				}
				params += arg
			}
			params += ")"

			msg = fmt.Sprintf("`%s%s -> %s`\n\n%s", hoverable.Node.Caller.Name, params, resolved.Return, resolved.Docs)
		default:
			utils.NonExhaustiveMatchPanic[any](resolved)
		}

		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Value: msg,
				Kind:  "markdown",
			},
			Range: toLspRange(hoverable.Range),
		}

	default:
		return nil
	}
}

func (state *State) handleGotoDefinition(params protocol.DefinitionParams) *protocol.Location {
	doc, ok := state.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	position := fromLspPosition(params.Position)
	res := analysis.GotoDefinition(doc.CheckResult.Program, position, doc.CheckResult)
	if res == nil {
		return nil
	}

	return &protocol.Location{
		Range: *toLspRange(res.Range),
		URI:   params.TextDocument.URI,
	}
}

func (state *State) handleGetSymbols(params protocol.DocumentSymbolParams) []protocol.DocumentSymbol {
	doc, ok := state.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	syms := doc.CheckResult.GetSymbols()
	lspDocumentSymbols := make([]protocol.DocumentSymbol, len(syms))
	for index, sym := range syms {
		lspDocumentSymbols[index] = protocol.DocumentSymbol{
			Name:           sym.Name,
			Detail:         sym.Detail,
			Kind:           protocol.SymbolKind(sym.Kind),
			Range:          *toLspRange(sym.Range),
			SelectionRange: *toLspRange(sym.SelectionRange),
		}
	}

	return lspDocumentSymbols
}

func (state *State) handleCodeAction(params protocol.CodeActionParams) []protocol.CodeAction {
	doc, ok := state.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	var actions []protocol.CodeAction
	for _, d := range doc.CheckResult.Diagnostics {
		index := slices.IndexFunc(params.Context.Diagnostics, func(lspDiagnostic protocol.Diagnostic) bool {
			id, ok := lspDiagnostic.Data.(float64)
			return ok && int32(id) == d.Id
		})

		var fixedDiagnostics []protocol.Diagnostic
		if index != -1 {
			fixedDiagnostics = append(fixedDiagnostics, params.Context.Diagnostics[index])
		}

		switch kind := d.Kind.(type) {
		case analysis.UnboundVariable:
			actions = append(actions, protocol.CodeAction{
				Title:       "Create variable",
				Kind:        protocol.QuickFix,
				Diagnostics: fixedDiagnostics,
				Edit: &protocol.WorkspaceEdit{
					Changes: map[protocol.DocumentURI][]protocol.TextEdit{
						params.TextDocument.URI: {
							CreateVar(kind, doc.CheckResult.Program),
						},
					},
				},
			})
		}
	}

	return actions
}

func fromLspPosition(p protocol.Position) parser.Position {
	return parser.Position{
		Line:      int(p.Line),
		Character: int(p.Character),
	}
}

func ParserToLspPosition(p parser.Position) protocol.Position {
	return protocol.Position{
		Line:      uint32(p.Line),
		Character: uint32(p.Character),
	}
}

func toLspRange(p parser.Range) *protocol.Range {
	return &protocol.Range{
		Start: ParserToLspPosition(p.Start),
		End:   ParserToLspPosition(p.End),
	}
}

func toLspDiagnostic(d analysis.Diagnostic) protocol.Diagnostic {
	return protocol.Diagnostic{
		Range:    *toLspRange(d.Range),
		Severity: protocol.DiagnosticSeverity(d.Kind.Severity()),
		Message:  d.Kind.Message(),
		Data:     float64(d.Id),
	}
}

var initializeResult protocol.InitializeResult = protocol.InitializeResult{
	Capabilities: protocol.ServerCapabilities{
		TextDocumentSync: protocol.TextDocumentSyncOptions{
			OpenClose: true,
			Change:    protocol.TextDocumentSyncKindFull,
		},
		HoverProvider:          true,
		DefinitionProvider:     true,
		DocumentSymbolProvider: true,
		CodeActionProvider:     true,
	},
	ServerInfo: &protocol.ServerInfo{
		Name:    "numscript-ls",
		Version: "0.0.1",
	},
}

func RunServer() error {
	stream := NewLsObjectStream(os.Stdin, os.Stdout)
	return NewConn(&stream).Wait()
}

func NewConn(objStream jsonrpc2.MessageStream) *jsonrpc2.Conn {
	state := State{
		documents: NewDocumentsStore[InMemoryDocument](),
	}

	return jsonrpc2.NewConn(objStream,
		jsonrpc2.NewRequestHandler("initialize", jsonrpc2.SyncHandling, func(_ any, conn *jsonrpc2.Conn) any {
			return initializeResult
		}),
		jsonrpc2.NewNotificationHandler("textDocument/didOpen", jsonrpc2.SyncHandling, func(p protocol.DidOpenTextDocumentParams, conn *jsonrpc2.Conn) {
			state.updateDocument(conn, p.TextDocument.URI, p.TextDocument.Text)
		}),
		jsonrpc2.NewNotificationHandler("textDocument/didChange", jsonrpc2.SyncHandling, func(p protocol.DidChangeTextDocumentParams, conn *jsonrpc2.Conn) {
			if len(p.ContentChanges) == 0 {
				return
			}
			text := p.ContentChanges[len(p.ContentChanges)-1].Text
			state.updateDocument(conn, p.TextDocument.URI, text)
		}),

		jsonrpc2.NewRequestHandler("textDocument/hover", jsonrpc2.AsyncHandling, func(p protocol.HoverParams, conn *jsonrpc2.Conn) any {
			return state.handleHover(p)
		}),
		jsonrpc2.NewRequestHandler("textDocument/codeAction", jsonrpc2.AsyncHandling, func(p protocol.CodeActionParams, _ *jsonrpc2.Conn) any {
			return state.handleCodeAction(p)
		}),
		jsonrpc2.NewRequestHandler("textDocument/definition", jsonrpc2.AsyncHandling, func(p protocol.DefinitionParams, conn *jsonrpc2.Conn) any {
			return state.handleGotoDefinition(p)
		}),
		jsonrpc2.NewRequestHandler("textDocument/documentSymbol", jsonrpc2.AsyncHandling, func(p protocol.DocumentSymbolParams, conn *jsonrpc2.Conn) any {
			return state.handleGetSymbols(p)
		}),

		jsonrpc2.NewRequestHandler("shutdown", jsonrpc2.SyncHandling, func(_ any, conn *jsonrpc2.Conn) any {
			if err := conn.SendNotification("exit", nil); err != nil {
				log.Printf("lsp: error sending exit notification: %v", err)
			}
			return nil
		}),
	)
}
