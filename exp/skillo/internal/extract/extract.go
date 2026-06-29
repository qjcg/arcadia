package extract

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/qjcg/arcadia/exp/skillo/internal/types"
	"github.com/qjcg/arcadia/exp/skillo/internal/validate"
)

// Result describes a single extracted skill.
type Result struct {
	Name string // skill name from SKILL.md frontmatter
}

// ExtractAll walks a module directory and extracts every directory containing
// a valid SKILL.md into the target skills directory. Returns the list of
// extracted skill names.
func ExtractAll(moduleDir, skillsDir string) ([]string, error) {
	return extractFiltered(moduleDir, skillsDir, nil)
}

// ExtractFiltered walks a module directory and extracts only the named skills.
// Returns the list of extracted skill names (subset of requested).
func ExtractFiltered(moduleDir, skillsDir string, requested []string) ([]string, error) {
	return extractFiltered(moduleDir, skillsDir, requested)
}

func extractFiltered(moduleDir, skillsDir string, requested []string) ([]string, error) {
	requestedSet := make(map[string]bool, len(requested))
	for _, r := range requested {
		requestedSet[r] = true
	}

	var extracted []string
	skillsFound := 0

	err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		skillMDPath := filepath.Join(path, "SKILL.md")
		mdInfo, err := os.Stat(skillMDPath)
		if err != nil || mdInfo.IsDir() {
			return nil
		}

		if err := validate.Validate(path); err != nil {
			return fmt.Errorf("validate %s: %w", path, err)
		}

		data, err := os.ReadFile(skillMDPath)
		if err != nil {
			return err
		}

		var fm struct {
			Name string `yaml:"name"`
		}
		if err := yaml.Unmarshal(data, &fm); err != nil {
			return fmt.Errorf("parse YAML in %s: %w", skillMDPath, err)
		}

		if fm.Name == "" {
			// Try extracting frontmatter manually
			content := string(data)
			if len(content) < 3 || content[:3] != "---" {
				return fmt.Errorf("no YAML frontmatter found in %s", skillMDPath)
			}
			rest := content[3:]
			end := 0
			for i := 0; i < len(rest); i++ {
				if rest[i] == '-' && i+2 < len(rest) && rest[i+1] == '-' && rest[i+2] == '-' {
					end = i
					break
				}
			}
			if end > 0 {
				fmStr := rest[:end]
				if err := yaml.Unmarshal([]byte(fmStr), &fm); err != nil {
					return fmt.Errorf("parse YAML in %s: %w", skillMDPath, err)
				}
			}
		}
		if fm.Name == "" {
			return fmt.Errorf("skill in %s has no name", path)
		}

		// Filter by requested list if provided
		if len(requestedSet) > 0 && !requestedSet[fm.Name] {
			skillsFound++
			return nil
		}

		dest := filepath.Join(skillsDir, fm.Name)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := copyDir(path, dest); err != nil {
			return fmt.Errorf("copy %s to %s: %w", path, dest, err)
		}
		fmt.Printf("Extracted skill: %s\n", fm.Name)
		extracted = append(extracted, fm.Name)
		skillsFound++
		return nil
	})

	if skillsFound == 0 {
		fmt.Printf("Warning: No SKILL.md found in %s\n", moduleDir)
	}

	if len(requested) > 0 && len(extracted) != len(requested) {
		// Some requested skills were not found
		found := make(map[string]bool, len(extracted))
		for _, e := range extracted {
			found[e] = true
		}
		for _, r := range requested {
			if !found[r] {
				return extracted, fmt.Errorf("skill %q not found in module", r)
			}
		}
	}

	return extracted, err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	info, err := s.Stat()
	if err != nil {
		return err
	}

	d, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}

// ListAvailableSkills returns skill names from SKILL.md frontmatter in a module dir.
func ListAvailableSkills(moduleDir string) ([]string, error) {
	var names []string
	err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) != "SKILL.md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// Parse just the name using the types.Skill struct
		content := string(data)
		if len(content) < 3 || content[:3] != "---" {
			return nil
		}
		rest := content[3:]
		end := stringsIndex(rest, "\n---")
		if end < 0 {
			return nil
		}
		fmStr := rest[:end]
		var skill types.Skill
		if err := yaml.Unmarshal([]byte(fmStr), &skill); err != nil {
			return nil
		}
		if skill.Name != "" {
			names = append(names, skill.Name)
		}
		return nil
	})
	return names, err
}

func stringsIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Ensure the unused import is used
var _ = types.Skill{}
