package expand

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Glob expands glob patterns in a word using filepath.Match.
// Supports *, ?, [...], and ** (recursive).
// Returns the word unchanged if no match found or no glob chars present.
func Glob(word string) []string {
	if !hasGlobChars(word) {
		return []string{word}
	}

	// Special case: ** for recursive matching
	if strings.Contains(word, "**") {
		return globStar(word)
	}

	matches, err := filepath.Glob(word)
	if err != nil || len(matches) == 0 {
		return []string{word}
	}

	sort.Strings(matches)
	return matches
}

// GlobExpand expands a word through the glob pipeline.
// If the word contains glob characters and matches files, returns matches.
// Otherwise returns the original word.
func GlobExpand(word string) []string {
	return Glob(word)
}

// hasGlobChars checks if a string contains glob metacharacters.
func hasGlobChars(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '*', '?', '[':
			return true
		}
	}
	return false
}

// globStar handles ** (recursive glob) patterns.
// ** matches zero or more directory levels.
func globStar(pattern string) []string {
	// Split on **, keeping the parts
	parts := strings.Split(pattern, "**")
	if len(parts) < 2 {
		// No ** found, fall back to regular glob
		matches, err := filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			return []string{pattern}
		}
		sort.Strings(matches)
		return matches
	}

	prefix := parts[0]
	suffix := strings.Join(parts[1:], "**")

	// Clean the prefix and suffix: strip leading/trailing slashes for matching
	prefix = strings.TrimSuffix(prefix, "/")
	suffix = strings.TrimPrefix(suffix, "/")

	// Determine the search root directory
	searchDir := "."
	pathPrefix := ""
	if prefix != "" {
		searchDir = filepath.Dir(prefix)
		pathPrefix = filepath.Base(prefix)
	}

	var results []string

	filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}

		// Compute the relative path from the search directory
		rel := path
		if searchDir != "." {
			rel = strings.TrimPrefix(path, searchDir+"/")
		}

		// Skip the root itself
		if rel == "." || rel == "" || rel == searchDir {
			return nil
		}

		// If there's a prefix to match, check it
		if pathPrefix != "" && !strings.HasPrefix(rel, pathPrefix+"/") && rel != pathPrefix {
			return nil
		}

		// Strip the prefix portion for suffix matching
		matchFrom := rel
		if pathPrefix != "" {
			matchFrom = strings.TrimPrefix(rel, pathPrefix+"/")
			matchFrom = strings.TrimPrefix(matchFrom, pathPrefix)
		}

		// ** at the end: match everything
		if suffix == "" {
			results = append(results, path)
			return nil
		}

		// Check if the base name matches the suffix pattern
		base := filepath.Base(matchFrom)
		if matched, _ := filepath.Match(suffix, base); matched {
			results = append(results, path)
			return nil
		}

		// Also check if the full relative path matches (for patterns like dir/*.go)
		if matched, _ := filepath.Match(suffix, matchFrom); matched {
			results = append(results, path)
		}

		return nil
	})

	if len(results) == 0 {
		return []string{pattern}
	}

	sort.Strings(results)
	return results
}
