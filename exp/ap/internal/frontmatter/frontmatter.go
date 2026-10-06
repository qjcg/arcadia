// Package frontmatter parses, normalizes, and renders SKILL.md YAML frontmatter
// to conform to the Agent Skills specification:
// https://agentskills.io/specification
package frontmatter

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Document is a parsed SKILL.md: YAML frontmatter plus an untouched Markdown body.
type Document struct {
	Meta map[string]any
	Body string
}

// Fields permitted by the Agent Skills specification, in render order.
var permittedFields = []string{"name", "description", "license", "compatibility", "metadata", "allowed-tools"}

// Permitted reports whether field is permitted in frontmatter by the
// Agent Skills specification.
func Permitted(field string) bool {
	return slices.Contains(permittedFields, field)
}

// Parse splits SKILL.md content into frontmatter and body. The content MUST
// start with a "---" line and contain a closing "---" line.
func Parse(content string) (Document, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return Document{}, fmt.Errorf("missing opening frontmatter delimiter")
	}
	closing := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closing = i
			break
		}
	}
	if closing < 0 {
		return Document{}, fmt.Errorf("missing closing frontmatter delimiter")
	}
	raw := strings.Join(lines[1:closing], "\n")
	var meta map[string]any
	if err := yaml.Unmarshal([]byte(raw), &meta); err != nil {
		return Document{}, fmt.Errorf("invalid frontmatter YAML: %w", err)
	}
	if meta == nil {
		meta = map[string]any{}
	}
	body := strings.Join(lines[closing+1:], "\n")
	return Document{Meta: meta, Body: body}, nil
}

// Normalize rewrites frontmatter to Agent Skills form:
//   - permitted fields are kept;
//   - every other key is moved into metadata;
//   - all metadata values are flattened to strings, nested maps becoming
//     dot-separated keys (e.g. metadata.credits.skill -> "credits.skill").
//
// The result contains only permitted fields.
func Normalize(meta map[string]any) (map[string]any, error) {
	out := map[string]any{}
	metadata := map[string]string{}

	for key, value := range meta {
		if key == "metadata" {
			nested, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("metadata must be a mapping")
			}
			if err := flatten(metadata, "", nested); err != nil {
				return nil, err
			}
			continue
		}
		if Permitted(key) {
			out[key] = value
			continue
		}
		if err := flatten(metadata, key, map[string]any{"": value}); err != nil {
			return nil, fmt.Errorf("field %q: %w", key, err)
		}
	}
	if len(metadata) > 0 {
		out["metadata"] = metadata
	}
	return out, nil
}

// flatten folds a nested mapping into string-valued entries with
// dot-separated keys.
func flatten(dst map[string]string, prefix string, in map[string]any) error {
	for key, value := range in {
		full := key
		if prefix != "" && key != "" {
			full = prefix + "." + key
		} else if key == "" {
			full = prefix
		}
		switch v := value.(type) {
		case map[string]any:
			if err := flatten(dst, full, v); err != nil {
				return err
			}
		case string:
			dst[full] = v
		case bool:
			dst[full] = fmt.Sprintf("%t", v)
		case int:
			dst[full] = fmt.Sprintf("%d", v)
		case int64:
			dst[full] = fmt.Sprintf("%d", v)
		case uint64:
			dst[full] = fmt.Sprintf("%d", v)
		case float64:
			dst[full] = fmt.Sprintf("%g", v)
		case nil:
			dst[full] = ""
		default:
			return fmt.Errorf("unsupported metadata value type %T for %q", value, full)
		}
	}
	return nil
}

// Render serializes frontmatter deterministically: permitted fields in
// specification order, then any remaining keys sorted, followed by the body.
func Render(meta map[string]any, body string) (string, error) {
	var sb strings.Builder
	sb.WriteString("---\n")

	seen := map[string]bool{}
	writeField := func(key string) error {
		value, ok := meta[key]
		if !ok {
			return nil
		}
		seen[key] = true
		out, err := yaml.Marshal(map[string]any{key: value})
		if err != nil {
			return fmt.Errorf("render %q: %w", key, err)
		}
		sb.Write(out)
		return nil
	}
	for _, key := range permittedFields {
		if err := writeField(key); err != nil {
			return "", err
		}
	}
	var rest []string
	for key := range meta {
		if !seen[key] {
			rest = append(rest, key)
		}
	}
	sort.Strings(rest)
	for _, key := range rest {
		if err := writeField(key); err != nil {
			return "", err
		}
	}

	sb.WriteString("---\n")
	sb.WriteString(body)
	return sb.String(), nil
}

// NormalizeDocument normalizes frontmatter and re-renders the full document.
func NormalizeDocument(doc Document) (string, error) {
	meta, err := Normalize(doc.Meta)
	if err != nil {
		return "", err
	}
	return Render(meta, doc.Body)
}
