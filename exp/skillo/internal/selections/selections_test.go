package selections

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()

	// Empty load
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 0 {
		t.Fatalf("expected empty, got %v", s)
	}

	// Save and reload
	s["github.com/user/repo"] = []string{"skill-a", "skill-b"}
	if err := Save(dir, s); err != nil {
		t.Fatal(err)
	}

	s2, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s2) != 1 {
		t.Fatalf("expected 1 module, got %d", len(s2))
	}
	if len(s2["github.com/user/repo"]) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(s2["github.com/user/repo"]))
	}
}

func TestInit(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}

	// File should exist with empty object
	data, err := os.ReadFile(filepath.Join(dir, "selections.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}" {
		t.Fatalf("expected {}, got %s", string(data))
	}

	// Second init should be idempotent
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}
}

func TestAddModule(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}

	// Add module with skills
	if err := AddModule(dir, "github.com/user/repo", []string{"skill-a", "skill-b"}); err != nil {
		t.Fatal(err)
	}

	s, _ := Load(dir)
	if len(s) != 1 {
		t.Fatalf("expected 1 module, got %d", len(s))
	}
	if len(s["github.com/user/repo"]) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(s["github.com/user/repo"]))
	}

	// Add module with empty skills list
	if err := AddModule(dir, "github.com/user/other", []string{}); err != nil {
		t.Fatal(err)
	}
	s, _ = Load(dir)
	if len(s["github.com/user/other"]) != 0 {
		t.Fatalf("expected empty skills, got %v", s["github.com/user/other"])
	}

	// Replace existing module
	if err := AddModule(dir, "github.com/user/repo", []string{"skill-c"}); err != nil {
		t.Fatal(err)
	}
	s, _ = Load(dir)
	if len(s["github.com/user/repo"]) != 1 || s["github.com/user/repo"][0] != "skill-c" {
		t.Fatalf("expected [skill-c], got %v", s["github.com/user/repo"])
	}
}

func TestRemoveSkill(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}

	AddModule(dir, "github.com/user/repo", []string{"skill-a", "skill-b", "skill-c"})

	// Remove a skill
	if err := RemoveSkill(dir, "skill-b"); err != nil {
		t.Fatal(err)
	}

	s, _ := Load(dir)
	if len(s["github.com/user/repo"]) != 2 {
		t.Fatalf("expected 2 skills, got %v", s["github.com/user/repo"])
	}

	// Remove last skill should also remove module
	if err := RemoveSkill(dir, "skill-a"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveSkill(dir, "skill-c"); err != nil {
		t.Fatal(err)
	}
	s, _ = Load(dir)
	if _, ok := s["github.com/user/repo"]; ok {
		t.Fatalf("expected module to be removed")
	}

	// Remove nonexistent skill
	if err := RemoveSkill(dir, "nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent skill")
	}
}

func TestRemoveModule(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}

	AddModule(dir, "github.com/user/repo", []string{"skill-a"})
	AddModule(dir, "github.com/user/other", []string{"skill-z"})

	if err := RemoveModule(dir, "github.com/user/repo"); err != nil {
		t.Fatal(err)
	}

	s, _ := Load(dir)
	if len(s) != 1 {
		t.Fatalf("expected 1 module, got %d", len(s))
	}
	if _, ok := s["github.com/user/other"]; !ok {
		t.Fatalf("expected other to remain")
	}
}

func TestFindModule(t *testing.T) {
	s := Selections{
		"github.com/user/repo":  {"skill-a", "skill-b"},
		"github.com/user/other": {"skill-c"},
	}

	if m := FindModule(s, "skill-a"); m != "github.com/user/repo" {
		t.Fatalf("expected github.com/user/repo, got %s", m)
	}
	if m := FindModule(s, "skill-c"); m != "github.com/user/other" {
		t.Fatalf("expected github.com/user/other, got %s", m)
	}
	if m := FindModule(s, "nonexistent"); m != "" {
		t.Fatalf("expected '', got %s", m)
	}

	// Empty skills list (all skills) — returns the module
	s2 := Selections{
		"github.com/user/all": {},
	}
	if m := FindModule(s2, "anything"); m != "github.com/user/all" {
		t.Fatalf("expected github.com/user/all, got %s", m)
	}
}

func TestModuleSkills(t *testing.T) {
	s := Selections{
		"github.com/user/repo": {"skill-a"},
	}
	if skills := ModuleSkills(s, "github.com/user/repo"); len(skills) != 1 || skills[0] != "skill-a" {
		t.Fatalf("expected [skill-a], got %v", skills)
	}
	if skills := ModuleSkills(s, "nonexistent"); skills != nil {
		t.Fatalf("expected nil, got %v", skills)
	}
}

func TestConvertLegacyManifest(t *testing.T) {
	legacy := map[string][]string{
		"github.com/user/repo": {"skill-a", "skill-b"},
	}
	sel := ConvertLegacyManifest(legacy)
	if len(sel) != 1 {
		t.Fatalf("expected 1 module, got %d", len(sel))
	}
	if len(sel["github.com/user/repo"]) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(sel["github.com/user/repo"]))
	}
}
