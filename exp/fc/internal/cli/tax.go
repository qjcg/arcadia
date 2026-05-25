package cli

import (
	"github.com/spf13/cobra"
)

func TaxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tax",
		Short: "Tax-related calculations",
	}

	cmd.AddCommand(PicCmd())

	return cmd
}
