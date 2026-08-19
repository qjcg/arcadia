package cli

import (
	"testing"

	"github.com/qjcg/arcadia/cmd/sv/internal/discovery"
)

func TestFilterModules(t *testing.T) {
	mods := []discovery.Module{
		{Name: ".", Path: "/repo"},
		{Name: "cmd/sv", Path: "/repo/cmd/sv"},
		{Name: "exp/slidesdeck", Path: "/repo/exp/slidesdeck"},
		{Name: "exp/roubaix", Path: "/repo/exp/roubaix"},
		{Name: "exp/roubaix/internal", Path: "/repo/exp/roubaix/internal"},
		{Name: "skills/go-gherkin-testing/references/godog-examples", Path: "/repo/skills/go-gherkin-testing/references/godog-examples"},
	}

	tests := []struct {
		name     string
		excludes []string
		want     []string
	}{
		{"no excludes", nil, []string{".", "cmd/sv", "exp/slidesdeck", "exp/roubaix", "exp/roubaix/internal", "skills/go-gherkin-testing/references/godog-examples"}},
		{"empty excludes", []string{}, []string{".", "cmd/sv", "exp/slidesdeck", "exp/roubaix", "exp/roubaix/internal", "skills/go-gherkin-testing/references/godog-examples"}},
		{"single exclude", []string{"exp/roubaix"}, []string{".", "cmd/sv", "exp/slidesdeck", "skills/go-gherkin-testing/references/godog-examples"}},
		{"multiple excludes", []string{"exp/roubaix", "cmd/sv"}, []string{".", "exp/slidesdeck", "skills/go-gherkin-testing/references/godog-examples"}},
		{"exclude root", []string{"."}, []string{"cmd/sv", "exp/slidesdeck", "exp/roubaix", "exp/roubaix/internal", "skills/go-gherkin-testing/references/godog-examples"}},
		{"exclude subtree prunes all", []string{"exp/roubaix/internal"}, []string{".", "cmd/sv", "exp/slidesdeck", "exp/roubaix", "skills/go-gherkin-testing/references/godog-examples"}},
		{"exclude skills prefix prunes all", []string{"skills"}, []string{".", "cmd/sv", "exp/slidesdeck", "exp/roubaix", "exp/roubaix/internal"}},
		{"exclude not found", []string{"nonexistent"}, []string{".", "cmd/sv", "exp/slidesdeck", "exp/roubaix", "exp/roubaix/internal", "skills/go-gherkin-testing/references/godog-examples"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterModules(mods, tt.excludes)
			if len(got) != len(tt.want) {
				t.Fatalf("filterModules() = %d modules %v, want %d %v", len(got), names(got), len(tt.want), tt.want)
			}
			for i, name := range tt.want {
				if got[i].Name != name {
					t.Errorf("filterModules() module[%d] = %q, want %q", i, got[i].Name, name)
				}
			}
		})
	}
}

func names(modules []discovery.Module) []string {
	out := make([]string, len(modules))
	for i, m := range modules {
		out[i] = m.Name
	}
	return out
}
