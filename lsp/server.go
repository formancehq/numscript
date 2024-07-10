package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"strconv"

	"github.com/sourcegraph/jsonrpc2"
)

type Handler interface {
	Handle(r jsonrpc2.Request) any
}

func RunServer(handler Handler) {
	buf := NewMessageBuffer(os.Stdin)

	for {
		request := buf.Read()

		bytes, err := json.Marshal(handler.Handle(request))
		if err != nil {
			panic(err)
		}

		rawMsg := json.RawMessage(bytes)
		jsonRes := jsonrpc2.Response{
			ID:     request.ID,
			Result: &rawMsg,
		}

		response, err := jsonRes.MarshalJSON()
		if err != nil {
			panic(err)
		}

		// TODO is the number of bytes correct?
		_, err = fmt.Printf(`Content-Length: %v\r\n\r\n%v`, len(response), string(response))
		if err != nil {
			panic(err)
		}

		os.Stderr.WriteString(fmt.Sprintf(`Content-Length: %v\r\n\r\n%v`, len(response), string(response)))

	}

}

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
