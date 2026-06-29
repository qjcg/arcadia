// Package selections manages the selections.json file that records which
// modules (and optionally which subset of their skills) are registered in
// a skillo scope (user or project).
package selections

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// Selections maps module path → list of skill names.
// An empty slice means "all skills from this module".
// A non-empty slice means "only these skills from this module".
type Selections map[string][]string

// Load reads selections.json from the given skillo directory.
// Returns an empty Selections if the file does not exist.
func Load(dir string) (Selections, error) {
	path := filepath.Join(dir, "selections.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(Selections), nil
		}
		return nil, fmt.Errorf("read selections: %w", err)
	}
	var s Selections
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse selections: %w", err)
	}
	if s == nil {
		s = make(Selections)
	}
	return s, nil
}

// Save writes selections.json to the given skillo directory.
func Save(dir string, s Selections) error {
	path := filepath.Join(dir, "selections.json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal selections: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write selections: %w", err)
	}
	return nil
}

// Init creates an empty selections.json in the given skillo directory.
// The directory must already exist.
func Init(dir string) error {
	path := filepath.Join(dir, "selections.json")
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}
	return Save(dir, make(Selections))
}

// AddModule registers a module with the given skill names in selections.
// If the module already exists, its skills are replaced with the new list.
func AddModule(dir, module string, skills []string) error {
	s, err := Load(dir)
	if err != nil {
		return err
	}
	if skills == nil {
		skills = []string{}
	}
	s[module] = skills
	return Save(dir, s)
}

// RemoveSkill removes a single skill name from its owning module's array.
// If the module's array becomes empty, the module entry is removed.
func RemoveSkill(dir, name string) error {
	s, err := Load(dir)
	if err != nil {
		return err
	}
	module := FindModule(s, name)
	if module == "" {
		return fmt.Errorf("skill %q not found in selections", name)
	}
	current := s[module]
	if len(current) > 0 {
		current = slices.DeleteFunc(current, func(n string) bool { return n == name })
	}
	if len(current) == 0 {
		delete(s, module)
	} else {
		s[module] = current
	}
	return Save(dir, s)
}

// RemoveModule removes an entire module entry from selections.
func RemoveModule(dir, module string) error {
	s, err := Load(dir)
	if err != nil {
		return err
	}
	delete(s, module)
	return Save(dir, s)
}

// FindModule scans the selections for which module owns the given skill name.
// Returns empty string if not found.
func FindModule(s Selections, name string) string {
	for module, skills := range s {
		if len(skills) == 0 {
			// Module with no filter implicitly owns the skill.
			// But we can't verify it here — this requires scanning the module dir.
			// Return the module and let the caller resolve ambiguity.
			return module
		}
		if slices.Contains(skills, name) {
			return module
		}
	}
	return ""
}

// ModuleSkills returns the skill list for a module, or nil if not found.
func ModuleSkills(s Selections, module string) []string {
	return s[module]
}

// ConvertLegacyManifest converts an old-style manifest.json (module→skills map)
// into selections format.
func ConvertLegacyManifest(moduleSkills map[string][]string) Selections {
	s := make(Selections, len(moduleSkills))
	for mod, skills := range moduleSkills {
		s[mod] = skills
	}
	return s
}
