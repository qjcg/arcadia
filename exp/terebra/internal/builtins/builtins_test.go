package builtins

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCdHandlerToTempDir(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	var out, errb bytes.Buffer
	code := cdHandler([]string{dir}, strings.NewReader(""), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if os.Getenv("OLDPWD") == "" {
		t.Fatal("expected OLDPWD to be set")
	}
}

func TestCdHandlerBadDir(t *testing.T) {
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	var out, errb bytes.Buffer
	code := cdHandler([]string{"/nonexistent/dir/xyz"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "cd:") {
		t.Fatalf("expected cd error, got %q", errb.String())
	}
}

func TestCdHandlerDashNoOLDPWD(t *testing.T) {
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Unsetenv("OLDPWD")
	var out, errb bytes.Buffer
	code := cdHandler([]string{"-"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "OLDPWD not set") {
		t.Fatalf("expected OLDPWD error, got %q", errb.String())
	}
}
