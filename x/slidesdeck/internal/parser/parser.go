package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/niklasfasching/go-org/org"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	goldmarkparser "github.com/yuin/goldmark/parser"
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
		goldmark.WithParserOptions(goldmarkparser.WithAutoHeadingID()),
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

	deck := &Deck{Title: extractTitle(content)}
	if deck.Title == "" {
		deck.Title = "Markdown Presentation"
	}

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

	deck := &Deck{Title: extractOrgTitle(content)}
	if deck.Title == "" {
		deck.Title = "Org Presentation"
	}

	for i, rs := range rawSlides {
		config := org.New()
		doc := config.Parse(strings.NewReader("#+OPTIONS: toc:nil\n"+rs+"\n"), "")

		writer := org.NewHTMLWriter()
		writer.HighlightCodeBlock = func(code, lang string, inline bool, params map[string]string) string {
			lexer := lexers.Get(lang)
			if lexer == nil {
				lexer = lexers.Fallback
			}
			style := styles.Get("dracula")
			if style == nil {
				style = styles.Fallback
			}
			formatter := chromahtml.New(chromahtml.WithLineNumbers(true), chromahtml.TabWidth(2))

			iterator, err := lexer.Tokenise(nil, code)
			if err != nil {
				return code
			}

			var buf bytes.Buffer
			if err := formatter.Format(&buf, style, iterator); err != nil {
				return code
			}
			return buf.String()
		}

		html, err := doc.Write(writer)
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

	inCodeBlock := false
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
		}

		isSeparator := false
		if !inCodeBlock {
			if sep != "" {
				isSeparator = strings.TrimSpace(line) == sep
			} else {
				isSeparator = strings.HasPrefix(line, "# ") || (strings.TrimSpace(line) == "---")
			}
		}

		if isSeparator && currentSlide.Len() > 0 {
			content := currentSlide.String()
			trimmed := strings.TrimSpace(content)
			// Check if it's just Org metadata or blank
			isMetadataOrBlank := true
			checkLines := strings.Split(trimmed, "\n")
			for _, l := range checkLines {
				l = strings.TrimSpace(l)
				if l != "" && !strings.HasPrefix(l, "#+") {
					isMetadataOrBlank = false
					break
				}
			}

			if !isMetadataOrBlank {
				slides = append(slides, trimmed)
			}
			currentSlide.Reset()
		}

		if isSeparator && (strings.TrimSpace(line) == "---" || (sep != "" && strings.TrimSpace(line) == sep)) {
			continue
		}

		if currentSlide.Len() > 0 {
			currentSlide.WriteString("\n")
		}
		currentSlide.WriteString(line)
	}

	if currentSlide.Len() > 0 {
		content := currentSlide.String()
		trimmed := strings.TrimSpace(content)
		isMetadataOrBlank := true
		checkLines := strings.Split(trimmed, "\n")
		for _, l := range checkLines {
			l = strings.TrimSpace(l)
			if l != "" && !strings.HasPrefix(l, "#+") {
				isMetadataOrBlank = false
				break
			}
		}
		if !isMetadataOrBlank {
			slides = append(slides, trimmed)
		}
	}

	return slides
}

func splitOrg(content string, sep string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")

	var slides []string
	var currentSlide strings.Builder

	inBlock := false
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(strings.ToLower(line))

		if strings.HasPrefix(trimmed, "#+begin_") || strings.HasPrefix(trimmed, "#+end_") {
			if strings.HasPrefix(trimmed, "#+begin_") {
				inBlock = true
			} else {
				inBlock = false
			}
		}

		isSeparator := false
		if !inBlock {
			if sep != "" {
				isSeparator = strings.TrimSpace(line) == sep
			} else {
				isSeparator = strings.HasPrefix(line, "* ") || (strings.TrimSpace(line) == "-----")
			}
		}

		if isSeparator && currentSlide.Len() > 0 {
			content := currentSlide.String()
			trimmed := strings.TrimSpace(content)
			// Check if it's just Org metadata or blank
			isMetadataOrBlank := true
			checkLines := strings.Split(trimmed, "\n")
			for _, l := range checkLines {
				l = strings.TrimSpace(l)
				if l != "" && !strings.HasPrefix(l, "#+") {
					isMetadataOrBlank = false
					break
				}
			}

			if !isMetadataOrBlank {
				slides = append(slides, trimmed)
			}
			currentSlide.Reset()
		}

		if isSeparator && (strings.TrimSpace(line) == "-----" || (sep != "" && strings.TrimSpace(line) == sep)) {
			continue
		}

		if currentSlide.Len() > 0 {
			currentSlide.WriteString("\n")
		}
		currentSlide.WriteString(line)
	}

	if currentSlide.Len() > 0 {
		content := currentSlide.String()
		trimmed := strings.TrimSpace(content)
		isMetadataOrBlank := true
		checkLines := strings.Split(trimmed, "\n")
		for _, l := range checkLines {
			l = strings.TrimSpace(l)
			if l != "" && !strings.HasPrefix(l, "#+") {
				isMetadataOrBlank = false
				break
			}
		}
		if !isMetadataOrBlank {
			slides = append(slides, trimmed)
		}
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
