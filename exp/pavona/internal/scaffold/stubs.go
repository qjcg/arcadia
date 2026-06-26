package scaffold

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

//go:embed templates/site/theme/default.templ.tmpl
var siteThemeDefaultTmpl string

//go:embed templates/site/build.go.tmpl
var siteBuildGoTmpl string

//go:embed templates/site/go.mod.tmpl
var siteGoModTmpl string

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
		contentName:           contentTmpl,
		"build.go":            siteBuildGoTmpl,
		"go.mod":              siteGoModTmpl,
		"features/.gitkeep":   siteFeaturesGitkeep,
		"theme/default.templ": siteThemeDefaultTmpl,
	}

	if err := writeTemplates(opts, files); err != nil {
		return err
	}

	return scaffoldPages(opts)
}

// scaffoldPages creates the user-specified content pages from --pages flag.
func scaffoldPages(opts Options) error {
	if len(opts.Pages) == 0 {
		return nil
	}

	// pflag splits on commas, which breaks brace syntax like "services/{foo,bar}".
	// Rejoin the fragments and re-split respecting braces to reconstruct the
	// original user intent, then expand any braces.
	raw := strings.Join(opts.Pages, ",")
	parts := SplitCommasRespectingBraces(raw)

	ext := ".md"
	if opts.Format == "org" {
		ext = ".org"
	}

	for _, raw := range parts {
		expanded := ExpandBraces(raw)
		for _, page := range expanded {
			content := pageContent(page, opts.Format)
			path := "content/" + page + ext
			fullPath := filepath.Join(opts.Dir, path)
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating directory for page %s: %w", page, err)
			}
			if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("writing page %s: %w", page, err)
			}
		}
	}
	return nil
}

// pageContent returns the body content for a scaffolded page.
func pageContent(page, format string) string {
	title := pageTitle(page)
	if format == "org" {
		return "#+TITLE: " + title + "\n#+OPTIONS: toc:nil\n\n* " + title + "\n\nThis is the " + page + " page.\n"
	}
	return "# " + title + "\n\nThis is the " + page + " page.\n"
}

// pageTitle derives a display title from a page path.
// "about" → "About", "services/foo" → "Foo", "index" → "Index"
func pageTitle(page string) string {
	parts := strings.Split(page, "/")
	last := parts[len(parts)-1]
	if len(last) == 0 {
		return "Untitled"
	}
	return strings.ToUpper(last[:1]) + last[1:]
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
	funcMap := template.FuncMap{
		"sanitize": func(s string) string {
			return strings.ReplaceAll(s, "-", "_")
		},
	}

	for path, content := range files {
		fullPath := filepath.Join(opts.Dir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}

		tmpl, err := template.New(path).Funcs(funcMap).Parse(content)
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
