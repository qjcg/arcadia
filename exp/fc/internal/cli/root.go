package cli

import (
	"github.com/spf13/cobra"
)

func PICmd() *cobra.Command {
	cmd := piCmd()
	cmd.AddCommand(ratesCmd())
	cmd.AddCommand(updateCmd())
	return cmd
}
