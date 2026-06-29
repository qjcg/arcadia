// Package normalize converts user-provided module path inputs into full Go
// module paths with explicit version specifications.
package normalize

import (
	"fmt"
	"strings"
)

// ModulePath converts user input into a full Go module path and version.
//
// Input formats:
//
//	"user/repo"           → ("github.com/user/repo", "latest")
//	"user/repo@v1.2.3"    → ("github.com/user/repo", "v1.2.3")
//	"https://github.com/user/repo" → ("github.com/user/repo", "latest")
//	"github.com/user/repo" → ("github.com/user/repo", "latest")
func ModulePath(input string) (module, version string, err error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", "", fmt.Errorf("empty module path")
	}

	// Strip https:// or http:// prefix
	if strings.HasPrefix(s, "https://") {
		s = s[len("https://"):]
	} else if strings.HasPrefix(s, "http://") {
		s = s[len("http://"):]
	}

	// Split version suffix
	module, version = s, "latest"
	if at := strings.LastIndex(s, "@"); at > 0 && at < len(s)-1 {
		module = s[:at]
		version = s[at+1:]
	} else if at == 0 {
		return "", "", fmt.Errorf("module path cannot start with '@'")
	}

	// Normalize short form: org/repo → github.com/org/repo
	// Full import paths always contain a dot.
	if !strings.Contains(module, ".") {
		// Check it's a short form like org/repo (has a slash) not just "foo"
		if strings.Contains(module, "/") {
			module = "github.com/" + module
		} else {
			return "", "", fmt.Errorf("module path %q is not a full import path or org/repo short form", input)
		}
	}

	return module, version, nil
}

// LooksLikeModulePath returns true if the input looks like a module path
// (contains '.' or '/') rather than a bare skill name.
func LooksLikeModulePath(s string) bool {
	return strings.Contains(s, ".") || strings.Contains(s, "/")
}
