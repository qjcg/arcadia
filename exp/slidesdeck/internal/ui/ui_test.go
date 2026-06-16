package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/qjcg/arcadia/exp/slidesdeck/internal/parser"
)

func TestSearchPaletteStructure(t *testing.T) {
	component := SearchPalette()
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render SearchPalette: %v", err)
	}
	rendered := buf.String()

	// Test functional elements and integration points
	functionalElements := []string{
		"Search slides...",
		"x-model=\"searchQuery\"",
		"x-for", // Ensure list rendering is present
		"search-results",
		"No matches found",
		"Navigate",
		"Select",
		"Close",
	}

	for _, element := range functionalElements {
		if !strings.Contains(rendered, element) {
			t.Errorf("SearchPalette missing expected functional element: %q", element)
		}
	}

	// Verify jump functionality is wired
	if !strings.Contains(rendered, "jumpToSlide") {
		t.Error("Missing jumpToSlide handler")
	}
}

func TestUI(t *testing.T) {
	deck := &parser.Deck{
		Title: "Test Deck",
		Slides: []parser.Slide{
			{ID: 0, Title: "Slide 1", Content: "<h1>Slide 1 Content</h1>"},
			{ID: 1, Title: "Slide 2", Content: "<h2>Slide 2 Content</h2>"},
		},
	}

	tests := []struct {
		name      string
		component templ.Component
		contains  []string
	}{
		{
			name:      "Layout integration",
			component: Layout(deck, "dark", "modern", "body { color: red; }", ".custom { }", "console.log('hi');"),
			contains: []string{
				"<title>Test Deck</title>",
				"data-theme=\"dark\"",
				"body { color: red; }",
				".custom { }",
				"console.log('hi');",
				"Slide 1 Content",
				"Slide 2 Content",
				"x-data=\"slideshow\"",
			},
		},
		{
			name:      "Theme Palette",
			component: ThemePalette(),
			contains: []string{
				"Search themes...",
				"theme-list",
			},
		},
		{
			name:      "Pause Mode",
			component: PauseMode(),
			contains: []string{
				"Countdown Message",
				"x-text=\"timeRemaining\"",
				"START",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tt.component.Render(context.Background(), &buf); err != nil {
				t.Fatalf("failed to render %s: %v", tt.name, err)
			}
			rendered := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(rendered, s) {
					t.Errorf("%s: rendered output missing expected string %q", tt.name, s)
				}
			}
		})
	}
}

func TestLayoutThemeInjection(t *testing.T) {
	themes := []string{"light", "dark", "cupcake"}
	for _, theme := range themes {
		t.Run(theme, func(t *testing.T) {
			deck := &parser.Deck{Title: "T", Slides: []parser.Slide{{ID: 0, Title: "T", Content: "C"}}}
			component := Layout(deck, theme, "modern", "", "", "")
			var buf bytes.Buffer
			_ = component.Render(context.Background(), &buf)
			if !strings.Contains(buf.String(), "data-theme=\""+theme+"\"") {
				t.Errorf("expected theme %q not found in layout", theme)
			}
		})
	}
}
