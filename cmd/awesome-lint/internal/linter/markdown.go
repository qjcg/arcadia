package linter

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// MarkdownDoc represents a parsed markdown document.
type MarkdownDoc struct {
	Root   ast.Node
	Source []byte
}

// TextOf returns the text content of a node.
func (d *MarkdownDoc) TextOf(node ast.Node) string {
	// Handle leaf text nodes directly
	if text, ok := node.(*ast.Text); ok {
		return string(text.Segment.Value(d.Source))
	}

	var buf bytes.Buffer
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		switch n := c.(type) {
		case *ast.Text:
			segment := n.Segment
			buf.Write(segment.Value(d.Source))
		case *ast.CodeSpan:
			for cc := n.FirstChild(); cc != nil; cc = cc.NextSibling() {
				if text, ok := cc.(*ast.Text); ok {
					buf.Write(text.Segment.Value(d.Source))
				}
			}
		case *ast.Link:
			buf.WriteString(d.TextOf(n))
		case *ast.Image:
			buf.WriteString(d.TextOf(n))
		case *ast.Emphasis:
			buf.WriteString(d.TextOf(n))
		case *ast.String:
			buf.Write(n.Value)
		default:
			buf.WriteString(d.TextOf(c))
		}
	}
	return buf.String()
}

// LineColOf returns the 1-based line and column for a node.
func (d *MarkdownDoc) LineColOf(node ast.Node) (int, int) {
	// Only block nodes have Lines(); inline nodes panic.
	if node.Type() == ast.TypeBlock {
		lines := node.Lines()
		if lines.Len() > 0 {
			seg := lines.At(0)
			line := d.LineFromPos(seg.Start)
			col := seg.Start - d.LineStart(line-1) + 1
			return line, col
		}
	}

	// For inline nodes, find the first text child with a segment
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		if text, ok := c.(*ast.Text); ok {
			seg := text.Segment
			line := d.LineFromPos(seg.Start)
			col := seg.Start - d.LineStart(line-1) + 1
			return line, col
		}
		line, col := d.LineColOf(c)
		if line > 1 || col > 1 {
			return line, col
		}
	}

	// Walk up to find a block-level ancestor with lines
	parent := node.Parent()
	for parent != nil {
		if parent.Type() == ast.TypeBlock {
			lines := parent.Lines()
			if lines.Len() > 0 {
				seg := lines.At(0)
				line := d.LineFromPos(seg.Start)
				col := seg.Start - d.LineStart(line-1) + 1
				return line, col
			}
		}
		parent = parent.Parent()
	}

	return 1, 1
}

// LineFromPos returns the 1-based line number for a byte position.
func (d *MarkdownDoc) LineFromPos(pos int) int {
	line := 1
	for i, b := range d.Source {
		if i >= pos {
			break
		}
		if b == '\n' {
			line++
		}
	}
	return line
}

// LineStart returns the byte offset of the start of a line (0-based).
func (d *MarkdownDoc) LineStart(lineIdx int) int {
	line := 0
	for i, b := range d.Source {
		if line == lineIdx {
			return i
		}
		if b == '\n' {
			line++
		}
	}
	return len(d.Source)
}

func parseMarkdown(source []byte) *MarkdownDoc {
	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)
	return &MarkdownDoc{
		Root:   doc,
		Source: source,
	}
}
