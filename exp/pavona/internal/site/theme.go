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
	Pages   []PageRef
}

// PageRef references a page in the site for navigation.
type PageRef struct {
	URL   string
	Title string
}

// WrapWithTheme renders content inside the theme layout.
func WrapWithTheme(theme Theme, content []byte, title string, pages []PageRef) ([]byte, error) {
	tmpl, err := template.New("theme").Parse(theme.Layout)
	if err != nil {
		return nil, fmt.Errorf("parsing theme template: %w", err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, LayoutParams{
		Title:   title,
		Content: template.HTML(content),
		Pages:   pages,
	})
	if err != nil {
		return nil, fmt.Errorf("executing theme template: %w", err)
	}

	return []byte(buf.String()), nil
}

func detectTitle(content []byte) string {
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Markdown H1
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimPrefix(trimmed, "# ")
		}
		// Org title keyword
		if strings.HasPrefix(trimmed, "#+TITLE:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "#+TITLE:"))
		}
	}
	return "Site"
}

// BuildWithTheme is like Build but wraps each page in the given theme.
func BuildWithTheme(contentDir, outputDir string, theme Theme) error {
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return fmt.Errorf("reading content dir: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	// Collect all pages for navigation
	type page struct {
		name    string
		content []byte
		title   string
		outPath string
	}
	var pages []page
	var pageRefs []PageRef

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		baseName := strings.TrimSuffix(entry.Name(), ext)
		content, err := os.ReadFile(filepath.Join(contentDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		var html []byte
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".org") {
			html, err = RenderOrg(content, entry.Name())
		} else {
			html, err = RenderMarkdown(content)
		}
		if err != nil {
			return fmt.Errorf("rendering %s: %w", entry.Name(), err)
		}

		title := detectTitle(content)
		outPath := filepath.Join(outputDir, baseName+".html")
		pages = append(pages, page{entry.Name(), html, title, outPath})
		pageRefs = append(pageRefs, PageRef{URL: baseName + ".html", Title: title})
	}

	for _, p := range pages {
		wrapped, err := WrapWithTheme(theme, p.content, p.title, pageRefs)
		if err != nil {
			return fmt.Errorf("wrapping %s: %w", p.name, err)
		}
		if err := os.WriteFile(p.outPath, wrapped, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", p.outPath, err)
		}
	}

	return nil
}

// Keep the original Build for backward compatibility
func Build(contentDir, outputDir string) error {
	return BuildWithTheme(contentDir, outputDir, DefaultTheme)
}
