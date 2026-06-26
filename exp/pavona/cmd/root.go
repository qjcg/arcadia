package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pavona",
		Short: "A Go framework that grows with you",
		Long: `Pavona is a Go framework for building CLI tools, libraries,
static sites, TUIs, web apps, and agents. Named after leaf coral
of the Pavona genus: layered, branching, and symbiotic.`,
		SilenceUsage: true,
	}

	cmd.AddCommand(NewNewCmd())
	cmd.AddCommand(NewAddCmd())
	cmd.AddCommand(NewRemoveCmd())

	return cmd
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
