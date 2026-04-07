package git

import (
	"os/exec"
	"strings"
)

// LatestTag returns the latest tag for a given module prefix
func LatestTag(root, prefix string) (string, error) {
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
		return "", err
	}

	tags := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(tags) == 0 || tags[0] == "" {
		return "", nil // No tags found
	}

	return tags[0], nil
}

// CommitsSince returns commit messages since a tag, scoped to a path.
// When excludePaths is provided and path is ".", commits that only touch
// files in excluded paths are filtered out.
func CommitsSince(root, tag, path string, excludePaths []string) ([]string, error) {
	// Get commit messages (one per line)
	args := []string{"log", "--format=%s"}
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

	// Split by newlines to get individual commit messages
	commits := strings.Split(trimmed, "\n")

	// If path is "." and we have exclusions, filter by file
	if path == "." && len(excludePaths) > 0 {
		var filtered []string
		for _, msg := range commits {
			if msg == "" {
				continue
			}

			// Get files for this commit using git show
			fileCmd := exec.Command("git", "show", "--no-patch", "--format=%s", "--name-only", msg)
			fileCmd.Dir = root
			fileOut, _ := fileCmd.CombinedOutput()

			lines := strings.Split(strings.TrimSpace(string(fileOut)), "\n")
			// lines[0] = commit message, lines[1:] = files

			include := true
			for _, f := range lines[1:] {
				if f == "" {
					continue
				}
				for _, excl := range excludePaths {
					if strings.HasPrefix(f, excl+"/") || f == excl {
						include = false
						break
					}
				}
				if !include {
					break
				}
			}

			if include {
				filtered = append(filtered, msg)
			}
		}
		return filtered, nil
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
