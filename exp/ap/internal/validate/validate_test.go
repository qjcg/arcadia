package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSkillName(t *testing.T) {
	t.Parallel()

	for _, good := range []string{"a", "demo", "go-gherkin-testing", "skill2"} {
		if err := checkSkillName(good); err != nil {
			t.Errorf("checkSkillName(%q) = %v, want nil", good, err)
		}
	}
	for _, bad := range []string{"", "-demo", "demo-", "Demo", "two--dashes", "under_score", "dot.name"} {
		if err := checkSkillName(bad); err == nil {
			t.Errorf("checkSkillName(%q) = nil, want error", bad)
		}
	}
}

func TestCheckPluginName(t *testing.T) {
	t.Parallel()

	for _, good := range []string{"a", "mattpocock-skills", "acme.tools", "lint3r"} {
		if err := checkPluginName(good); err != nil {
			t.Errorf("checkPluginName(%q) = %v, want nil", good, err)
		}
	}
	for _, bad := range []string{"", "-start", "has--double", "too.many..dots", "My-Plugin"} {
		if err := checkPluginName(bad); err == nil {
			t.Errorf("checkPluginName(%q) = nil, want error", bad)
		}
	}
}

func TestValidateSkillDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skill := filepath.Join(dir, "demo")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("---\nname: demo\ndescription: A demo skill.\n---\n# Demo\n")
	if errs := validateSkillDir(skill); len(errs) > 0 {
		t.Errorf("valid skill rejected: %v", errs)
	}

	write("---\nname: other\ndescription: A demo skill.\n---\n# Demo\n")
	assertErrorContains(t, validateSkillDir(skill), "must match directory name")

	write("---\nname: demo\ndescription: A demo skill.\ndisable-model-invocation: true\n---\n# Demo\n")
	assertErrorContains(t, validateSkillDir(skill), "unknown frontmatter field")

	write("---\nname: demo\ndescription: " + strings.Repeat("d", 1025) + "\n---\n# Demo\n")
	assertErrorContains(t, validateSkillDir(skill), "maximum is 1024")

	write("---\nname: demo\ndescription: A demo skill.\nmetadata:\n  nested:\n    key: value\n---\n# Demo\n")
	assertErrorContains(t, validateSkillDir(skill), "metadata.nested must be a string")
}

func TestValidatePlugin(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("plugin.json", `{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json",
  "name": "demo-plugin",
  "version": "1.0.0",
  "description": "A demo plugin.",
  "author": {"name": "Someone", "url": "https://example.com"},
  "license": "MIT"
}`)
	write("skills/demo/SKILL.md", "---\nname: demo\ndescription: A demo skill.\n---\n# Demo\n")
	write("README.md", "See [the skill](skills/demo/SKILL.md) and [docs](../docs).\n")

	errs := validatePlugin(root)
	assertErrorContains(t, errs, "escapes the plugin root")

	write("plugin.json", `{
  "$schema": "https://example.com/plugin.schema.json",
  "name": "demo-plugin",
  "skills": ["demo"],
  "author": {"team": "x"}
}`)
	errs = validatePlugin(root)
	assertErrorContains(t, errs, `unknown field "skills"`)
	assertErrorContains(t, errs, "author: unknown field")
	assertErrorContains(t, errs, "$schema must be")
}

func TestValidateDispatch(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skill := filepath.Join(root, "skills", "demo")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("---\nname: demo\ndescription: D.\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	errs, err := Validate(filepath.Join(root, "skills"))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(errs) > 0 {
		t.Errorf("skills tree rejected: %v", errs)
	}
}

func assertErrorContains(t *testing.T, errs []error, want string) {
	t.Helper()
	for _, err := range errs {
		if strings.Contains(err.Error(), want) {
			return
		}
	}
	t.Errorf("no error containing %q in %v", want, errs)
}
