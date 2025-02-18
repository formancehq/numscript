package json_rpc

import (
	"encoding/json"

	"github.com/sourcegraph/jsonrpc2"
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

func NewError(code int64, message string) responseError {
	return responseError{
		Code:    code,
		Message: message,
	}
}

type Message interface {
	message()
}

// https://www.jsonrpc.org/specification#response_object
type Response struct {
	// result is the content of the response.
	Result json.RawMessage
	// err is set only if the call failed.
	Error *responseError
	// id of the request this is a response to.
	ID jsonrpc2.ID
}

func (Response) message() {}

var _ Message = (*Response)(nil)

type Request struct {
	// ID of this request, used to tie the Response back to the request.
	// This will be nil for notifications.
	ID *jsonrpc2.ID
	// Method is a string containing the method name to invoke.
	Method string
	// Params is either a struct or an array with the parameters of the method.
	Params json.RawMessage
}

func (Request) message() {}

var _ Message = (*Request)(nil)

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

func UnmarshalMessage(data []byte) (Message, error) {
	combined := messageCombined{}
	err := json.Unmarshal(data, &combined)
	if err != nil {
		return nil, err
	}

	if combined.Method != "" {
		req := Request{
			ID:     &combined.ID,
			Method: combined.Method,
			Params: combined.Params,
		}
		return req, nil
	} else {
		res := Response{
			ID:     combined.ID,
			Result: combined.Result,
			Error:  combined.Error,
		}
		return res, nil
	}
}

type messageCombined struct {
	VersionTag string          `json:"jsonrpc"`
	ID         jsonrpc2.ID     `json:"id,omitempty"`
	Method     string          `json:"method,omitempty"`
	Params     json.RawMessage `json:"params,omitempty"`
	Result     json.RawMessage `json:"result,omitempty"`
	Error      *responseError  `json:"error,omitempty"`
}

// https://www.jsonrpc.org/specification#error_object
type responseError struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}
