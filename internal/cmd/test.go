package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/spf13/cobra"
)

func readSpecsFiles() []specs_format.RawSpec {
	var specs []specs_format.RawSpec

	for _, root := range opts.paths {
		root = strings.TrimSuffix(root, "/")

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if d.IsDir() {
				return nil
			}

			if !strings.HasSuffix(path, ".num.specs.json") {
				return nil
			}

			numscriptFileName := strings.TrimSuffix(path, ".specs.json")

			numscriptContent, err := os.ReadFile(numscriptFileName)
			if err != nil {
				return err
			}

			specsFileContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			specs = append(specs, specs_format.RawSpec{
				NumscriptPath:    numscriptFileName,
				SpecsPath:        path,
				NumscriptContent: string(numscriptContent),
				SpecsFileContent: specsFileContent,
			})

			return nil
		})

		if err != nil {
			_, _ = os.Stderr.Write([]byte(err.Error()))
			os.Exit(1)
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
