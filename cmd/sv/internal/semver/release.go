package semver

import (
	"fmt"
	"strings"

	"github.com/qjcg/arcadia/cmd/sv/internal/git"
)

// Pending describes the next release for a module: the version that
// would be tagged and the commits that would be included in that release.
type Pending struct {
	Module  string           // module path, e.g. "." or "x/mod"
	Version string           // full version string, e.g. "v1.2.3" or "x/mod/v1.2.3"
	Commits []git.CommitInfo // commits since the last non-retracted tag, scoped to the module
	Warning string           // warning about retracted versions, if any
}

// PendingRelease computes the next version for a module and the commits that
// would be included in that release. It is the single source of truth shared by
// `sv next` and `sv changelog --release`, so both agree on the pending version.
// allModuleNames is the list of all module paths in the repo (used to exclude
// submodule commits from the root module's release). Returns nil when there is
// no version bump (no commits since the last tag).
func PendingRelease(root, mName string, allModuleNames []string, defaultPatch bool) (*Pending, error) {
	tag, warning, err := LatestNonRetractedTag(root, mName)
	if err != nil {
		return nil, err
	}

	// Build exclude paths: all modules except the current one
	var excludePaths []string
	if mName == "." {
		for _, n := range allModuleNames {
			if n != "." {
				excludePaths = append(excludePaths, n)
			}
		}
	}

	commits, err := git.CommitsSince(root, tag, mName, excludePaths)
	if err != nil {
		return nil, err
	}

	// CommitsSince does not populate Short; derive it from the full hash so
	// changelog entries render the abbreviated hash like other entries.
	for i := range commits {
		if commits[i].Short == "" && len(commits[i].Hash) >= 7 {
			commits[i].Short = commits[i].Hash[:7]
		}
	}

	// Convert git.CommitInfo to semver.Commit
	semverCommits := make([]Commit, len(commits))
	for i, c := range commits {
		semverCommits[i] = Commit{Message: c.Message, Files: c.Files}
	}
	next, err := CalculateNext(tag, mName, semverCommits, defaultPatch)
	if err != nil {
		return nil, err
	}

	if next == tag && tag != "" {
		return nil, nil // No change
	}

	return &Pending{Module: mName, Version: next, Commits: commits, Warning: warning}, nil
}

// LatestNonRetractedTag returns the latest tag for a module, skipping retracted versions.
// It collects retract directives from the working tree go.mod and from all tagged versions.
// Returns the tag, a warning message if the original latest was retracted, and any error.
func LatestNonRetractedTag(root string, mName string) (tag string, warning string, err error) {
	allTags, err := git.Tags(root, mName)
	if err != nil {
		return "", "", err
	}
	if len(allTags) == 0 {
		return "", "", nil
	}

	// Collect retractions from working tree go.mod and from all tagged versions
	retractions := collectRetractions(root, mName, allTags)

	// If no retractions, return the latest tag
	if len(retractions) == 0 {
		return allTags[0], "", nil
	}

	// Find the first non-retracted tag
	for _, t := range allTags {
		// Extract version part: strip module path prefix
		versionPart := t
		if mName != "." && strings.HasPrefix(t, mName+"/") {
			versionPart = strings.TrimPrefix(t, mName+"/")
		}
		if !IsVersionRetracted(versionPart, retractions) {
			if t != allTags[0] {
				return t, fmt.Sprintf("version %s is retracted, using %s", allTags[0], t), nil
			}
			return t, "", nil
		}
	}

	// All tags are retracted
	return "", fmt.Sprintf("all tagged versions for module %s are retracted", mName), nil
}

// collectRetractions gathers all retract directives from the working tree go.mod
// and from go.mod files at each tagged version.
func collectRetractions(root, mName string, tags []string) []Retraction {
	var allRetractions []Retraction

	// Start with working tree go.mod
	goMod, err := git.ReadGoMod(root, mName)
	if err == nil {
		allRetractions = append(allRetractions, ParseRetractions(goMod)...)
	}

	// Check go.mod at each tag
	for _, tag := range tags {
		goModAtTag, err := git.ReadGoModAtTag(root, tag, mName)
		if err != nil {
			continue // skip tags that don't have go.mod at that path
		}
		allRetractions = append(allRetractions, ParseRetractions(goModAtTag)...)
	}

	return allRetractions
}
