package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestSetupCLI_Help(t *testing.T) {
	var out, err bytes.Buffer
	cmd := setupCLI(&out, &err)
	cmd.SetArgs([]string{"--help"})
	errOut := cmd.Execute()
	if errOut != nil {
		t.Fatalf("unexpected error: %v", errOut)
	}
	if !strings.Contains(out.String(), "awesome-lint") {
		t.Errorf("expected help to contain 'awesome-lint', got: %s", out.String())
	}
	if !strings.Contains(out.String(), "filename") {
		t.Errorf("expected help to contain 'filename', got: %s", out.String())
	}
}

func TestSetupCLI_Version(t *testing.T) {
	var out, err bytes.Buffer
	cmd := setupCLI(&out, &err)
	cmd.SetArgs([]string{"--version"})
	errOut := cmd.Execute()
	if errOut != nil {
		t.Fatalf("unexpected error: %v", errOut)
	}
}

func TestSetupCLI_RunE_NoArgs_DefaultFilename(t *testing.T) {
	var out, err bytes.Buffer
	cmd := setupCLI(&out, &err)
	// Use a non-existent file to trigger lint error path
	cmd.SetArgs([]string{"--filename", "/nonexistent/testfile.md"})
	errOut := cmd.Execute()
	if errOut == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestSetupCLI_RunE_WithFileArg(t *testing.T) {
	var out, err bytes.Buffer
	cmd := setupCLI(&out, &err)
	// Use a non-existent file argument to trigger lint error path
	cmd.SetArgs([]string{"/nonexistent/other.md"})
	errOut := cmd.Execute()
	if errOut == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestSetupCLI_JSONOutput(t *testing.T) {
	var out, err bytes.Buffer
	cmd := setupCLI(&out, &err)
	// Use non-existent file with JSON flag
	cmd.SetArgs([]string{"--json", "--filename", "/nonexistent/testfile.md"})
	errOut := cmd.Execute()
	if errOut == nil {
		t.Fatal("expected error for non-existent file")
	}
}
