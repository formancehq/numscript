package lsp

import "sync"

type documentStore[Doc any] struct {
	mu        *sync.RWMutex
	documents map[DocumentURI]Doc
}

func NewDocumentsStore[Doc any]() documentStore[Doc] {
	return documentStore[Doc]{
		mu:        &sync.RWMutex{},
		documents: make(map[DocumentURI]Doc),
	}
}

func (s documentStore[Doc]) Get(uri DocumentURI) (Doc, bool) {
	s.mu.RLock()
	doc, ok := s.documents[uri]
	s.mu.RUnlock()
	return doc, ok
}

func (s documentStore[Doc]) Set(uri DocumentURI, doc Doc) {
	s.mu.Lock()
	s.documents[uri] = doc
	s.mu.Unlock()
}
