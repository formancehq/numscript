package language_server

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
)

type ServerArgs[State any] struct {
	InitialState func(notify func(method string, params any)) State
	Handler      func(r jsonrpc2.Request, state State) any
}

func sendNotification(writeMutex *sync.Mutex, method string, params any) {
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

	writeMutex.Lock()
	os.Stdout.Write([]byte(encoded))
	writeMutex.Unlock()
}

func RunServer[State any](args ServerArgs[State]) {
	buf := newMessageBuffer(os.Stdin)
	mu := sync.Mutex{}

	state := args.InitialState(func(method string, params any) {
		sendNotification(&mu, method, params)
	})

	for {
		request := buf.Read()

		go func() {
			bytes, err := json.Marshal(args.Handler(request, state))
			if err != nil {
				panic(err)
			}

			rawMsg := json.RawMessage(bytes)
			encoded := encodeMessage(jsonrpc2.Response{
				ID:     request.ID,
				Result: &rawMsg,
			})

			mu.Lock()
			os.Stdout.Write([]byte(encoded))
			mu.Unlock()
		}()
	}
}

func encodeMessage(msg any) string {
	bytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(bytes), bytes)
}
