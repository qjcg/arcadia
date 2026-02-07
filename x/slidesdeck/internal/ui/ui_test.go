package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/charmbracelet/arcadia/x/slidesdeck/internal/parser"
)

func TestSearchPaletteStructure(t *testing.T) {
	component := SearchPalette()
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render SearchPalette: %v", err)
	}
	rendered := buf.String()

	// Test basic structural elements
	structuralElements := []string{
		"Search slides...",
		"x-ref=\"slideSearch\"",
		"x-model=\"searchQuery\"",
		"@input=\"performSearch",
		"x-for=\"(result, index) in searchResults\"",
		"search-results",
		"selectedPaletteIndex",
		"ring-2 ring-primary/50",
		"badge badge-sm badge-outline",
		"font-bold text-xl",
		"text-sm opacity-60 line-clamp-1",
		"'#' + (result.id + 1)",
		"x-text=\"result.title\"",
		"x-text=\"result.content\"",
		"x-show=\"searchQuery && searchResults.length === 0\"",
		"No slides found",
		"↑↓ Navigate",
		"↵ Jump to slide",
		"Esc to close",
	}

	for _, element := range structuralElements {
		if !strings.Contains(rendered, element) {
			t.Errorf("SearchPalette missing expected structural element: %q", element)
		}
	}

	// Test x-show conditions for empty state
	if !strings.Contains(rendered, "searchQuery && searchResults.length === 0") {
		t.Error("Empty state x-show condition missing")
	}

	// Test highlighting classes
	highlightClass := "bg-primary/10 ring-2 ring-primary/50"
	if !strings.Contains(rendered, highlightClass) {
		t.Errorf("Missing highlight class: %s", highlightClass)
	}

	// Test result item structure
	if !strings.Contains(rendered, "jumpToSlide(result.id)") {
		t.Error("Missing jumpToSlide click handler")
	}

	// Test mouse enter handler for navigation
	if !strings.Contains(rendered, "@mouseenter=\"selectedPaletteIndex = index\"") {
		t.Error("Missing mouse enter handler for palette navigation")
	}
}

func TestSearchPaletteInteractions(t *testing.T) {
	// Test different result states rendering
	testCases := []struct {
		name     string
		contains []string
		missing  []string
	}{
		{
			name: "empty query shows all text",
			contains: []string{
				"bg-black/50 backdrop-blur-sm",
				"absolute inset-0",
				"x-show=\"showSearch\"",
				"@keydown.escape.window=\"showSearch = false\"",
				"bg-base-100 rounded-2xl shadow-2xl",
				"max-w-2xl",
				"Search slides...",
			},
		},
		{
			name: "search functionality complete",
			contains: []string{
				"px-6 py-4 rounded-xl hover:bg-base-200 cursor-pointer",
				"selectedPaletteIndex === index ? 'bg-primary/10 ring-2 ring-primary/50' : ''",
				"@click=\"jumpToSlide(result.id)\"",
				"@mouseenter=\"selectedPaletteIndex = index\"",
			},
		},
		{
			name: "search result content preview",
			contains: []string{
				"flex-1 min-w-0",
				"badge badge-sm badge-outline opacity-50 font-mono",
				"<h3 class=\"font-bold text-xl truncate\"",
				"p class=\"text-sm opacity-60 line-clamp-1\"",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			component := SearchPalette()
			var buf bytes.Buffer
			if err := component.Render(context.Background(), &buf); err != nil {
				t.Fatalf("failed to render SearchPalette: %v", err)
			}
			rendered := buf.String()

			for _, expected := range tc.contains {
				if !strings.Contains(rendered, expected) {
					t.Errorf("Expected string not found: %q", expected)
				}
			}

			for _, notExpected := range tc.missing {
				if strings.Contains(rendered, notExpected) {
					t.Errorf("Unexpected string found: %q", notExpected)
				}
			}
		})
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
			name:      "Theme Palette structural elements",
			component: ThemePalette(),
			contains: []string{
				"Search themes...",
				"x-ref=\"themeSearch\"",
				"theme-list",
				"Esc to close",
			},
		},
		{
			name:      "Search Palette structural elements",
			component: SearchPalette(),
			contains: []string{
				"Search slides...",
				"x-ref=\"slideSearch\"",
				"search-results",
				"No slides found matching",
				"Jump to slide",
			},
		},
		{
			name:      "Pause Mode structural elements",
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
