package cmd

import (
	"fmt"
	"os"

	"github.com/bolocera/pavona/scaffold"
	"github.com/spf13/cobra"
)

func NewNewCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "new <type> <name>",
		Short: "Scaffold a new project",
		Long: `Scaffold a new project of the given type.

Types:
  tool      CLI tool with subcommand routing
  lib       Go library with test helpers and BDD scaffold
  site      Static site with Markdown or org-mode content
  tui       Terminal UI with bubbletea components
  app       Full-stack web app with templ, HTMX, Tailwind
  agent     NATS Agent Protocol service

Examples:
  pavona new tool gh-deploy
  pavona new site blog --format org
  pavona new agent triagebot`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectType := args[0]
			name := args[1]

			dir := name
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				return fmt.Errorf("directory %q already exists", dir)
			}

			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}

			opts := scaffold.Options{
				Name:   name,
				Dir:    dir,
				Format: format,
			}

			if err := scaffold.Generate(projectType, opts); err != nil {
				os.RemoveAll(dir)
				return fmt.Errorf("scaffolding %s: %w", projectType, err)
			}

			fmt.Fprintf(os.Stderr, "Created %s at ./%s\n", projectType, dir)
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "markdown", "Content format for site type (markdown or org)")

	return cmd
}
