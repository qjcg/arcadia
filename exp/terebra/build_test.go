package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCmdNoArgs(t *testing.T) {
	if got := buildCmd(nil); got != 1 {
		t.Fatalf("expected exit 1, got %d", got)
	}
}

func TestBuildCmdMissingFile(t *testing.T) {
	if got := buildCmd([]string{"/nonexistent/script.trb"}); got != 1 {
		t.Fatalf("expected exit 1, got %d", got)
	}
}

func TestBuildCmdValidScript(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "script.trb")
	if err := os.WriteFile(script, []byte("echo hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write the output binary inside the temp dir so we don't pollute the
	// source tree. The embedded script (echo hi) is valid, so the build
	// must succeed and produce an executable file.
	out := filepath.Join(dir, "out")
	if got := buildCmd([]string{script, out}); got != 0 {
		t.Fatalf("expected exit 0, got %d", got)
	}
	fi, err := os.Stat(out)
	if err != nil {
		t.Fatalf("expected compiled binary at %s: %v", out, err)
	}
	if fi.Size() <= 0 {
		t.Fatalf("expected non-empty binary at %s", out)
	}
}
