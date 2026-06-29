package semver

import "testing"

func TestParseRetractions(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []Retraction
	}{
		{
			name:    "empty",
			content: "",
			want:    nil,
		},
		{
			name: "single version inline",
			content: `module example.com/m
go 1.21
retract v1.0.0
`,
			want: []Retraction{{Low: "v1.0.0", High: "v1.0.0"}},
		},
		{
			name: "range single line",
			content: `module example.com/m
retract [v1.0.0, v1.9.9]
`,
			want: []Retraction{{Low: "v1.0.0", High: "v1.9.9"}},
		},
		{
			name: "block form",
			content: `module example.com/m
retract (
    v1.0.0
    [v1.1.0, v1.5.0]
)
`,
			want: []Retraction{
				{Low: "v1.0.0", High: "v1.0.0"},
				{Low: "v1.1.0", High: "v1.5.0"},
			},
		},
		{
			name: "with rationale comments",
			content: `module example.com/m

// Published accidentally.
retract v1.0.0

// Contains severe bug.
retract [v1.1.0, v1.2.0]
`,
			want: []Retraction{
				{Low: "v1.0.0", High: "v1.0.0"},
				{Low: "v1.1.0", High: "v1.2.0"},
			},
		},
		{
			name: "self-retraction",
			content: `module example.com/m
retract v1.0.0
retract v1.0.1
`,
			want: []Retraction{
				{Low: "v1.0.0", High: "v1.0.0"},
				{Low: "v1.0.1", High: "v1.0.1"},
			},
		},
		{
			name: "no retract directives",
			content: `module example.com/m
go 1.21
`,
			want: nil,
		},
		{
			name: "range with block comment above",
			content: `module example.com/m

// Retract all v0.x versions after accidentally releasing v1.0.0
retract [v0.0.0, v1.0.1]
`,
			want: []Retraction{{Low: "v0.0.0", High: "v1.0.1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRetractions(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("ParseRetractions() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i].Low != tt.want[i].Low || got[i].High != tt.want[i].High {
					t.Errorf("ParseRetractions()[%d] = {%s, %s}, want {%s, %s}",
						i, got[i].Low, got[i].High, tt.want[i].Low, tt.want[i].High)
				}
			}
		})
	}
}

func TestIsVersionRetracted(t *testing.T) {
	retractions := []Retraction{
		{Low: "v1.0.0", High: "v1.0.0"},
		{Low: "v1.2.0", High: "v1.5.0"},
		{Low: "v2.0.0", High: "v2.0.0"},
	}

	tests := []struct {
		version string
		want    bool
	}{
		{"v1.0.0", true},
		{"v1.0.1", false},
		{"v1.2.0", true},
		{"v1.3.0", true},
		{"v1.5.0", true},
		{"v1.6.0", false},
		{"v2.0.0", true},
		{"v2.0.1", false},
		{"v0.9.0", false},
		{"v3.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := IsVersionRetracted(tt.version, retractions)
			if got != tt.want {
				t.Errorf("IsVersionRetracted(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestIsVersionRetracted_InvalidVersion(t *testing.T) {
	retractions := []Retraction{{Low: "v1.0.0", High: "v1.0.0"}}
	if IsVersionRetracted("not-a-version", retractions) {
		t.Error("IsVersionRetracted should return false for invalid versions")
	}
}
