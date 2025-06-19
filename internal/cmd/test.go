package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/spf13/cobra"
)

func test(path string) {
	numscriptContent, err := os.ReadFile(path)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}

	parseResult := parser.Parse(string(numscriptContent))
	// TODO assert no parse err
	// TODO we might want to do static checking

	specsFileContent, err := os.ReadFile(path + ".specs.json")
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}

	var specs specs_format.Specs
	err = json.Unmarshal([]byte(specsFileContent), &specs)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}

	out, err := specs_format.Run(parseResult.Value, specs)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}

	if !out.Success {
		fmt.Printf("Postings mismatch.\n\tExpected: %v\n\tGot:%v\n", out.ExpectedPostings, out.ActualPostings)
	}
}

// TODO test directory instead
var testCmd = &cobra.Command{
	Use:   "test <path>",
	Short: "Test a numscript file, using the corresponding spec file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		test(path)
	},
}
