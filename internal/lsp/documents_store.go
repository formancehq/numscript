package lsp

import (
	"sync"

	"github.com/formancehq/numscript/internal/lsp/lsp_types"
)

type documentStore[Doc any] struct {
	mu        *sync.RWMutex
	documents map[lsp_types.DocumentURI]Doc
}

func NewDocumentsStore[Doc any]() documentStore[Doc] {
	return documentStore[Doc]{
		mu:        &sync.RWMutex{},
		documents: make(map[lsp_types.DocumentURI]Doc),
	}
}

func (s documentStore[Doc]) Get(uri lsp_types.DocumentURI) (Doc, bool) {
	s.mu.RLock()
	doc, ok := s.documents[uri]
	s.mu.RUnlock()
	return doc, ok
}

func (s documentStore[Doc]) Set(uri lsp_types.DocumentURI, doc Doc) {
	s.mu.Lock()
	s.documents[uri] = doc
	s.mu.Unlock()
}
