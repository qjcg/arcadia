package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/qjcg/arcadia/exp/skillo/internal/types"
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

func Validate(dir string) error {
	skillMD := filepath.Join(dir, "SKILL.md")
	data, err := os.ReadFile(skillMD)
	if err != nil {
		return fmt.Errorf("read SKILL.md: %w", err)
	}
	frontmatter := extractFrontmatter(string(data))
	if frontmatter == "" {
		return fmt.Errorf("no YAML frontmatter found")
	}
	var skill types.Skill
	if err := yaml.Unmarshal([]byte(frontmatter), &skill); err != nil {
		return fmt.Errorf("parse YAML: %w", err)
	}
	if skill.Name == "" || skill.Description == "" {
		return fmt.Errorf("missing name or description")
	}
	return nil
}
