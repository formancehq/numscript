package lsp_test

import (
	"encoding/json"
	"testing"

	"github.com/formancehq/numscript/internal/jsonrpc2"
	"github.com/formancehq/numscript/internal/lsp"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
)

func TestServerReadWrite(t *testing.T) {
	in := make(chan jsonrpc2.Message)
	out := make(chan jsonrpc2.Message)
	lsp.NewConn(jsonrpc2.NewChanObjStream(in, out))

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

	var init protocol.InitializeResult
	err := json.Unmarshal(response.Result, &init)
	require.Nil(t, err)

	b, ok := init.Capabilities.HoverProvider.(bool)
	require.True(t, ok, "cast init.Capabilities.HoverProvider to bool")
	require.True(t, b)
}
