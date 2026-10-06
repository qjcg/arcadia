// Package sync generates Agent Plugin packages under agents/plugins from
// sync specs: skills are copied from a pinned git ref or a local tree,
// flattened into the spec-mandated skills/<name>/ layout, and their
// frontmatter is normalized to Agent Skills form.
package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/qjcg/arcadia/cmd/ap/internal/frontmatter"
	"github.com/qjcg/arcadia/cmd/ap/internal/spec"
)

// Run regenerates every package described by the specs and returns the
// package roots it wrote. repoRoot is used to resolve local sources;
// packages are written under outDir/<spec.Name>.
func Run(specs []spec.Spec, repoRoot, outDir string) ([]string, error) {
	var roots []string
	for _, s := range specs {
		root, err := buildOne(s, repoRoot, outDir)
		if err != nil {
			return roots, fmt.Errorf("sync %s: %w", s.Name, err)
		}
		roots = append(roots, root)
	}
	return roots, nil
}

func buildOne(s spec.Spec, repoRoot, outDir string) (string, error) {
	srcDir, cleanup, err := materializeSource(s, repoRoot)
	if err != nil {
		return "", err
	}
	if cleanup != nil {
		defer cleanup()
	}

	root := filepath.Join(outDir, s.Name)
	if err := os.RemoveAll(root); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(root, "skills"), 0o755); err != nil {
		return "", err
	}

	skillPaths, err := resolveSkills(srcDir, s.Source)
	if err != nil {
		return "", err
	}
	for _, rel := range skillPaths {
		name := filepath.Base(rel)
		if err := copyDir(filepath.Join(srcDir, rel), filepath.Join(root, "skills", name)); err != nil {
			return "", fmt.Errorf("copy skill %s: %w", rel, err)
		}
	}
	for _, file := range s.Source.CopyRoot {
		if err := copyFile(filepath.Join(srcDir, file), filepath.Join(root, filepath.Base(file))); err != nil {
			return "", fmt.Errorf("copy %s: %w", file, err)
		}
	}

	if err := normalizeSkills(root); err != nil {
		return "", err
	}
	if err := writeManifest(root, s); err != nil {
		return "", err
	}
	return root, nil
}

// materializeSource returns the source tree root and an optional cleanup.
func materializeSource(s spec.Spec, repoRoot string) (string, func(), error) {
	switch s.Source.Type {
	case "local":
		path := s.Source.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(repoRoot, path)
		}
		return path, nil, nil
	case "git":
		tmp, err := os.MkdirTemp("", "ap-sync-*")
		if err != nil {
			return "", nil, err
		}
		cleanup := func() { os.RemoveAll(tmp) }
		cmd := exec.Command("git", "clone", "--depth", "1", "--branch", s.Source.Ref, "--quiet", s.Source.Repo, tmp)
		if out, err := cmd.CombinedOutput(); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("git clone %s@%s: %w\n%s", s.Source.Repo, s.Source.Ref, err, out)
		}
		return tmp, cleanup, nil
	default:
		return "", nil, fmt.Errorf("unknown source type %q", s.Source.Type)
	}
}

// resolveSkills returns source-tree-relative skill paths, sorted.
func resolveSkills(srcDir string, source spec.Source) ([]string, error) {
	if len(source.Skills) > 0 {
		paths := append([]string(nil), source.Skills...)
		sort.Strings(paths)
		for _, rel := range paths {
			if _, err := os.Stat(filepath.Join(srcDir, rel, "SKILL.md")); err != nil {
				return nil, fmt.Errorf("skill %s: %w", rel, err)
			}
		}
		return paths, nil
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			if _, err := os.Stat(filepath.Join(srcDir, entry.Name(), "SKILL.md")); err == nil {
				paths = append(paths, entry.Name())
			}
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no skills found in %s", srcDir)
	}
	sort.Strings(paths)
	return paths, nil
}

// normalizeSkills rewrites every SKILL.md under root/skills to Agent Skills form.
func normalizeSkills(root string) error {
	skillsDir := filepath.Join(root, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		doc, err := frontmatter.Parse(string(raw))
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		normalized, err := frontmatter.NormalizeDocument(doc)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(normalized), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// manifest is the generated plugin.json (Agent Plugins v1.0.0, §5.2).
type manifest struct {
	Schema      string      `json:"$schema"`
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Description string      `json:"description"`
	Author      spec.Author `json:"author"`
	Homepage    string      `json:"homepage,omitempty"`
	Repository  string      `json:"repository,omitempty"`
	License     string      `json:"license,omitempty"`
	Keywords    []string    `json:"keywords,omitempty"`
}

func writeManifest(root string, s spec.Spec) error {
	m := manifest{
		Schema:      "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json",
		Name:        s.Name,
		Version:     s.Version,
		Description: s.Description,
		Author:      s.Author,
		Homepage:    s.Homepage,
		Repository:  s.Repository,
		License:     s.License,
		Keywords:    s.Keywords,
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(filepath.Join(root, "plugin.json"), out, 0o644)
}

// copyDir recursively copies src into dst, rejecting symlinks.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: symlinks are not supported", path)
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return copyFileMode(path, target, info.Mode().Perm())
	})
}

func copyFile(src, dst string) error {
	return copyFileMode(src, dst, 0o644)
}

func copyFileMode(src, dst string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// String describes a spec source for logging.
func String(s spec.Spec) string {
	switch s.Source.Type {
	case "git":
		return fmt.Sprintf("%s@%s", s.Source.Repo, s.Source.Ref)
	default:
		return strings.TrimPrefix(s.Source.Path, "./")
	}
}
