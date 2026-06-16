package git

import (
	"os/exec"
	"strings"
)

// ReadGoModAtTag reads the go.mod content from a specific tag in the repo.
func ReadGoModAtTag(root, tag, modulePath string) (string, error) {
	path := "go.mod"
	if modulePath != "." {
		path = modulePath + "/go.mod"
	}

	cmd := exec.Command("git", "show", tag+":"+path)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ReadGoMod reads the go.mod content from the working tree.
func ReadGoMod(root, modulePath string) (string, error) {
	path := "go.mod"
	if modulePath != "." {
		path = modulePath + "/go.mod"
	}

	cmd := exec.Command("git", "show", "HEAD:"+path)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Tags returns all tags matching the given prefix, sorted by version descending.
func Tags(root, prefix string) ([]string, error) {
	var pattern string
	if prefix == "." {
		pattern = "v*"
	} else {
		pattern = prefix + "/v*"
	}

	cmd := exec.Command("git", "tag", "--list", pattern, "--sort=-v:refname")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return nil, nil
	}

	tags := strings.Split(trimmed, "\n")
	return tags, nil
}

// LatestTag returns the latest tag for a given module prefix.
func LatestTag(root, prefix string) (string, error) {
	tags, err := Tags(root, prefix)
	if err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", nil
	}
	return tags[0], nil
}

// CommitsSince returns commit messages since a tag, scoped to a path.
// When excludePaths is provided and path is ".", commits that only touch
// files in excluded paths are filtered out.
func CommitsSince(root, tag, path string, excludePaths []string) ([]string, error) {
	// Get commit hash and message (one per line: hash\nmessage)
	args := []string{"log", "--format=%H%n%s"}
	if tag != "" {
		args = append(args, tag+"..HEAD")
	}

	// For non-root paths, use git's path filtering
	if path != "." {
		args = append(args, "--", path)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return nil, nil
	}

	// Split into commit entries (each is 2 lines: hash + message)
	// Use "\n" to split, then group pairs
	lines := strings.Split(trimmed, "\n")
	var entries [][]string
	for i := 0; i < len(lines)-1; i += 2 {
		if i+1 < len(lines) {
			entries = append(entries, []string{lines[i], lines[i+1]})
		}
	}

	// If path is "." and we have exclusions, filter by file
	if path == "." && len(excludePaths) > 0 {
		var filtered []string
		for _, entry := range entries {
			if len(entry) < 2 {
				continue
			}
			hash := entry[0]
			msg := entry[1]
			if msg == "" {
				continue
			}

			// Get files for this commit using git show with hash
			fileCmd := exec.Command("git", "show", "--format=%s", "--name-only", hash)
			fileCmd.Dir = root
			fileOut, _ := fileCmd.CombinedOutput()

			fileLines := strings.Split(strings.TrimSpace(string(fileOut)), "\n")
			// fileLines[0] = commit message, fileLines[1:] = files

			hasNonExcluded := false
			hasFiles := false
			for _, f := range fileLines[1:] {
				f = strings.TrimSpace(f)
				if f == "" {
					continue
				}
				hasFiles = true
				isExcluded := false
				for _, excl := range excludePaths {
					if strings.HasPrefix(f, excl+"/") || f == excl {
						isExcluded = true
						break
					}
				}
				if !isExcluded {
					hasNonExcluded = true
					break
				}
			}

			// Include commit if:
			// - It has no files (e.g., empty commits via --allow-empty)
			// - It has at least one non-excluded file
			if !hasFiles || hasNonExcluded {
				filtered = append(filtered, msg)
			}
		}
		return filtered, nil
	}

	// No exclusions: just return the messages
	var commits []string
	for _, entry := range entries {
		if len(entry) >= 2 {
			commits = append(commits, entry[1])
		}
	}
	return commits, nil
}

// Root returns the absolute path to the git root
func Root() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
