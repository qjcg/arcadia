package cli

import (
	"os"
	"path/filepath"
	"strings"

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

// filterModules removes modules that match an entry in excludes, either by
// exact repo-relative path (e.g. "." or "exp/foo") or by directory subtree
// prefix (e.g. "exp" also prunes "exp/roubaix", "exp/roubaix/internal", ...).
func filterModules(modules []discovery.Module, excludes []string) []discovery.Module {
	if len(excludes) == 0 {
		return modules
	}
	filtered := make([]discovery.Module, 0, len(modules))
	for _, m := range modules {
		if !isExcluded(m.Name, excludes) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// isExcluded reports whether moduleName falls under any excluded path.
// An exclude matches the module if the module's repo-relative path equals
// the exclude, or begins with the exclude followed by a path separator.
func isExcluded(moduleName string, excludes []string) bool {
	for _, e := range excludes {
		if moduleName == e || strings.HasPrefix(moduleName, e+"/") {
			return true
		}
	}
	return false
}
