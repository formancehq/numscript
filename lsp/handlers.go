package lsp

import (
	"encoding/json"
	"fmt"
	"numscript/analysis"
	"numscript/parser"

	"github.com/sourcegraph/jsonrpc2"
)

type InMemoryDocument struct {
	Text        string
	ParseResult parser.ParseResult[parser.Program]
	CheckResult analysis.CheckResult
}

type State struct {
	documents map[DocumentURI]InMemoryDocument
}

func (state *State) updateDocument(uri DocumentURI, text string) {
	parseResult := parser.Parse(text)
	checkResult := analysis.Check(parseResult.Value)

	state.documents[uri] = InMemoryDocument{
		Text:        text,
		ParseResult: parseResult,
		CheckResult: checkResult,
	}

	var diagnostics []Diagnostic = make([]Diagnostic, 0)
	for _, parseErr := range parseResult.Errors {
		diagnostics = append(diagnostics, Diagnostic{
			Message: parseErr.Msg,
			Range:   toLspRange(parseErr.Range),
		})
	}

	for _, diagnostic := range checkResult.Diagnostics {
		diagnostics = append(diagnostics, toLspDiagnostic(diagnostic))
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

func (state *State) handleHover(params HoverParams) *Hover {
	position := fromLspPosition(params.Position)

	doc, ok := state.documents[params.TextDocument.URI]
	if !ok {
		return nil
	}

	hoverable := analysis.HoverOn(doc.ParseResult.Value, position)

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
	default:
		return nil
	}
}

func (state *State) handleGotoDefinition(params DefinitionParams) *Location {
	doc, ok := state.documents[params.TextDocument.URI]
	if !ok {
		return nil
	}

	position := fromLspPosition(params.Position)
	res := analysis.GotoDefinition(doc.ParseResult.Value, position, doc.CheckResult)
	if res == nil {
		return nil
	}

	return &Location{
		Range: toLspRange(res.Range),
		URI:   params.TextDocument.URI,
	}
}

func (state *State) handleCompletion(params CompletionParams) *CompletionList {
	doc, ok := state.documents[params.TextDocument.URI]
	if !ok {
		return nil
	}
	position := fromLspPosition(params.Position)
	completion := analysis.HandleCompletion(doc.ParseResult.Value, position)

	ret := CompletionList{IsIncomplete: false}
	for _, item := range completion.Items {
		ret.Items = append(ret.Items, CompletionItem{
			Label: item.Label,
			Kind:  VariableCompletion,
		})
	}

	return &ret
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
				HoverProvider:      true,
				DefinitionProvider: true,
				CompletionProvider: &CompletionOptions{
					TriggerCharacters: []string{"$"},
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

	case "textDocument/hover":
		var p HoverParams
		json.Unmarshal([]byte(*r.Params), &p)
		return state.handleHover(p)

	case "textDocument/definition":
		var p DefinitionParams
		json.Unmarshal([]byte(*r.Params), &p)
		return state.handleGotoDefinition(p)

	case "textDocument/completion":
		var p CompletionParams
		json.Unmarshal([]byte(*r.Params), &p)
		return state.handleCompletion(p)

	default:
		// Unhandled method
		// TODO should it panic?
		return nil
	}
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
