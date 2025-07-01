package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/formancehq/numscript/internal/ansi"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/formancehq/numscript/internal/utils"
	"github.com/spf13/cobra"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func showFailingTestCase(specsFilePath string, result specs_format.TestCaseResult) {
	if result.Pass {
		return
	}

	fmt.Print("\n\n")

	failColor := ansi.Compose(ansi.BgRed, ansi.ColorLight, ansi.Bold)
	fmt.Print(failColor(" FAIL "))
	fmt.Println(ansi.ColorRed(" " + specsFilePath + " > " + result.It))

	showGiven := len(result.Balances) != 0 || len(result.Meta) != 0 || len(result.Vars) != 0
	if showGiven {
		fmt.Println(ansi.Underline("\nGIVEN:"))
	}

	if len(result.Balances) != 0 {
		fmt.Println()
		fmt.Println(result.Balances.PrettyPrint())
		fmt.Println()
	}

	if len(result.Meta) != 0 {
		fmt.Println()
		fmt.Println(result.Meta.PrettyPrint())
		fmt.Println()
	}

	if len(result.Vars) != 0 {
		fmt.Println()
		fmt.Println(utils.CsvPrettyMap("Name", "Value", result.Vars))
		fmt.Println()
	}

	fmt.Print(ansi.Underline("EXPECT:\n\n"))

	fmt.Println(ansi.ColorGreen("- Expected"))
	fmt.Println(ansi.ColorRed("+ Received\n"))

	dmp := diffmatchpatch.New()

	expected, _ := json.MarshalIndent(result.ExpectedPostings, "", "  ")
	actual, _ := json.MarshalIndent(result.ActualPostings, "", "  ")

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

func test(specsFilePath string) specs_format.SpecsResult {
	if !strings.HasSuffix(specsFilePath, ".num.specs.json") {
		panic("Wrong name")
	}

	numscriptFileName := strings.TrimSuffix(specsFilePath, ".specs.json")

	numscriptContent, err := os.ReadFile(numscriptFileName)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}

	parseResult := parser.Parse(string(numscriptContent))
	// TODO assert no parse err
	// TODO we might want to do static checking

	specsFileContent, err := os.ReadFile(specsFilePath)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}

	var specs specs_format.Specs
	err = json.Unmarshal([]byte(specsFileContent), &specs)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}

	out := specs_format.Check(parseResult.Value, specs)

	if out.Total == 0 {
		fmt.Println(ansi.ColorRed("Empty test suite!"))
		os.Exit(1)
	} else if out.Failing == 0 {
		testsCount := ansi.ColorBrightBlack(fmt.Sprintf("(%d tests)", out.Total))
		fmt.Printf("%s %s %s\n", ansi.ColorGreen("✓"), numscriptFileName, testsCount)
	} else {
		failedTestsCount := ansi.ColorRed(fmt.Sprintf("%d failed", out.Failing))

		testsCount := ansi.ColorBrightBlack(fmt.Sprintf("(%d tests | %s)", out.Total, failedTestsCount))
		fmt.Printf("%s %s %s\n", ansi.ColorRed("❯"), numscriptFileName, testsCount)

		for _, result := range out.Cases {
			if result.Pass {
				continue
			}

			fmt.Printf("  %s %s\n", ansi.ColorRed("×"), result.It)
		}

	}

	return out
}

func testPaths(paths []string) {
	for _, path := range paths {
		path = strings.TrimSuffix(path, "/")

		glob := fmt.Sprintf(path + "/*.num.specs.json")

		files, err := filepath.Glob(glob)
		if err != nil {
			panic(err)
		}

		type FailingSpec struct {
			File   string
			Result specs_format.TestCaseResult
		}

		var failingTests []FailingSpec

		for _, file := range files {
			out := test(file)

			for _, testCase := range out.Cases {
				if testCase.Pass {
					continue
				}

				failingTests = append(failingTests, FailingSpec{
					File:   file,
					Result: testCase,
				})
			}
		}

		if len(failingTests) == 0 {
			return
		}

		for _, failedTest := range failingTests {
			showFailingTestCase(failedTest.File, failedTest.Result)
		}
		os.Exit(1)
	}
}

var testCmd = &cobra.Command{
	Use:   "test <path>",
	Short: "Test a numscript file, using the corresponding spec file",
	Args:  cobra.MatchAll(),
	Run: func(cmd *cobra.Command, paths []string) {
		if len(paths) == 0 {
			paths = []string{"."}
		}

		testPaths(paths)
	},
}
