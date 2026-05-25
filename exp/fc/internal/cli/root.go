package cli

import (
	"github.com/spf13/cobra"
)

func PicCmd() *cobra.Command {
	cmd := picCmd()
	cmd.AddCommand(ratesCmd())
	cmd.AddCommand(updateCmd())
	return cmd
}
