package lsp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/formancehq/numscript/internal/jsonrpc2"
	"github.com/formancehq/numscript/internal/lsp"
	"github.com/stretchr/testify/require"
)

func TestObjectStreamWrite(t *testing.T) {
	out := NewRwCloser()

	stream := lsp.NewLsObjectStream(NewRwCloser(), out)

	msg := jsonrpc2.Request{
		Method: "updatedateConfig",
		Params: []byte("{}"),
	}
	strMsg, _ := json.Marshal(msg)

	err := stream.WriteMessage(msg)
	require.Nil(t, err)

	expectedMsg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(strMsg), strMsg)
	bs := make([]byte, len(expectedMsg))
	_, err = out.Read(bs)
	require.Nil(t, err)

	require.Equal(t, []byte(expectedMsg), bs)
}

func TestObjectStreamRead(t *testing.T) {
	in := NewRwCloser()

	stream := lsp.NewLsObjectStream(in, NewRwCloser())

	sentMsg := jsonrpc2.Request{
		Method: "updatedateConfig",
		Params: []byte("{}"),
	}
	strMsg, _ := json.Marshal(sentMsg)

	in.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(strMsg), strMsg)))

	receivedMsg, err := stream.ReadMessage()
	require.Nil(t, err)

	require.Equal(t, receivedMsg, sentMsg)
}

func TestObjectStreamSimmetric(t *testing.T) {
	in := NewRwCloser()
	out := NewRwCloser()

	serverStream := lsp.NewLsObjectStream(in, out)
	clientStream := lsp.NewLsObjectStream(out, in)

	msg := jsonrpc2.Request{Method: "got-data", Params: []byte("{}")}

	err := serverStream.WriteMessage(msg)
	require.Nil(t, err)

	obj, err := clientStream.ReadMessage()
	require.Nil(t, err)

	require.Equal(t, msg, obj)

}

type MockReadWriteCloser struct {
	io.ReadWriter
}

func NewRwCloser() MockReadWriteCloser {
	return MockReadWriteCloser{ReadWriter: &bytes.Buffer{}}
}

func (c MockReadWriteCloser) Close() error {
	return nil
}
