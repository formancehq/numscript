package cmd

import (
	"io/fs"
	"os"

	"github.com/formancehq/numscript/internal/format"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/spf13/cobra"
)

func formatFile(path string) {
	source, err := os.ReadFile(path)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}
	result := parser.Parse(string(source))
	// TODO proper error
	if len(result.Errors) != 0 {
		panic("Got errors while parsing")
	}

	formatted := format.Format(result.Value)
	errw := os.WriteFile(path, ([]byte)(formatted), fs.ModePerm)
	if errw != nil {
		panic(errw)
	}
}

func getFmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "format",
		Short:  "Format a numscript file",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			formatFile(path)
		},
	}
}
