package scaffold

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/tool/main.go.tmpl
var toolMainTmpl string

//go:embed templates/tool/go.mod.tmpl
var toolGoModTmpl string

//go:embed templates/tool/Taskfile.yaml.tmpl
var toolTaskfileTmpl string

//go:embed templates/tool/gitignore.tmpl
var toolGitignoreTmpl string

//go:embed templates/tool/features/.gitkeep
var toolFeaturesGitkeep string

type ToolGenerator struct{}

func (g ToolGenerator) Generate(opts Options) error {
	files := map[string]string{
		"main.go":           toolMainTmpl,
		"go.mod":            toolGoModTmpl,
		"Taskfile.yaml":     toolTaskfileTmpl,
		".gitignore":        toolGitignoreTmpl,
		"features/.gitkeep": toolFeaturesGitkeep,
	}

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
