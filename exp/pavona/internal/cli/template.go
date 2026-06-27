package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/pavona/internal/scaffold"
	"github.com/spf13/cobra"
)

type TemplateParams struct {
	Template string `short:"t" descr:"Template source: built-in name or path to template directory" default:"" optional:"true"`
	Output   string `short:"o" descr:"Output directory (default: current directory)" default:"" optional:"true"`
	Name     string `short:"n" descr:"Project name (skips the name prompt if provided)" default:"" optional:"true"`
	Quiet    bool   `short:"q" descr:"Non-interactive mode — use defaults" optional:"true"`
	List     bool   `short:"l" descr:"List available built-in templates" optional:"true"`
}

func TemplateCmd() boa.CmdT[TemplateParams] {
	return boa.CmdT[TemplateParams]{
		Use:   "pavona",
		Short: "A cookiecutter-inspired template engine",
		Long: `Pavona hydrates templates — point it at a template directory
(or use a built-in), answer a few questions, and get a fully
rendered project in seconds.

Built-in templates:
  tool     Go CLI tool with cobra subcommands and BDD tests
  lib      Minimal Go library module with test helpers
  site     Static site with Markdown or org-mode content
  tui      Terminal UI app using bubbletea
  app      Full-stack web app with templ, SQLite, HTMX, Tailwind
  agent    NATS Agent Protocol service

Examples:
  pavona -t tool -o ./my-cli
  pavona -t /path/to/template -o ./my-project
  pavona -t tool -o ./my-cli --name my-cli -q
  pavona -l`,
		RunFunc: RunTemplate,
	}
}

func RunTemplate(p *TemplateParams, cmd *cobra.Command, args []string) {
	// --list flag
	if p.List {
		templates := scaffold.ListBuiltin()
		if len(templates) == 0 {
			fmt.Fprintln(os.Stderr, "No built-in templates available.")
			return
		}
		fmt.Println("Built-in templates:")
		for _, t := range templates {
			desc := t.Description
			if desc == "" {
				desc = "(no description)"
			}
			fmt.Printf("  %-8s  %s\n", t.Name, desc)
		}
		return
	}

	// -t is required when not listing
	if p.Template == "" {
		cmd.Help()
		os.Exit(1)
	}

	// Resolve template source
	templateDir, err := scaffold.Resolve(p.Template)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if !strings.HasPrefix(err.Error(), "template ") {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "\nAvailable built-in templates:")
		for _, t := range scaffold.ListBuiltin() {
			fmt.Fprintf(os.Stderr, "  %s\n", t.Name)
		}
		os.Exit(1)
	}

	// Parse config
	cfg, err := scaffold.ParseConfig(templateDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config.cue: %v\n", err)
		os.Exit(1)
	}

	// Pre-fill name from --name flag if provided
	if p.Name != "" {
		for i, v := range cfg.Variables {
			if v.Name == "project_name" {
				cfg.Variables[i].Default = p.Name
			}
		}
	}

	// Prompt for variables
	values := scaffold.PromptForVariables(cfg.Variables, p.Quiet)

	// Validate required values
	var missing []string
	for _, v := range cfg.Variables {
		if v.Required {
			val := values[v.Name]
			if strings.TrimSpace(val) == "" {
				missing = append(missing, v.Prompt)
			}
		}
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "Error: required values missing: %s\n", strings.Join(missing, ", "))
		os.Exit(1)
	}

	// Determine output directory
	outputDir := p.Output
	if outputDir == "" {
		outputDir = values["project_name"]
	}
	if outputDir == "" {
		outputDir = "output"
	}

	// Resolve relative paths to absolute
	if !filepath.IsAbs(outputDir) {
		cwd, _ := os.Getwd()
		outputDir = filepath.Join(cwd, outputDir)
	}

	// Hydrate
	if err := scaffold.Hydrate(templateDir, outputDir, values); err != nil {
		fmt.Fprintf(os.Stderr, "Error hydrating template: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Created project at %s\n", outputDir)
}
