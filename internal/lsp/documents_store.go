package lsp

import (
	"sync"

	"go.lsp.dev/protocol"
)

type documentStore[Doc any] struct {
	mu        *sync.RWMutex
	documents map[protocol.DocumentURI]Doc
}

func NewDocumentsStore[Doc any]() documentStore[Doc] {
	return documentStore[Doc]{
		mu:        &sync.RWMutex{},
		documents: make(map[protocol.DocumentURI]Doc),
	}
}

func (s documentStore[Doc]) Get(uri protocol.DocumentURI) (Doc, bool) {
	s.mu.RLock()
	doc, ok := s.documents[uri]
	s.mu.RUnlock()
	return doc, ok
}

// Atomically fill the store with the given action when elem is not present
func (s documentStore[Doc]) GetOrPut(uri protocol.DocumentURI, getDefault func() Doc) Doc {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, ok := s.documents[uri]
	if ok {
		return doc
	}

	doc = getDefault()
	s.documents[uri] = doc
	return doc
}

func (s documentStore[Doc]) Set(uri protocol.DocumentURI, doc Doc) {
	s.mu.Lock()
	s.documents[uri] = doc
	s.mu.Unlock()
}

// Noop if uri does not exist
func (s documentStore[Doc]) Update(uri protocol.DocumentURI, update func(doc *Doc)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, ok := s.documents[uri]
	if !ok {
		return false
	}

	update(&doc)
	s.documents[uri] = doc
	return true
}

func (s documentStore[Doc]) Delete(uri protocol.DocumentURI) {
	s.mu.Lock()
	delete(s.documents, uri)
	s.mu.Unlock()
}
