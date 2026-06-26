package site

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter holds parsed YAML metadata from a content file.
type Frontmatter struct {
	Title string `yaml:"title"`
	Order int    `yaml:"order"`
	Draft bool   `yaml:"draft"`
}

// ParseFrontmatter extracts YAML frontmatter and the remaining body from content.
// Returns nil Frontmatter if no frontmatter block is found.
func ParseFrontmatter(content []byte) (*Frontmatter, []byte) {
	s := string(content)
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return nil, content
	}

	// Find the closing ---
	rest := s[4:] // skip opening ---\n
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return nil, content
	}

	yamlBlock := rest[:idx]
	body := rest[idx+5:] // skip \n---\n

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, content
	}

	return &fm, []byte(body)
}

// DetectTitleFromContent extracts a title from content (first # heading or #+TITLE).
func DetectTitleFromContent(content []byte) string {
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimPrefix(trimmed, "# ")
		}
		if strings.HasPrefix(trimmed, "#+TITLE:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "#+TITLE:"))
		}
	}
	return "Untitled"
}
