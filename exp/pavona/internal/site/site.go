package site

import (
	"bytes"
	"fmt"

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
