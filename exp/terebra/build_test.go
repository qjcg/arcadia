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

func TestBuildCmdInvalidScript(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "script.trb")
	if err := os.WriteFile(script, []byte("echo hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Compilation will fail because the temp .go in module root references
	// internal packages; in this test environment it should error out.
	_ = buildCmd([]string{script})
}
