package jsonrpc2_test

import (
	"encoding/json"
	"testing"

	"github.com/formancehq/numscript/internal/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func TestHandleRequest(t *testing.T) {
	type SumParams struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	client := newClient(
		jsonrpc2.NewRequestHandler("sum", func(p SumParams, conn *jsonrpc2.Conn) any {
			return p.X + p.Y
		}),
	)

	raw, err := client.SendRequest("sum", SumParams{X: 100, Y: 42})
	require.Nil(t, err)

	var res int
	json.Unmarshal(raw, &res)
	require.Equal(t, 142, res)
}

func TestHandleNotification(t *testing.T) {
	type NotifParams struct {
		Value string `json:"value"`
	}

	ch := make(chan string)

	client := newClient(
		jsonrpc2.NewNotificationHandler("greet", func(p NotifParams, conn *jsonrpc2.Conn) {
			ch <- p.Value
		}),
	)

	err := client.SendNotification("greet", NotifParams{
		Value: "Hello!",
	})
	require.Nil(t, err)

	require.Equal(t, "Hello!", <-ch)
}

func newClient(serverHandlers ...jsonrpc2.Handler) *jsonrpc2.Conn {
	in := make(chan jsonrpc2.Message)
	out := make(chan jsonrpc2.Message)

	jsonrpc2.NewConn(jsonrpc2.NewChanObjStream(in, out), serverHandlers...)
	client := jsonrpc2.NewConn(jsonrpc2.NewChanObjStream(out, in))

	return client
}
