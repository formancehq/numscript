package jsonrpc2

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/formancehq/numscript/internal/utils"
)

type MessageStream interface {
	io.Closer
	WriteMessage(obj Message) error
	ReadMessage() (Message, error)
}

type requestHandler func(raw json.RawMessage) any
type notificationHandler func(raw json.RawMessage)

type Server struct {
	currentId            int64
	opened               bool
	stream               MessageStream
	requestsHandlers     map[string]requestHandler
	notificationHandlers map[string]notificationHandler
	pendingRequestMu     sync.RWMutex
	pendingRequests      map[ID](chan Response)
}

// Create a new Server
//
// By default, the server will try write concurrently to the ObjectStream
func NewServer(objStream MessageStream) *Server {
	return &Server{
		opened:               true,
		stream:               objStream,
		requestsHandlers:     map[string]requestHandler{},
		notificationHandlers: map[string]notificationHandler{},
		pendingRequests:      map[ID](chan Response){},
	}
}

// Add a request handler for the given method. Not thread-safe: only add handlers synchronously before server.Listen() call
//
// The handler will be called asynchronously
func HandleRequest[Params any](s *Server, method string, handler func(params Params) any) {
	s.requestsHandlers[method] = func(raw json.RawMessage) any {
		var payload Params
		json.Unmarshal([]byte(raw), &payload)
		return handler(payload)
	}
}

// Add a notification handler for the given method. Not thread-safe: only add handlers synchronously before server.Listen() call
//
// The handler will be called asynchronously
func HandleNotification[Params any](s *Server, method string, handler func(params Params)) {
	s.notificationHandlers[method] = func(raw json.RawMessage) {
		var payload Params
		json.Unmarshal([]byte(raw), &payload)
		handler(payload)
	}
}

// Send a json rpc request and wait for the response. Thread safe.
func SendRequest(s *Server, method string, params any) (any, error) {
	bytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	freshId := NewIntId(atomic.AddInt64(&s.currentId, 1))

	s.stream.WriteMessage(Request{
		ID:     freshId,
		Method: method,
		Params: bytes,
	})

	ch := make(chan Response)

	s.pendingRequestMu.Lock()
	s.pendingRequests[freshId] = ch
	s.pendingRequestMu.Unlock()

	response := <-ch

	s.pendingRequestMu.Lock()
	delete(s.pendingRequests, freshId)
	s.pendingRequestMu.Unlock()

	return response, nil
}

// Send a json rpc request and wait for the message to be sent. Thread safe
func SendNotification(s *Server, method string, params any) error {
	bytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	err = s.stream.WriteMessage(Request{
		Method: method,
		Params: bytes,
	})

	return err
}

func (s *Server) Close() {
	// TODO also stop reading input stream
	// TODO should also close the pendingRequests channels
	s.opened = false
}

func (s *Server) handleRequest(request Request) error {
	if request.IsNotification() {
		handler, ok := s.notificationHandlers[request.Method]
		if !ok {
			return nil
		}

		go handler(request.Params)
	} else {
		handler, ok := s.requestsHandlers[request.Method]
		if !ok {
			return nil
		}

		go func() {
			out := handler(request.Params)

			bytes, _ := json.Marshal(out)

			s.stream.WriteMessage(Response{
				ID:     request.ID,
				Result: bytes,
			})
		}()
	}
	return nil
}

func (s *Server) handleResponse(response Response) error {
	s.pendingRequestMu.RLock()
	request := s.pendingRequests[response.ID]
	s.pendingRequestMu.RUnlock()

	go func() {
		request <- response
	}()

	return nil
}

func (s *Server) handleMessage(msg Message) error {
	switch msg := msg.(type) {
	case Request:
		return s.handleRequest(msg)
	case Response:
		return s.handleResponse(msg)
	default:

		// This should never happen
		return utils.NonExhaustiveMatchPanic[error](fmt.Sprintf("Invalid msg: %#v", msg))
	}
}

// blocks while listening to incoming requests, until Close() is called
//
// returns an error, if any
func (s *Server) Listen() error {
	for s.opened {
		msg, err := s.stream.ReadMessage()
		if err != nil {
			return err
		}

		err = s.handleMessage(msg)
		if err != nil {
			return err
		}
	}

	return nil
}
