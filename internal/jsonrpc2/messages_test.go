package jsonrpc2_test

import (
	"encoding/json"
	"testing"

	"github.com/formancehq/numscript/internal/jsonrpc2"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalRequest(t *testing.T) {

	msg, err := jsonrpc2.UnmarshalMessage([]byte(`
		{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}
	`))

	require.Nil(t, err)

	require.Equal(t, jsonrpc2.Request{
		ID:     jsonrpc2.NewIntId(1),
		Method: "subtract",
		Params: json.RawMessage("[42, 23]"),
	}, msg)
}

func TestUnmarshalRequestStringId(t *testing.T) {
	msg, err := jsonrpc2.UnmarshalMessage([]byte(`
		{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": "string-id"}
	`))

	require.Nil(t, err)

	require.Equal(t, jsonrpc2.Request{
		ID:     jsonrpc2.NewStringId("string-id"),
		Method: "subtract",
		Params: json.RawMessage("[42, 23]"),
	}, msg)
}

func TestUnmarshalNotification(t *testing.T) {
	msg, err := jsonrpc2.UnmarshalMessage([]byte(`
		{"jsonrpc": "2.0", "method": "update", "params": [1,2,3,4,5]}
	`))

	require.Nil(t, err)

	require.Equal(t, jsonrpc2.Request{
		Method: "update",
		Params: json.RawMessage("[1,2,3,4,5]"),
	}, msg)
}

func TestUnmarshalNotificationNoParams(t *testing.T) {
	msg, err := jsonrpc2.UnmarshalMessage([]byte(`
		{"jsonrpc": "2.0", "method": "update"}
	`))

	require.Nil(t, err)

	require.Equal(t, jsonrpc2.Request{
		Method: "update",
	}, msg)
}

func TestUnmarshalResponse(t *testing.T) {
	msg, err := jsonrpc2.UnmarshalMessage([]byte(`
		 {"jsonrpc": "2.0", "result": 19, "id": -42}
	`))

	require.Nil(t, err)

	require.Equal(t, jsonrpc2.Response{
		ID:     jsonrpc2.NewIntId(-42),
		Result: json.RawMessage("19"),
	}, msg)
}

func TestMarshalRequests(t *testing.T) {

	req := jsonrpc2.Request{
		ID:     jsonrpc2.NewIntId(1),
		Method: "subtract",
		Params: json.RawMessage("[42, 23]"),
	}

	bytes, err := json.Marshal(req)
	require.Nil(t, err)

	snaps.MatchJSON(t, bytes)
}

func TestMarshalNotifications(t *testing.T) {
	req := jsonrpc2.Request{
		Method: "updateCounter",
		Params: json.RawMessage("42"),
	}

	bytes, err := json.Marshal(req)
	require.Nil(t, err)

	snaps.MatchJSON(t, bytes)
}

func TestMarshalOkResponse(t *testing.T) {
	req := jsonrpc2.Response{
		ID:     jsonrpc2.NewIntId(42),
		Result: json.RawMessage(`{"x": 42}`),
	}

	bytes, err := json.Marshal(req)
	require.Nil(t, err)

	snaps.MatchJSON(t, bytes)
}

func TestMarshalErrResponse(t *testing.T) {
	req := jsonrpc2.Response{
		ID:    jsonrpc2.NewIntId(42),
		Error: &jsonrpc2.ErrInvalidRequest,
	}

	bytes, err := json.Marshal(req)
	require.Nil(t, err)

	snaps.MatchJSON(t, bytes)
}
