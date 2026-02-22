package extract

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/qjcg/arcadia/exp/skillo/internal/manifest"
	"github.com/qjcg/arcadia/exp/skillo/internal/types"
	"github.com/qjcg/arcadia/exp/skillo/internal/validate"
)

func extractFrontmatter(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || lines[0] != "---" {
		return ""
	}
	var frontmatter []string
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			break
		}
		frontmatter = append(frontmatter, lines[i])
	}
	return strings.Join(frontmatter, "\n")
}

func ExtractSkills(moduleDir, skillsDir, modulesDir, module string) error {
	var extractedSkills []string
	skillsFound := 0
	err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		skillMD := filepath.Join(path, "SKILL.md")
		if _, err := os.Stat(skillMD); err == nil {
			if err := validate.Validate(path); err != nil {
				// Log or skip invalid skills? Skipping for now but reporting.
				return fmt.Errorf("validate %s: %w", path, err)
			}

			data, err := os.ReadFile(skillMD)
			if err != nil {
				return err
			}

			frontmatter := extractFrontmatter(string(data))
			if frontmatter == "" {
				return fmt.Errorf("no YAML frontmatter found in %s", skillMD)
			}

			var skill types.Skill
			if err := yaml.Unmarshal([]byte(frontmatter), &skill); err != nil {
				return fmt.Errorf("parse YAML in %s: %w", skillMD, err)
			}

			if skill.Name == "" {
				return fmt.Errorf("skill in %s has no name", path)
			}

			dest := filepath.Join(skillsDir, skill.Name)
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}

			if err := copyDir(path, dest); err != nil {
				return fmt.Errorf("copy %s to %s: %w", path, dest, err)
			}

			fmt.Printf("Extracted skill: %s\n", skill.Name)
			extractedSkills = append(extractedSkills, skill.Name)
			skillsFound++
		}
		return nil
	})

	if skillsFound == 0 {
		fmt.Printf("Warning: No SKILL.md found in %s\n", moduleDir)
	}

	// Update manifest with extracted skills
	if err == nil && skillsFound > 0 && modulesDir != "" && module != "" {
		m, err := manifest.Load(modulesDir)
		if err != nil {
			return fmt.Errorf("load manifest: %w", err)
		}
		m.AddSkills(module, extractedSkills)
		if err := m.Save(modulesDir); err != nil {
			return fmt.Errorf("save manifest: %w", err)
		}
	}

	return err
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
