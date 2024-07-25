package cmd

import (
	"numscript/lsp"

	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "run the lsp server",
	Run: func(cmd *cobra.Command, args []string) {
		lsp.RunServer(lsp.ServerArgs[lsp.State]{
			InitialState: lsp.InitialState(),
			Handler:      lsp.Handle,
		})
	},
}
