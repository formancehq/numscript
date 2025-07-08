package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/ansi"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/formancehq/numscript/internal/utils"
	"github.com/spf13/cobra"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func showDiff(expected_ any, got_ any) {
	dmp := diffmatchpatch.New()

	expected, _ := json.MarshalIndent(expected_, "", "  ")
	actual, _ := json.MarshalIndent(got_, "", "  ")

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
				fmt.Println(ansi.ColorBrightBlack("  " + line))
			}
		}
	}
}

func showFailingTestCase(testResult testResult) (rerun bool) {
	specsFilePath := testResult.File
	result := testResult.Result

	if result.Pass {
		return false
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

	fmt.Println()
	fmt.Println(ansi.ColorGreen("- Expected"))
	fmt.Println(ansi.ColorRed("+ Received\n"))

	for _, failedAssertion := range result.FailedAssertions {

		fmt.Println(ansi.Underline(failedAssertion.Assertion))
		fmt.Println()
		showDiff(failedAssertion.Expected, failedAssertion.Got)

		if interactiveMode {
			fmt.Println(ansi.ColorBrightBlack(
				fmt.Sprintf("\nPress %s to update snapshot, %s to go the the next one",
					ansi.ColorBrightYellow("u"),
					ansi.ColorBrightYellow("n"),
				)))

			reader := bufio.NewReader(os.Stdin)
			line, _, err := reader.ReadLine()
			if err != nil {
				panic(err)
			}

			switch string(line) {
			case "u":
				testResult.Specs.TestCases = utils.Map(testResult.Specs.TestCases, func(t specs_format.TestCase) specs_format.TestCase {
					// TODO check there are no duplicate "It"
					if t.It == testResult.Result.It {
						switch failedAssertion.Expected {
						case "expect.postings":
							t.ExpectedPostings = failedAssertion.Expected.([]interpreter.Posting)

						default:
							panic("TODO implement")

						}

					}

					return t
				})

				newSpecs, err := json.MarshalIndent(testResult.Specs, "", "  ")
				if err != nil {
					panic(err)
				}

				err = os.WriteFile(testResult.File, newSpecs, os.ModePerm)
				if err != nil {
					panic(err)
				}
				return true

			case "n":
				return false

			default:
				panic("TODO invalid command")
			}

		}

	}

	return false
}

func test(specsFilePath string) (specs_format.Specs, specs_format.SpecsResult) {
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

	return specs, out
}

type testResult struct {
	Specs  specs_format.Specs
	File   string
	Result specs_format.TestCaseResult
}

func testPaths(paths []string) {
	testFiles := 0
	failedTestFiles := 0

	var allTests []testResult
	for _, path := range paths {
		path = strings.TrimSuffix(path, "/")

		glob := fmt.Sprintf(path + "/*.num.specs.json")

		files, err := filepath.Glob(glob)
		if err != nil {
			panic(err)
		}
		testFiles += len(files)

		for _, file := range files {
			specs, out := test(file)

			for _, testCase := range out.Cases {
				allTests = append(allTests, testResult{
					Specs:  specs,
					File:   file,
					Result: testCase,
				})
			}

			// Count tests
			isTestFailed := slices.ContainsFunc(out.Cases, func(tc specs_format.TestCaseResult) bool {
				return tc.Pass
			})
			if isTestFailed {
				failedTestFiles += 1
			}
		}
	}

	for _, test_ := range allTests {
		rerun := showFailingTestCase(test_)
		if rerun {
			fmt.Print("\033[H\033[2J")
			testPaths(paths)
			return
		}
	}

	// Stats
	printFilesStats(allTests)

}

var interactiveMode = false

func printFilesStats(allTests []testResult) {
	failedTests := utils.Filter(allTests, func(t testResult) bool {
		return !t.Result.Pass
	})

	testFilesLabel := "Test files"
	testsLabel := "Tests"

	paddedLabel := func(s string) string {
		maxLen := max(len(testFilesLabel), len(testsLabel)) // yeah, ok, this could be hardcoded, I know
		return ansi.ColorBrightBlack(fmt.Sprintf(" %*s ", maxLen, s))
	}

	fmt.Println()

	// Files stats
	{
		filesCount := len(slices.CompactFunc(allTests, func(t1 testResult, t2 testResult) bool {
			return t1.File == t2.File
		}))
		failedTestsFilesCount := len(slices.CompactFunc(failedTests, func(t1 testResult, t2 testResult) bool {
			return t1.File == t2.File
		}))
		passedTestsFilesCount := filesCount - failedTestsFilesCount

		var testFilesUIParts []string
		if failedTestsFilesCount != 0 {
			testFilesUIParts = append(testFilesUIParts,
				ansi.Compose(ansi.ColorBrightRed, ansi.Bold)(fmt.Sprintf("%d failed", failedTestsFilesCount)),
			)
		}
		if passedTestsFilesCount != 0 {
			testFilesUIParts = append(testFilesUIParts,
				ansi.Compose(ansi.ColorBrightGreen, ansi.Bold)(fmt.Sprintf("%d passed", passedTestsFilesCount)),
			)
		}
		testFilesUI := strings.Join(testFilesUIParts, ansi.ColorBrightBlack(" | "))
		totalTestFilesUI := ansi.ColorBrightBlack(fmt.Sprintf("(%d)", filesCount))
		fmt.Print(paddedLabel(testFilesLabel) + " " + testFilesUI + " " + totalTestFilesUI)
	}

	fmt.Println()

	// Tests stats
	{

		testsCount := len(allTests)
		failedTestsCount := len(failedTests)
		passedTestsCount := testsCount - failedTestsCount

		var testUIParts []string
		if failedTestsCount != 0 {
			testUIParts = append(testUIParts,
				ansi.Compose(ansi.ColorBrightRed, ansi.Bold)(fmt.Sprintf("%d failed", failedTestsCount)),
			)
		}
		if passedTestsCount != 0 {
			testUIParts = append(testUIParts,
				ansi.Compose(ansi.ColorBrightGreen, ansi.Bold)(fmt.Sprintf("%d passed", passedTestsCount)),
			)
		}

		testsUI := strings.Join(testUIParts, ansi.ColorBrightBlack(" | "))
		totalTestsUI := ansi.ColorBrightBlack(fmt.Sprintf("(%d)", testsCount))

		fmt.Print(paddedLabel(testsLabel) + " " + testsUI + " " + totalTestsUI)

		if failedTestsCount != 0 {
			os.Exit(1)
		}
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
