package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/formancehq/numscript/internal/ansi"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/spf13/cobra"
)

func readSpecsFiles() []specs_format.RawSpec {
	var specs []specs_format.RawSpec

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

			specs = append(specs, specs_format.RawSpec{
				NumscriptPath:    numscriptFileName,
				SpecsPath:        specsFilePath,
				NumscriptContent: string(numscriptContent),
				SpecsFileContent: specsFileContent,
			})
		}
	}

	return specs
}

type testArgs struct {
	paths []string
}

var opts = testArgs{}

func runTestCmd() {
	files := readSpecsFiles()
	pass := specs_format.RunSpecs(os.Stdout, os.Stderr, files)
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
