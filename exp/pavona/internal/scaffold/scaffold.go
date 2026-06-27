package scaffold

import (
	"fmt"
)

// Options configures project scaffolding.
type Options struct {
	Name        string   // project name (used as dir name, Go module path)
	Dir         string   // output directory
	Format      string   // markdown or org (for site type)
	PackageName string   // sanitized Go package name (no hyphens)
	Pages       []string // page paths to scaffold (for site type)
	Demo        bool     // scaffold the demo variant (URL shortener for app type)
}

// Generator creates a project of a specific type.
type Generator interface {
	Generate(opts Options) error
}

var generators = map[string]Generator{
	"tool":  ToolGenerator{},
	"lib":   LibGenerator{},
	"site":  SiteGenerator{},
	"tui":   TuiGenerator{},
	"app":   AppGenerator{},
	"agent": AgentGenerator{},
}

// Generate scaffolds a new project of the given type.
func Generate(projectType string, opts Options) error {
	gen, ok := generators[projectType]
	if !ok {
		types := make([]string, 0, len(generators))
		for t := range generators {
			types = append(types, t)
		}
		return fmt.Errorf("unknown project type %q (valid: %v)", projectType, types)
	}

	if err := gen.Generate(opts); err != nil {
		return err
	}

	// Apply demo overlay for app type
	if opts.Demo && projectType == "app" {
		demoGen := AppDemoGenerator{}
		return demoGen.Generate(opts)
	}

	return nil
}
