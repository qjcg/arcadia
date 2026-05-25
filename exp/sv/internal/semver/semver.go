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

// CalculateNext calculates the next version based on commits.
// defaultPatch controls whether non-feat/fix commits bump the patch version.
func CalculateNext(current, path string, commits []string, defaultPatch bool) (string, error) {
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
		commit = strings.ToLower(commit)
		if strings.Contains(commit, "breaking change:") || strings.Contains(commit, "!") && strings.Contains(commit, ":") {
			bump = BumpMajor
			break // Major is highest
		}
		if strings.HasPrefix(commit, "feat") {
			if bump < BumpMinor {
				bump = BumpMinor
			}
		} else if strings.HasPrefix(commit, "fix") {
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
