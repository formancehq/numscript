package cmd

import (
	"encoding/json"
	"fmt"
	"io"
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

type rawSpec struct {
	NumscriptPath    string
	SpecsPath        string
	NumscriptContent string
	SpecsFileContent []byte
}

func readSpecsFiles() []rawSpec {
	var specs []rawSpec

	for _, path := range opts.paths {
		path = strings.TrimSuffix(path, "/")

		specsFilePaths, err := filepath.Glob(path + "/*.num.specs.json")
		if err != nil {
			panic(err)
		}

		if len(specsFilePaths) == 0 {
			_, _ = os.Stderr.Write([]byte(ansi.ColorRed("No specs files found\n")))
			os.Exit(1)
		}

		for _, specsFilePath := range specsFilePaths {
			numscriptFileName := strings.TrimSuffix(specsFilePath, ".specs.json")

			// TODO Improve err message ("no matching numscript for specsfile")
			numscriptContent, err := os.ReadFile(numscriptFileName)
			if err != nil {
				_, _ = os.Stderr.Write([]byte(err.Error()))
				os.Exit(1)
			}

			specsFileContent, err := os.ReadFile(specsFilePath)
			if err != nil {
				_, _ = os.Stderr.Write([]byte(err.Error()))
				os.Exit(1)
			}

			specs = append(specs, rawSpec{
				NumscriptPath:    numscriptFileName,
				SpecsPath:        specsFilePath,
				NumscriptContent: string(numscriptContent),
				SpecsFileContent: specsFileContent,
			})
		}
	}

	return specs
}

func runRawSpecs(stdout io.Writer, stderr io.Writer, rawSpecs []rawSpec) bool {
	if len(rawSpecs) == 0 {
		_, _ = stderr.Write([]byte(ansi.ColorRed("No specs files found\n")))
		return false
	}

	failedTestFiles := 0

	var allTests []testResult

	for _, rawSpec := range rawSpecs {
		specs, out, ok := runRawSpec(stdout, stderr, rawSpec)
		if !ok {
			return false
		}

		// Count tests
		isTestFailed := slices.ContainsFunc(out.Cases, func(tc specs_format.TestCaseResult) bool {
			return tc.Pass
		})
		if isTestFailed {
			failedTestFiles += 1
		}

		for _, caseResult := range out.Cases {
			allTests = append(allTests, testResult{
				Specs:  specs,
				Result: caseResult,
				File:   rawSpec.SpecsPath,
			})
		}

	}

	for _, test_ := range allTests {
		showFailingTestCase(stderr, test_)
	}

	// Stats
	return printFilesStats(stdout, allTests)

}

func runRawSpec(stdout io.Writer, stderr io.Writer, rawSpec rawSpec) (specs_format.Specs, specs_format.SpecsResult, bool) {
	parseResult := parser.Parse(rawSpec.NumscriptContent)
	if len(parseResult.Errors) != 0 {
		for _, err := range parseResult.Errors {
			showErr(stderr, rawSpec.NumscriptPath, rawSpec.NumscriptContent, err)
		}
		return specs_format.Specs{}, specs_format.SpecsResult{}, false
	}

	var specs specs_format.Specs
	err := json.Unmarshal(rawSpec.SpecsFileContent, &specs)
	if err != nil {
		_, _ = stderr.Write([]byte(ansi.ColorRed(fmt.Sprintf("\nError: %s.specs.json\n\n", rawSpec.NumscriptPath))))
		_, _ = stderr.Write([]byte(err.Error() + "\n"))
		return specs_format.Specs{}, specs_format.SpecsResult{}, false
	}

	out, iErr := specs_format.Check(parseResult.Value, specs)

	if iErr != nil {
		showErr(stderr, rawSpec.NumscriptPath, rawSpec.NumscriptContent, iErr)
		return specs_format.Specs{}, specs_format.SpecsResult{}, false
	}

	if out.Total == 0 {
		fmt.Fprintln(stdout, ansi.ColorRed("Empty test suite: "+rawSpec.SpecsPath))
		return specs_format.Specs{}, specs_format.SpecsResult{}, false
	} else if out.Failing == 0 {
		testsCount := ansi.ColorBrightBlack(fmt.Sprintf("(%d tests)", out.Total))
		fmt.Fprintf(stdout, "%s %s %s\n", ansi.ColorGreen("✓"), rawSpec.NumscriptPath, testsCount)
	} else {
		failedTestsCount := ansi.ColorRed(fmt.Sprintf("%d failed", out.Failing))

		testsCount := ansi.ColorBrightBlack(fmt.Sprintf("(%d tests | %s)", out.Total, failedTestsCount))
		fmt.Fprintf(stdout, "%s %s %s\n", ansi.ColorRed("❯"), rawSpec.NumscriptPath, testsCount)

		for _, result := range out.Cases {
			if result.Pass {
				continue
			}

			fmt.Fprintf(stdout, "  %s %s\n", ansi.ColorRed("×"), result.It)
		}
	}

	return specs, out, true
}

func showDiff(w io.Writer, expected_ any, got_ any) {
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
				fmt.Fprintln(w, ansi.ColorGreen("- "+line))
			case diffmatchpatch.DiffInsert:
				fmt.Fprintln(w, ansi.ColorRed("+ "+line))
			case diffmatchpatch.DiffEqual:
				fmt.Fprintln(w, ansi.ColorBrightBlack("  "+line))
			}
		}
	}
}

func showFailingTestCase(w io.Writer, testResult testResult) {
	if testResult.Result.Pass {
		return
	}

	specsFilePath := testResult.File
	result := testResult.Result

	fmt.Fprint(w, "\n\n")

	failColor := ansi.Compose(ansi.BgRed, ansi.ColorLight, ansi.Bold)
	fmt.Fprint(w, failColor(" FAIL "))
	fmt.Fprintln(w, ansi.ColorRed(" "+specsFilePath+" > "+result.It))

	showGiven := len(result.Balances) != 0 || len(result.Meta) != 0 || len(result.Vars) != 0
	if showGiven {
		fmt.Fprintln(w, ansi.Underline("\nGIVEN:"))
	}

	if len(result.Balances) != 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, result.Balances.PrettyPrint())
		fmt.Fprintln(w)
	}

	if len(result.Meta) != 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, result.Meta.PrettyPrint())
		fmt.Fprintln(w)
	}

	if len(result.Vars) != 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, utils.CsvPrettyMap("Name", "Value", result.Vars))
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, ansi.ColorGreen("- Expected"))
	fmt.Fprintln(w, ansi.ColorRed("+ Received\n"))

	for _, failedAssertion := range result.FailedAssertions {
		fmt.Fprintln(w, ansi.Underline(failedAssertion.Assertion))
		fmt.Fprintln(w)
		showDiff(w, failedAssertion.Expected, failedAssertion.Got)
	}
}

// TODO take writer
func showErr(stderr io.Writer, filename string, script string, err interpreter.InterpreterError) {
	rng := err.GetRange()

	errFile := fmt.Sprintf("\nError: %s:%d:%d\n\n", filename, rng.Start.Line+1, rng.Start.Character+1)
	_, _ = stderr.Write([]byte(ansi.ColorRed(errFile)))
	_, _ = stderr.Write([]byte(err.Error() + "\n\n"))

	if rng.Start != rng.End {
		_, _ = stderr.Write([]byte("\n"))
		_, _ = stderr.Write([]byte(rng.ShowOnSource(script) + "\n"))
	}
}

type testResult struct {
	Specs  specs_format.Specs
	File   string
	Result specs_format.TestCaseResult
}

func printFilesStats(w io.Writer, allTests []testResult) bool {
	failedTests := utils.Filter(allTests, func(t testResult) bool {
		return !t.Result.Pass
	})

	testFilesLabel := "Test files"
	testsLabel := "Tests"

	paddedLabel := func(s string) string {
		maxLen := max(len(testFilesLabel), len(testsLabel)) // yeah, ok, this could be hardcoded, I know
		return ansi.ColorBrightBlack(fmt.Sprintf(" %*s ", maxLen, s))
	}

	fmt.Fprintln(w)

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
		fmt.Fprint(w, paddedLabel(testFilesLabel)+" "+testFilesUI+" "+totalTestFilesUI)
	}

	fmt.Fprintln(w)

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

		fmt.Fprintln(w, paddedLabel(testsLabel)+" "+testsUI+" "+totalTestsUI)

		return failedTestsCount == 0
	}

}

type testArgs struct {
	paths []string
}

var opts = testArgs{}

func runTestCmd() {
	files := readSpecsFiles()
	pass := runRawSpecs(os.Stdout, os.Stderr, files)
	if !pass {
		os.Exit(1)
	}
}

func getTestCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "test folder...",
		Short: "Test numscript file using the numscript specs format",
		Long: `Searches for any <file>.num.specs files in the given directory (or directories),
and tests the corresponding <file>.num file (if any).
Defaults to "." if there are no given paths`,
		Args: cobra.MatchAll(),
		Run: func(cmd *cobra.Command, paths []string) {

			if len(paths) == 0 {
				paths = []string{"."}
			}

			opts.paths = paths
			runTestCmd()
		},
	}

	return cmd
}
