package extract

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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

		// Extract the skill name from the SKILL.md frontmatter.
		name, err := extractSkillName(path)
		if err != nil {
			fmt.Printf("Warning: skipping %s: %v\n", path, err)
			return nil
		}

		if name == "" {
			return nil
		}

		// Only count toward skillsFound after we have the name
		skillsFound++

		// Filter by requested list if provided
		if len(requestedSet) > 0 && !requestedSet[name] {
			return nil
		}

		dest := filepath.Join(skillsDir, name)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := copyDir(path, dest); err != nil {
			return fmt.Errorf("copy %s to %s: %w", path, dest, err)
		}
		fmt.Printf("Extracted skill: %s\n", name)
		extracted = append(extracted, name)
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

// extractSkillName reads a SKILL.md and returns the skill name from frontmatter.
// Returns ("", nil) if the file has no frontmatter (not a skill).
// Returns ("", error) if the frontmatter exists but is invalid.
func extractSkillName(skillDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return "", err
	}
	content := string(data)
	if len(content) < 3 || content[:3] != "---" {
		return "", nil
	}
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", nil
	}
	fmStr := rest[:end]
	var fm struct {
		Name string `yaml:"name"`
	}
	if err := yaml.Unmarshal([]byte(fmStr), &fm); err != nil {
		return "", fmt.Errorf("invalid frontmatter: %w", err)
	}
	if fm.Name == "" {
		return "", fmt.Errorf("missing name in frontmatter")
	}
	return fm.Name, nil
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
		name, err := extractSkillName(filepath.Dir(path))
		if err == nil && name != "" {
			names = append(names, name)
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

var _ = stringsIndex // keep string utility available
