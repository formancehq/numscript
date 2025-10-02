package cmd

import (
	"github.com/formancehq/numscript/internal/mcp_impl"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:    "mcp",
	Short:  "Run the mcp server",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := mcp_impl.RunServer()
		if err != nil {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			return err
		}

		return nil
	},
}
