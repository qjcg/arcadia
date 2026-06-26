package scaffold

import (
	"fmt"
)

// Options configures project scaffolding.
type Options struct {
	Name   string
	Dir    string
	Format string // markdown or org (for site type)
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
	return gen.Generate(opts)
}
