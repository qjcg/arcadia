package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
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

// HistoricalPaths returns all paths (current and historical) that a module has
// been located at, by following file rename history. Returns the current
// modulePath only if no renaming is detected or an error occurs.
func HistoricalPaths(root, modulePath string) ([]string, error) {
	if modulePath == "." {
		return []string{"."}, nil
	}

	goModPath := modulePath + "/go.mod"
	cmd := exec.Command("git", "log", "--follow", "--name-only", "--format=", "--", goModPath)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return []string{modulePath}, nil
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return []string{modulePath}, nil
	}

	pathSet := map[string]bool{modulePath: true}
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		dir := filepath.Dir(line)
		if dir != "." && dir != "" {
			pathSet[dir] = true
		}
	}

	result := make([]string, 0, len(pathSet))
	for p := range pathSet {
		result = append(result, p)
	}
	return result, nil
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

	// For non-root paths, use git's path filtering with rename history
	if path != "." {
		paths, err := HistoricalPaths(root, path)
		if err == nil {
			for _, p := range paths {
				args = append(args, "--", p)
			}
		} else {
			args = append(args, "--", path)
		}
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

// TagDate returns the date of a tag's commit in YYYY-MM-DD format.
func TagDate(root, tag string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%aI", tag)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	// Extract just the date portion from ISO timestamp
	date := strings.TrimSpace(string(out))
	if len(date) >= 10 {
		date = date[:10]
	}
	return date, nil
}

// CommitsSinceDate returns commit messages since a given date (inclusive).
// The date should be in a format git understands (e.g., "2024-01-15" or "2 weeks ago").
func CommitsSinceDate(root, since, path string) ([]string, error) {
	args := []string{"log", "--format=%s", "--since=" + since}
	if path != "." {
		paths, err := HistoricalPaths(root, path)
		if err == nil {
			for _, p := range paths {
				args = append(args, "--", p)
			}
		} else {
			args = append(args, "--", path)
		}
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
	return strings.Split(trimmed, "\n"), nil
}

// CommitsBetween returns commit messages between two tags (exclusive of fromTag, inclusive of toTag).
// If fromTag is empty, returns commits up to toTag.
func CommitsBetween(root, fromTag, toTag, path string) ([]string, error) {
	var rangeArg string
	if fromTag == "" {
		rangeArg = toTag
	} else {
		rangeArg = fromTag + ".." + toTag
	}

	args := []string{"log", "--format=%s", rangeArg}
	if path != "." {
		paths, hErr := HistoricalPaths(root, path)
		if hErr == nil {
			for _, p := range paths {
				args = append(args, "--", p)
			}
		} else {
			args = append(args, "--", path)
		}
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
	return strings.Split(trimmed, "\n"), nil
}

// CommitInfo holds a parsed git commit result.
type CommitInfo struct {
	Hash    string // full commit hash
	Short   string // short hash (7 characters)
	Message string // commit message subject line
}

// CommitsBetweenDetail returns commits between two tags (exclusive of fromTag, inclusive of toTag).
// If fromTag is empty, returns commits up to toTag. Returns structured commit info.
func CommitsBetweenDetail(root, fromTag, toTag, path string) ([]CommitInfo, error) {
	var rangeArg string
	if fromTag == "" {
		rangeArg = toTag
	} else {
		rangeArg = fromTag + ".." + toTag
	}

	// --format=%H for full hash, %h for short hash, %s for subject
	args := []string{"log", "--format=%H%n%h%n%s", rangeArg}
	if path != "." {
		paths, hErr := HistoricalPaths(root, path)
		if hErr == nil {
			for _, p := range paths {
				args = append(args, "--", p)
			}
		} else {
			args = append(args, "--", path)
		}
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

	lines := strings.Split(trimmed, "\n")
	var commits []CommitInfo
	for i := 0; i+2 < len(lines); i += 3 {
		commits = append(commits, CommitInfo{
			Hash:    strings.TrimSpace(lines[i]),
			Short:   strings.TrimSpace(lines[i+1]),
			Message: strings.TrimSpace(lines[i+2]),
		})
	}
	return commits, nil
}

// CommitsSinceDateDetail returns structured commit info since a given date.
func CommitsSinceDateDetail(root, since, path string) ([]CommitInfo, error) {
	args := []string{"log", "--format=%H%n%h%n%s", "--since=" + since}
	if path != "." {
		paths, hErr := HistoricalPaths(root, path)
		if hErr == nil {
			for _, p := range paths {
				args = append(args, "--", p)
			}
		} else {
			args = append(args, "--", path)
		}
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

	lines := strings.Split(trimmed, "\n")
	var commits []CommitInfo
	for i := 0; i+2 < len(lines); i += 3 {
		commits = append(commits, CommitInfo{
			Hash:    strings.TrimSpace(lines[i]),
			Short:   strings.TrimSpace(lines[i+1]),
			Message: strings.TrimSpace(lines[i+2]),
		})
	}
	return commits, nil
}

// TagAnnotated creates an annotated git tag with the given name and message.
func TagAnnotated(root, tagName, message string) error {
	cmd := exec.Command("git", "tag", "-a", tagName, "-m", message)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create annotated tag %s: %w\n%s", tagName, err, string(out))
	}
	return nil
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
