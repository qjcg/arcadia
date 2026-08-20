package cli

import (
	"github.com/qjcg/arcadia/cmd/sv/internal/semver"
)

// latestNonRetractedTag returns the latest tag for a module, skipping retracted versions.
// It is a thin wrapper over semver.LatestNonRetractedTag, kept for callers in this package.
func latestNonRetractedTag(root string, mName string) (tag string, warning string, err error) {
	return semver.LatestNonRetractedTag(root, mName)
}
