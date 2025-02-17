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

	var init lsp_types.InitializeResult
	err := json.Unmarshal(response.Result, &init)
	require.Nil(t, err)

	require.True(t, init.Capabilities.HoverProvider)
}
