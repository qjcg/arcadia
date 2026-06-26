package cmd

import (
	"github.com/spf13/cobra"
)

func NewRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <type> <name>",
		Short: "Remove a component from the project",
		Long: `Remove a previously added component.

Types:
  handler   Remove an HTTP handler and its route
  feature   Remove a gherkin feature file and its step definitions
  tool      Remove a CLI subcommand
  page      Remove a content page
  stream    Remove a JetStream stream definition
  job       Remove a background worker
  migration Remove a database migration`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement remove logic per type
			return nil
		},
	}

	return cmd
}
