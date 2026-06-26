package cli

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/pavona/internal/scaffold"
	"github.com/spf13/cobra"
)

type NewParams struct {
	Type   string   `positional:"true" descr:"Project type (tool, lib, site, tui, app, agent)"`
	Name   string   `positional:"true" descr:"Project name"`
	Format string   `short:"f" descr:"Content format for site type (markdown or org)" default:"markdown" optional:"true"`
	Pages  []string `short:"p" descr:"Pages to scaffold (comma-separated paths, supports brace expansion like services/{foo,bar})" optional:"true"`
}

func NewCmd() boa.CmdT[NewParams] {
	return boa.CmdT[NewParams]{
		Use:   "new",
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
		RunFunc: func(p *NewParams, cmd *cobra.Command, args []string) {
			dir := p.Name
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error: directory %q already exists\n", dir)
				os.Exit(1)
			}

			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
				os.Exit(1)
			}

			opts := scaffold.Options{
				Name:   p.Name,
				Dir:    dir,
				Format: p.Format,
				Pages:  p.Pages,
			}

			if err := scaffold.Generate(p.Type, opts); err != nil {
				os.RemoveAll(dir)
				fmt.Fprintf(os.Stderr, "Error scaffolding %s: %v\n", p.Type, err)
				os.Exit(1)
			}

			fmt.Fprintf(os.Stderr, "Created %s at ./%s\n", p.Type, dir)
		},
	}
}
