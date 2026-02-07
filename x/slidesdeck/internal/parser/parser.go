package parser

import (
	"bufio"
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

type Options struct {
	Separator string
}

func ParseMarkdown(r io.Reader, opts Options) (*Deck, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	content := string(data)
	rawSlides := splitMarkdown(content, opts.Separator)

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

func ParseOrg(r io.Reader, opts Options) (*Deck, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	content := string(data)
	rawSlides := splitOrg(content, opts.Separator)

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

func splitMarkdown(content string, sep string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")

	var slides []string
	var currentSlide strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		isSeparator := false
		if sep != "" {
			isSeparator = strings.TrimSpace(line) == sep
		} else {
			// Default logic: # Heading or ---
			isSeparator = strings.HasPrefix(line, "# ") || (strings.TrimSpace(line) == "---")
		}

		if isSeparator && currentSlide.Len() > 0 {
			slides = append(slides, currentSlide.String())
			currentSlide.Reset()
		}

		// If it was a '---' separator, don't include it in content
		if isSeparator && strings.TrimSpace(line) == "---" && sep == "" {
			continue
		}

		// If it was a custom separator, don't include it
		if isSeparator && sep != "" {
			continue
		}

		if currentSlide.Len() > 0 {
			currentSlide.WriteString("\n")
		}
		currentSlide.WriteString(line)
	}

	if currentSlide.Len() > 0 {
		slides = append(slides, currentSlide.String())
	}

	return slides
}

func splitOrg(content string, sep string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")

	var slides []string
	var currentSlide strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		isSeparator := false
		if sep != "" {
			isSeparator = strings.TrimSpace(line) == sep
		} else {
			// Default logic: * Heading or -----
			isSeparator = strings.HasPrefix(line, "* ") || (strings.TrimSpace(line) == "-----")
		}

		if isSeparator && currentSlide.Len() > 0 {
			slides = append(slides, currentSlide.String())
			currentSlide.Reset()
		}

		// If it was a '-----' separator, don't include it in content
		if isSeparator && strings.TrimSpace(line) == "-----" && sep == "" {
			continue
		}

		// If it was a custom separator, don't include it
		if isSeparator && sep != "" {
			continue
		}

		if currentSlide.Len() > 0 {
			currentSlide.WriteString("\n")
		}
		currentSlide.WriteString(line)
	}

	if currentSlide.Len() > 0 {
		slides = append(slides, currentSlide.String())
	}

	return slides
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
