package lsp_types_extra

import "go.lsp.dev/protocol"

type InlayHintParams struct {
	TextDocument protocol.TextDocumentIdentifier `json:"textDocument"`
	Range        protocol.Range                  `json:"range"`
}

type InlayHintKind string

const (
	InlayHintKindType  InlayHintKind = "type"
	InlayHintKindParam InlayHintKind = "parameter"
)

type InlayHint struct {
	Position  protocol.Position   `json:"position"`
	Label     string              `json:"label"`
	Kind      *InlayHintKind      `json:"kind,omitempty"`
	Tooltip   *string             `json:"tooltip,omitempty"`
	TextEdits []protocol.TextEdit `json:"textEdits,omitempty"`
}

type InlayHintOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

type InlayHintRegistrationOptions struct {
	protocol.TextDocumentRegistrationOptions
	InlayHintOptions
}

type InlayHintParamsResult []InlayHint

// -- Initialize
type ServerCapabilities struct {
	protocol.ServerCapabilities
	InlayHintProvider bool `json:"inlayHintProvider,omitempty"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities   `json:"capabilities"`
	ServerInfo   *protocol.ServerInfo `json:"serverInfo,omitempty"`
}
