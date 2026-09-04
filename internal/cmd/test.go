package cmd

import (
	"fmt"
	"os"

	"github.com/formancehq/numscript/internal/ansi"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/spf13/cobra"
)

type testArgs struct {
	paths   []string
	migrate bool
}

func runTestCmd(opts testArgs) {
	files, err := specs_format.ReadSpecsFiles(opts.paths)
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
		return
	}

	if opts.migrate {
		migrateSpecsFiles(files)
	}

	pass := specs_format.RunSpecs(os.Stdout, os.Stderr, files)
	if !pass {
		os.Exit(1)
	}
}

// migrateSpecsFiles rewrites any stale specs file to the current format in
// place.
//
// It also updates each migrated file's in-memory content, so the RunSpecs
// call right after sees the already-current version instead of re-reading
// the pre-migration bytes.
func migrateSpecsFiles(files []specs_format.RawSpec) {
	for i, file := range files {
		migrated, changed, err := specs_format.MigrateSpecsContent(file.SpecsFileContent)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, ansi.ColorRed(fmt.Sprintf("✗ cannot migrate %s: %s", file.SpecsPath, err)))
			continue
		}
		if !changed {
			continue
		}

		if err := os.WriteFile(file.SpecsPath, migrated, 0644); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to migrate %s: %s\n", file.SpecsPath, err)
			os.Exit(1)
		}
		files[i].SpecsFileContent = migrated
		_, _ = fmt.Fprintln(os.Stdout, ansi.ColorYellow("↻ migrated "+file.SpecsPath))
	}
}

func getTestCmd() *cobra.Command {
	var opts testArgs

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
			runTestCmd(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.migrate, "migrate", false, "rewrite specs files that use an outdated format to the current one")

	return cmd
}
