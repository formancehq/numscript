package lsp

import (
	"fmt"
	"os"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/json_rpc"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

type InMemoryDocument struct {
	Text        string
	CheckResult analysis.CheckResult
}

type State struct {
	notify    func(method string, params any)
	documents documentStore[InMemoryDocument]
}

func (state *State) updateDocument(uri DocumentURI, text string) {
	checkResult := analysis.CheckSource(text)

	state.documents.Set(uri, InMemoryDocument{
		Text:        text,
		CheckResult: checkResult,
	})

	var diagnostics []Diagnostic = make([]Diagnostic, 0)
	for _, diagnostic := range checkResult.Diagnostics {
		diagnostics = append(diagnostics, toLspDiagnostic(diagnostic))
	}

	state.notify("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func initialState(notify func(method string, params any)) State {
	return State{
		notify:    notify,
		documents: NewDocumentsStore[InMemoryDocument](),
	}
}

func (state *State) handleHover(params HoverParams) *Hover {
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

		return &Hover{
			Contents: MarkupContent{
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

		return &Hover{
			Contents: MarkupContent{
				Value: msg,
				Kind:  "markdown",
			},
			Range: toLspRange(hoverable.Range),
		}

	default:
		return nil
	}
}

func (state *State) handleGotoDefinition(params DefinitionParams) *Location {
	doc, ok := state.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	position := fromLspPosition(params.Position)
	res := analysis.GotoDefinition(doc.CheckResult.Program, position, doc.CheckResult)
	if res == nil {
		return nil
	}

	return &Location{
		Range: toLspRange(res.Range),
		URI:   params.TextDocument.URI,
	}
}

func (state *State) handleGetSymbols(params DocumentSymbolParams) []DocumentSymbol {
	doc, ok := state.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}

	var lspDocumentSymbols []DocumentSymbol
	for _, sym := range doc.CheckResult.GetSymbols() {
		lspDocumentSymbols = append(lspDocumentSymbols, DocumentSymbol{
			Name:           sym.Name,
			Detail:         sym.Detail,
			Kind:           SymbolKind(sym.Kind),
			Range:          toLspRange(sym.Range),
			SelectionRange: toLspRange(sym.SelectionRange),
		})
	}

	return lspDocumentSymbols
}

func fromLspPosition(p Position) parser.Position {
	return parser.Position{
		Line:      int(p.Line),
		Character: int(p.Character),
	}
}

func toLspPosition(p parser.Position) Position {
	return Position{
		Line:      uint32(p.Line),
		Character: uint32(p.Character),
	}
}

func toLspRange(p parser.Range) Range {
	return Range{
		Start: toLspPosition(p.Start),
		End:   toLspPosition(p.End),
	}
}

func toLspDiagnostic(d analysis.Diagnostic) Diagnostic {
	return Diagnostic{
		Range:    toLspRange(d.Range),
		Severity: DiagnosticSeverity(d.Kind.Severity()),
		Message:  d.Kind.Message(),
	}
}

var initializeResult InitializeResult = InitializeResult{
	Capabilities: ServerCapabilities{
		TextDocumentSync: TextDocumentSyncOptions{
			OpenClose: true,
			Change:    Full,
		},
		HoverProvider:          true,
		DefinitionProvider:     true,
		DocumentSymbolProvider: true,
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

func RunServer() error {
	stream := NewLsObjectStream(os.Stdin, os.Stdout)
	return RunServerWith(&stream)
}

func RunServerWith(objStream json_rpc.ObjectStream) error {
	s := json_rpc.NewServer(objStream)

	state := initialState(func(method string, params any) {
		json_rpc.SendNotification(s, method, params)
	})

	json_rpc.HandleRequest(s, "initialize", func(_ any) any {
		return initializeResult
	})

	json_rpc.HandleNotification(s, "textDocument/didOpen", func(p DidOpenTextDocumentParams) {
		state.updateDocument(p.TextDocument.URI, p.TextDocument.Text)
	})

	json_rpc.HandleNotification(s, "textDocument/didChange", func(p DidChangeTextDocumentParams) {
		text := p.ContentChanges[len(p.ContentChanges)-1].Text
		state.updateDocument(p.TextDocument.URI, text)
	})

	json_rpc.HandleRequest(s, "textDocument/hover", func(p HoverParams) any {
		return state.handleHover(p)
	})

	json_rpc.HandleRequest(s, "textDocument/definition", func(p DefinitionParams) any {
		return state.handleGotoDefinition(p)
	})

	json_rpc.HandleRequest(s, "textDocument/documentSymbol", func(p DocumentSymbolParams) any {
		sm := state.handleGetSymbols(p)
		return sm
	})

	return s.Listen()
}
