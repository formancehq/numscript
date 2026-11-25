package analysis_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
)

// Simulates a user typing each program of the corpus, one character at a time:
// the analysis of every intermediate (and mostly invalid) program must not panic.
func TestAnalysisOfIncompleteProgramsDoesNotPanic(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	for path, src := range numscriptCorpus(t) {
		for end := range len(src) + 1 {
			partial := src[:end]
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic analysing a prefix of %s: %v\n---\n%s", path, r, partial)
					}
				}()
				analyseEverywhere(partial)
			}()
		}
	}
}

func FuzzAnalysis(f *testing.F) {
	for _, src := range numscriptCorpus(f) {
		f.Add(src)
	}
	f.Fuzz(func(t *testing.T, src string) {
		analyseEverywhere(src)
	})
}

// Runs the whole analysis, and the position-based queries on each line of the program
func analyseEverywhere(src string) {
	res := analysis.CheckSource(src)
	res.GetSymbols()

	lines := strings.Split(src, "\n")
	for line, text := range lines {
		for _, char := range []int{0, len(text) / 2, len(text)} {
			pos := parser.Position{Line: line, Character: char}
			analysis.HoverOn(res.Program, pos)
			analysis.GotoDefinition(res.Program, pos, res)
			analysis.Complete(src, pos)
			analysis.SignatureHelpAt(src, pos)
		}
	}
}

func numscriptCorpus(t testing.TB) map[string]string {
	t.Helper()
	corpus := map[string]string{}
	err := filepath.WalkDir("../interpreter/testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".num") {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		corpus[path] = string(content)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return corpus
}
