package cmd

import (
	"github.com/spf13/cobra"
)

func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <type> <name>",
		Short: "Add a component to the project",
		Long: `Add a component to the current project.

Types:
  handler   Add an HTTP handler with route
  feature   Add a gherkin feature file with godog step definitions
  tool      Add a CLI subcommand
  page      Add a content page (site projects)
  stream    Add a JetStream stream definition (NATS projects)
  job       Add a background worker
  migration Add a database migration`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement add logic per type
			return nil
		},
	}

	return cmd
}
