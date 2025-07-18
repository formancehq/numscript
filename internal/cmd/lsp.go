package cmd

import (
	"fmt"
	"os"

	"github.com/formancehq/numscript/internal/lsp"

	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:    "lsp",
	Short:  "Run the lsp server",
	Long:   "Run the lsp server. This command is usually meant to be used for editors integration.",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := lsp.RunServer()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}
