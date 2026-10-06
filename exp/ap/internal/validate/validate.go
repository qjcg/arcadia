// Package validate checks plugin packages and skill trees against the
// Agent Plugins specification v1.0.0 and the Agent Skills specification.
package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/qjcg/arcadia/exp/ap/internal/frontmatter"
)

const (
	// PluginSchema is the canonical plugin manifest schema identifier
	// (Agent Plugins v1.0.0, §5.2).
	PluginSchema = "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json"
	// MCPSchema is the canonical MCP configuration schema identifier
	// (Agent Plugins v1.0.0, §7.2.1).
	MCPSchema = "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json"
)

var (
	nameRe      = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)
	pluginName  = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?$`)
	linkPattern = regexp.MustCompile(`\]\(([^)#\s]+)`)
)

// manifestFields are the only top-level fields permitted in plugin.json
// (Agent Plugins v1.0.0, §5.2).
var manifestFields = map[string]bool{
	"$schema": true, "name": true, "version": true, "description": true,
	"author": true, "homepage": true, "repository": true, "license": true,
	"keywords": true, "extensions": true,
}

// Validate validates the filesystem tree at path:
//   - a plugin package root (contains plugin.json),
//   - a bare skills tree (children contain SKILL.md), or
//   - a directory of plugin package roots (children contain plugin.json).
//
// It returns one error per violation.
func Validate(path string) ([]error, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", path)
	}
	if _, err := os.Stat(filepath.Join(path, "plugin.json")); err == nil {
		return validatePlugin(path), nil
	}
	children, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var packages, skillDirs []string
	for _, child := range children {
		if !child.IsDir() {
			continue
		}
		childPath := filepath.Join(path, child.Name())
		if _, err := os.Stat(filepath.Join(childPath, "plugin.json")); err == nil {
			packages = append(packages, childPath)
			continue
		}
		if _, err := os.Stat(filepath.Join(childPath, "SKILL.md")); err == nil {
			skillDirs = append(skillDirs, childPath)
		}
	}
	var errs []error
	switch {
	case len(skillDirs) > 0:
		for _, dir := range skillDirs {
			errs = append(errs, validateSkillDir(dir)...)
		}
	case len(packages) > 0:
		for _, pkg := range packages {
			errs = append(errs, validatePlugin(pkg)...)
		}
	default:
		return nil, fmt.Errorf("%s: no plugin.json and no skills found", path)
	}
	return errs, nil
}

// validatePlugin validates one plugin package root.
func validatePlugin(root string) []error {
	var errs []error

	raw, err := os.ReadFile(filepath.Join(root, "plugin.json"))
	if err != nil {
		return []error{fmt.Errorf("%s: %w", root, err)}
	}
	var manifest map[string]any
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return []error{fmt.Errorf("%s/plugin.json: invalid JSON: %w", root, err)}
	}
	errs = append(errs, validateManifest(root, manifest)...)

	if _, err := os.Stat(filepath.Join(root, "mcp.json")); err == nil {
		errs = append(errs, validateMCP(root)...)
	}
	if info, err := os.Stat(filepath.Join(root, "skills")); err == nil {
		if !info.IsDir() {
			errs = append(errs, fmt.Errorf("%s/skills: must be a directory", root))
		} else {
			children, err := os.ReadDir(filepath.Join(root, "skills"))
			if err != nil {
				return append(errs, err)
			}
			for _, child := range children {
				if child.IsDir() {
					errs = append(errs, validateSkillDir(filepath.Join(root, "skills", child.Name()))...)
				} else {
					errs = append(errs, fmt.Errorf("%s/skills/%s: not a directory", root, child.Name()))
				}
			}
		}
	}
	errs = append(errs, validateContainment(root)...)
	return errs
}

func validateManifest(root string, manifest map[string]any) []error {
	var errs []error
	bad := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("%s/plugin.json: %s", root, fmt.Sprintf(format, args...)))
	}

	for field := range manifest {
		if !manifestFields[field] {
			bad("unknown field %q", field)
		}
	}
	if schema, _ := manifest["$schema"].(string); schema != PluginSchema {
		bad("$schema must be %q", PluginSchema)
	}
	name, _ := manifest["name"].(string)
	if err := checkPluginName(name); err != nil {
		bad("name: %v", err)
	}
	if version, ok := manifest["version"]; ok {
		if _, isString := version.(string); !isString {
			bad("version must be a string")
		}
	}
	for _, field := range []string{"description", "homepage", "repository", "license"} {
		if value, ok := manifest[field]; ok {
			if _, isString := value.(string); !isString {
				bad("%s must be a string", field)
			}
		}
	}
	if author, ok := manifest["author"]; ok {
		fields, isObject := author.(map[string]any)
		if !isObject {
			bad("author must be an object")
		} else {
			for key, value := range fields {
				switch key {
				case "name", "email", "url":
					if _, isString := value.(string); !isString {
						bad("author.%s must be a string", key)
					}
				default:
					bad("author: unknown field %q", key)
				}
			}
		}
	}
	if keywords, ok := manifest["keywords"]; ok {
		list, isList := keywords.([]any)
		if !isList {
			bad("keywords must be an array of strings")
		} else {
			for i, value := range list {
				if _, isString := value.(string); !isString {
					bad("keywords[%d] must be a string", i)
				}
			}
		}
	}
	if extensions, ok := manifest["extensions"]; ok {
		namespaces, isObject := extensions.(map[string]any)
		if !isObject {
			bad("extensions must be an object")
		} else {
			for ns, value := range namespaces {
				if _, isObject := value.(map[string]any); !isObject {
					bad("extensions.%s must be an object", ns)
				}
			}
		}
	}
	return errs
}

func validateMCP(root string) []error {
	raw, err := os.ReadFile(filepath.Join(root, "mcp.json"))
	if err != nil {
		return []error{fmt.Errorf("%s: %w", root, err)}
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		return []error{fmt.Errorf("%s/mcp.json: invalid JSON: %w", root, err)}
	}
	var errs []error
	for field := range config {
		if field != "$schema" && field != "mcpServers" {
			errs = append(errs, fmt.Errorf("%s/mcp.json: unknown field %q", root, field))
		}
	}
	if schema, _ := config["$schema"].(string); schema != MCPSchema {
		errs = append(errs, fmt.Errorf("%s/mcp.json: $schema must be %q", root, MCPSchema))
	}
	if servers, ok := config["mcpServers"]; ok {
		if _, isObject := servers.(map[string]any); !isObject {
			errs = append(errs, fmt.Errorf("%s/mcp.json: mcpServers must be an object", root))
		}
	}
	return errs
}

// validateSkillDir validates one skill directory containing SKILL.md.
func validateSkillDir(dir string) []error {
	raw, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return []error{fmt.Errorf("%s: %w", dir, err)}
	}
	doc, err := frontmatter.Parse(string(raw))
	if err != nil {
		return []error{fmt.Errorf("%s/SKILL.md: %w", dir, err)}
	}
	var errs []error
	bad := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("%s/SKILL.md: %s", dir, fmt.Sprintf(format, args...)))
	}

	for field := range doc.Meta {
		if !frontmatter.Permitted(field) {
			bad("unknown frontmatter field %q (use metadata)", field)
		}
	}
	name, _ := doc.Meta["name"].(string)
	if err := checkSkillName(name); err != nil {
		bad("name: %v", err)
	}
	if name != filepath.Base(dir) {
		bad("name %q must match directory name %q", name, filepath.Base(dir))
	}
	description, _ := doc.Meta["description"].(string)
	if description == "" {
		bad("description is required")
	} else if len(description) > 1024 {
		bad("description is %d characters, maximum is 1024", len(description))
	}
	if metadata, ok := doc.Meta["metadata"]; ok {
		fields, isObject := metadata.(map[string]any)
		if !isObject {
			bad("metadata must be a mapping")
		} else {
			for key, value := range fields {
				if _, isString := value.(string); !isString {
					bad("metadata.%s must be a string", key)
				}
			}
		}
	}
	return errs
}

// validateContainment rejects symlinks and Markdown link targets that
// resolve outside the plugin root (Agent Plugins v1.0.0, §4.1).
func validateContainment(root string) []error {
	var errs []error
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return []error{err}
	}
	err = filepath.WalkDir(rootAbs, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			errs = append(errs, fmt.Errorf("%s: symlinks are not allowed in plugin packages", path))
			return nil
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		for _, match := range linkPattern.FindAllStringSubmatch(string(raw), -1) {
			target := match[1]
			if strings.Contains(target, "://") || strings.HasPrefix(target, "/") {
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(path), target))
			if resolved != rootAbs && !strings.HasPrefix(resolved, rootAbs+string(filepath.Separator)) {
				errs = append(errs, fmt.Errorf("%s: link %q escapes the plugin root", rel, match[1]))
			}
		}
		return nil
	})
	if err != nil {
		errs = append(errs, fmt.Errorf("%s: %w", root, err))
	}
	return errs
}

func checkSkillName(name string) error {
	if name == "" {
		return fmt.Errorf("required")
	}
	if len(name) > 64 {
		return fmt.Errorf("must be at most 64 characters")
	}
	if !nameRe.MatchString(name) {
		return fmt.Errorf("must be lowercase alphanumeric and hyphens, not starting or ending with a hyphen")
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("must not contain consecutive hyphens")
	}
	return nil
}

func checkPluginName(name string) error {
	if name == "" {
		return fmt.Errorf("required")
	}
	if len(name) > 64 {
		return fmt.Errorf("must be at most 64 characters")
	}
	if !pluginName.MatchString(name) {
		return fmt.Errorf("must be lowercase alphanumeric, hyphens and periods, not starting or ending with a non-alphanumeric")
	}
	if strings.Contains(name, "--") || strings.Contains(name, "..") {
		return fmt.Errorf("must not contain consecutive hyphens or periods")
	}
	return nil
}
