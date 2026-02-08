package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindModules(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sv-discovery-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy monorepo structure
	dirs := []string{
		"root-mod",
		"x/mod1",
		"x/mod2",
		"pkg/internal",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Create go.mod files
	goMods := []string{
		"go.mod",
		"x/mod1/go.mod",
		"x/mod2/go.mod",
	}
	for _, f := range goMods {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("module test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	modules, err := FindModules(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{
		".":      true,
		"x/mod1": true,
		"x/mod2": true,
	}

	if len(modules) != len(expected) {
		t.Errorf("expected %d modules, got %d", len(expected), len(modules))
	}

	for _, m := range modules {
		if !expected[m.Name] {
			t.Errorf("found unexpected module: %s", m.Name)
		}
	}
}

func TestGetCurrentModule(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sv-current-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "x/mod1"), 0o755)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module root"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "x/mod1/go.mod"), []byte("module mod1"), 0o644)

	tests := []struct {
		name     string
		dir      string
		wantName string
	}{
		{"at root", tmpDir, "."},
		{"inside nested", filepath.Join(tmpDir, "x/mod1"), "x/mod1"},
		{"inside nested sub-path", filepath.Join(tmpDir, "x/mod1/some/deep/path"), "x/mod1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := GetCurrentModule(tmpDir, tt.dir)
			if err != nil {
				t.Fatal(err)
			}
			if m.Name != tt.wantName {
				t.Errorf("GetCurrentModule() = %v, want %v", m.Name, tt.wantName)
			}
		})
	}
}
