package cli

import (
	"os"
	"path/filepath"

	"github.com/qjcg/arcadia/exp/sv/internal/discovery"
)

func getModules(root, modulePath string, allModules bool) ([]discovery.Module, error) {
	if allModules {
		return discovery.FindModules(root)
	}

	var m discovery.Module
	var err error
	if modulePath != "" {
		absPath, err := filepath.Abs(modulePath)
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(root, absPath)
		if err != nil {
			return nil, err
		}
		m = discovery.Module{Name: filepath.ToSlash(rel), Path: absPath}
	} else {
		cwd, _ := os.Getwd()
		m, err = discovery.GetCurrentModule(root, cwd)
		if err != nil {
			return nil, err
		}
	}
	return []discovery.Module{m}, nil
}