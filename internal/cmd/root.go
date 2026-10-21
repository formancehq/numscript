package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type CliOptions struct {
	Version string
}

var rootCmd = &cobra.Command{
	Use:   "numscript",
	Short: "Numscript cli",
	Long:  "Numscript cli",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute(options CliOptions) {
	rootCmd.Version = options.Version

	rootCmd.AddCommand(lspCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(getTestCmd())
	rootCmd.AddCommand(getTestInitCmd())
	rootCmd.AddCommand(getRunCmd())

	// The ir/bytecode tooling tracks an unstable instruction set, so it stays out
	// of --help unless NUMSCRIPT_EXPERIMENTAL_CLI is set. It is always registered
	// and runnable, like lsp and mcp.
	hidden := os.Getenv("NUMSCRIPT_EXPERIMENTAL_CLI") == ""
	for _, experimentalCmd := range []*cobra.Command{getAssembleCmd(), getBytecodeRunCmd()} {
		experimentalCmd.Hidden = hidden
		rootCmd.AddCommand(experimentalCmd)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
