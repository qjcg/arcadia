package scaffold

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// TemplateInfo describes a built-in template.
type TemplateInfo struct {
	Name        string
	Description string
}

// builtinEntry defines a built-in template with its embedded FS and extra files not in embed.
type builtinEntry struct {
	Name       string
	Desc       string
	FS         embed.FS
	Root       string
	ExtraFiles map[string]string // filename → content for hidden files embed.FS skips
	ExtraDirs  []string          // empty directories embed.FS skips
}

var builtinRegistry []builtinEntry

// RegisterBuiltin registers a built-in template so ListBuiltin and Resolve can find it.
func RegisterBuiltin(name, desc string, fs embed.FS, root string, extraFiles map[string]string, extraDirs []string) {
	builtinRegistry = append(builtinRegistry, builtinEntry{
		Name: name, Desc: desc, FS: fs, Root: root,
		ExtraFiles: extraFiles,
		ExtraDirs:  extraDirs,
	})
}

// ListBuiltin returns info about all registered built-in templates.
func ListBuiltin() []TemplateInfo {
	var list []TemplateInfo
	for _, e := range builtinRegistry {
		list = append(list, TemplateInfo{Name: e.Name, Description: e.Desc})
	}
	return list
}

// Resolve finds a template source: built-in, path, or XDG data.
func Resolve(name string) (string, error) {
	for _, e := range builtinRegistry {
		if e.Name == name {
			return extractEmbedded(e)
		}
	}

	if fi, err := os.Stat(name); err == nil && fi.IsDir() {
		if _, err := os.Stat(filepath.Join(name, "config.cue")); err == nil {
			return name, nil
		}
		return "", fmt.Errorf("directory %q exists but has no config.cue", name)
	}

	xdgDir := os.Getenv("XDG_DATA_HOME")
	if xdgDir == "" {
		home, _ := os.UserHomeDir()
		xdgDir = filepath.Join(home, ".local", "share")
	}
	tmplDir := filepath.Join(xdgDir, "pavona", "templates", name)
	if fi, err := os.Stat(tmplDir); err == nil && fi.IsDir() {
		if _, err := os.Stat(filepath.Join(tmplDir, "config.cue")); err == nil {
			return tmplDir, nil
		}
	}

	return "", fmt.Errorf("template %q not found", name)
}

// extractEmbedded copies embedded template files and extra content to a temp directory.
func extractEmbedded(entry builtinEntry) (string, error) {
	tmpDir, err := os.MkdirTemp("", "pavona-"+entry.Name+"-")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	// Extract files from embed.FS
	err = fs.WalkDir(entry.FS, entry.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(entry.Root, path)
		dest := filepath.Join(tmpDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		data, err := entry.FS.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("extracting embedded template: %w", err)
	}

	// Write extra hidden files as .tmpl files so they get rendered during hydration
	for name, content := range entry.ExtraFiles {
		dest := filepath.Join(tmpDir, name+".tmpl")
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("creating dir for extra file %s: %w", name, err)
		}
		if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("writing extra file %s: %w", name, err)
		}
	}

	// Create extra empty directories
	for _, dir := range entry.ExtraDirs {
		dest := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(dest, 0o755); err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("creating extra dir %s: %w", dir, err)
		}
	}

	return tmpDir, nil
}

// sanitize converts a string to a valid Go package name: lowercase, no hyphens.
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '-' || r == '_' {
			b.WriteRune('_')
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// Hydrate renders a template directory into an output directory.
func Hydrate(templateDir, outputDir string, vars map[string]string) error {
	if fi, err := os.Stat(outputDir); err == nil {
		if fi.IsDir() {
			entries, err := os.ReadDir(outputDir)
			if err == nil && len(entries) > 0 {
				return fmt.Errorf("output directory %q exists and is not empty", outputDir)
			}
		}
	}

	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(templateDir, path)
		if err != nil {
			return err
		}

		// Skip config.cue
		if rel == "config.cue" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Render the relative path through template
		renderedRel, err := renderPath(rel, vars)
		if err != nil {
			return fmt.Errorf("rendering path %q: %w", rel, err)
		}

		renderedRel = strings.TrimSuffix(renderedRel, ".tmpl")
		destPath := filepath.Join(outputDir, renderedRel)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".tmpl") {
			return renderFile(path, destPath, vars)
		}
		return copyFile(path, destPath)
	})
}

// renderPath applies template rendering to a file/directory path.
func renderPath(path string, vars map[string]string) (string, error) {
	tmpl, err := template.New("path").Option("missingkey=error").Parse(path)
	if err != nil {
		return "", fmt.Errorf("parsing path template: %w", err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("executing path template: %w", err)
	}
	return buf.String(), nil
}

// renderFile reads a template file, renders it, and writes to dest.
func renderFile(src, dest string, vars map[string]string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"sanitize": sanitize,
	}

	tmpl, err := template.New(filepath.Base(src)).
		Funcs(funcMap).
		Option("missingkey=error").
		Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", src, err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, vars); err != nil {
		return fmt.Errorf("executing template %s: %w", src, err)
	}
	return nil
}

// copyFile copies a file byte-for-byte.
func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
