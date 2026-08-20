package expand

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlobNoGlobChars(t *testing.T) {
	got, err := Glob("plain.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "plain.txt" {
		t.Fatalf("expected unchanged word, got %v", got)
	}
}

func TestGlobNoMatch(t *testing.T) {
	got, err := Glob("nonexistent-*.xyz", nil)
	if err != nil {
		t.Fatal(err)
	}
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
	got, err := Glob(filepath.Join(dir, "*.txt"), nil)
	if err != nil {
		t.Fatal(err)
	}
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
	got, err := Glob(filepath.Join(dir, "*.txt"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != filepath.Join(dir, "a.txt") {
		t.Fatalf("expected sorted results, got %v", got)
	}
}

func TestGlobExpand(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := GlobExpand(filepath.Join(dir, "*.txt"), nil)
	if err != nil {
		t.Fatal(err)
	}
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

func TestHasUnquotedGlobChars(t *testing.T) {
	cases := []struct {
		s    string
		mask []bool
		want bool
	}{
		{"*.go", nil, true},
		{"*.go", []bool{false, true, true, true}, true},
		{"*.go", []bool{true, true, true, true}, false},
		{"a*b", []bool{false, true, false}, false},
		{"a*b", []bool{false, false, false}, true},
		{"plain", nil, false},
	}
	for _, c := range cases {
		if got := hasUnquotedGlobChars(c.s, c.mask); got != c.want {
			t.Errorf("hasUnquotedGlobChars(%q, %v) = %v, want %v", c.s, c.mask, got, c.want)
		}
	}
}

func TestGlobQuotedNoMatch(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// All bytes quoted: the glob char must not expand.
	pattern := filepath.Join(dir, "*.txt")
	mask := make([]bool, len(pattern))
	for i := range mask {
		mask[i] = true
	}
	got, err := Glob(pattern, mask)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != pattern {
		t.Fatalf("expected literal pattern, got %v", got)
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
	got, err := Glob(filepath.Join(dir, "**", "*.go"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 recursive matches, got %v", got)
	}
}

func TestGlobStarNoMatch(t *testing.T) {
	got, err := Glob("nonexistent/**/*.go", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "nonexistent/**/*.go" {
		t.Fatalf("expected unchanged pattern on no match, got %v", got)
	}
}

func TestGlobStarMultiple(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "a", "b"), 0o755)
	os.MkdirAll(filepath.Join(dir, "x", "y"), 0o755)
	os.WriteFile(filepath.Join(dir, "a", "b", "deep.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "x", "y", "deep.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "a", "top.txt"), []byte("x"), 0o644)

	got, err := Glob(filepath.Join(dir, "**", "**", "deep.txt"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 matches for **/**, got %v", got)
	}

	got, err = Glob(filepath.Join(dir, "a", "**", "b", "**", "deep.txt"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != filepath.Join(dir, "a", "b", "deep.txt") {
		t.Fatalf("expected a/b/deep.txt, got %v", got)
	}
}

func TestGlobStarZeroDirs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "deep.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "sub", "deep.txt"), []byte("x"), 0o644)

	// **/deep.txt must match deep.txt in the root (zero directories).
	got, err := Glob(filepath.Join(dir, "**", "deep.txt"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 matches (zero-dir + nested), got %v", got)
	}
}

func TestGlobNullGlob(t *testing.T) {
	got, err := GlobWithOptions("nonexistent-*.xyz", nil, GlobOptions{NullGlob: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result with nullglob, got %v", got)
	}
}

func TestGlobDotGlob(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "visible"), []byte("x"), 0o644)

	// Without dotglob, * does not match .hidden.
	got, err := Glob(filepath.Join(dir, "*"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != filepath.Join(dir, "visible") {
		t.Fatalf("expected only visible without dotglob, got %v", got)
	}

	// With dotglob, * matches .hidden too.
	got, err = GlobWithOptions(filepath.Join(dir, "*"), nil, GlobOptions{DotGlob: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 matches with dotglob, got %v", got)
	}
}

func TestGlobMalformedPattern(t *testing.T) {
	_, err := Glob("a[", nil)
	if err == nil {
		t.Fatal("expected error for malformed pattern")
	}
}
