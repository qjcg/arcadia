package semver

import "testing"

func TestCalculateNext(t *testing.T) {
	tests := []struct {
		name    string
		current string
		path    string
		commits []string
		want    string
	}{
		{
			name:    "root patch",
			current: "v1.0.0",
			path:    ".",
			commits: []string{"fix: bug"},
			want:    "v1.0.1",
		},
		{
			name:    "root minor",
			current: "v1.0.0",
			path:    ".",
			commits: []string{"feat: feature"},
			want:    "v1.1.0",
		},
		{
			name:    "module path-based tag",
			current: "x/slidedeck/v2.3.1",
			path:    "x/slidedeck",
			commits: []string{"feat: feature"},
			want:    "x/slidedeck/v2.4.0",
		},
		{
			name:    "breaking change",
			current: "v1.0.0",
			path:    ".",
			commits: []string{"feat!: break"},
			want:    "v2.0.0",
		},
		{
			name:    "breaking change footer",
			current: "v1.0.0",
			path:    ".",
			commits: []string{"feat: something", "BREAKING CHANGE: something broke"},
			want:    "v2.0.0",
		},
		{
			name:    "multiple commits",
			current: "v1.0.0",
			path:    ".",
			commits: []string{"fix: bug1", "fix: bug2", "feat: feat1"},
			want:    "v1.1.0",
		},
		{
			name:    "no relevant commits",
			current: "v1.0.0",
			path:    ".",
			commits: []string{"chore: docs", "style: lint"},
			want:    "v1.0.0",
		},
		{
			name:    "initial version",
			current: "",
			path:    ".",
			commits: []string{"feat: initial"},
			want:    "v0.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateNext(tt.current, tt.path, tt.commits)
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
