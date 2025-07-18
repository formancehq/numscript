package cmd

import (
	"os"

	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/spf13/cobra"
)

type testArgs struct {
	paths []string
}

var opts = testArgs{}

func runTestCmd() {
	files, err := specs_format.ReadSpecsFiles(opts.paths)
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
