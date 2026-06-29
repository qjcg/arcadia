package cli

import (
	"fmt"
	"strings"

	"github.com/qjcg/arcadia/cmd/sv/internal/git"
	"github.com/qjcg/arcadia/cmd/sv/internal/semver"
)

// latestNonRetractedTag returns the latest tag for a module, skipping retracted versions.
// It collects retract directives from the working tree go.mod and from all tagged versions.
// Returns the tag, a warning message if the original latest was retracted, and any error.
func latestNonRetractedTag(root string, mName string) (tag string, warning string, err error) {
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
		if !semver.IsVersionRetracted(versionPart, retractions) {
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
func collectRetractions(root, mName string, tags []string) []semver.Retraction {
	var allRetractions []semver.Retraction

	// Start with working tree go.mod
	goMod, err := git.ReadGoMod(root, mName)
	if err == nil {
		allRetractions = append(allRetractions, semver.ParseRetractions(goMod)...)
	}

	// Check go.mod at each tag
	for _, tag := range tags {
		goModAtTag, err := git.ReadGoModAtTag(root, tag, mName)
		if err != nil {
			continue // skip tags that don't have go.mod at that path
		}
		allRetractions = append(allRetractions, semver.ParseRetractions(goModAtTag)...)
	}

	return allRetractions
}
