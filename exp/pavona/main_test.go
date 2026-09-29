package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var testBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "pavona-test-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating test directory: %v\n", err)
		os.Exit(1)
	}

	testBinary = filepath.Join(dir, "pavona")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "building pavona: %v\n%s", err, out)
		_ = os.RemoveAll(dir)
		os.Exit(1)
	}

	exitCode := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(exitCode)
}

func runPavona(t *testing.T, bin string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runPavonaInDir(t *testing.T, bin, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestBuild(t *testing.T) {
	if _, err := os.Stat(testBinary); err != nil {
		t.Fatalf("binary not found: %v", err)
	}
}

func TestListTemplates(t *testing.T) {
	bin := testBinary
	out, err := runPavona(t, bin, "list")
	if err != nil {
		t.Fatalf("pavona list failed: %v\n%s", err, out)
	}
	for _, name := range []string{"tool", "lib", "site", "tui", "app", "agent", "pavona", "monorepo-go"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected built-in template %q in output, got:\n%s", name, out)
		}
	}

	aliasOut, err := runPavona(t, bin, "ls")
	if err != nil {
		t.Fatalf("pavona ls failed: %v\n%s", err, aliasOut)
	}
	if !strings.Contains(aliasOut, "monorepo-go") {
		t.Errorf("expected list alias to show built-in templates, got:\n%s", aliasOut)
	}
}

func TestNewTemplateCompletion(t *testing.T) {
	bin := testBinary
	out, err := runPavona(t, bin, "__complete", "new", "")
	if err != nil {
		t.Fatalf("pavona new completion failed: %v\n%s", err, out)
	}
	for _, name := range []string{"tool", "lib", "site", "tui", "app", "agent", "pavona", "monorepo-go"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected built-in template %q in completion output, got:\n%s", name, out)
		}
	}

	out, err = runPavona(t, bin, "__complete", "new", "mo")
	if err != nil {
		t.Fatalf("pavona new filtered completion failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "monorepo-go") || strings.Contains(out, "tool\t") {
		t.Errorf("expected completion filtered by prefix, got:\n%s", out)
	}
}

func TestMonorepoGoTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := testBinary
	outDir := filepath.Join(tmp, "my-monorepo")

	out, err := runPavonaInDir(t, bin, tmp, "new", "monorepo-go", "my-monorepo", "-q")
	if err != nil {
		t.Fatalf("pavona new monorepo-go failed: %v\n%s", err, out)
	}

	for _, path := range []string{
		"go.work",
		"Taskfile.yaml",
		"README.md",
		"AGENTS.md",
		".editorconfig",
		".editorconfig-checker.json",
		".lefthook.yaml",
		".github/CODEOWNERS",
		".github/workflows/sv-release.yml",
		"docs",
	} {
		if _, err := os.Stat(filepath.Join(outDir, path)); err != nil {
			t.Errorf("expected %q to exist: %v", path, err)
		}
	}

	checks := map[string]string{
		"go.work":                          "go 1.27.1",
		"Taskfile.yaml":                    "TODO build",
		".github/workflows/sv-release.yml": "actions/checkout",
	}
	for path, expected := range checks {
		data, err := os.ReadFile(filepath.Join(outDir, path))
		if err != nil {
			t.Errorf("reading %q: %v", path, err)
			continue
		}
		if !strings.Contains(string(data), expected) {
			t.Errorf("expected %q to contain %q, got:\n%s", path, expected, data)
		}
	}
}

func TestPavonaTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := testBinary
	templateDir := filepath.Join(tmp, "my-template")

	out, err := runPavonaInDir(t, bin, tmp, "new", "pavona", "my-template", "-q")
	if err != nil {
		t.Fatalf("pavona new pavona failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(filepath.Join(templateDir, "config.cue")); err != nil {
		t.Errorf("expected config.cue to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(templateDir, "main.go.tmpl")); !os.IsNotExist(err) {
		t.Errorf("expected main.go.tmpl not to exist, got error: %v", err)
	}

	projectDir := filepath.Join(tmp, "generated-project")
	out, err = runPavonaInDir(t, bin, tmp, "new", templateDir, "generated-project", "-q")
	if err != nil {
		t.Fatalf("creating project from generated template failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "main.go")); !os.IsNotExist(err) {
		t.Errorf("expected generated main.go not to exist, got error: %v", err)
	}
}

func TestToolTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := testBinary
	outDir := filepath.Join(tmp, "my-cli")

	out, err := runPavonaInDir(t, bin, tmp, "new", "tool", "my-cli", "-q")
	if err != nil {
		t.Fatalf("pavona new tool failed: %v\n%s", err, out)
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
	bin := testBinary
	outDir := filepath.Join(tmp, "go-csvstream")

	out, err := runPavonaInDir(t, bin, tmp, "new", "lib", "go-csvstream", "-q")
	if err != nil {
		t.Fatalf("pavona new lib failed: %v\n%s", err, out)
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
	bin := testBinary
	outDir := filepath.Join(tmp, "blog")

	out, err := runPavonaInDir(t, bin, tmp, "new", "site", "blog", "-q")
	if err != nil {
		t.Fatalf("pavona new site failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(filepath.Join(outDir, "content/index.md")); os.IsNotExist(err) {
		t.Errorf("expected content/index.md to exist")
	}
}

func TestTuiTemplate(t *testing.T) {
	tmp := t.TempDir()
	bin := testBinary
	outDir := filepath.Join(tmp, "chatmonitor")

	out, err := runPavonaInDir(t, bin, tmp, "new", "tui", "chatmonitor", "-q")
	if err != nil {
		t.Fatalf("pavona new tui failed: %v\n%s", err, out)
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
	bin := testBinary
	outDir := filepath.Join(tmp, "acmecorp")

	out, err := runPavonaInDir(t, bin, tmp, "new", "app", "acmecorp", "-q")
	if err != nil {
		t.Fatalf("pavona new app failed: %v\n%s", err, out)
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
	bin := testBinary
	outDir := filepath.Join(tmp, "triagebot")

	out, err := runPavonaInDir(t, bin, tmp, "new", "agent", "triagebot", "-q")
	if err != nil {
		t.Fatalf("pavona new agent failed: %v\n%s", err, out)
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
	bin := testBinary

	customDir := filepath.Join(tmp, "custom-tmpl")
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := `package template

name:        "custom"
description: "A custom template for testing"

variables: {
	// Project name
	project_name: string

	// Greeting message
	message?: string | *"Hello, World!"
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

	outDir := filepath.Join(tmp, "custom-test")
	out, err := runPavonaInDir(t, bin, tmp, "new", customDir, "custom-test", "-q")
	if err != nil {
		t.Fatalf("pavona new custom failed: %v\n%s", err, out)
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
	bin := testBinary
	out, err := runPavona(t, bin)
	if err != nil {
		t.Fatalf("pavona with no arguments failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "---|---.") || !strings.Contains(out, "\x1b[38;2;255;127;80m") {
		t.Errorf("expected no-argument help to show branching coral art in coral color, got:\n%s", out)
	}

	out, err = runPavona(t, bin, "--help")
	if err != nil {
		t.Fatalf("pavona --help failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "---|---.") || !strings.Contains(out, "\x1b[38;2;255;127;80m") {
		t.Errorf("expected --help to show branching coral art in coral color, got:\n%s", out)
	}
	if !strings.Contains(out, "template") {
		t.Errorf("expected --help output to mention templates, got:\n%s", out)
	}
}
