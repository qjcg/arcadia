package locale

import (
	"sort"
	"unicode"
)

// registry holds the registered profiles in insertion order.
var registry []*Profile

// Register adds a profile to the registry. It replaces any profile with the
// same name so built-in profiles are easy to override.
func Register(p *Profile) {
	for i, existing := range registry {
		if existing.Name == p.Name {
			registry[i] = p
			return
		}
	}
	registry = append(registry, p)
}

// Default returns the fallback profile used when no language can be detected.
// It is the first registered profile (English).
func Default() *Profile {
	if len(registry) == 0 {
		registerBuiltins()
	}
	return registry[0]
}

// ForDoc detects the language of a document and returns its profile, falling
// back to Default() when no non-default profile matches.
func ForDoc(src []byte) *Profile {
	if len(registry) == 0 {
		registerBuiltins()
	}

	runes := []rune(string(src))
	best := Default()
	bestScore := 0
	bestPriority := best.Priority
	for _, p := range registry {
		if p.IsDefault {
			continue
		}
		score := score(p, runes)
		if score == 0 {
			continue
		}
		if score > bestScore || (score == bestScore && p.Priority > bestPriority) {
			best = p
			bestScore = score
			bestPriority = p.Priority
		}
	}
	return best
}

// Score returns how many code points in runes belong to the profile's scripts.
func score(p *Profile, runes []rune) int {
	count := 0
	for _, r := range runes {
		for _, table := range p.Scripts {
			if unicode.Is(table, r) {
				count++
				break
			}
		}
	}
	return count
}

// registerBuiltins populates the registry on first use so the package has no
// initialization-order dependencies.
func registerBuiltins() {
	registerBuiltinEnglish()
	registerBuiltinChinese()
}

// SortedNames returns the registered profile names in a stable order.
func SortedNames() []string {
	var names []string
	for _, p := range registry {
		names = append(names, p.Name)
	}
	sort.Strings(names)
	return names
}
