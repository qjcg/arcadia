package discovery

import (
	"os"
	"path/filepath"
	"strings"
)

// Module represents a Go module found in the repository
type Module struct {
	Name string // Path from root to module directory (e.g., "." or "x/slidesdeck")
	Path string // Absolute path to module directory
}

// FindModules searches for Go modules within the given root directory
func FindModules(root string) ([]Module, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var modules []Module
	err = filepath.Walk(rootAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && (info.Name() == ".git" || info.Name() == "vendor") {
			return filepath.SkipDir
		}
		if !info.IsDir() && info.Name() == "go.mod" {
			dir := filepath.Dir(path)
			rel, err := filepath.Rel(rootAbs, dir)
			if err != nil {
				return err
			}
			// Normalize Windows paths if any
			rel = filepath.ToSlash(rel)
			modules = append(modules, Module{
				Name: rel,
				Path: dir,
			})
		}
		return nil
	})
	return modules, err
}

// GetCurrentModule returns the module that contains the given directory
func GetCurrentModule(root, currentDir string) (Module, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return Module{}, err
	}

	modules, err := FindModules(rootAbs)
	if err != nil {
		return Module{}, err
	}

	absCurrent, err := filepath.Abs(currentDir)
	if err != nil {
		return Module{}, err
	}

	var bestMatch Module
	for _, m := range modules {
		absMPath, err := filepath.Abs(m.Path)
		if err != nil {
			continue
		}
		if absCurrent == absMPath || strings.HasPrefix(absCurrent, absMPath+string(os.PathSeparator)) {
			if bestMatch.Path == "" || len(m.Path) > len(bestMatch.Path) {
				bestMatch = m
			}
		}
	}

	if bestMatch.Path == "" {
		return Module{Name: ".", Path: rootAbs}, nil
	}

	return bestMatch, nil
}
