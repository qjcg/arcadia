package site

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niklasfasching/go-org/org"
	"github.com/yuin/goldmark"
)

// RenderMarkdown converts markdown content to HTML.
func RenderMarkdown(content []byte) ([]byte, error) {
	md := goldmark.New()
	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return nil, fmt.Errorf("rendering markdown: %w", err)
	}
	return buf.Bytes(), nil
}

// RenderOrg converts org-mode content to HTML.
func RenderOrg(content []byte, path string) ([]byte, error) {
	// Prepend option to disable auto-generated table of contents.
	// Org files should control their own output, not get an implicit nav.
	prefixed := append([]byte("#+OPTIONS: toc:nil\n"), content...)
	doc := org.New().Parse(bytes.NewReader(prefixed), path)
	if doc.Error != nil {
		return nil, fmt.Errorf("parsing org: %w", doc.Error)
	}
	html, err := doc.Write(org.NewHTMLWriter())
	if err != nil {
		return nil, fmt.Errorf("writing org HTML: %w", err)
	}
	return []byte(html), nil
}

// Build scans a content directory, renders all files to HTML in dist/.
// Files with .org extension use go-org; all others use goldmark.
func Build(contentDir, outputDir string) error {
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return fmt.Errorf("reading content dir: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

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

		outPath := filepath.Join(outputDir, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))+".html")
		if err := os.WriteFile(outPath, html, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
	}

	return nil
}
