package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T, dir string) string {
	t.Helper()
	bin := filepath.Join(dir, "pavona")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("building pavona: %v\n%s", err, out)
	}
	return bin
}

func runPavona(t *testing.T, bin string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestBuild(t *testing.T) {
	bin := buildBinary(t, t.TempDir())
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("binary not found: %v", err)
	}
}

func TestListTemplates(t *testing.T) {
	bin := buildBinary(t, t.TempDir())
	out, err := runPavona(t, bin, "-l")
	if err != nil {
		t.Fatalf("pavona -l failed: %v\n%s", err, out)
	}
	for _, name := range []string{"tool", "lib", "site", "tui", "app", "agent"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected built-in template %q in output, got:\n%s", name, out)
		}
	}
}

func TestToolTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)
	outDir := filepath.Join(tmp, "my-cli")

	out, err := runPavona(t, bin, "-t", "tool", "-o", outDir, "-n", "my-cli", "-q")
	if err != nil {
		t.Fatalf("pavona -t tool failed: %v\n%s", err, out)
	}

	checks := []string{"main.go", "go.mod", "Taskfile.yaml", ".gitignore", "features"}
	for _, f := range checks {
		p := filepath.Join(outDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected %q to exist", p)
		}
	}

	// Verify template rendering
	data, err := os.ReadFile(filepath.Join(outDir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "my-cli") {
		t.Errorf("expected main.go to contain project name, got:\n%s", string(data))
	}
}

func TestLibTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)
	outDir := filepath.Join(tmp, "go-csvstream")

	out, err := runPavona(t, bin, "-t", "lib", "-o", outDir, "-n", "go-csvstream", "-q")
	if err != nil {
		t.Fatalf("pavona -t lib failed: %v\n%s", err, out)
	}

	for _, f := range []string{"lib.go", "lib_test.go", "go.mod", "Taskfile.yaml", ".gitignore"} {
		p := filepath.Join(outDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected %q to exist", p)
		}
	}
}

func TestSiteTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)
	outDir := filepath.Join(tmp, "blog")

	out, err := runPavona(t, bin, "-t", "site", "-o", outDir, "-n", "blog", "-q")
	if err != nil {
		t.Fatalf("pavona -t site failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(filepath.Join(outDir, "content/index.md")); os.IsNotExist(err) {
		t.Errorf("expected content/index.md to exist")
	}
}

func TestTuiTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)
	outDir := filepath.Join(tmp, "chatmonitor")

	out, err := runPavona(t, bin, "-t", "tui", "-o", outDir, "-n", "chatmonitor", "-q")
	if err != nil {
		t.Fatalf("pavona -t tui failed: %v\n%s", err, out)
	}

	for _, f := range []string{"main.go", "go.mod", ".gitignore"} {
		p := filepath.Join(outDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected %q to exist", p)
		}
	}
}

func TestAppTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)
	outDir := filepath.Join(tmp, "acmecorp")

	out, err := runPavona(t, bin, "-t", "app", "-o", outDir, "-n", "acmecorp", "-q")
	if err != nil {
		t.Fatalf("pavona -t app failed: %v\n%s", err, out)
	}

	for _, f := range []string{"main.go", "main_test.go", "go.mod", "Dockerfile", ".gitignore", "internal/handlers/health.go"} {
		p := filepath.Join(outDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected %q to exist", p)
		}
	}
}

func TestAgentTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)
	outDir := filepath.Join(tmp, "triagebot")

	out, err := runPavona(t, bin, "-t", "agent", "-o", outDir, "-n", "triagebot", "-q")
	if err != nil {
		t.Fatalf("pavona -t agent failed: %v\n%s", err, out)
	}

	for _, f := range []string{"main.go", "go.mod", ".gitignore"} {
		p := filepath.Join(outDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected %q to exist", p)
		}
	}
}

func TestCustomTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)

	customDir := filepath.Join(tmp, "custom-tmpl")
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := `package template

name:        "custom"
description: "A custom template for testing"

variables: {
	project_name: {
		prompt:    "Project name"
		default:   ""
		required:  true
	}
	message: {
		prompt:    "Greeting message"
		default:   "Hello, World!"
		required:  false
	}
}
`
	if err := os.WriteFile(filepath.Join(customDir, "config.cue"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	mainTmpl := `package main

import "fmt"

func main() {
	fmt.Println("{{.message}}")
}
`
	if err := os.WriteFile(filepath.Join(customDir, "main.go.tmpl"), []byte(mainTmpl), 0o644); err != nil {
		t.Fatal(err)
	}

	staticFile := []byte("static content\n")
	if err := os.WriteFile(filepath.Join(customDir, "README.md"), staticFile, 0o644); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(tmp, "custom-output")
	out, err := runPavona(t, bin, "-t", customDir, "-o", outDir, "-n", "custom-test", "-q")
	if err != nil {
		t.Fatalf("pavona -t custom failed: %v\n%s", err, out)
	}

	// Verify rendered file
	data, err := os.ReadFile(filepath.Join(outDir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Hello, World!") {
		t.Errorf("expected main.go to contain message, got:\n%s", string(data))
	}

	// Verify static file copied as-is
	data, err = os.ReadFile(filepath.Join(outDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "static content\n" {
		t.Errorf("expected README.md to be copied verbatim, got: %q", string(data))
	}
}

func TestHelpFlag(t *testing.T) {
	bin := buildBinary(t, t.TempDir())
	out, err := runPavona(t, bin, "--help")
	if err != nil {
		t.Fatalf("pavona --help failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "template") {
		t.Errorf("expected --help output to mention templates, got:\n%s", out)
	}
}
