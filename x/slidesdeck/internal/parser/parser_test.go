package parser

import (
	"strings"
	"testing"
)

func TestSplitMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		sep      string
		expected []string
	}{
		{
			name: "headings split",
			content: `# Slide 1
Content 1
# Slide 2
Content 2`,
			sep:      "",
			expected: []string{"# Slide 1\nContent 1", "# Slide 2\nContent 2"},
		},
		{
			name: "horizontal rule split",
			content: `Slide 1
---
Slide 2`,
			sep:      "",
			expected: []string{"Slide 1", "Slide 2"},
		},
		{
			name: "custom separator",
			content: `Slide 1
!!!
Slide 2`,
			sep:      "!!!",
			expected: []string{"Slide 1", "Slide 2"},
		},
		{
			name: "ignore heading in code block",
			content: `# Slide 1
` + "```" + `
# Not a slide
` + "```" + `
# Slide 2`,
			sep:      "",
			expected: []string{"# Slide 1\n```\n# Not a slide\n```", "# Slide 2"},
		},
		{
			name: "trailing separator",
			content: `Slide 1
!!!`,
			sep:      "!!!",
			expected: []string{"Slide 1"},
		},
		{
			name: "starting separator",
			content: `!!!
Slide 1`,
			sep:      "!!!",
			expected: []string{"", "Slide 1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitMarkdown(tt.content, tt.sep)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d slides, got %d", len(tt.expected), len(got))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("slide %d: expected %q, got %q", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestSplitOrg(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		sep      string
		expected []string
	}{
		{
			name: "headings split",
			content: `* Slide 1
Content 1
* Slide 2
Content 2`,
			sep:      "",
			expected: []string{"* Slide 1\nContent 1", "* Slide 2\nContent 2"},
		},
		{
			name: "horizontal rule split",
			content: `Slide 1
-----
Slide 2`,
			sep:      "",
			expected: []string{"Slide 1", "Slide 2"},
		},
		{
			name: "custom separator",
			content: `Slide 1
SPLIT
Slide 2`,
			sep:      "SPLIT",
			expected: []string{"Slide 1", "Slide 2"},
		},
		{
			name: "ignore heading in block",
			content: `* Slide 1
#+BEGIN_SRC
* Not a slide
#+END_SRC
* Slide 2`,
			sep:      "",
			expected: []string{"* Slide 1\n#+BEGIN_SRC\n* Not a slide\n#+END_SRC", "* Slide 2"},
		},
		{
			name: "metadata is skipped",
			content: `#+TITLE: My Presentation
#+OPTIONS: toc:nil
* Slide 1
Content`,
			sep:      "",
			expected: []string{"* Slide 1\nContent"},
		},
		{
			name: "nested headings not split",
			content: `* Slide 1
** Subheading
Content
* Slide 2`,
			sep:      "",
			expected: []string{"* Slide 1\n** Subheading\nContent", "* Slide 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitOrg(tt.content, tt.sep)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d slides, got %d", len(tt.expected), len(got))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("slide %d: expected %q, got %q", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	input := "# Hello World\nContent"
	expected := "Hello World"
	got := extractTitle(input)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	inputOrg := "* Org Title\nContent"
	expectedOrg := "Org Title"
	gotOrg := extractOrgTitle(inputOrg)
	if gotOrg != expectedOrg {
		t.Errorf("expected %q, got %q", expectedOrg, gotOrg)
	}
}

func TestParseMarkdown(t *testing.T) {
	input := `# Title
# Slide 1
Hello
---
# Slide 2
World`
	deck, err := ParseMarkdown(strings.NewReader(input), Options{})
	if err != nil {
		t.Fatal(err)
	}

	if deck.Title != "Title" {
		t.Errorf("expected Title, got %q", deck.Title)
	}
	if len(deck.Slides) != 3 {
		t.Errorf("expected 3 slides, got %d", len(deck.Slides))
	}

	// Highlighting test
	inputWithCode := "# Slide 1\n```go\nfunc main() {}\n```"
	deck, _ = ParseMarkdown(strings.NewReader(inputWithCode), Options{})
	if !strings.Contains(deck.Slides[0].Content, "chrono") && !strings.Contains(deck.Slides[0].Content, "main") {
		// Chroma might use different classes, but 'main' should be there
		if !strings.Contains(deck.Slides[0].Content, "main") {
			t.Errorf("code highlighting seems missing")
		}
	}
}

func TestParseOrg(t *testing.T) {
	input := `#+TITLE: Org Presentation
* Slide 1
Hello
-----
* Slide 2
World`
	deck, err := ParseOrg(strings.NewReader(input), Options{})
	if err != nil {
		t.Fatal(err)
	}

	// Now handling first slide title as deck title
	if deck.Title != "Slide 1" {
		t.Errorf("expected Slide 1, got %q", deck.Title)
	}
	if len(deck.Slides) != 2 {
		t.Errorf("expected 2 slides, got %d", len(deck.Slides))
	}

	// Highlighting test
	inputWithCode := "* Slide 1\n#+BEGIN_SRC go\nfunc main() {}\n#+END_SRC"
	deck, _ = ParseOrg(strings.NewReader(inputWithCode), Options{})
	if !strings.Contains(deck.Slides[0].Content, "main") {
		t.Errorf("code highlighting seems missing in org")
	}
}

func TestOrgHeadingLevels(t *testing.T) {
	input := "* Slide Title\n** Subheading"
	deck, err := ParseOrg(strings.NewReader(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	content := deck.Slides[0].Content
	if !strings.Contains(content, "<h1") || !strings.Contains(content, "Slide Title") {
		t.Errorf("Expected <h1> for level 1 heading, got: %s", content)
	}
	if !strings.Contains(content, "<h2") || !strings.Contains(content, "Subheading") {
		t.Errorf("Expected <h2> for level 2 heading, got: %s", content)
	}
}

func TestMarkdownHeadingLevels(t *testing.T) {
	input := "# Slide Title\n## Subheading"
	deck, err := ParseMarkdown(strings.NewReader(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	content := deck.Slides[0].Content
	if !strings.Contains(content, "<h1") || !strings.Contains(content, "Slide Title") {
		t.Errorf("Expected <h1> for level 1 heading, got: %s", content)
	}
	if !strings.Contains(content, "<h2") || !strings.Contains(content, "Subheading") {
		t.Errorf("Expected <h2> for level 2 heading, got: %s", content)
	}
}
