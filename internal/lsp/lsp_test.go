package lsp_test

import (
	"encoding/json"
	"testing"

	"github.com/formancehq/numscript/internal/json_rpc"
	"github.com/formancehq/numscript/internal/lsp"
	"github.com/formancehq/numscript/internal/lsp/lsp_types"
	"github.com/stretchr/testify/require"
)

func TestServerReadWrite(t *testing.T) {
	in := make(chan json_rpc.Message)
	out := make(chan json_rpc.Message)
	objStream := NewChanObjStream(in, out)
	go lsp.RunServerWith(&objStream)

	in <- json_rpc.Request{
		ID:     json_rpc.NewIntId(0),
		Method: "initialize",
		Params: []byte("{}"),
	}

	response := (<-out).(json_rpc.Response)

	require.Equal(t,
		json_rpc.NewIntId(0),
		response.ID,
	)

	var init lsp_types.InitializeResult
	err := json.Unmarshal(response.Result, &init)
	require.Nil(t, err)

	require.True(t, init.Capabilities.HoverProvider)
}

type ChanObjStream struct {
	in  <-chan json_rpc.Message
	out chan<- json_rpc.Message
}

var _ json_rpc.MessageStream = (*ChanObjStream)(nil)

func NewChanObjStream(in <-chan json_rpc.Message, out chan<- json_rpc.Message) ChanObjStream {
	return ChanObjStream{
		in:  in,
		out: out,
	}
}

func (c *ChanObjStream) Close() error {
	close(c.out)
	return nil
}

func (c *ChanObjStream) ReadMessage() (json_rpc.Message, error) {
	msg := <-c.in
	return msg, nil
}

func (c *ChanObjStream) WriteMessage(obj json_rpc.Message) error {
	c.out <- obj
	return nil
}
