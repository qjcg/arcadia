package cli

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/pavona/internal/scaffold"
	"github.com/spf13/cobra"
)

func scaffoldProject(projectType, name string, opts scaffold.Options) {
	dir := name
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory %q already exists\n", dir)
		os.Exit(1)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	opts.Name = name
	opts.Dir = dir

	if err := scaffold.Generate(projectType, opts); err != nil {
		os.RemoveAll(dir)
		fmt.Fprintf(os.Stderr, "Error scaffolding %s: %v\n", projectType, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Created %s at ./%s\n", projectType, dir)
}

func NewCmd() boa.CmdT[boa.NoParams] {
	type toolParams struct {
		Name string `positional:"true" descr:"Project name"`
	}
	type libParams struct {
		Name string `positional:"true" descr:"Project name"`
	}
	type siteParams struct {
		Name   string   `positional:"true" descr:"Project name"`
		Format string   `short:"f" descr:"Content format (markdown or org)" default:"markdown" optional:"true"`
		Pages  []string `short:"p" descr:"Pages to scaffold (comma-separated paths, supports brace expansion like services/{foo,bar})" optional:"true"`
	}
	type tuiParams struct {
		Name string `positional:"true" descr:"Project name"`
	}
	type appParams struct {
		Name string `positional:"true" descr:"Project name"`
	}
	type agentParams struct {
		Name string `positional:"true" descr:"Project name"`
	}

	return boa.CmdT[boa.NoParams]{
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
		SubCmds: boa.SubCmds(
			boa.CmdT[toolParams]{
				Use:   "tool",
				Short: "Scaffold a CLI tool",
				RunFunc: func(p *toolParams, cmd *cobra.Command, args []string) {
					scaffoldProject("tool", p.Name, scaffold.Options{})
				},
			},
			boa.CmdT[libParams]{
				Use:   "lib",
				Short: "Scaffold a Go library",
				RunFunc: func(p *libParams, cmd *cobra.Command, args []string) {
					scaffoldProject("lib", p.Name, scaffold.Options{})
				},
			},
			boa.CmdT[siteParams]{
				Use:   "site",
				Short: "Scaffold a static site",
				Long: `Scaffold a static site with Markdown or org-mode content.

Examples:
  pavona new site blog
  pavona new site blog --format org
  pavona new site docs --format org --pages "about,services/{foo,bar}"`,
				RunFunc: func(p *siteParams, cmd *cobra.Command, args []string) {
					opts := scaffold.Options{
						Format: p.Format,
						Pages:  p.Pages,
					}
					scaffoldProject("site", p.Name, opts)
				},
			},
			boa.CmdT[tuiParams]{
				Use:   "tui",
				Short: "Scaffold a TUI app",
				RunFunc: func(p *tuiParams, cmd *cobra.Command, args []string) {
					scaffoldProject("tui", p.Name, scaffold.Options{})
				},
			},
			boa.CmdT[appParams]{
				Use:   "app",
				Short: "Scaffold a web app",
				RunFunc: func(p *appParams, cmd *cobra.Command, args []string) {
					scaffoldProject("app", p.Name, scaffold.Options{})
				},
			},
			boa.CmdT[agentParams]{
				Use:   "agent",
				Short: "Scaffold a NATS agent",
				RunFunc: func(p *agentParams, cmd *cobra.Command, args []string) {
					scaffoldProject("agent", p.Name, scaffold.Options{})
				},
			},
		),
	}
}
