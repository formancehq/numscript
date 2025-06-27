package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/formancehq/numscript/internal/ansi"
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

	out := specs_format.Run(parseResult.Value, specs)
	for _, result := range out.Cases {
		if !result.Pass {
			fmt.Println(ansi.Underline(`it: ` + result.It))

			fmt.Println("\nExpected:")
			expected, _ := json.MarshalIndent(result.ExpectedPostings, "", "  ")
			fmt.Println(ansi.ColorGreen(string(expected)))

			fmt.Println("\nGot:")
			actual, _ := json.MarshalIndent(result.ActualPostings, "", "  ")
			fmt.Println(ansi.ColorRed(string(actual)))

		}
	}

	if out.Total == 0 {
		fmt.Println(ansi.ColorRed("Empty test suite!"))
		os.Exit(1)
	} else if out.Failing == 0 {
		fmt.Printf("All tests passing ✅\n")
		return
	} else {
		os.Exit(1)
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
