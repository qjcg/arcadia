package expand

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlobNoGlobChars(t *testing.T) {
	got := Glob("plain.txt")
	if len(got) != 1 || got[0] != "plain.txt" {
		t.Fatalf("expected unchanged word, got %v", got)
	}
}

func TestGlobNoMatch(t *testing.T) {
	got := Glob("nonexistent-*.xyz")
	if len(got) != 1 || got[0] != "nonexistent-*.xyz" {
		t.Fatalf("expected unchanged word on no match, got %v", got)
	}
}

func TestGlobMatches(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", "c.log"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got := Glob(filepath.Join(dir, "*.txt"))
	if len(got) != 2 {
		t.Fatalf("expected 2 matches, got %v", got)
	}
}

func TestGlobSorted(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"b.txt", "a.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got := Glob(filepath.Join(dir, "*.txt"))
	if got[0] != filepath.Join(dir, "a.txt") {
		t.Fatalf("expected sorted results, got %v", got)
	}
}

func TestGlobExpand(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := GlobExpand(filepath.Join(dir, "*.txt"))
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %v", got)
	}
}

func TestHasGlobChars(t *testing.T) {
	cases := map[string]bool{
		"plain":   false,
		"*.go":    true,
		"a?b":     true,
		"a[b]":    true,
		"a/b/c":   false,
		"**/*.go": true,
	}
	for in, want := range cases {
		if got := hasGlobChars(in); got != want {
			t.Errorf("hasGlobChars(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestGlobStarRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "root.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "nested.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Glob(filepath.Join(dir, "**", "*.go"))
	if len(got) != 2 {
		t.Fatalf("expected 2 recursive matches, got %v", got)
	}
}

func TestGlobStarNoMatch(t *testing.T) {
	got := Glob("nonexistent/**/*.go")
	if len(got) != 1 || got[0] != "nonexistent/**/*.go" {
		t.Fatalf("expected unchanged pattern on no match, got %v", got)
	}
}
