package parser

import (
	"fmt"
	"io"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/niklasfasching/go-org/org"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

type Slide struct {
	ID      int
	Title   string
	Content string // HTML content
}

type Deck struct {
	Title  string
	Slides []Slide
}

func ParseMarkdown(r io.Reader) (*Deck, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	content := string(data)
	// Split by first level headings or ---
	rawSlides := splitMarkdown(content)

	md := goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
					chromahtml.WithLinkableLineNumbers(true, ""),
				),
			),
		),
	)

	deck := &Deck{Title: "Markdown Presentation"}
	for i, rs := range rawSlides {
		var buf strings.Builder
		if err := md.Convert([]byte(rs), &buf); err != nil {
			return nil, err
		}

		title := extractTitle(rs)
		if title == "" {
			title = fmt.Sprintf("Slide %d", i+1)
		}

		deck.Slides = append(deck.Slides, Slide{
			ID:      i,
			Title:   title,
			Content: buf.String(),
		})
	}

	return deck, nil
}

func ParseOrg(r io.Reader) (*Deck, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	content := string(data)
	// Simple org splitting by * headings
	rawSlides := splitOrg(content)

	deck := &Deck{Title: "Org Presentation"}
	for i, rs := range rawSlides {
		config := org.New()
		doc := config.Parse(strings.NewReader(rs), "")
		html, err := doc.Write(org.NewHTMLWriter())
		if err != nil {
			return nil, err
		}

		title := extractOrgTitle(rs)
		if title == "" {
			title = fmt.Sprintf("Slide %d", i+1)
		}

		deck.Slides = append(deck.Slides, Slide{
			ID:      i,
			Title:   title,
			Content: html,
		})
	}

	return deck, nil
}

func splitMarkdown(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")

	// Split by ---
	parts := strings.Split(content, "\n---\n")

	var finalBlocks []string
	for _, p := range parts {
		finalBlocks = append(finalBlocks, strings.TrimSpace(p))
	}

	return finalBlocks
}

func splitOrg(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")

	parts := strings.Split(content, "\n* ")
	var finalBlocks []string
	for i, p := range parts {
		block := p
		if i > 0 {
			block = "* " + p
		}
		finalBlocks = append(finalBlocks, strings.TrimSpace(block))
	}
	return finalBlocks
}

func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "# ") {
			return strings.TrimPrefix(l, "# ")
		}
	}
	return ""
}

func extractOrgTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "* ") {
			return strings.TrimPrefix(l, "* ")
		}
	}
	return ""
}
