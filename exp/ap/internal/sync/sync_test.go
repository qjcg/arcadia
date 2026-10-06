package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qjcg/arcadia/exp/ap/internal/spec"
)

func TestRunLocalSource(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(repo, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("agents/skills/demo/SKILL.md", "---\nname: demo\ndescription: A demo skill.\ndisable-model-invocation: true\n---\n# Demo\n")
	write("agents/skills/demo/references/notes.md", "notes\n")
	write("agents/skills/LICENSE", "license text\n")

	s := spec.Spec{
		Name:        "demo-plugin",
		Version:     "0.1.0",
		Description: "A demo plugin.",
		License:     "MIT",
		Author:      spec.Author{Name: "Someone"},
		Source: spec.Source{
			Type:     "local",
			Path:     "agents/skills",
			Skills:   []string{"demo"},
			CopyRoot: []string{"LICENSE"},
		},
	}

	out := filepath.Join(repo, "out")
	roots, err := Run([]spec.Spec{s}, repo, out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(roots) != 1 {
		t.Fatalf("roots = %v", roots)
	}
	root := roots[0]

	for _, rel := range []string{
		"plugin.json",
		"LICENSE",
		"skills/demo/SKILL.md",
		"skills/demo/references/notes.md",
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}

	raw, err := os.ReadFile(filepath.Join(root, "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"$schema": "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json"`,
		`"name": "demo-plugin"`,
		`"version": "0.1.0"`,
	} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("plugin.json missing %s:\n%s", want, raw)
		}
	}

	skill, err := os.ReadFile(filepath.Join(root, "skills/demo/SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(skill), "disable-model-invocation: true") {
		t.Errorf("frontmatter not normalized:\n%s", skill)
	}
	if !strings.Contains(string(skill), "disable-model-invocation: \"true\"") {
		t.Errorf("metadata missing moved field:\n%s", skill)
	}
	if !strings.HasSuffix(string(skill), "# Demo\n") {
		t.Errorf("body not preserved:\n%s", skill)
	}
}
