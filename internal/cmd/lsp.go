package cmd

import (
	"github.com/formancehq/numscript/internal/lsp"

	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:    "lsp",
	Short:  "Run the lsp server",
	Long:   "Run the lsp server. This command is usually meant to be used for editors integration.",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := lsp.RunServer()
		if err != nil {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			return err
		}

		return nil
	},
}
