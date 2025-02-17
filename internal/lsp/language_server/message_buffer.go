package language_server

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"strconv"

	"github.com/sourcegraph/jsonrpc2"
)

type messageBuffer struct{ reader *bufio.Reader }

func newMessageBuffer(r io.Reader) messageBuffer {
	return messageBuffer{
		reader: bufio.NewReader(r),
	}
}

func (mb *messageBuffer) Read() jsonrpc2.Request {
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
