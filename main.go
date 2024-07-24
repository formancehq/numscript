package main

import (
	"flag"
	"fmt"
	lsp "numscript/lsp"
	"numscript/parser"
	"os"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
)

var pathFlag = flag.String("path", "", "path to execute")

func recoverPanic() {
	r := recover()
	if r == nil {
		return
	}

	errMsg := fmt.Sprintf("[uncaught panic]: %s\n%s\n", r, string(debug.Stack()))
	os.Stderr.Write([]byte(errMsg))
	sentry.CaptureMessage(errMsg)
	sentry.Flush(2 * time.Second)
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn:              "https://b8b6cfd5dab95e1258d80963c3db73bf@o4504394442539008.ingest.us.sentry.io/4507623538884608",
		AttachStacktrace: true,
	})

	defer recoverPanic()
	defer sentry.Flush(2 * time.Second)

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
			os.Stderr.Write([]byte(err.Error()))
			return
		}

		parsed := parser.Parse(string(dat))

		fmt.Printf("Parser errors: %d\n\n", len(parsed.Errors))
		for _, err := range parsed.Errors {
			fmt.Printf("%v,  (line=%d, char=%d) ", err.Msg, err.Range.Start.Line, err.Range.Start.Character)
		}
		return
	default:
		os.Stderr.Write([]byte(fmt.Sprintf("Invalid argument: '%s'", firstArg)))
	}
}
