package cli

import (
	"os"
	"path/filepath"

	"github.com/qjcg/arcadia/exp/sv/internal/discovery"
)

func getModules(root string, modulePaths []string, allModules bool) ([]discovery.Module, error) {
	if allModules {
		return discovery.FindModules(root)
	}

	if len(modulePaths) > 0 {
		var modules []discovery.Module
		for _, p := range modulePaths {
			absPath, err := filepath.Abs(p)
			if err != nil {
				return nil, err
			}
			rel, err := filepath.Rel(root, absPath)
			if err != nil {
				return nil, err
			}
			modules = append(modules, discovery.Module{Name: filepath.ToSlash(rel), Path: absPath})
		}
		return modules, nil
	}

	cwd, _ := os.Getwd()
	m, err := discovery.GetCurrentModule(root, cwd)
	if err != nil {
		return nil, err
	}
	return []discovery.Module{m}, nil
}