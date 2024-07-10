package main

import (
	"flag"
	"fmt"
	"numscript/parser"
	"os"
)

var pathFlag = flag.String("path", "", "path to execute")

func main() {
	flag.Parse()

	if *pathFlag == "" {
		panic("Path argument is required")
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

}
