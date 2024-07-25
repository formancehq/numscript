package cmd

import (
	"fmt"
	"numscript/parser"
	"os"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check a numscript file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		dat, err := os.ReadFile(path)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}

		parsed := parser.Parse(string(dat))

		fmt.Printf("Parser errors: %d\n\n", len(parsed.Errors))
		for _, err := range parsed.Errors {
			fmt.Printf("%v,  (line=%d, char=%d) ", err.Msg, err.Range.Start.Line, err.Range.Start.Character)
		}
	},
}
