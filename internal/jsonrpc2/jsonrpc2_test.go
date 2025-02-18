package jsonrpc2_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func TestHandleRequest(t *testing.T) {
	type SumParams struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	in := make(chan jsonrpc2.Message)
	out := make(chan jsonrpc2.Message)

	server := jsonrpc2.NewServer(jsonrpc2.NewChanObjStream(in, out))
	jsonrpc2.HandleRequest(server, "sum", func(p SumParams) any {
		return p.X + p.Y
	})
	go server.Listen()

	client := jsonrpc2.NewServer(jsonrpc2.NewChanObjStream(out, in))
	go client.Listen()

	res, err := jsonrpc2.SendRequest[int](client, "sum", SumParams{X: 100, Y: 42})
	require.Nil(t, err)

	require.Equal(t, 142, *res)
}
