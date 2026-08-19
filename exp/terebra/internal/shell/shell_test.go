package shell

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chzyer/readline"
	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
)

func TestRunScriptFromStringSimple(t *testing.T) {
	if err := RunScriptFromString("echo hi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunScriptFromStringWithNewline(t *testing.T) {
	if err := RunScriptFromString("echo a\necho b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunScriptFromStringShebang(t *testing.T) {
	if err := RunScriptFromString("#!/bin/sh\necho a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuiltinReadonlyList(t *testing.T) {
	s := newTestShell()
	s.vars["a"] = "1"
	s.vars["b"] = "2"
	s.readonly["a"] = true
	var out, errb bytes.Buffer
	if code := s.builtinReadonly(nil, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if out.Len() == 0 {
		t.Fatal("expected readonly listing in output")
	}
}

func TestBuiltinReadonlySet(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	if code := s.builtinReadonly([]string{"x=5", "y"}, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !s.readonly["x"] || !s.readonly["y"] {
		t.Fatal("expected x and y to be readonly")
	}
}

func TestReadHistory(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	os.Setenv("HOME", dir)
	os.WriteFile(filepath.Join(dir, historyFile), []byte("cmd1\ncmd2\n"), 0o644)
	got := s.readHistory()
	if len(got) != 2 {
		t.Fatalf("expected 2 history entries, got %d: %v", len(got), got)
	}
}

func TestReadHistoryMissing(t *testing.T) {
	s := newTestShell()
	os.Setenv("HOME", t.TempDir())
	if got := s.readHistory(); got != nil {
		t.Fatalf("expected nil for missing history, got %v", got)
	}
}

func TestBuiltinExecMissingCommand(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	if code := s.builtinExec(nil, &out, &errb); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestBuiltinExecDashANoArg(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	if code := s.builtinExec([]string{"-a"}, &out, &errb); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestBuiltinExecNotFound(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	if code := s.builtinExec([]string{"nonexistent-cmd-zzz"}, &out, &errb); code != 127 {
		t.Fatalf("expected exit 127, got %d", code)
	}
}

func TestFillHeredocsPrefilled(t *testing.T) {
	s := newTestShell()
	redir := &parser.Redirect{Type: parser.RedirectHeredoc, File: "EOF", Content: "already"}
	script := &parser.Script{
		Pipelines: []*parser.Pipeline{
			{Commands: []*parser.Command{{Redirects: []*parser.Redirect{redir}}}},
		},
	}
	// Content is non-empty so the readline loop is skipped; should not crash.
	s.fillHeredocs(script)
	if redir.Content != "already" {
		t.Fatalf("expected content preserved, got %q", redir.Content)
	}
}

func TestFillHeredocsReadline(t *testing.T) {
	s := newTestShell()
	// Create a pipe to feed readline
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	cfg := &readline.Config{
		Prompt: "",
		Stdin:  r,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	rl, err := readline.NewEx(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer rl.Close()
	s.rl = rl

	// Feed heredoc lines: "line1\nEOF\n"
	go func() {
		w.Write([]byte("line1\nEOF\n"))
		w.Close()
	}()

	redir := &parser.Redirect{Type: parser.RedirectHeredoc, File: "EOF"}
	script := &parser.Script{
		Pipelines: []*parser.Pipeline{
			{Commands: []*parser.Command{{Redirects: []*parser.Redirect{redir}}}},
		},
	}
	s.fillHeredocs(script)
	if !strings.Contains(redir.Content, "line1") {
		t.Fatalf("expected heredoc content, got %q", redir.Content)
	}
}
