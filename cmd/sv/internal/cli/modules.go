package cli

import (
	"os"
	"path/filepath"

	"github.com/qjcg/arcadia/cmd/sv/internal/discovery"
)

func getModules(root string, modulePaths, excludes []string, allModules bool) ([]discovery.Module, error) {
	var modules []discovery.Module
	var err error

	if allModules {
		modules, err = discovery.FindModules(root)
	} else if len(modulePaths) > 0 {
		for _, p := range modulePaths {
			absPath, pathErr := filepath.Abs(p)
			if pathErr != nil {
				return nil, pathErr
			}
			rel, pathErr := filepath.Rel(root, absPath)
			if pathErr != nil {
				return nil, pathErr
			}
			modules = append(modules, discovery.Module{Name: filepath.ToSlash(rel), Path: absPath})
		}
	} else {
		cwd, _ := os.Getwd()
		m, pathErr := discovery.GetCurrentModule(root, cwd)
		if pathErr != nil {
			return nil, pathErr
		}
		modules = []discovery.Module{m}
	}
	if err != nil {
		return nil, err
	}

	return filterModules(modules, excludes), nil
}

// filterModules removes any module whose name matches an entry in excludes.
// Names are matched exactly against the module's repo-relative path (e.g. "." or "exp/foo").
func filterModules(modules []discovery.Module, excludes []string) []discovery.Module {
	if len(excludes) == 0 {
		return modules
	}
	excludeSet := make(map[string]bool, len(excludes))
	for _, e := range excludes {
		excludeSet[e] = true
	}
	filtered := make([]discovery.Module, 0, len(modules))
	for _, m := range modules {
		if !excludeSet[m.Name] {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
