package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest tracks which skills were extracted from which module
type Manifest struct {
	// ModuleSkills maps module path to list of skill names
	ModuleSkills map[string][]string `json:"module_skills"`
}

// Load reads the manifest from the modules directory
func Load(modulesDir string) (*Manifest, error) {
	manifestPath := filepath.Join(modulesDir, ".skillo-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{ModuleSkills: make(map[string][]string)}, nil
		}
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if m.ModuleSkills == nil {
		m.ModuleSkills = make(map[string][]string)
	}
	return &m, nil
}

// Save writes the manifest to the modules directory
func (m *Manifest) Save(modulesDir string) error {
	manifestPath := filepath.Join(modulesDir, ".skillo-manifest.json")
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

// AddSkills records that a module provides specific skills
func (m *Manifest) AddSkills(module string, skills []string) {
	m.ModuleSkills[module] = skills
}

// GetSkills returns the skills provided by a module
func (m *Manifest) GetSkills(module string) []string {
	return m.ModuleSkills[module]
}

// RemoveModule removes a module and returns its associated skills
func (m *Manifest) RemoveModule(module string) []string {
	skills := m.ModuleSkills[module]
	delete(m.ModuleSkills, module)
	return skills
}
