package scaffold

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/lib/go.mod.tmpl
var libGoModTmpl string

//go:embed templates/lib/lib.go.tmpl
var libLibTmpl string

//go:embed templates/lib/lib_test.go.tmpl
var libTestTmpl string

//go:embed templates/lib/Taskfile.yaml.tmpl
var libTaskfileTmpl string

//go:embed templates/lib/features/.gitkeep
var libFeaturesGitkeep string

type LibGenerator struct{}

func (g LibGenerator) Generate(opts Options) error {
	files := map[string]string{
		"go.mod":            libGoModTmpl,
		"lib.go":            libLibTmpl,
		"lib_test.go":       libTestTmpl,
		"Taskfile.yaml":     libTaskfileTmpl,
		".gitignore":        "# library\n/bin/\n",
		"features/.gitkeep": libFeaturesGitkeep,
	}
	return writeTemplates(opts, files)
}

//go:embed templates/site/content/index.md.tmpl
var siteIndexMdTmpl string

//go:embed templates/site/content/index.org.tmpl
var siteIndexOrgTmpl string

//go:embed templates/site/features/.gitkeep
var siteFeaturesGitkeep string

type SiteGenerator struct{}

func (g SiteGenerator) Generate(opts Options) error {
	var contentTmpl string
	contentName := "content/index.md"
	if opts.Format == "org" {
		contentTmpl = siteIndexOrgTmpl
		contentName = "content/index.org"
	} else {
		contentTmpl = siteIndexMdTmpl
	}

	files := map[string]string{
		contentName:         contentTmpl,
		"features/.gitkeep": siteFeaturesGitkeep,
	}
	return writeTemplates(opts, files)
}

//go:embed templates/tui/main.go.tmpl
var tuiMainTmpl string

//go:embed templates/tui/go.mod.tmpl
var tuiGoModTmpl string

//go:embed templates/tui/Taskfile.yaml.tmpl
var tuiTaskfileTmpl string

//go:embed templates/tui/gitignore.tmpl
var tuiGitignoreTmpl string

//go:embed templates/tui/features/.gitkeep
var tuiFeaturesGitkeep string

type TuiGenerator struct{}

func (g TuiGenerator) Generate(opts Options) error {
	files := map[string]string{
		"main.go":           tuiMainTmpl,
		"go.mod":            tuiGoModTmpl,
		"Taskfile.yaml":     tuiTaskfileTmpl,
		".gitignore":        tuiGitignoreTmpl,
		"features/.gitkeep": tuiFeaturesGitkeep,
	}
	return writeTemplates(opts, files)
}

//go:embed templates/app/main.go.tmpl
var appMainTmpl string

//go:embed templates/app/go.mod.tmpl
var appGoModTmpl string

//go:embed templates/app/features/.gitkeep
var appFeaturesGitkeep string

type AppGenerator struct{}

func (g AppGenerator) Generate(opts Options) error {
	files := map[string]string{
		"main.go":           appMainTmpl,
		"go.mod":            appGoModTmpl,
		"features/.gitkeep": appFeaturesGitkeep,
	}
	return writeTemplates(opts, files)
}

//go:embed templates/agent/main.go.tmpl
var agentMainTmpl string

//go:embed templates/agent/go.mod.tmpl
var agentGoModTmpl string

//go:embed templates/agent/features/.gitkeep
var agentFeaturesGitkeep string

type AgentGenerator struct{}

func (g AgentGenerator) Generate(opts Options) error {
	files := map[string]string{
		"main.go":           agentMainTmpl,
		"go.mod":            agentGoModTmpl,
		"features/.gitkeep": agentFeaturesGitkeep,
	}
	return writeTemplates(opts, files)
}

func writeTemplates(opts Options, files map[string]string) error {
	for path, content := range files {
		fullPath := filepath.Join(opts.Dir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}

		tmpl, err := template.New(path).Parse(content)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", path, err)
		}

		f, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("creating file %s: %w", path, err)
		}

		if err := tmpl.Execute(f, opts); err != nil {
			f.Close()
			return fmt.Errorf("executing template %s: %w", path, err)
		}
		f.Close()
	}

	return nil
}
