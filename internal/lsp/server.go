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

type ServerArgs[State any] struct {
	InitialState State
	Handler      func(r jsonrpc2.Request, state *State) any
}

func SendNotification(method string, params interface{}) {
	bytes, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	rawParams := json.RawMessage(bytes)
	encoded := encodeMessage(jsonrpc2.Request{
		Notif:  true,
		Method: method,
		Params: &rawParams,
	})
	os.Stderr.Write([]byte(encoded))
	_, err = fmt.Print(encoded)
	if err != nil {
		panic(err)
	}
}

func RunServer[State any](args ServerArgs[State]) {
	buf := NewMessageBuffer(os.Stdin)

	for {
		request := buf.Read()

		bytes, err := json.Marshal(args.Handler(request, &args.InitialState))
		if err != nil {
			panic(err)
		}

		rawMsg := json.RawMessage(bytes)
		encoded := encodeMessage(jsonrpc2.Response{
			ID:     request.ID,
			Result: &rawMsg,
		})
		_, err = fmt.Print(encoded)
		if err != nil {
			panic(err)
		}
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
		if err.Error() == "EOF" {
			os.Exit(0)
		}
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

func encodeMessage(msg any) string {
	bytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(bytes), bytes)
}
