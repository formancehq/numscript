package lsp_test

import (
	"encoding/json"
	"testing"

	"github.com/formancehq/numscript/internal/jsonrpc2"
	"github.com/formancehq/numscript/internal/lsp"
	"github.com/formancehq/numscript/internal/lsp/lsp_types"
	"github.com/stretchr/testify/require"
)

func TestServerReadWrite(t *testing.T) {
	in := make(chan jsonrpc2.Message)
	out := make(chan jsonrpc2.Message)
	objStream := NewChanObjStream(in, out)
	go lsp.RunServerWith(&objStream)

	in <- jsonrpc2.Request{
		ID:     jsonrpc2.NewIntId(0),
		Method: "initialize",
		Params: []byte("{}"),
	}

	response := (<-out).(jsonrpc2.Response)

	require.Equal(t,
		jsonrpc2.NewIntId(0),
		response.ID,
	)

	var init lsp_types.InitializeResult
	err := json.Unmarshal(response.Result, &init)
	require.Nil(t, err)

	require.True(t, init.Capabilities.HoverProvider)
}

type ChanObjStream struct {
	in  <-chan jsonrpc2.Message
	out chan<- jsonrpc2.Message
}

var _ jsonrpc2.MessageStream = (*ChanObjStream)(nil)

func NewChanObjStream(in <-chan jsonrpc2.Message, out chan<- jsonrpc2.Message) ChanObjStream {
	return ChanObjStream{
		in:  in,
		out: out,
	}
}

func (c *ChanObjStream) Close() error {
	close(c.out)
	return nil
}

func (c *ChanObjStream) ReadMessage() (jsonrpc2.Message, error) {
	msg := <-c.in
	return msg, nil
}

func (c *ChanObjStream) WriteMessage(obj jsonrpc2.Message) error {
	c.out <- obj
	return nil
}
