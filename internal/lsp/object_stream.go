package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"sync"

	"github.com/formancehq/numscript/internal/jsonrpc2"
)

type LsObjectStream struct {
	mu     sync.Mutex
	reader *bufio.Reader

	in  io.ReadCloser
	out io.WriteCloser
}

var _ jsonrpc2.MessageStream = (*LsObjectStream)(nil)

func NewLsObjectStream(in io.ReadCloser, out io.WriteCloser) LsObjectStream {
	reader := bufio.NewReader(in)
	return LsObjectStream{
		reader: reader,
		in:     in,
		out:    out,
	}
}

func (s *LsObjectStream) Close() error {
	err := s.out.Close()
	err2 := s.in.Close()

	if err != nil {
		return err
	}

	if err2 != nil {
		return err2
	}

	return nil
}

func (s *LsObjectStream) WriteMessage(obj jsonrpc2.Message) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	encoded := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(bytes), bytes)

	s.mu.Lock()
	_, err = s.out.Write([]byte(encoded))
	s.mu.Unlock()
	if err != nil {
		return err
	}

	return nil
}

func (s *LsObjectStream) ReadMessage() (jsonrpc2.Message, error) {
	tpr := textproto.NewReader(s.reader)

	headers, err := tpr.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	contentLenHeader := headers.Get("Content-Length")
	len, err := strconv.ParseInt(contentLenHeader, 10, 0)
	if err != nil {
		return nil, err
	}

	bytes := make([]byte, len)
	readBytes, err := io.ReadFull(tpr.R, bytes)
	if err != nil {
		return nil, err
	}

	if readBytes != int(len) {
		return nil, fmt.Errorf("missing bytes to read. Read: %d, total: %d", len, readBytes)
	}

	return jsonrpc2.UnmarshalMessage(bytes)
}
