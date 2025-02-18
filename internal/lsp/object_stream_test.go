package lsp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/formancehq/numscript/internal/lsp"
	"github.com/stretchr/testify/require"
)

func TestObjectStreamWrite(t *testing.T) {
	out := NewRwCloser()

	stream := lsp.NewLsObjectStream(NewRwCloser(), out)

	outmsg := `{}`
	expectedMsg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(outmsg), outmsg)

	err := stream.WriteObject(json.RawMessage(outmsg))
	require.Nil(t, err)

	bs := make([]byte, len(expectedMsg))
	_, err = out.Read(bs)
	require.Nil(t, err)

	require.Equal(t, []byte(expectedMsg), bs)
}

func TestObjectStreamRead(t *testing.T) {
	in := NewRwCloser()

	stream := lsp.NewLsObjectStream(in, NewRwCloser())

	outmsg := `{"x": 42}`
	msg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(outmsg), outmsg)

	in.Write([]byte(msg))

	raw, err := stream.ReadObject()
	require.Nil(t, err)

	var m any
	err = json.Unmarshal(*raw, &m)
	require.Nil(t, err)

	require.Equal(t, map[string]any{"x": float64(42)}, m)
}

func TestObjectStreamSimmetric(t *testing.T) {
	in := NewRwCloser()
	out := NewRwCloser()

	serverStream := lsp.NewLsObjectStream(in, out)
	clientStream := lsp.NewLsObjectStream(out, in)

	err := serverStream.WriteObject(42)
	require.Nil(t, err)

	obj, err := clientStream.ReadObject()
	require.Nil(t, err)

	var decoded float64
	json.Unmarshal(*obj, &decoded)
	require.Equal(t, float64(42), decoded)

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
