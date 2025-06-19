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
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(getRunCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
