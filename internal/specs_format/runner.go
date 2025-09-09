package specs_format

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/ansi"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type RawSpec struct {
	NumscriptPath    string
	SpecsPath        string
	NumscriptContent string
	SpecsFileContent []byte
}

type TestResult struct {
	Specs  Specs
	File   string
	Result TestCaseResult
}

func readSpecFile(path string) (RawSpec, error) {
	numscriptFileName := strings.TrimSuffix(path, ".specs.json")

	numscriptContent, err := os.ReadFile(numscriptFileName)
	if err != nil {
		return RawSpec{}, err
	}

	specsFileContent, err := os.ReadFile(path)
	if err != nil {
		return RawSpec{}, err
	}

	return RawSpec{
		NumscriptPath:    numscriptFileName,
		SpecsPath:        path,
		NumscriptContent: string(numscriptContent),
		SpecsFileContent: specsFileContent,
	}, nil
}

func ReadSpecsFiles(paths []string) ([]RawSpec, error) {
	var specs []RawSpec

	for _, root := range paths {
		root = strings.TrimSuffix(root, "/")

		info, err := os.Stat(root)
		if err != nil {
			_, _ = os.Stderr.Write([]byte(err.Error()))
			os.Exit(1)
		}

		if !info.IsDir() {
			rawSpec, err := readSpecFile(root)
			if err != nil {
				return nil, err
			}

			specs = append(specs, rawSpec)
			continue
		}

		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if d.IsDir() || !strings.HasSuffix(path, ".num.specs.json") {
				return nil
			}

			rawSpec, err := readSpecFile(path)
			if err != nil {
				return err
			}

			specs = append(specs, rawSpec)
			return nil
		})

		if err != nil {
			return nil, err
		}

	}

	return specs, nil
}

func RunSpecs(stdout io.Writer, stderr io.Writer, rawSpecs []RawSpec) bool {
	if len(rawSpecs) == 0 {
		_, _ = stderr.Write([]byte(ansi.ColorRed("No specs files found\n")))
		return false
	}

	var allTests []TestResult

	for _, rawSpec := range rawSpecs {
		specs, out, ok := runRawSpec(stdout, stderr, rawSpec)
		if !ok {
			return false
		}

		for _, caseResult := range out.Cases {
			allTests = append(allTests, TestResult{
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

func runRawSpec(stdout io.Writer, stderr io.Writer, rawSpec RawSpec) (Specs, SpecsResult, bool) {
	parseResult := parser.Parse(rawSpec.NumscriptContent)
	if len(parseResult.Errors) != 0 {
		for _, err := range parseResult.Errors {
			showErr(stderr, rawSpec.NumscriptPath, rawSpec.NumscriptContent, err)
		}
		return Specs{}, SpecsResult{}, false
	}

	var specs Specs
	err := json.Unmarshal(rawSpec.SpecsFileContent, &specs)
	if err != nil {
		_, _ = stderr.Write([]byte(ansi.ColorRed(fmt.Sprintf("\nError: %s.specs.json\n\n", rawSpec.NumscriptPath))))
		_, _ = stderr.Write([]byte(err.Error() + "\n"))
		return Specs{}, SpecsResult{}, false
	}

	out, iErr := Check(parseResult.Value, specs)

	if iErr != nil {
		showErr(stderr, rawSpec.NumscriptPath, rawSpec.NumscriptContent, iErr)
		return Specs{}, SpecsResult{}, false
	}

	if out.Total == 0 {
		_, _ = fmt.Fprintln(stdout, ansi.ColorRed("Empty test suite: "+rawSpec.SpecsPath))
		return Specs{}, SpecsResult{}, false
	} else if out.Failing == 0 {
		testsCount := ansi.ColorBrightBlack(fmt.Sprintf("(%d tests)", out.Total))
		_, _ = fmt.Fprintf(stdout, "%s %s %s\n", ansi.ColorGreen("✓"), rawSpec.NumscriptPath, testsCount)
	} else {
		failedTestsCount := ansi.ColorRed(fmt.Sprintf("%d failed", out.Failing))

		testsCount := ansi.ColorBrightBlack(fmt.Sprintf("(%d tests | %s)", out.Total, failedTestsCount))
		_, _ = fmt.Fprintf(stdout, "%s %s %s\n", ansi.ColorRed("❯"), rawSpec.NumscriptPath, testsCount)

		for _, result := range out.Cases {
			if result.Pass {
				continue
			}

			_, _ = fmt.Fprintf(stdout, "  %s %s\n", ansi.ColorRed("×"), result.It)
		}
	}

	return specs, out, true
}

func ShowDiff(w io.Writer, expected_ any, got_ any) {
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
				_, _ = fmt.Fprintln(w, ansi.ColorGreen("- "+line))
			case diffmatchpatch.DiffInsert:
				_, _ = fmt.Fprintln(w, ansi.ColorRed("+ "+line))
			case diffmatchpatch.DiffEqual:
				_, _ = fmt.Fprintln(w, ansi.ColorBrightBlack("  "+line))
			}
		}
	}
}

func showFailingTestCase(w io.Writer, testResult TestResult) {
	if testResult.Result.Pass {
		return
	}

	specsFilePath := testResult.File
	result := testResult.Result

	_, _ = fmt.Fprint(w, "\n\n")

	failColor := ansi.Compose(ansi.BgRed, ansi.ColorLight, ansi.Bold)
	_, _ = fmt.Fprint(w, failColor(" FAIL "))
	_, _ = fmt.Fprintln(w, ansi.ColorRed(" "+specsFilePath+" > "+result.It))

	//  --- Preconditions
	showGiven := len(result.Balances) != 0 || len(result.Meta) != 0 || len(result.Vars) != 0
	if showGiven {
		_, _ = fmt.Fprintln(w, ansi.Underline("\nGIVEN:"))
	}

	if len(result.Balances) != 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, result.Balances.PrettyPrint())
		_, _ = fmt.Fprintln(w)
	}

	if len(result.Meta) != 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, result.Meta.PrettyPrint())
		_, _ = fmt.Fprintln(w)
	}

	if len(result.Vars) != 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, utils.CsvPrettyMap("Name", "Value", result.Vars))
		_, _ = fmt.Fprintln(w)
	}

	//  --- Outputs
	_, _ = fmt.Fprintln(w, ansi.Underline("\nGOT:"))
	if len(result.Postings) != 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, interpreter.PrettyPrintPostings(result.Postings))
		_, _ = fmt.Fprintln(w)
	} else {
		_, _ = fmt.Fprintln(w)

		_, _ = fmt.Fprintln(w, ansi.ColorBrightBlack("<no postings>"))
		_, _ = fmt.Fprintln(w)
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, ansi.ColorGreen("- Expected"))
	_, _ = fmt.Fprintln(w, ansi.ColorRed("+ Received\n"))

	for _, failedAssertion := range result.FailedAssertions {
		_, _ = fmt.Fprintln(w, ansi.Underline(failedAssertion.Assertion))
		_, _ = fmt.Fprintln(w)
		ShowDiff(w, failedAssertion.Expected, failedAssertion.Got)
		_, _ = fmt.Fprintln(w)
	}
}

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

func printFilesStats(w io.Writer, allTests []TestResult) bool {
	failedTests := utils.Filter(allTests, func(t TestResult) bool {
		return !t.Result.Pass
	})

	testFilesLabel := "Test files"
	testsLabel := "Tests"

	paddedLabel := func(s string) string {
		maxLen := max(len(testFilesLabel), len(testsLabel)) // yeah, ok, this could be hardcoded, I know
		return ansi.ColorBrightBlack(fmt.Sprintf(" %*s ", maxLen, s))
	}

	_, _ = fmt.Fprintln(w)

	// Files stats
	{
		filesCount := len(slices.CompactFunc(allTests, func(t1 TestResult, t2 TestResult) bool {
			return t1.File == t2.File
		}))
		failedTestsFilesCount := len(slices.CompactFunc(failedTests, func(t1 TestResult, t2 TestResult) bool {
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
		_, _ = fmt.Fprint(w, paddedLabel(testFilesLabel)+" "+testFilesUI+" "+totalTestFilesUI)
	}

	_, _ = fmt.Fprintln(w)

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

		_, _ = fmt.Fprintln(w, paddedLabel(testsLabel)+" "+testsUI+" "+totalTestsUI)

		return failedTestsCount == 0
	}

}
