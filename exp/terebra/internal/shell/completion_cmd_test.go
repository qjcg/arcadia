package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommonPrefixLen(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"abc", "abd", 2},
		{"abc", "abc", 3},
		{"abc", "xyz", 0},
		{"", "abc", 0},
		{"abc", "", 0},
	}
	for _, c := range cases {
		if got := commonPrefixLen(c.a, c.b); got != c.want {
			t.Errorf("commonPrefixLen(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestCompleteCommandNameBuiltin(t *testing.T) {
	s := newTestShell()
	got := s.completeCommandName("ec")
	found := false
	for _, c := range got {
		if c == "echo" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'echo' in completions, got %v", got)
	}
}

func TestCompleteCommandNameNoMatch(t *testing.T) {
	s := newTestShell()
	got := s.completeCommandName("zzzzzz")
	if len(got) != 0 {
		t.Fatalf("expected no completions, got %v", got)
	}
}

func TestCompleteFileOrArg(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "beta.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("x"), 0o644)

	s := newTestShell()
	got := s.completeFileOrArg([]string{"cat"}, dir+"/al")
	if len(got) != 1 || got[0] != dir+"/alpha.txt" {
		t.Fatalf("expected [%s/alpha.txt], got %v", dir, got)
	}
}

func TestCompleteFileOrArgHiddenSkipped(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("x"), 0o644)
	s := newTestShell()
	got := s.completeFileOrArg([]string{"cat"}, dir+"/")
	if len(got) != 0 {
		t.Fatalf("expected no hidden files, got %v", got)
	}
}

func TestCompleteFileOrArgDirSuffix(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	os.Mkdir(sub, 0o755)
	s := newTestShell()
	got := s.completeFileOrArg([]string{"cat"}, dir+"/sub")
	if len(got) != 1 || got[0] != dir+"/subdir/" {
		t.Fatalf("expected dir with trailing slash, got %v", got)
	}
}

func TestCompleteFileOrArgBadDir(t *testing.T) {
	s := newTestShell()
	got := s.completeFileOrArg([]string{"cat"}, "/nonexistent/xyz")
	if got != nil {
		t.Fatalf("expected nil for bad dir, got %v", got)
	}
}

func TestCompleteCommandEmpty(t *testing.T) {
	s := newTestShell()
	if got := s.completeCommand(""); got != nil {
		t.Fatalf("expected nil for empty line, got %v", got)
	}
}

func TestCompleteCommandFirstWord(t *testing.T) {
	s := newTestShell()
	got := s.completeCommand("ec")
	found := false
	for _, c := range got {
		if c == "echo" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'echo' in completions, got %v", got)
	}
}

func TestCompleteCommandAfterSpace(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644)
	s := newTestShell()
	got := s.completeCommand("cat " + dir + "/fi")
	if len(got) != 1 || got[0] != dir+"/file.txt" {
		t.Fatalf("expected file completion, got %v", got)
	}
}

func TestCompleteFileOrArgBareTilde(t *testing.T) {
	s := newTestShell()
	got := s.completeFileOrArg([]string{"cat"}, "~")
	if len(got) != 1 || !strings.HasSuffix(got[0], "/") {
		t.Fatalf("expected home dir with trailing slash, got %v", got)
	}
}

func TestCompleteFileOrArgTildeHome(t *testing.T) {
	home := t.TempDir()
	os.WriteFile(filepath.Join(home, "alpha.txt"), []byte("x"), 0o644)

	oldHome, hadHome := os.LookupEnv("HOME")
	os.Setenv("HOME", home)
	defer func() {
		if hadHome {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	s := newTestShell()
	got := s.completeFileOrArg([]string{"cat"}, "~/al")
	if len(got) != 1 || got[0] != filepath.Join(home, "alpha.txt") {
		t.Fatalf("expected ~/ expansion to %q, got %v", filepath.Join(home, "alpha.txt"), got)
	}
}

func TestShellCompleterDo(t *testing.T) {
	s := newTestShell()
	c := &shellCompleter{shell: s}
	suffixes, commonLen := c.Do([]rune("ec"), 2)
	if len(suffixes) == 0 {
		t.Fatal("expected suffixes")
	}
	if commonLen != 2 {
		t.Fatalf("expected commonLen 2, got %d", commonLen)
	}
}

func TestShellCompleterDoNoMatch(t *testing.T) {
	s := newTestShell()
	c := &shellCompleter{shell: s}
	suffixes, commonLen := c.Do([]rune("zzzz"), 4)
	if suffixes != nil || commonLen != 0 {
		t.Fatalf("expected (nil, 0), got (%v, %d)", suffixes, commonLen)
	}
}
