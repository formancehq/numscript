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

type requestHandler func(id ID, raw json.RawMessage)
type notificationHandler func(raw json.RawMessage)

type Conn struct {
	listenErr            chan error
	currentId            int64
	opened               bool
	stream               MessageStream
	requestsHandlers     map[string]requestHandler
	notificationHandlers map[string]notificationHandler
	pendingRequestMu     sync.RWMutex
	pendingRequests      map[ID](chan Response)
}

type Handler struct {
	register func(*Conn)
}

// Create a request handler for the given method
//
// The handler will be called asynchronously
func NewRequestHandler[Params any](method string, handler func(params Params, conn *Conn) any) Handler {
	return Handler{
		register: func(conn *Conn) {
			conn.requestsHandlers[method] = func(id ID, raw json.RawMessage) {
				var payload Params
				if raw != nil {
					err := json.Unmarshal([]byte(raw), &payload)
					if err != nil {
						conn.stream.WriteMessage(Response{
							ID:    id,
							Error: &ErrInvalidParams,
						})
						return
					}
				}

				out := handler(payload, conn)

				bytes, _ := json.Marshal(out)

				conn.stream.WriteMessage(Response{
					ID:     id,
					Result: bytes,
				})

			}
		},
	}
}

// Create a notification handler for the given method
//
// The handler will be called asynchronously
func NewNotificationHandler[Params any](method string, handler func(params Params, conn *Conn)) Handler {
	return Handler{
		register: func(conn *Conn) {
			conn.notificationHandlers[method] = func(raw json.RawMessage) {
				var payload Params
				json.Unmarshal([]byte(raw), &payload)
				handler(payload, conn)
			}
		},
	}
}

// Starts listening asynchronously to the MessageStream and returns the connection.
//
// By default, the server will try write concurrently to the ObjectStream
func NewConn(objStream MessageStream, handlers ...Handler) *Conn {
	conn := Conn{
		listenErr:            make(chan error),
		opened:               true,
		stream:               objStream,
		requestsHandlers:     map[string]requestHandler{},
		notificationHandlers: map[string]notificationHandler{},
		pendingRequests:      map[ID](chan Response){},
	}

	// Register the handlers BEFORE listening to the messages stream
	for _, handler := range handlers {
		handler.register(&conn)
	}

	// listen to the incoming messages
	go func() {
		for conn.opened {
			msg, err := conn.stream.ReadMessage()
			if err != nil {
				conn.listenErr <- err
				return
			}

			err = conn.handleMessage(msg)
			if err != nil {
				conn.listenErr <- err
				return
			}
		}
	}()

	return &conn
}

// Send a json rpc request and wait for the response. Thread safe.
//
// Will panick whenever the params object fails json.Marshal-ing
func (s *Conn) SendRequest(method string, params any) (json.RawMessage, *ResponseError) {
	bytes, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}

	freshId := NewIntId(atomic.AddInt64(&s.currentId, 1))

	ch := make(chan Response)

	s.pendingRequestMu.Lock()
	s.pendingRequests[freshId] = ch
	s.pendingRequestMu.Unlock()

	go s.stream.WriteMessage(Request{
		ID:     freshId,
		Method: method,
		Params: bytes,
	})

	response := <-ch

	s.pendingRequestMu.Lock()
	delete(s.pendingRequests, freshId)
	s.pendingRequestMu.Unlock()

	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// Send a json rpc request and wait for the message to be sent. Thread safe
//
// Will panick whenever the params object fails json.Marshal-ing
func (s *Conn) SendNotification(method string, params any) error {
	bytes, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}

	return s.stream.WriteMessage(Request{
		Method: method,
		Params: bytes,
	})
}

func (s *Conn) Close() {
	s.stream.Close()
	for _, ch := range s.pendingRequests {
		close(ch)
	}
	s.opened = false
}

func (s *Conn) handleRequest(request Request) error {
	if request.IsNotification() {
		handler, ok := s.notificationHandlers[request.Method]
		if !ok {
			return nil
		}

		go handler(request.Params)
	} else {
		handler, ok := s.requestsHandlers[request.Method]
		if !ok {
			go s.stream.WriteMessage(Response{
				ID:    request.ID,
				Error: &ErrMethodNotFound,
			})
			return nil
		}

		go handler(request.ID, request.Params)
	}
	return nil
}

func (s *Conn) handleResponse(response Response) error {
	s.pendingRequestMu.RLock()
	request, ok := s.pendingRequests[response.ID]
	s.pendingRequestMu.RUnlock()

	if !ok {
		return nil
	}

	go func() {
		request <- response
	}()

	return nil
}

func (s *Conn) handleMessage(msg Message) error {
	switch msg := msg.(type) {
	case Request:
		return s.handleRequest(msg)
	case Response:
		return s.handleResponse(msg)
	case nil:
		return nil
	default:

		// This should never happen
		return utils.NonExhaustiveMatchPanic[error](fmt.Sprintf("Invalid msg: %#v", msg))
	}
}

// blocks until connections is closed, returning its error (or nil)
func (c *Conn) Wait() error {
	return <-c.listenErr
}
