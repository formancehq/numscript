package main

import (
	"encoding/json"
	"flag"
	"fmt"
	lsp "numscript/lsp"
	"numscript/parser"
	"os"

	"github.com/sourcegraph/jsonrpc2"
)

var pathFlag = flag.String("path", "", "path to execute")

func main() {
	flag.Parse()
	firstArg := flag.Arg(0)

	switch firstArg {
	case "lsp":
		runServer()
		return

	case "parse":
		if *pathFlag == "" {
			fmt.Println("Err: Path argument is required")
			return
		}

		dat, err := os.ReadFile(*pathFlag)
		if err != nil {
			panic(err)
		}

		parsed := parser.Parse(string(dat))

		fmt.Printf("Lexing errors: %d\nParser errors: %d\n\n", len(parsed.LexerErrors), len(parsed.Errors))
		for _, err := range parsed.Errors {
			fmt.Printf("%v,  (line=%d, char=%d) ", err.Msg, err.Range.Start.Line, err.Range.Start.Character)
		}
		return
	default:
		fmt.Printf("Invalid argument: '%s'", firstArg)
	}
}

func handle(r jsonrpc2.Request) any {
	switch r.Method {
	case "initialize":
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{},
			// This is ugly. Is there a shortcut?
			ServerInfo: struct {
				Name    string "json:\"name\""
				Version string "json:\"version,omitempty\""
			}{
				Name:    "numscript-ls",
				Version: "0.0.1",
			},
		}

	default:
		// Unhandled method
		// TODO should it panic?
		return nil
	}

}

func runServer() {
	buf := lsp.NewMessageBuffer(os.Stdin)

	for {
		request := buf.Read()

		bytes, err := json.Marshal(handle(request))
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
		fmt.Printf(`Content-Length: %d\r\n\r\n%s`, len(response), response)

		// os.Stderr.WriteString("RES " + string(res))
	}
}
