package parser

import (
	"strings"
	"testing"
)

func TestParseRedirectStdout(t *testing.T) {
	l := NewLexer("> out.txt")
	tok := l.NextToken()
	r, err := parseRedirect(tok, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != RedirectStdout {
		t.Fatalf("expected RedirectStdout, got %v", r.Type)
	}
	if r.File != "out.txt" {
		t.Fatalf("expected 'out.txt', got %q", r.File)
	}
}

func TestParseRedirectStderrToStdout(t *testing.T) {
	l := NewLexer("2>&1")
	tok := l.NextToken()
	r, err := parseRedirect(tok, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != RedirectStderrToStdout {
		t.Fatalf("expected RedirectStderrToStdout, got %v", r.Type)
	}
	if r.File != "1" {
		t.Fatalf("expected file '1', got %q", r.File)
	}
}

func TestParseRedirectHereString(t *testing.T) {
	l := NewLexer("<<< hello")
	tok := l.NextToken()
	r, err := parseRedirect(tok, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != RedirectHereString {
		t.Fatalf("expected RedirectHereString, got %v", r.Type)
	}
	if r.Content != "hello" {
		t.Fatalf("expected content 'hello', got %q", r.Content)
	}
}

func TestParseRedirectInvalid(t *testing.T) {
	l := NewLexer("echo")
	tok := l.NextToken()
	_, err := parseRedirect(tok, l)
	if err == nil {
		t.Fatal("expected error for non-redirect token")
	}
}

func TestParseRedirectMissingFile(t *testing.T) {
	l := NewLexer(">")
	tok := l.NextToken()
	_, err := parseRedirect(tok, l)
	if err == nil {
		t.Fatal("expected error for missing filename")
	}
}

func TestParseHeredoc(t *testing.T) {
	input := "<<EOF\nline1\nline2\nEOF\n"
	l := NewLexer(input)
	tok := l.NextToken()
	r, err := parseRedirect(tok, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != RedirectHeredoc {
		t.Fatalf("expected RedirectHeredoc, got %v", r.Type)
	}
	if r.File != "EOF" {
		t.Fatalf("expected delimiter 'EOF', got %q", r.File)
	}
	if !strings.Contains(r.Content, "line1") || !strings.Contains(r.Content, "line2") {
		t.Fatalf("expected heredoc content, got %q", r.Content)
	}
}

func TestParseHeredocDash(t *testing.T) {
	input := "<<-EOF\n\tline\nEOF\n"
	l := NewLexer(input)
	tok := l.NextToken()
	r, err := parseRedirect(tok, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != RedirectHeredocDash {
		t.Fatalf("expected RedirectHeredocDash, got %v", r.Type)
	}
}

func TestParseHeredocMissingDelimiter(t *testing.T) {
	l := NewLexer("<<")
	tok := l.NextToken()
	_, err := parseRedirect(tok, l)
	if err == nil {
		t.Fatal("expected error for missing heredoc delimiter")
	}
}

func TestParseHeredocEmptyContent(t *testing.T) {
	input := "<<EOF\nEOF\n"
	l := NewLexer(input)
	tok := l.NextToken()
	r, err := parseRedirect(tok, l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Content != "" {
		t.Fatalf("expected empty content, got %q", r.Content)
	}
}
