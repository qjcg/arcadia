package semver

import "testing"

func TestCalculateNext(t *testing.T) {
	tests := []struct {
		name         string
		current      string
		path         string
		commits      []Commit
		defaultPatch bool
		want         string
	}{
		{
			name:    "root patch",
			current: "v1.0.0",
			path:    ".",
			commits: []Commit{{Message: "fix: bug"}},
			want:    "v1.0.1",
		},
		{
			name:    "root minor",
			current: "v1.0.0",
			path:    ".",
			commits: []Commit{{Message: "feat: feature"}},
			want:    "v1.1.0",
		},
		{
			name:    "module path-based tag",
			current: "x/slidedeck/v2.3.1",
			path:    "x/slidedeck",
			commits: []Commit{{Message: "feat: feature"}},
			want:    "x/slidedeck/v2.4.0",
		},
		{
			name:    "breaking change",
			current: "v1.0.0",
			path:    ".",
			commits: []Commit{{Message: "feat!: break"}},
			want:    "v2.0.0",
		},
		{
			name:    "breaking change footer",
			current: "v1.0.0",
			path:    ".",
			commits: []Commit{{Message: "feat: something"}, {Message: "BREAKING CHANGE: something broke"}},
			want:    "v2.0.0",
		},
		{
			name:    "multiple commits",
			current: "v1.0.0",
			path:    ".",
			commits: []Commit{{Message: "fix: bug1"}, {Message: "fix: bug2"}, {Message: "feat: feat1"}},
			want:    "v1.1.0",
		},
		{
			name:         "no relevant commits with default patch",
			current:      "v1.0.0",
			path:         ".",
			commits:      []Commit{{Message: "chore: docs"}, {Message: "style: lint"}},
			defaultPatch: true,
			want:         "v1.0.1",
		},
		{
			name:    "initial version",
			current: "",
			path:    ".",
			commits: []Commit{{Message: "feat: initial"}},
			want:    "v0.1.0",
		},
		{
			name:    "pre-release patch bump",
			current: "v1.0.0-alpha",
			path:    ".",
			commits: []Commit{{Message: "fix: bug"}},
			want:    "v1.0.0",
		},
		{
			name:    "pre-release minor bump",
			current: "v1.0.0-rc.1",
			path:    ".",
			commits: []Commit{{Message: "feat: feature"}},
			want:    "v1.1.0",
		},
		{
			name:    "zero patch version",
			current: "v0.0.1",
			path:    ".",
			commits: []Commit{{Message: "fix: bug"}},
			want:    "v0.0.2",
		},
		{
			name:    "zero minor version bump",
			current: "v0.1.0",
			path:    ".",
			commits: []Commit{{Message: "feat: feature"}},
			want:    "v0.2.0",
		},
		{
			name:    "zero major version",
			current: "v0.1.0",
			path:    ".",
			commits: []Commit{{Message: "feat!: breaking"}},
			want:    "v1.0.0",
		},
		{
			name:    "scoped breaking change ignored on root with no source files",
			current: "v0.1.0",
			path:    ".",
			commits: []Commit{{Message: "feat(sv)!: major change in sv", Files: []string{"go.work", "README.md"}}},
			want:    "v0.2.0", // feat without ! (no .go files in scope) → minor
		},
		{
			name:    "scoped breaking change applies with source files",
			current: "v0.1.0",
			path:    ".",
			commits: []Commit{{Message: "feat(sv)!: major change", Files: []string{"main.go"}}},
			want:    "v1.0.0", // ! with .go file → major
		},
		{
			name:    "scoped breaking change on submodule always applies",
			current: "v0.1.0",
			path:    "cmd/sv",
			commits: []Commit{{Message: "feat(sv)!: major change"}},
			want:    "cmd/sv/v1.0.0", // no file info → backward compatible, ! counts
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateNext(tt.current, tt.path, tt.commits, tt.defaultPatch)
			if err != nil {
				t.Errorf("CalculateNext() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateNext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIncrement(t *testing.T) {
	tests := []struct {
		name    string
		current string
		path    string
		bump    Bump
		want    string
	}{
		{"root major", "v1.0.0", ".", BumpMajor, "v2.0.0"},
		{"root minor", "v1.0.0", ".", BumpMinor, "v1.1.0"},
		{"root patch", "v1.0.0", ".", BumpPatch, "v1.0.1"},
		{"module major", "x/mod/v1.0.0", "x/mod", BumpMajor, "x/mod/v2.0.0"},
		{"empty current", "", ".", BumpMinor, "v0.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Increment(tt.current, tt.path, tt.bump)
			if err != nil {
				t.Errorf("Increment() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Increment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateNext_Errors(t *testing.T) {
	tests := []struct {
		name    string
		current string
		path    string
		commits []Commit
		wantErr bool
	}{
		{
			name:    "invalid semver",
			current: "not-a-version",
			path:    ".",
			commits: []Commit{{Message: "feat: feature"}},
			wantErr: true,
		},
		{
			name:    "invalid semver with module path",
			current: "x/mod/bad-version",
			path:    "x/mod",
			commits: []Commit{{Message: "fix: bug"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateNext(tt.current, tt.path, tt.commits, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateNext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIncrement_Errors(t *testing.T) {
	tests := []struct {
		name    string
		current string
		path    string
		bump    Bump
		wantErr bool
	}{
		{
			name:    "invalid semver",
			current: "abc",
			path:    ".",
			bump:    BumpMinor,
			wantErr: true,
		},
		{
			name:    "invalid bump type",
			current: "v1.0.0",
			path:    ".",
			bump:    BumpNone,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Increment(tt.current, tt.path, tt.bump)
			if (err != nil) != tt.wantErr {
				t.Errorf("Increment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
