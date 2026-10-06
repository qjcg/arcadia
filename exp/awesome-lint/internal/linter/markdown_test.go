package linter

import (
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestTextOf(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		doc := parseMarkdown([]byte("Hello world"))
		got := doc.TextOf(doc.Root)
		// Paragraph wraps the text, space preserved
		if got != "Hello world" {
			t.Errorf("TextOf() = %q, want %q", got, "Hello world")
		}
	})

	t.Run("text with emphasis", func(t *testing.T) {
		doc := parseMarkdown([]byte("Hello *world*"))
		got := doc.TextOf(doc.Root)
		// Spaces between inline elements preserved by goldmark
		if got != "Hello world" {
			t.Errorf("TextOf() = %q, want %q", got, "Hello world")
		}
	})

	t.Run("text with link", func(t *testing.T) {
		doc := parseMarkdown([]byte("[Example](https://example.com)"))
		got := doc.TextOf(doc.Root)
		if got != "Example" {
			t.Errorf("TextOf() = %q, want %q", got, "Example")
		}
	})

	t.Run("text with image", func(t *testing.T) {
		doc := parseMarkdown([]byte("![Alt](image.png)"))
		got := doc.TextOf(doc.Root)
		if got != "Alt" {
			t.Errorf("TextOf() = %q, want %q", got, "Alt")
		}
	})

	t.Run("text with code span", func(t *testing.T) {
		doc := parseMarkdown([]byte("Use `code` here"))
		got := doc.TextOf(doc.Root)
		// Spaces preserved
		if got != "Use code here" {
			t.Errorf("TextOf() = %q, want %q", got, "Use code here")
		}
	})

	t.Run("heading text", func(t *testing.T) {
		markdown := "# Heading Text\n\nContent."
		doc := parseMarkdown([]byte(markdown))
		// Find the heading node specifically
		var heading ast.Node
		ast.Walk(doc.Root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering && n.Kind() == ast.KindHeading {
				heading = n
				return ast.WalkStop, nil
			}
			return ast.WalkContinue, nil
		})
		if heading == nil {
			t.Fatal("no heading found")
		}
		got := doc.TextOf(heading)
		if got != "Heading Text" {
			t.Errorf("TextOf() = %q, want %q", got, "Heading Text")
		}
	})

	t.Run("empty document", func(t *testing.T) {
		doc := parseMarkdown([]byte(""))
		got := doc.TextOf(doc.Root)
		if got != "" {
			t.Errorf("TextOf() = %q, want %q", got, "")
		}
	})
}

func TestLineColOf(t *testing.T) {
	t.Run("heading node", func(t *testing.T) {
		markdown := "# Hello\n\nContent.\n"
		doc := parseMarkdown([]byte(markdown))
		// Find the first heading
		var heading ast.Node
		ast.Walk(doc.Root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering && n.Kind() == ast.KindHeading {
				heading = n
				return ast.WalkStop, nil
			}
			return ast.WalkContinue, nil
		})
		if heading == nil {
			t.Fatal("no heading found")
		}
		line, col := doc.LineColOf(heading)
		// Heading is at column 3 because "# " is 2 characters
		if line != 1 || col != 3 {
			t.Errorf("LineColOf() = (%d,%d), want (1,3)", line, col)
		}
	})

	t.Run("inline node", func(t *testing.T) {
		markdown := "# Hello\n"
		doc := parseMarkdown([]byte(markdown))
		// Find the first text node
		var textNode ast.Node
		ast.Walk(doc.Root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering && n.Kind() == ast.KindText {
				textNode = n
				return ast.WalkStop, nil
			}
			return ast.WalkContinue, nil
		})
		if textNode == nil {
			t.Fatal("no text node found")
		}
		line, col := doc.LineColOf(textNode)
		if line != 1 || col != 3 {
			t.Errorf("LineColOf() = (%d,%d), want (1,3)", line, col)
		}
	})

	t.Run("unknown inline node falls back to parent", func(t *testing.T) {
		// Use a node type that has no text children but has a block parent
		markdown := "- List item\n"
		doc := parseMarkdown([]byte(markdown))
		// Find the first list item
		var listItem ast.Node
		ast.Walk(doc.Root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering && n.Kind() == ast.KindListItem {
				listItem = n
				return ast.WalkStop, nil
			}
			return ast.WalkContinue, nil
		})
		if listItem == nil {
			t.Fatal("no list item found")
		}
		line, col := doc.LineColOf(listItem)
		if col < 1 || line < 1 {
			t.Errorf("LineColOf() = (%d,%d), want positive values", line, col)
		}
	})
}

func TestLineFromPos(t *testing.T) {
	doc := &MarkdownDoc{Source: []byte("line1\nline2\nline3\n")}
	tests := []struct {
		pos  int
		want int
	}{
		{0, 1},
		{6, 2},
		{12, 3},
		{18, 4},
		{100, 4},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := doc.LineFromPos(tt.pos); got != tt.want {
				t.Errorf("LineFromPos(%d) = %d, want %d", tt.pos, got, tt.want)
			}
		})
	}
}

func TestLineStart(t *testing.T) {
	doc := &MarkdownDoc{Source: []byte("line1\nline2\nline3\n")}
	tests := []struct {
		lineIdx int
		want    int
	}{
		{0, 0},
		{1, 6},
		{2, 12},
		{3, 18},
		{10, 18},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := doc.LineStart(tt.lineIdx); got != tt.want {
				t.Errorf("LineStart(%d) = %d, want %d", tt.lineIdx, got, tt.want)
			}
		})
	}
}

func TestParseMarkdown(t *testing.T) {
	source := []byte("# Hello\n\nWorld.\n")
	doc := parseMarkdown(source)
	if doc == nil {
		t.Fatal("parseMarkdown() returned nil")
	}
	if doc.Root == nil {
		t.Fatal("parseMarkdown() Root is nil")
	}
}
