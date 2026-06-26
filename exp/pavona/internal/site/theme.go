package site

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

//go:embed layouts/default.gohtml
var defaultLayout string

// Theme wraps rendered content in a complete HTML page layout.
type Theme struct {
	Name   string
	Layout string // Go template string for the HTML shell
	Dir    string // path to theme directory in the project
}

// DefaultTheme is the built-in theme used when no custom theme is specified.
var DefaultTheme = Theme{
	Name:   "default",
	Layout: defaultLayout,
}

// LayoutParams holds the data passed to the theme template.
type LayoutParams struct {
	Title   string
	Content template.HTML
	Nav     *TreeNode
}

// WrapWithTheme renders content inside the theme layout.
func WrapWithTheme(theme Theme, content []byte, title string, nav *TreeNode) ([]byte, error) {
	tmpl, err := template.New("theme").Funcs(template.FuncMap{
		"hasChildren": func(n *TreeNode) bool { return len(n.Children) > 0 },
		"add":         func(a, b int) int { return a + b },
	}).Parse(theme.Layout)
	if err != nil {
		return nil, fmt.Errorf("parsing theme template: %w", err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, LayoutParams{
		Title:   title,
		Content: template.HTML(content),
		Nav:     nav,
	})
	if err != nil {
		return nil, fmt.Errorf("executing theme template: %w", err)
	}

	return []byte(buf.String()), nil
}

// BuildWithTheme walks contentDir, renders all files, wraps them in the theme.
func BuildWithTheme(contentDir, outputDir string, theme Theme) error {
	flat, nav, err := BuildPageTree(contentDir)
	if err != nil {
		return fmt.Errorf("building page tree: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	for _, p := range flat {
		if p.Draft {
			continue
		}

		outPath := filepath.Join(outputDir, p.URL)

		// Create parent directories for section pages
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		wrapped, err := WrapWithTheme(theme, p.HTML, p.Title, nav)
		if err != nil {
			return fmt.Errorf("wrapping %s: %w", p.URL, err)
		}
		if err := os.WriteFile(outPath, wrapped, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
	}

	return nil
}

// Build wraps BuildWithTheme using the default theme.
func Build(contentDir, outputDir string) error {
	return BuildWithTheme(contentDir, outputDir, DefaultTheme)
}
