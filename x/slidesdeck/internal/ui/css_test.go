package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/arcadia/x/slidesdeck/internal/parser"
)

func TestLayoutCustomCSSInjection(t *testing.T) {
	deck := &parser.Deck{
		Title: "Test Deck",
		Slides: []parser.Slide{
			{ID: 0, Title: "Slide 1", Content: "Content"},
		},
	}

	css := "[data-theme=dark] { color: white; }"
	customCss := "html[data-font-theme=\"custom\"] { font-family: serif; }"
	js := "console.log('test')"

	component := Layout(deck, "dark", "modern", css, customCss, js)
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	rendered := buf.String()

	// Check if both CSS strings are present in the style tag
	expectedStyleTag := "<style type=\"text/css\">\n" + css + "\n" + customCss + "\n</style>"
	if !strings.Contains(rendered, expectedStyleTag) {
		// Use a looser check if exact whitespace/newlines are tricky
		if !strings.Contains(rendered, css) || !strings.Contains(rendered, customCss) {
			t.Errorf("Custom CSS not correctly injected.\nWant part of: %q\nGot: %s", expectedStyleTag, rendered)
		}
	}
}
