package cmd

import (
	"github.com/PagoPlus/numscript-wasm/internal/lsp"

	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:    "lsp",
	Short:  "Run the lsp server",
	Long:   "Run the lsp server. This command is usually meant to be used for editors integration.",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		lsp.RunServer(lsp.ServerArgs[lsp.State]{
			InitialState: lsp.InitialState(),
			Handler:      lsp.Handle,
		})
	},
}
