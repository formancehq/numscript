package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/formancehq/numscript/internal/ansi"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/spf13/cobra"

	"github.com/sergi/go-diff/diffmatchpatch"
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

	out := specs_format.Check(parseResult.Value, specs)
	for _, result := range out.Cases {
		if result.Pass {
			continue
		}

		expected, _ := json.MarshalIndent(result.ExpectedPostings, "", "  ")
		actual, _ := json.MarshalIndent(result.ActualPostings, "", "  ")

		fmt.Println(ansi.Underline("it: " + result.It))

		if len(result.Balances) != 0 {
			fmt.Println()
			fmt.Println(result.Balances.PrettyPrint())
			fmt.Println()
		}

		fmt.Println(ansi.ColorGreen("- Expected"))
		fmt.Println(ansi.ColorRed("+ Received\n"))

		dmp := diffmatchpatch.New()

		aChars, bChars, lineArray := dmp.DiffLinesToChars(string(expected), string(actual))
		diffs := dmp.DiffMain(aChars, bChars, true)
		diffs = dmp.DiffCharsToLines(diffs, lineArray)

		for _, diff := range diffs {
			lines := strings.Split(diff.Text, "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				switch diff.Type {
				case diffmatchpatch.DiffDelete:
					fmt.Println(ansi.ColorGreen("- " + line))
				case diffmatchpatch.DiffInsert:
					fmt.Println(ansi.ColorRed("+ " + line))
				case diffmatchpatch.DiffEqual:
					fmt.Println("  " + line)
				}
			}
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
