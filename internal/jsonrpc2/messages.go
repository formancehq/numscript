package jsonrpc2

import (
	"encoding/json"
	"fmt"
)

// https://www.jsonrpc.org/specification#error_object
var (
	ErrUnknown        = NewError(-32001, "JSON RPC unknown error")
	ErrParse          = NewError(-32700, "JSON RPC parse error")
	ErrInvalidRequest = NewError(-32600, "JSON RPC invalid request")
	ErrMethodNotFound = NewError(-32601, "JSON RPC method not found")
	ErrInvalidParams  = NewError(-32602, "JSON RPC invalid params")
	ErrInternal       = NewError(-32603, "JSON RPC internal error")
)

func NewError(code int64, message string) ResponseError {
	return ResponseError{
		Code:    code,
		Message: message,
	}
}

type ID interface {
	id()
}

// int id
type intId int64

func NewIntId(value int64) ID { return intId(value) }

func (intId) id() {}

var _ ID = (*intId)(nil)

// string id
type stringId string

func NewStringId(value string) ID { return stringId(value) }

func (stringId) id() {}

var _ ID = (*stringId)(nil)

type Message interface {
	message()
}

// https://www.jsonrpc.org/specification#response_object
type Response struct {
	// result is the content of the response.
	Result json.RawMessage
	// err is set only if the call failed.
	Error *ResponseError
	// id of the request this is a response to.
	ID ID
}

func (Response) message() {}

var _ Message = (*Response)(nil)

type Request struct {
	// ID of this request, used to tie the Response back to the request.
	// This will be nil for notifications.
	ID ID
	// Method is a string containing the method name to invoke.
	Method string
	// Params is either a struct or an array with the parameters of the method.
	Params json.RawMessage
}

func (Request) message() {}

func (r Request) IsNotification() bool {
	return r.ID == nil
}

var _ Message = (*Request)(nil)

// Marshaling
const versionTag = "2.0"

func (r Response) MarshalJSON() ([]byte, error) {
	combined := messageCombined{
		VersionTag: versionTag,
		ID:         r.ID,
	}

	if r.Error != nil {
		combined.Error = r.Error
	} else {
		combined.Result = r.Result
	}

	return json.Marshal(combined)
}

func (r Request) MarshalJSON() ([]byte, error) {
	combined := messageCombined{
		VersionTag: versionTag,
		ID:         r.ID,
		Method:     r.Method,
		Params:     r.Params,
	}

	return json.Marshal(combined)
}

func unmarshalID(raw any) (ID, error) {
	switch raw := raw.(type) {
	case float64:
		return NewIntId(int64(raw)), nil
	case int64:
		return NewIntId(raw), nil
	case string:
		return NewStringId(raw), nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid id type: %s", raw)
	}
}

func UnmarshalMessage(data []byte) (Message, error) {
	combined := messageCombined{}
	err := json.Unmarshal(data, &combined)
	if err != nil {
		return nil, err
	}

	id, err := unmarshalID(combined.ID)
	if err != nil {
		return nil, err
	}

	if combined.Method != "" {
		req := Request{
			ID:     id,
			Method: combined.Method,
			Params: combined.Params,
		}
		return req, nil
	} else {
		res := Response{
			ID:     id,
			Result: combined.Result,
			Error:  combined.Error,
		}
		return res, nil
	}
}

type messageCombined struct {
	VersionTag string          `json:"jsonrpc"`
	ID         any             `json:"id,omitempty"`
	Method     string          `json:"method,omitempty"`
	Params     json.RawMessage `json:"params,omitempty"`
	Result     json.RawMessage `json:"result,omitempty"`
	Error      *ResponseError  `json:"error,omitempty"`
}

// https://www.jsonrpc.org/specification#error_object
type ResponseError struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}
