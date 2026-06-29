package semver

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

type Bump uint8

const (
	BumpNone Bump = iota
	BumpPatch
	BumpMinor
	BumpMajor
)

// Commit represents a single commit for version calculation purposes.
type Commit struct {
	Message string
	Files   []string // changed files relative to repo root (optional)
}

// CalculateNext calculates the next version based on commits.
// defaultPatch controls whether non-feat/fix commits bump the patch version.
// When path is "." and a commit has populated Files, the "!" breaking change
// marker is only respected if at least one file is a .go file or go.mod,
// so that submodule-scoped breaking changes don't bump the root module.
func CalculateNext(current, path string, commits []Commit, defaultPatch bool) (string, error) {
	vStr := current
	if vStr == "" {
		vStr = "v0.1.0"
		if path != "." {
			return path + "/" + vStr, nil
		}
		return vStr, nil
	}

	// Strip module path if present
	versionPart := vStr
	if path != "." && strings.HasPrefix(vStr, path+"/") {
		versionPart = strings.TrimPrefix(vStr, path+"/")
	}

	v, err := semver.NewVersion(versionPart)
	if err != nil {
		return "", err
	}

	bump := BumpNone
	hasCommits := len(commits) > 0
	for _, commit := range commits {
		msg := strings.ToLower(commit.Message)
		if strings.Contains(msg, "breaking change:") {
			bump = BumpMajor
			break // Major is highest
		}
		if strings.Contains(msg, "!") && strings.Contains(msg, ":") {
			// Only count ! as breaking if the commit touches relevant source files
			if commitHasSourceCode(commit.Files, path) {
				bump = BumpMajor
				break // Major is highest
			}
		}
		if strings.HasPrefix(msg, "feat") {
			if bump < BumpMinor {
				bump = BumpMinor
			}
		} else if strings.HasPrefix(msg, "fix") {
			if bump < BumpPatch {
				bump = BumpPatch
			}
		}
	}

	// If there are commits but no semver-relevant type, optionally default to patch bump
	if hasCommits && bump == BumpNone && defaultPatch {
		bump = BumpPatch
	}

	var nextV semver.Version
	switch bump {
	case BumpNone:
		return vStr, nil
	case BumpPatch:
		nextV = v.IncPatch()
	case BumpMinor:
		nextV = v.IncMinor()
	case BumpMajor:
		nextV = v.IncMajor()
	}

	nextStr := "v" + nextV.String()
	if path != "." {
		return path + "/" + nextStr, nil
	}
	return nextStr, nil
}

// commitHasSourceCode checks if a commit's changed files include source code
// (.go or go.mod) relevant to the module at the given path.
// An empty files slice is treated as having source code for backward compatibility.
func commitHasSourceCode(files []string, path string) bool {
	if len(files) == 0 {
		return true // fall back to counting the ! when no file info
	}
	for _, f := range files {
		if strings.HasSuffix(f, ".go") || f == "go.mod" {
			return true
		}
	}
	return false
}

// Increment forces a specific bump on the current version
func Increment(current, path string, bump Bump) (string, error) {
	vStr := current
	if vStr == "" {
		vStr = "v0.1.0"
		if path != "." {
			return path + "/" + vStr, nil
		}
		return vStr, nil
	}

	// Strip module path if present
	versionPart := vStr
	if path != "." && strings.HasPrefix(vStr, path+"/") {
		versionPart = strings.TrimPrefix(vStr, path+"/")
	}

	v, err := semver.NewVersion(versionPart)
	if err != nil {
		return "", err
	}

	var nextV semver.Version
	switch bump {
	case BumpPatch:
		nextV = v.IncPatch()
	case BumpMinor:
		nextV = v.IncMinor()
	case BumpMajor:
		nextV = v.IncMajor()
	default:
		return vStr, nil
	}

	nextStr := "v" + nextV.String()
	if path != "." {
		return path + "/" + nextStr, nil
	}
	return nextStr, nil
}
