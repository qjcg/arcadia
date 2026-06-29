package changelog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qjcg/arcadia/cmd/sv/internal/git"
)

func TestCategorizeCommit(t *testing.T) {
	tests := []struct {
		msg     string
		cat     Category
		cleaned string
		include bool
	}{
		{
			msg:     "feat: add new feature",
			cat:     Added,
			cleaned: "add new feature",
			include: true,
		},
		{
			msg:     "fix: resolve crashing bug",
			cat:     Fixed,
			cleaned: "resolve crashing bug",
			include: true,
		},
		{
			msg:     "feat(api): add user endpoint",
			cat:     Added,
			cleaned: "add user endpoint",
			include: true,
		},
		{
			msg:     "feat!: breaking change",
			cat:     Changed,
			cleaned: "breaking change",
			include: true,
		},
		{
			msg:     "feat(api)!: breaking api change",
			cat:     Changed,
			cleaned: "breaking api change",
			include: true,
		},
		{
			msg:     "docs: update readme",
			cat:     Changed,
			cleaned: "update readme",
			include: true,
		},
		{
			msg:     "chore: bump dependencies",
			cat:     Changed,
			cleaned: "bump dependencies",
			include: true,
		},
		{
			msg:     "refactor: simplify logic",
			cat:     Changed,
			cleaned: "simplify logic",
			include: true,
		},
		{
			msg:     "deprecated: old endpoint",
			cat:     Deprecated,
			cleaned: "old endpoint",
			include: true,
		},
		{
			msg:     "security: fix XSS",
			cat:     Security,
			cleaned: "fix XSS",
			include: true,
		},
		{
			msg:     "unknown: something random",
			cat:     "",
			cleaned: "",
			include: false,
		},
		{
			msg:     "BREAKING CHANGE: complete overhaul",
			cat:     Changed,
			cleaned: "complete overhaul",
			include: true,
		},
		{
			msg:     "",
			cat:     "",
			cleaned: "",
			include: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			cat, cleaned, include := categorizeCommit(tt.msg)
			if cat != tt.cat {
				t.Errorf("category = %q, want %q", cat, tt.cat)
			}
			if cleaned != tt.cleaned {
				t.Errorf("cleaned = %q, want %q", cleaned, tt.cleaned)
			}
			if include != tt.include {
				t.Errorf("include = %v, want %v", include, tt.include)
			}
		})
	}
}

func TestFormatItem(t *testing.T) {
	t.Run("without url prefix", func(t *testing.T) {
		URLPrefix = ""
		got := FormatItem("abc1234", "add login page")
		want := "- abc1234 - add login page"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("with url prefix", func(t *testing.T) {
		URLPrefix = "https://github.com/org/repo/commit/"
		defer func() { URLPrefix = "" }()
		got := FormatItem("abc1234", "add login page")
		want := "- [abc1234](https://github.com/org/repo/commit/abc1234) - add login page"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestFormatChangelog(t *testing.T) {
	cl := &Changelog{
		Entries: []Entry{
			{
				Version: "v1.0.0",
				Date:    "2024-01-15",
				Items: map[Category][]Item{
					Added: {{Hash: "abc1234", Message: "New CLI tool"}, {Hash: "def5678", Message: "API endpoint"}},
					Fixed: {{Hash: "ghi9012", Message: "Crash on startup"}},
				},
			},
			{
				Version:  "v0.1.0",
				Date:     "2023-12-01",
				Overview: "Initial release of the project.",
				Items: map[Category][]Item{
					Added: {{Hash: "jkl3456", Message: "Basic functionality"}},
				},
			},
		},
	}

	output := FormatChangelog(cl)

	if !contains(output, "# Changelog") {
		t.Error("missing main header")
	}
	if !contains(output, "Keep a Changelog") {
		t.Error("missing keepachangelog reference")
	}
	if !contains(output, "## [v1.0.0] - 2024-01-15") {
		t.Error("missing v1.0.0 header")
	}
	if !contains(output, "### Added") {
		t.Error("missing Added section")
	}
	if !contains(output, "### Fixed") {
		t.Error("missing Fixed section")
	}
	if !contains(output, "- abc1234 - New CLI tool") {
		t.Error("missing formatted item")
	}
	if !contains(output, "- ghi9012 - Crash on startup") {
		t.Error("missing formatted fixed item")
	}
	if !contains(output, "## [v0.1.0] - 2023-12-01") {
		t.Error("missing v0.1.0 header")
	}
	if !contains(output, "Initial release of the project.") {
		t.Error("missing overview text")
	}

	v1Idx := indexOf(output, "v1.0.0")
	v0Idx := indexOf(output, "v0.1.0")
	if v1Idx > v0Idx {
		t.Error("entries not in newest-first order")
	}
}

func TestFormatChangelog_Unreleased(t *testing.T) {
	cl := &Changelog{
		Entries: []Entry{
			{
				Version: "unreleased",
				Items: map[Category][]Item{
					Added: {{Hash: "xyz7890", Message: "Work in progress feature"}},
				},
			},
		},
	}

	output := FormatChangelog(cl)

	if !contains(output, "## [unreleased]") {
		t.Error("missing unreleased header")
	}
	if contains(output, "2024") {
		t.Error("unreleased should not have a date")
	}
	if !contains(output, "- xyz7890 - Work in progress feature") {
		t.Error("missing unreleased item")
	}
}

func TestFormatEntry(t *testing.T) {
	entry := Entry{
		Version: "v1.0.0",
		Date:    "2024-06-01",
		Items: map[Category][]Item{
			Fixed: {{Hash: "aaa1111", Message: "Bug fix"}},
		},
	}

	output := FormatEntry(entry)

	if !contains(output, "## [v1.0.0] - 2024-06-01") {
		t.Error("missing entry header")
	}
	if !contains(output, "- aaa1111 - Bug fix") {
		t.Error("missing formatted bug fix item")
	}
}

func TestFormatEntryWithOverview(t *testing.T) {
	entry := Entry{
		Version:  "v2.0.0",
		Date:     "2024-06-15",
		Overview: "Major rewrite of the core engine.",
		Items: map[Category][]Item{
			Changed: {{Hash: "bbb2222", Message: "Complete rewrite"}},
		},
	}

	output := FormatEntry(entry)

	if !contains(output, "Major rewrite of the core engine.") {
		t.Error("missing overview")
	}
}

func TestParseSince(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2025", "2025-01-01"},
		{"2024", "2024-01-01"},
		{"2024-06-15", "2024-06-15"},
		{"", ""},
		{"abc", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSince(tt.input)
			if tt.want == "" {
				if err == nil {
					t.Errorf("expected error, got %q", result)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.want {
				t.Errorf("got %q, want %q", result, tt.want)
			}
		})
	}

	t.Run("8w", func(t *testing.T) {
		result, err := parseSince("8w")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 10 {
			t.Errorf("expected date format, got %q", result)
		}
	})
}

func TestBuildEntry(t *testing.T) {
	commits := []git.CommitInfo{
		{Short: "aaa1111", Message: "feat: add login page"},
		{Short: "bbb2222", Message: "fix: fix broken link"},
		{Short: "ccc3333", Message: "docs: update readme"},
		{Short: "ddd4444", Message: "chore: bump deps"},
		{Short: "eee5555", Message: "unknown: skip this"},
	}

	entry := buildEntry("v1.0.0", "2024-01-15", commits)

	if entry.Version != "v1.0.0" {
		t.Errorf("version = %q, want %q", entry.Version, "v1.0.0")
	}
	if entry.Date != "2024-01-15" {
		t.Errorf("date = %q, want %q", entry.Date, "2024-01-15")
	}

	if len(entry.Items[Added]) != 1 || entry.Items[Added][0].Message != "add login page" {
		t.Errorf("Added items = %v", entry.Items[Added])
	}
	if len(entry.Items[Fixed]) != 1 || entry.Items[Fixed][0].Message != "fix broken link" {
		t.Errorf("Fixed items = %v", entry.Items[Fixed])
	}
	if len(entry.Items[Changed]) != 2 {
		t.Errorf("Changed items = %v, want 2 items", entry.Items[Changed])
	}
}

func TestWriteEntryDir(t *testing.T) {
	dir := t.TempDir()

	cl := &Changelog{
		Entries: []Entry{
			{
				Version:  "v1.0.0",
				Date:     "2024-01-15",
				Overview: "First stable release.",
				Items: map[Category][]Item{
					Added: {{Hash: "fff6666", Message: "Feature A"}},
				},
			},
			{
				Version: "unreleased",
				Items: map[Category][]Item{
					Added: {{Hash: "ggg7777", Message: "Feature B"}},
				},
			},
		},
	}

	if err := WriteEntryDir(dir, cl); err != nil {
		t.Fatalf("WriteEntryDir failed: %v", err)
	}

	v1Path := filepath.Join(dir, "v1.0.0.md")
	if _, err := os.Stat(v1Path); os.IsNotExist(err) {
		t.Error("v1.0.0.md not written")
	}
	content, err := os.ReadFile(v1Path)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(content), "- fff6666 - Feature A") {
		t.Error("v1.0.0.md missing formatted item")
	}

	overviewPath := filepath.Join(dir, "v1.0.0_overview.md")
	if _, err := os.Stat(overviewPath); os.IsNotExist(err) {
		t.Error("v1.0.0_overview.md not written")
	}

	unreleasedPath := filepath.Join(dir, "unreleased.md")
	if _, err := os.Stat(unreleasedPath); os.IsNotExist(err) {
		t.Error("unreleased.md not written")
	}
}

func TestLoadOverviewFiles(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "v1.0.0_overview.md"), []byte("First stable release.\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "v0.1.0_overview.md"), []byte("Initial release.\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "unreleased_overview.md"), []byte(""), 0o644)

	overviews, err := LoadOverviewFiles(dir)
	if err != nil {
		t.Fatalf("LoadOverviewFiles failed: %v", err)
	}

	if len(overviews) != 2 {
		t.Errorf("expected 2 overviews, got %d", len(overviews))
	}
	if overviews["v1.0.0"] != "First stable release." {
		t.Errorf("v1.0.0 overview = %q", overviews["v1.0.0"])
	}
	if overviews["v0.1.0"] != "Initial release." {
		t.Errorf("v0.1.0 overview = %q", overviews["v0.1.0"])
	}
	if _, ok := overviews["unreleased"]; ok {
		t.Error("unreleased overview should not be loaded (empty)")
	}
}

func TestWriteChangelogFile(t *testing.T) {
	dir := t.TempDir()

	cl := &Changelog{
		Entries: []Entry{
			{
				Version: "v1.0.0",
				Date:    "2024-01-15",
				Items: map[Category][]Item{
					Added: {{Hash: "abc1234", Message: "New CLI tool"}},
					Fixed: {{Hash: "def5678", Message: "Crash on startup"}},
				},
			},
			{
				Version: "v0.1.0",
				Date:    "2023-12-01",
				Items: map[Category][]Item{
					Added: {{Hash: "ghi9012", Message: "Basic functionality"}},
				},
			},
		},
	}

	if err := WriteChangelogFile(dir, cl); err != nil {
		t.Fatalf("WriteChangelogFile failed: %v", err)
	}

	path := filepath.Join(dir, "CHANGELOG.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("CHANGELOG.md not found: %v", err)
	}

	output := string(content)
	if !contains(output, "# Changelog") {
		t.Error("missing main header")
	}
	if !contains(output, "## [v1.0.0] - 2024-01-15") {
		t.Error("missing v1.0.0 header")
	}
	if !contains(output, "### Added") {
		t.Error("missing Added section")
	}
	if !contains(output, "- abc1234 - New CLI tool") {
		t.Error("missing formatted item")
	}
	if !contains(output, "## [v0.1.0] - 2023-12-01") {
		t.Error("missing v0.1.0 header")
	}
}

func TestWriteChangelogFile_EmptyEntries(t *testing.T) {
	dir := t.TempDir()

	cl := &Changelog{}

	if err := WriteChangelogFile(dir, cl); err != nil {
		t.Fatalf("WriteChangelogFile failed: %v", err)
	}

	path := filepath.Join(dir, "CHANGELOG.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("CHANGELOG.md not found: %v", err)
	}

	output := string(content)
	if !contains(output, "# Changelog") {
		t.Error("missing main header with empty entries")
	}
}

func TestLoadOverviewFiles_NoDir(t *testing.T) {
	overviews, err := LoadOverviewFiles("/nonexistent/path")
	if err != nil {
		t.Fatalf("LoadOverviewFiles should not fail for nonexistent dir: %v", err)
	}
	if len(overviews) != 0 {
		t.Errorf("expected empty, got %d", len(overviews))
	}
}

func TestFormatMultiModuleChangelog(t *testing.T) {
	modules := map[string]*Changelog{
		"pkg/foo": {
			Entries: []Entry{
				{
					Version: "v1.0.0",
					Date:    "2024-01-15",
					Items: map[Category][]Item{
						Added: {{Hash: "abc1234", Message: "New feature in foo"}},
					},
				},
			},
		},
		"pkg/bar": {
			Entries: []Entry{
				{
					Version: "v0.5.0",
					Date:    "2023-11-01",
					Items: map[Category][]Item{
						Fixed: {{Hash: "def5678", Message: "Bug fix in bar"}},
					},
				},
			},
		},
	}

	output := FormatMultiModuleChangelog(modules)

	// Single H1 header
	if !contains(output, "# Changelog") {
		t.Error("missing main header")
	}
	if !contains(output, "Keep a Changelog") {
		t.Error("missing keepachangelog reference")
	}

	// Module sections are H2
	if !contains(output, "## Module: pkg/foo") {
		t.Error("missing pkg/foo module header")
	}
	if !contains(output, "## Module: pkg/bar") {
		t.Error("missing pkg/bar module header")
	}

	// Version headers within modules are H3 (demoted one level)
	if !contains(output, "### [v1.0.0] - 2024-01-15") {
		t.Error("missing demoted v1.0.0 header")
	}
	if !contains(output, "### [v0.5.0] - 2023-11-01") {
		t.Error("missing demoted v0.5.0 header")
	}

	// Category sections within modules are H4
	if !contains(output, "#### Added") {
		t.Error("missing demoted Added section")
	}
	if !contains(output, "#### Fixed") {
		t.Error("missing demoted Fixed section")
	}

	// Module entries listed in alphabetical order (bar < foo)
	fooIdx := indexOf(output, "pkg/foo")
	barIdx := indexOf(output, "pkg/bar")
	if barIdx > fooIdx {
		t.Error("modules should be in alphabetical order: bar before foo")
	}
}

func TestFormatMultiModuleChangelog_SingleModule(t *testing.T) {
	modules := map[string]*Changelog{
		"pkg/baz": {
			Entries: []Entry{
				{
					Version: "v2.0.0",
					Date:    "2024-06-01",
					Items: map[Category][]Item{
						Changed: {{Hash: "ghi9012", Message: "Major refactor"}},
					},
				},
			},
		},
	}

	output := FormatMultiModuleChangelog(modules)

	// Should still use module wrapper even with one module
	if !contains(output, "## Module: pkg/baz") {
		t.Error("missing module header")
	}
	if !contains(output, "### [v2.0.0] - 2024-06-01") {
		t.Error("missing demoted version header")
	}
	if !contains(output, "#### Changed") {
		t.Error("missing demoted Changed section")
	}
}

func TestFormatMultiModuleChangelog_EmptyModules(t *testing.T) {
	modules := map[string]*Changelog{
		"pkg/empty":  {Entries: []Entry{}},
		"pkg/active": {Entries: []Entry{{Version: "v1.0.0", Date: "2024-01-01", Items: map[Category][]Item{Added: {{Hash: "aaa", Message: "Something"}}}}}},
	}

	output := FormatMultiModuleChangelog(modules)

	// Empty module should be skipped
	if contains(output, "Module: pkg/empty") {
		t.Error("empty module should be omitted")
	}
	if !contains(output, "## Module: pkg/active") {
		t.Error("non-empty module should appear")
	}
}

func TestFormatMultiModuleChangelog_Empty(t *testing.T) {
	output := FormatMultiModuleChangelog(map[string]*Changelog{})
	if !contains(output, "# Changelog") {
		t.Error("empty multi-module should still have header")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
