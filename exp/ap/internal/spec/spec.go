// Package spec loads Agent Plugin sync specs from YAML files.
//
// A sync spec describes one generated plugin package under agents/plugins:
// where its skills come from (a pinned git ref or a local tree) and the
// manifest metadata to write into plugin.json.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Spec is one plugin sync spec.
type Spec struct {
	// Name is the plugin name and the name of the generated package
	// directory (agents/plugins/<name>).
	Name string `yaml:"name"`
	// Version is the plugin manifest version. For vendored packages it
	// tracks the upstream version.
	Version string `yaml:"version"`
	// Description is the plugin manifest description.
	Description string `yaml:"description"`
	// Author is the plugin manifest author.
	Author Author `yaml:"author"`
	// Homepage is the plugin manifest homepage URL (optional).
	Homepage string `yaml:"homepage"`
	// Repository is the plugin manifest repository URL (optional).
	Repository string `yaml:"repository"`
	// License is the plugin manifest license (SPDX identifier recommended).
	License string `yaml:"license"`
	// Keywords are the plugin manifest keywords (optional).
	Keywords []string `yaml:"keywords"`
	// Source describes where the skills are copied from.
	Source Source `yaml:"source"`
}

// Author is the plugin manifest author object.
type Author struct {
	Name  string `yaml:"name" json:"name,omitempty"`
	Email string `yaml:"email" json:"email,omitempty"`
	URL   string `yaml:"url" json:"url,omitempty"`
}

// Source describes the skill source tree for a generated package.
type Source struct {
	// Type is "git" or "local".
	Type string `yaml:"type"`
	// Repo is the git repository URL (Type "git").
	Repo string `yaml:"repo"`
	// Ref is the pinned git ref (Type "git").
	Ref string `yaml:"ref"`
	// Path is a repo-relative path to the source tree (Type "local").
	Path string `yaml:"path"`
	// Skills lists the skills to copy, identified by their path relative to
	// the source tree (e.g. "engineering/ask-matt") or, for local sources,
	// their directory name under Path. Empty means every directory
	// containing a SKILL.md, walked only one level deep.
	Skills []string `yaml:"skills"`
	// CopyRoot lists source-tree-relative files to copy to the package
	// root (e.g. LICENSE, CHANGELOG.md).
	CopyRoot []string `yaml:"copy-root"`
}

// Load reads and validates every *.yaml spec in dir, sorted by name.
func Load(dir string) ([]Spec, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read specs dir: %w", err)
	}
	var specs []Spec
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		var s Spec
		if err := yaml.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		if err := s.Validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		specs = append(specs, s)
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Name < specs[j].Name })
	return specs, nil
}

// Validate checks that the spec is complete and self-consistent.
func (s Spec) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}
	if s.Version == "" {
		return fmt.Errorf("version is required")
	}
	if s.Description == "" {
		return fmt.Errorf("description is required")
	}
	if s.License == "" {
		return fmt.Errorf("license is required")
	}
	if s.Author.Name == "" {
		return fmt.Errorf("author.name is required")
	}
	switch s.Source.Type {
	case "git":
		if s.Source.Repo == "" || s.Source.Ref == "" {
			return fmt.Errorf("source: git requires repo and ref")
		}
	case "local":
		if s.Source.Path == "" {
			return fmt.Errorf("source: local requires path")
		}
	default:
		return fmt.Errorf("source: unknown type %q", s.Source.Type)
	}
	return nil
}
