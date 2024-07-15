package main

import (
	"flag"
	"fmt"
	lsp "numscript/lsp"
	"numscript/parser"
	"os"
)

var pathFlag = flag.String("path", "", "path to execute")

func main() {
	flag.Parse()
	firstArg := flag.Arg(0)

	switch firstArg {
	case "lsp":
		lsp.RunServer(lsp.ServerArgs[lsp.State]{
			InitialState: lsp.InitialState(),
			Handler:      lsp.Handle,
		})
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

		fmt.Printf("Parser errors: %d\n\n", len(parsed.Errors))
		for _, err := range parsed.Errors {
			fmt.Printf("%v,  (line=%d, char=%d) ", err.Msg, err.Range.Start.Line, err.Range.Start.Character)
		}
		return
	default:
		fmt.Printf("Invalid argument: '%s'", firstArg)
	}
}
