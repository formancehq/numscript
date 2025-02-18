package json_rpc

import (
	"encoding/json"
	"io"
	"sync"
	"sync/atomic"

	"github.com/sourcegraph/jsonrpc2"
)

type ObjectStream interface {
	io.Closer
	WriteObject(obj any) error
	ReadObject() (*json.RawMessage, error)
}

type requestHandler func(raw json.RawMessage) any
type notificationHandler func(raw json.RawMessage)

type Server struct {
	currentId            uint64
	opened               bool
	stream               ObjectStream
	requestsHandlers     map[string]requestHandler
	notificationHandlers map[string]notificationHandler
	pendingRequestMu     sync.RWMutex
	pendingRequests      map[uint64](chan jsonrpc2.Response)
}

// Create a new Server
//
// By default, the server will try write concurrently to the ObjectStream
func NewServer(objStream ObjectStream) *Server {
	return &Server{
		opened:               true,
		stream:               objStream,
		requestsHandlers:     map[string]requestHandler{},
		notificationHandlers: map[string]notificationHandler{},
		pendingRequests:      map[uint64](chan jsonrpc2.Response){},
	}
}

// Add a request handler for the given method
//
// The handler will be called asynchronously
func HandleRequest[Params any](s *Server, method string, handler func(params Params) any) {
	s.requestsHandlers[method] = func(raw json.RawMessage) any {
		var payload Params
		json.Unmarshal([]byte(raw), &payload)
		return handler(payload)
	}
}

// Add a notification handler for the given method
//
// The handler will be called asynchronously
func HandleNotification[Params any](s *Server, method string, handler func(params Params)) {
	s.notificationHandlers[method] = func(raw json.RawMessage) {
		var payload Params
		json.Unmarshal([]byte(raw), &payload)
		handler(payload)
	}
}

// Send a json rpc request and wait for the response
func SendRequest(s *Server, method string, params any) (any, error) {
	bytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	rawParams := json.RawMessage(bytes)

	freshId := atomic.AddUint64(&s.currentId, 1)

	s.stream.WriteObject(jsonrpc2.Request{
		ID:     jsonrpc2.ID{Num: freshId},
		Notif:  false,
		Method: method,
		Params: &rawParams,
	})

	ch := make(chan jsonrpc2.Response)

	s.pendingRequestMu.Lock()
	s.pendingRequests[freshId] = ch
	s.pendingRequestMu.Unlock()

	response := <-ch

	s.pendingRequestMu.Lock()
	delete(s.pendingRequests, freshId)
	s.pendingRequestMu.Unlock()

	return response, nil
}

// Send a json rpc request and wait for the message to be sent
func SendNotification(s *Server, method string, params any) error {
	bytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	rawParams := json.RawMessage(bytes)
	err = s.stream.WriteObject(jsonrpc2.Request{
		Notif:  true,
		Method: method,
		Params: &rawParams,
	})

	return err
}

func (s *Server) Close() {
	// TODO also stop reading input stream
	// TODO should also close the pendingRequests channels
	s.opened = false
}

func (s *Server) handleRequest(request jsonrpc2.Request) error {
	if request.Notif {
		handler, ok := s.notificationHandlers[request.Method]
		if !ok {
			return nil
		}

		go handler(*request.Params)
	} else {
		handler, ok := s.requestsHandlers[request.Method]
		if !ok {
			return nil
		}

		go func() {
			out := handler(*request.Params)

			bytes, _ := json.Marshal(out)

			var jsonRaw json.RawMessage = bytes

			s.stream.WriteObject(jsonrpc2.Response{
				ID:     request.ID,
				Result: &jsonRaw,
			})
		}()
	}
	return nil
}

func (s *Server) handleResponse(response jsonrpc2.Response) error {
	s.pendingRequestMu.RLock()
	request := s.pendingRequests[response.ID.Num]
	s.pendingRequestMu.RUnlock()

	go func() {
		request <- response
	}()

	return nil
}

func (s *Server) handleRawMessage(raw json.RawMessage) error {
	var request jsonrpc2.Request
	err := json.Unmarshal([]byte(raw), &request)

	if err == nil && request.Method != "" {
		return s.handleRequest(request)
	}

	var response jsonrpc2.Response
	err = json.Unmarshal([]byte(raw), &response)
	if err == nil {
		return s.handleResponse(response)
	}

	return nil
}

// blocks while listening to incoming requests, until Close() is called
//
// returns an error, if any
func (s *Server) Listen() error {
	for s.opened {
		raw, err := s.stream.ReadObject()
		if err != nil {
			return err
		}

		err = s.handleRawMessage(*raw)
		if err != nil {
			return err
		}
	}

	return nil
}
