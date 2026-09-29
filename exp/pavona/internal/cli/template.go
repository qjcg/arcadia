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
	Template string
	Output   string `short:"o" descr:"Output directory (default: current directory)" default:"" optional:"true"`
	Name     string `short:"n" descr:"Project name (skips the name prompt if provided)" default:"" optional:"true"`
	Quiet    bool   `short:"q" descr:"Non-interactive mode — use defaults" optional:"true"`
}

type NewParams struct {
	Output string `short:"o" descr:"Output directory (default: derived from project name)" default:"" optional:"true"`
	Name   string `short:"n" descr:"Project name (skips the name prompt if provided)" default:"" optional:"true"`
	Quiet  bool   `short:"q" descr:"Non-interactive mode — use defaults" optional:"true"`
}

func NewCmd() *cobra.Command {
	cmd := boa.CmdT[NewParams]{
		Use:   "new <template> [name]",
		Short: "Create a project from a template",
		Args:  cobra.RangeArgs(1, 2),
		RunFunc: func(p *NewParams, cmd *cobra.Command, args []string) {
			params := &TemplateParams{
				Template: args[0],
				Output:   p.Output,
				Name:     p.Name,
				Quiet:    p.Quiet,
			}
			if len(args) == 2 {
				if params.Output == "" {
					params.Output = args[1]
				}
				if params.Name == "" {
					params.Name = args[1]
				}
			}
			RunTemplate(params, cmd, nil)
		},
	}.ToCobra()
	cmd.ValidArgsFunction = func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var matches []string
		for _, template := range scaffold.ListBuiltin() {
			if strings.HasPrefix(template.Name, toComplete) {
				matches = append(matches, fmt.Sprintf("%s\t%s", template.Name, template.Description))
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	}
	return cmd
}

func ListCmd() *cobra.Command {
	return boa.CmdT[struct{}]{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available templates",
		Args:    cobra.NoArgs,
		RunFunc: func(_ *struct{}, _ *cobra.Command, _ []string) {
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
				fmt.Printf("  %-12s %s\n", t.Name, desc)
			}
		},
	}.ToCobra()
}

func RunTemplate(p *TemplateParams, cmd *cobra.Command, args []string) {
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
