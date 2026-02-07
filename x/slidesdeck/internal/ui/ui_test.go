package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/charmbracelet/arcadia/x/slidesdeck/internal/parser"
)

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
			name:      "Layout main elements",
			component: Layout(deck, "dark", "body { color: red; }", "console.log('hi');"),
			contains: []string{
				"<title>Test Deck</title>",
				"data-theme=\"dark\"",
				"body { color: red; }",
				"console.log('hi');",
				"<h1>Slide 1 Content</h1>",
				"<h2>Slide 2 Content</h2>",
				"x-data=\"slideshow\"",
			},
		},
		{
			name:      "Theme Palette elements",
			component: ThemePalette(),
			contains: []string{
				"Search themes...",
				"x-ref=\"themeSearch\"",
				"theme-list",
				"Esc to close",
			},
		},
		{
			name:      "Search Palette elements",
			component: SearchPalette(),
			contains: []string{
				"Search slides...",
				"x-ref=\"slideSearch\"",
				"No slides found",
				"Jump to slide",
			},
		},
		{
			name:      "Pause Mode elements",
			component: PauseMode(),
			contains: []string{
				"Break message...",
				"x-text=\"timeRemaining\"",
				"Start Break",
				"x-model=\"pauseMinutes\"",
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
	themes := []string{"light", "dark", "cupcake", "cyberpunk"}
	for _, theme := range themes {
		t.Run(theme, func(t *testing.T) {
			deck := &parser.Deck{Title: "Title", Slides: []parser.Slide{{ID: 0, Title: "T", Content: "C"}}}
			component := Layout(deck, theme, "", "")
			var buf bytes.Buffer
			_ = component.Render(context.Background(), &buf)
			expected := "data-theme=\"" + theme + "\""
			if !strings.Contains(buf.String(), expected) {
				t.Errorf("expected theme %q not found in layout", theme)
			}
		})
	}
}
