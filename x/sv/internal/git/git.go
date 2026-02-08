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

// CommitsSince returns commit messages since a tag, scoped to a path
func CommitsSince(root, tag, path string) ([]string, error) {
	args := []string{"log", "--format=%s"}
	if tag != "" {
		args = append(args, tag+"..HEAD")
	}
	if path != "." {
		args = append(args, "--", path)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	commits := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(commits) == 1 && commits[0] == "" {
		return nil, nil
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
