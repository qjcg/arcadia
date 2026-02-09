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

	component := Layout(deck, "dark", "modern", css, customCss, "")
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	rendered := buf.String()

	// Verify CSS is injected (observable behavior)
	if !strings.Contains(rendered, css) {
		t.Errorf("Base CSS not found")
	}
	if !strings.Contains(rendered, customCss) {
		t.Errorf("Custom CSS not found")
	}
}
