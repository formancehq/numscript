package lsp_test

import (
	"encoding/json"
	"testing"

	"github.com/formancehq/numscript/internal/json_rpc"
	"github.com/formancehq/numscript/internal/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	in := make(chan any)
	out := make(chan any)
	objStream := NewChanObjStream(in, out)
	go lsp.RunServerWith(&objStream)

	empyParams := json.RawMessage("{}")

	in <- jsonrpc2.Request{
		ID:     jsonrpc2.ID{Num: 0},
		Method: "initialize",
		Params: &empyParams,
	}

	res := <-out

	response := res.(jsonrpc2.Response)

	require.Equal(t,
		jsonrpc2.ID{Num: 0},
		response.ID,
	)

	var init lsp.InitializeResult
	err := json.Unmarshal(*response.Result, &init)
	require.Nil(t, err)

	require.True(t, init.Capabilities.HoverProvider)
}

type ChanObjStream struct {
	in  <-chan any
	out chan<- any
}

var _ json_rpc.ObjectStream = (*ChanObjStream)(nil)

func NewChanObjStream(in <-chan any, out chan<- any) ChanObjStream {
	return ChanObjStream{
		in:  in,
		out: out,
	}
}

func (c *ChanObjStream) Close() error {
	close(c.out)
	return nil
}

func (c *ChanObjStream) ReadObject() (*json.RawMessage, error) {
	x := <-c.in
	bytes, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}
	return (*json.RawMessage)(&bytes), nil
}

func (c *ChanObjStream) WriteObject(obj any) error {
	c.out <- obj
	return nil
}
