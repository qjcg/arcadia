package parser

import (
	"strings"
	"testing"
)

func TestParseMarkdown(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		sep           string
		expectedTitle string
		expectedCount int
	}{
		{
			name: "headings split",
			content: `# Slide 1
Content 1
# Slide 2
Content 2`,
			expectedTitle: "Slide 1",
			expectedCount: 2,
		},
		{
			name: "horizontal rule split",
			content: `Slide 1
---
Slide 2`,
			expectedCount: 2,
		},
		{
			name: "custom separator",
			content: `Slide 1
!!!
Slide 2`,
			sep:           "!!!",
			expectedCount: 2,
		},
		{
			name: "ignore heading in code block",
			content: `# Slide 1
` + "```" + `
# Not a slide
` + "```" + `
# Slide 2`,
			expectedCount: 2,
		},
		{
			name: "heading levels",
			content: `# Slide Title
## Subheading`,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck, err := ParseMarkdown(strings.NewReader(tt.content), Options{Separator: tt.sep})
			if err != nil {
				t.Fatalf("ParseMarkdown failed: %v", err)
			}
			if len(deck.Slides) != tt.expectedCount {
				t.Fatalf("expected %d slides, got %d", tt.expectedCount, len(deck.Slides))
			}
			if tt.expectedTitle != "" && deck.Title != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, deck.Title)
			}

			if tt.name == "heading levels" {
				content := deck.Slides[0].Content
				if !strings.Contains(content, "<h1") || !strings.Contains(content, "Slide Title") {
					t.Errorf("Expected <h1> for level 1 heading, got: %s", content)
				}
				if !strings.Contains(content, "<h2") || !strings.Contains(content, "Subheading") {
					t.Errorf("Expected <h2> for level 2 heading, got: %s", content)
				}
			}
		})
	}
}

func TestParseOrg(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		sep           string
		expectedTitle string
		expectedCount int
	}{
		{
			name: "headings split",
			content: `* Slide 1
Content 1
* Slide 2
Content 2`,
			expectedTitle: "Slide 1",
			expectedCount: 2,
		},
		{
			name: "horizontal rule split",
			content: `Slide 1
-----
Slide 2`,
			expectedCount: 2,
		},
		{
			name: "custom separator",
			content: `Slide 1
SPLIT
Slide 2`,
			sep:           "SPLIT",
			expectedCount: 2,
		},
		{
			name: "ignore heading in block",
			content: `* Slide 1
#+BEGIN_SRC
* Not a slide
#+END_SRC
* Slide 2`,
			expectedCount: 2,
		},
		{
			name: "metadata is skipped",
			content: `#+TITLE: My Presentation
#+OPTIONS: toc:nil
* Slide 1
Content`,
			expectedTitle: "Slide 1", // Now handling first slide title as deck title
			expectedCount: 1,
		},
		{
			name: "heading levels",
			content: `* Slide Title
** Subheading`,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck, err := ParseOrg(strings.NewReader(tt.content), Options{Separator: tt.sep})
			if err != nil {
				t.Fatalf("ParseOrg failed: %v", err)
			}
			if len(deck.Slides) != tt.expectedCount {
				t.Fatalf("expected %d slides, got %d", tt.expectedCount, len(deck.Slides))
			}
			if tt.expectedTitle != "" && deck.Title != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, deck.Title)
			}

			if tt.name == "heading levels" {
				content := deck.Slides[0].Content
				if !strings.Contains(content, "<h1") || !strings.Contains(content, "Slide Title") {
					t.Errorf("Expected <h1> for level 1 heading, got: %s", content)
				}
				if !strings.Contains(content, "<h2") || !strings.Contains(content, "Subheading") {
					t.Errorf("Expected <h2> for level 2 heading, got: %s", content)
				}
			}
		})
	}
}

func TestHighlighting(t *testing.T) {
	t.Run("Markdown Go", func(t *testing.T) {
		input := "# Slide 1\n```go\nfunc main() {}\n```"
		deck, _ := ParseMarkdown(strings.NewReader(input), Options{})
		if !strings.Contains(deck.Slides[0].Content, "main") {
			t.Errorf("code highlighting seems missing")
		}
	})

	t.Run("Org Go", func(t *testing.T) {
		input := "* Slide 1\n#+BEGIN_SRC go\nfunc main() {}\n#+END_SRC"
		deck, _ := ParseOrg(strings.NewReader(input), Options{})
		if !strings.Contains(deck.Slides[0].Content, "main") {
			t.Errorf("code highlighting seems missing in org")
		}
	})
}
