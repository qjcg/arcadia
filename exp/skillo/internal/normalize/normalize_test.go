package normalize

import (
	"testing"
)

func TestModulePath(t *testing.T) {
	tests := []struct {
		input       string
		wantModule  string
		wantVersion string
		wantErr     bool
	}{
		// Short form
		{"user/repo", "github.com/user/repo", "latest", false},
		{"user/repo@v1.2.3", "github.com/user/repo", "v1.2.3", false},
		{"org/repo", "github.com/org/repo", "latest", false},
		{"org/repo@main", "github.com/org/repo", "main", false},

		// Full Go import path
		{"github.com/user/repo", "github.com/user/repo", "latest", false},
		{"github.com/user/repo@v1.2.3", "github.com/user/repo", "v1.2.3", false},
		{"github.com/org/repo@latest", "github.com/org/repo", "latest", false},

		// HTTPS URL
		{"https://github.com/user/repo", "github.com/user/repo", "latest", false},
		{"https://github.com/user/repo@v1", "github.com/user/repo", "v1", false},
		{"http://github.com/user/repo", "github.com/user/repo", "latest", false},

		// Edge cases
		{"", "", "", true},
		{"@v1.2.3", "", "", true},
		{"justname", "", "", true},
	}

	for _, tt := range tests {
		mod, ver, err := ModulePath(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ModulePath(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			continue
		}
		if mod != tt.wantModule {
			t.Errorf("ModulePath(%q) module = %q, want %q", tt.input, mod, tt.wantModule)
		}
		if ver != tt.wantVersion {
			t.Errorf("ModulePath(%q) version = %q, want %q", tt.input, ver, tt.wantVersion)
		}
	}
}

func TestLooksLikeModulePath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"tester", false},
		{"my-skill", false},
		{"github.com/user/repo", true},
		{"org/repo", true},
		{"https://github.com/user/repo", true},
		{"", false},
		{"a.b", true},
	}

	for _, tt := range tests {
		got := LooksLikeModulePath(tt.input)
		if got != tt.want {
			t.Errorf("LooksLikeModulePath(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
