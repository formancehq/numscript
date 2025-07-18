package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/spf13/cobra"
)

func readSpecFile(path string) (specs_format.RawSpec, error) {
	numscriptFileName := strings.TrimSuffix(path, ".specs.json")

	numscriptContent, err := os.ReadFile(numscriptFileName)
	if err != nil {
		return specs_format.RawSpec{}, nil
	}

	specsFileContent, err := os.ReadFile(path)
	if err != nil {
		return specs_format.RawSpec{}, err
	}

	return specs_format.RawSpec{
		NumscriptPath:    numscriptFileName,
		SpecsPath:        path,
		NumscriptContent: string(numscriptContent),
		SpecsFileContent: specsFileContent,
	}, nil
}

func readSpecsFiles() ([]specs_format.RawSpec, error) {
	var specs []specs_format.RawSpec

	for _, root := range opts.paths {
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

type testArgs struct {
	paths []string
}

var opts = testArgs{}

func runTestCmd() {
	files, err := readSpecsFiles()
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
		return
	}

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
