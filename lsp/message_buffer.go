package lsp

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"strconv"

	"github.com/sourcegraph/jsonrpc2"
)

type MessageBuffer struct{ reader *bufio.Reader }

func NewMessageBuffer(r io.Reader) MessageBuffer {
	return MessageBuffer{
		reader: bufio.NewReader(r),
	}
}

func (mb *MessageBuffer) Read() jsonrpc2.Request {
	tpr := textproto.NewReader(mb.reader)

	headers, err := tpr.ReadMIMEHeader()
	if err != nil {
		panic(err)
	}

	contentLenHeader := headers.Get("Content-Length")
	len, err := strconv.ParseInt(contentLenHeader, 10, 0)
	if err != nil {
		panic(err)
	}

	bytes := make([]byte, len)
	readBytes, readErr := io.ReadFull(tpr.R, bytes)
	if readErr != nil {
		panic(readErr)
	}
	if readBytes != int(len) {
		panic(fmt.Sprint("Missing bytes to read. Read: ", readBytes, ", total: ", len))
	}

	var req jsonrpc2.Request
	err1 := req.UnmarshalJSON(bytes)
	if err1 != nil {
		panic(err1)
	}

	return req
}
