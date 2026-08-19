package shell

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltinSetDebug(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	if code := s.builtinSet([]string{"-x"}, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !s.debug {
		t.Fatal("expected debug mode enabled")
	}
	if code := s.builtinSet([]string{"+x"}, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if s.debug {
		t.Fatal("expected debug mode disabled")
	}
}

func TestBuiltinSetUnknownOption(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	if code := s.builtinSet([]string{"-z"}, &out, &errb); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown option") {
		t.Fatalf("expected unknown option error, got %q", errb.String())
	}
}

func TestBuiltinSetList(t *testing.T) {
	s := newTestShell()
	s.vars["foo"] = "bar"
	var out, errb bytes.Buffer
	if code := s.builtinSet(nil, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "foo=bar") {
		t.Fatalf("expected foo=bar in output, got %q", out.String())
	}
}

func TestBuiltinState(t *testing.T) {
	s := newTestShell()
	s.vars["x"] = "1"
	var out, errb bytes.Buffer
	if code := s.builtinState(nil, strings.NewReader(""), &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "x") {
		t.Fatalf("expected state output, got %q", out.String())
	}
}

func TestBuiltinStateSaveAndLoad(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	os.Setenv("HOME", dir)
	s.vars["savedvar"] = "hello"

	var out, errb bytes.Buffer
	if code := s.builtinStateSave([]string{"test"}, &out, &errb); code != 0 {
		t.Fatalf("save failed: %d (stderr: %s)", code, errb.String())
	}

	// Load into a fresh shell
	s2 := newTestShell()
	os.Setenv("HOME", dir)
	var out2, errb2 bytes.Buffer
	if code := s2.builtinStateLoad([]string{"test"}, &out2, &errb2); code != 0 {
		t.Fatalf("load failed: %d (stderr: %s)", code, errb2.String())
	}
	if got := s2.getVar("savedvar"); got != `"hello"` {
		t.Fatalf("expected savedvar=\"hello\" after load, got %q", got)
	}
}

func TestBuiltinStateLoadMissing(t *testing.T) {
	s := newTestShell()
	os.Setenv("HOME", t.TempDir())
	var out, errb bytes.Buffer
	if code := s.builtinStateLoad([]string{"nonexistent"}, &out, &errb); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestBuiltinHistory(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	os.Setenv("HOME", dir)
	os.WriteFile(filepath.Join(dir, historyFile), []byte("cmd1\ncmd2\n"), 0o644)
	var out, errb bytes.Buffer
	if code := s.builtinHistory(nil, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "cmd1") {
		t.Fatalf("expected cmd1 in history output, got %q", out.String())
	}
}

func TestBuiltinHistoryNoFile(t *testing.T) {
	s := newTestShell()
	os.Setenv("HOME", t.TempDir())
	var out, errb bytes.Buffer
	if code := s.builtinHistory(nil, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}
