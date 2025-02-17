package cmd

import (
	"github.com/formancehq/numscript/internal/lsp"
	"github.com/formancehq/numscript/internal/lsp/language_server"

	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:    "lsp",
	Short:  "Run the lsp server",
	Long:   "Run the lsp server. This command is usually meant to be used for editors integration.",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		language_server.RunServer(language_server.ServerArgs[lsp.State]{
			InitialState: lsp.InitialState,
			Handler:      lsp.Handle,
		})
	},
}
