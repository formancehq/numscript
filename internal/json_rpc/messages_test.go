package json_rpc_test

import (
	"encoding/json"
	"testing"

	"github.com/formancehq/numscript/internal/json_rpc"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalRequest(t *testing.T) {

	msg, err := json_rpc.UnmarshalMessage([]byte(`
		{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}
	`))

	require.Nil(t, err)

	require.Equal(t, json_rpc.Request{
		ID:     json_rpc.NewIntId(1),
		Method: "subtract",
		Params: json.RawMessage("[42, 23]"),
	}, msg)
}

func TestUnmarshalRequestStringId(t *testing.T) {
	msg, err := json_rpc.UnmarshalMessage([]byte(`
		{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": "string-id"}
	`))

	require.Nil(t, err)

	require.Equal(t, json_rpc.Request{
		ID:     json_rpc.NewStringId("string-id"),
		Method: "subtract",
		Params: json.RawMessage("[42, 23]"),
	}, msg)
}

func TestUnmarshalNotification(t *testing.T) {
	msg, err := json_rpc.UnmarshalMessage([]byte(`
		{"jsonrpc": "2.0", "method": "update", "params": [1,2,3,4,5]}
	`))

	require.Nil(t, err)

	require.Equal(t, json_rpc.Request{
		Method: "update",
		Params: json.RawMessage("[1,2,3,4,5]"),
	}, msg)
}

func TestUnmarshalNotificationNoParams(t *testing.T) {
	msg, err := json_rpc.UnmarshalMessage([]byte(`
		{"jsonrpc": "2.0", "method": "update"}
	`))

	require.Nil(t, err)

	require.Equal(t, json_rpc.Request{
		Method: "update",
	}, msg)
}

func TestUnmarshalResponse(t *testing.T) {
	msg, err := json_rpc.UnmarshalMessage([]byte(`
		 {"jsonrpc": "2.0", "result": 19, "id": -42}
	`))

	require.Nil(t, err)

	require.Equal(t, json_rpc.Response{
		ID:     json_rpc.NewIntId(-42),
		Result: json.RawMessage("19"),
	}, msg)
}

func TestMarshalRequests(t *testing.T) {

	req := json_rpc.Request{
		ID:     json_rpc.NewIntId(1),
		Method: "subtract",
		Params: json.RawMessage("[42, 23]"),
	}

	bytes, err := json.Marshal(req)
	require.Nil(t, err)

	snaps.MatchJSON(t, bytes)
}
