package lsp

import (
	"fmt"
	"os"
	"slices"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/jsonrpc2"
	"github.com/formancehq/numscript/internal/lsp/lsp_types"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

type InMemoryDocument struct {
	Text        string
	CheckResult analysis.CheckResult
}

type State struct {
	documents documentStore[InMemoryDocument]
}

func (state *State) updateDocument(conn *jsonrpc2.Conn, uri lsp_types.DocumentURI, text string) {
	checkResult := analysis.CheckSource(text)

	state.documents.Set(uri, InMemoryDocument{
		Text:        text,
		CheckResult: checkResult,
	})

	var diagnostics []lsp_types.Diagnostic = make([]lsp_types.Diagnostic, 0)
	for _, diagnostic := range checkResult.Diagnostics {
		diagnostics = append(diagnostics, toLspDiagnostic(diagnostic))
	}

	conn.SendNotification("textDocument/publishDiagnostics", lsp_types.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func (state *State) handleHover(params lsp_types.HoverParams) *lsp_types.Hover {
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

		return &lsp_types.Hover{
			Contents: lsp_types.MarkupContent{
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

		return &lsp_types.Hover{
			Contents: lsp_types.MarkupContent{
				Value: msg,
				Kind:  "markdown",
			},
			Range: toLspRange(hoverable.Range),
		}

	default:
		return nil
	}
}

func (state *State) handleGotoDefinition(params lsp_types.DefinitionParams) *lsp_types.Location {
	doc, ok := state.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	position := fromLspPosition(params.Position)
	res := analysis.GotoDefinition(doc.CheckResult.Program, position, doc.CheckResult)
	if res == nil {
		return nil
	}

	return &lsp_types.Location{
		Range: toLspRange(res.Range),
		URI:   params.TextDocument.URI,
	}
}

func (state *State) handleGetSymbols(params lsp_types.DocumentSymbolParams) []lsp_types.DocumentSymbol {
	doc, ok := state.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	syms := doc.CheckResult.GetSymbols()
	lspDocumentSymbols := make([]lsp_types.DocumentSymbol, len(syms))
	for index, sym := range syms {
		lspDocumentSymbols[index] = lsp_types.DocumentSymbol{
			Name:           sym.Name,
			Detail:         sym.Detail,
			Kind:           lsp_types.SymbolKind(sym.Kind),
			Range:          toLspRange(sym.Range),
			SelectionRange: toLspRange(sym.SelectionRange),
		}
	}

	return lspDocumentSymbols
}

func (state *State) handleCodeAction(params lsp_types.CodeActionParams) []lsp_types.CodeAction {
	doc, ok := state.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	var actions []lsp_types.CodeAction
	for _, d := range doc.CheckResult.Diagnostics {
		index := slices.IndexFunc(params.Context.Diagnostics, func(lspDiagnostic lsp_types.Diagnostic) bool {
			id, ok := lspDiagnostic.Data.(float64)
			return ok && int32(id) == d.Id
		})

		var fixedDiagnostics []lsp_types.Diagnostic
		if index != -1 {
			fixedDiagnostics = append(fixedDiagnostics, params.Context.Diagnostics[index])
		}

		switch kind := d.Kind.(type) {
		case analysis.UnboundVariable:
			actions = append(actions, lsp_types.CodeAction{
				Title:       "Create variable",
				Kind:        lsp_types.QuickFix,
				Diagnostics: fixedDiagnostics,
				Edit: lsp_types.WorkspaceEdit{
					Changes: map[string][]lsp_types.TextEdit{
						string(params.TextDocument.URI): {CreateVar(kind, doc.CheckResult.Program)},
					},
				},
			})
		}
	}

	return actions
}

func fromLspPosition(p lsp_types.Position) parser.Position {
	return parser.Position{
		Line:      int(p.Line),
		Character: int(p.Character),
	}
}

func ParserToLspPosition(p parser.Position) lsp_types.Position {
	return lsp_types.Position{
		Line:      uint32(p.Line),
		Character: uint32(p.Character),
	}
}

func toLspRange(p parser.Range) lsp_types.Range {
	return lsp_types.Range{
		Start: ParserToLspPosition(p.Start),
		End:   ParserToLspPosition(p.End),
	}
}

func toLspDiagnostic(d analysis.Diagnostic) lsp_types.Diagnostic {
	return lsp_types.Diagnostic{
		Range:    toLspRange(d.Range),
		Severity: lsp_types.DiagnosticSeverity(d.Kind.Severity()),
		Message:  d.Kind.Message(),
		Data:     float64(d.Id),
	}
}

var initializeResult lsp_types.InitializeResult = lsp_types.InitializeResult{
	Capabilities: lsp_types.ServerCapabilities{
		TextDocumentSync: lsp_types.TextDocumentSyncOptions{
			OpenClose: true,
			Change:    lsp_types.Full,
		},
		HoverProvider:          true,
		DefinitionProvider:     true,
		DocumentSymbolProvider: true,
		CodeActionProvider:     true,
	},
	// This is ugly. Is there a shortcut?
	ServerInfo: struct {
		Name    string `json:"name"`
		Version string `json:"version,omitempty"`
	}{
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
		jsonrpc2.NewNotificationHandler("textDocument/didOpen", jsonrpc2.SyncHandling, func(p lsp_types.DidOpenTextDocumentParams, conn *jsonrpc2.Conn) {
			state.updateDocument(conn, p.TextDocument.URI, p.TextDocument.Text)
		}),
		jsonrpc2.NewNotificationHandler("textDocument/didChange", jsonrpc2.SyncHandling, func(p lsp_types.DidChangeTextDocumentParams, conn *jsonrpc2.Conn) {
			if len(p.ContentChanges) == 0 {
				return
			}
			text := p.ContentChanges[len(p.ContentChanges)-1].Text
			state.updateDocument(conn, p.TextDocument.URI, text)
		}),

		jsonrpc2.NewRequestHandler("textDocument/hover", jsonrpc2.AsyncHandling, func(p lsp_types.HoverParams, conn *jsonrpc2.Conn) any {
			return state.handleHover(p)
		}),
		jsonrpc2.NewRequestHandler("textDocument/codeAction", jsonrpc2.AsyncHandling, func(p lsp_types.CodeActionParams, _ *jsonrpc2.Conn) any {
			return state.handleCodeAction(p)
		}),
		jsonrpc2.NewRequestHandler("textDocument/definition", jsonrpc2.AsyncHandling, func(p lsp_types.DefinitionParams, conn *jsonrpc2.Conn) any {
			return state.handleGotoDefinition(p)
		}),
		jsonrpc2.NewRequestHandler("textDocument/documentSymbol", jsonrpc2.AsyncHandling, func(p lsp_types.DocumentSymbolParams, conn *jsonrpc2.Conn) any {
			return state.handleGetSymbols(p)
		}),

		jsonrpc2.NewRequestHandler("shutdown", jsonrpc2.SyncHandling, func(_ any, conn *jsonrpc2.Conn) any {
			conn.SendNotification("exit", nil)
			return nil
		}),
	)
}
