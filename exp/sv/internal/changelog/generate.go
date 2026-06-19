package changelog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/qjcg/arcadia/exp/sv/internal/git"
)

// GenerateFromGit generates a Changelog from git history for the given module.
// If fromVersion is empty, starts from the first tag. If toVersion is empty,
// includes all tags up to the latest.
func GenerateFromGit(root, modulePath, fromVersion, toVersion string) (*Changelog, error) {
	allTags, err := git.Tags(root, modulePath)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}
	if len(allTags) == 0 {
		return &Changelog{}, nil
	}

	// Tags are sorted descending (newest first). Reverse for ascending.
	ascTags := make([]string, len(allTags))
	for i, t := range allTags {
		ascTags[len(ascTags)-1-i] = t
	}

	// Filter by from/to range
	startIdx, endIdx := 0, len(ascTags)-1
	if fromVersion != "" {
		found := false
		for i, t := range ascTags {
			if tagMatchesVersion(t, modulePath, fromVersion) {
				startIdx = i
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("version %s not found in tags", fromVersion)
		}
	}
	if toVersion != "" {
		found := false
		for i, t := range ascTags {
			if tagMatchesVersion(t, modulePath, toVersion) {
				endIdx = i
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("version %s not found in tags", toVersion)
		}
	}
	tags := ascTags[startIdx : endIdx+1]

	var entries []Entry
	for i, tag := range tags {
		var commits []git.CommitInfo
		var err error

		// Determine previous tag for commit range using the full ascending tag list
		prevIdx := startIdx + i - 1
		if prevIdx >= 0 {
			commits, err = git.CommitsBetweenDetail(root, ascTags[prevIdx], tag, modulePath)
		} else {
			commits, err = git.CommitsBetweenDetail(root, "", tag, modulePath)
		}
		if err != nil {
			return nil, fmt.Errorf("getting commits for %s: %w", tag, err)
		}

		if len(commits) == 0 {
			continue
		}

		date, err := git.TagDate(root, tag)
		if err != nil {
			date = "" // non-fatal; omit date
		}

		entry := buildEntry(tag, date, commits)
		entries = append(entries, entry)
	}

	// Add unreleased (commits from latest tag to HEAD)
	if toVersion == "" {
		latestTag := allTags[0] // newest tag
		unreleasedCommits, err := git.CommitsBetweenDetail(root, latestTag, "", modulePath)
		if err == nil && len(unreleasedCommits) > 0 {
			entry := buildEntry("unreleased", "", unreleasedCommits)
			entries = append(entries, entry)
		}
	}

	// Sort entries: newest first for output
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Version == "unreleased" {
			return true
		}
		if entries[j].Version == "unreleased" {
			return false
		}
		return entries[i].Version > entries[j].Version
	})

	return &Changelog{Entries: entries}, nil
}

// GenerateSinceDate generates a Changelog from git history since a given date.
func GenerateSinceDate(root, modulePath, since string) (*Changelog, error) {
	commits, err := git.CommitsSinceDateDetail(root, since, modulePath)
	if err != nil {
		return nil, fmt.Errorf("getting commits since %s: %w", since, err)
	}
	if err != nil {
		return nil, fmt.Errorf("getting commits since %s: %w", since, err)
	}

	if len(commits) == 0 {
		return &Changelog{}, nil
	}

	entry := buildEntry("unreleased", "", commits)
	return &Changelog{Entries: []Entry{entry}}, nil
}

// ParseSince parses the --since flag value.
// Supports:
//   - "2025" (year, starting Jan 1)
//   - "8w" (weeks ago)
//   - "3m" (months ago)
//   - "1y" (years ago)
//   - ISO date string like "2024-01-15"
func ParseSince(s string) (string, error) {
	// Try ISO date first
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return s, nil
	}

	// Try year (4 digits)
	if len(s) == 4 {
		if year, err := strconv.Atoi(s); err == nil && year >= 1000 && year <= 9999 {
			return fmt.Sprintf("%d-01-01", year), nil
		}
	}

	// Try duration like "8w", "3m", "1y"
	if len(s) >= 2 {
		unit := s[len(s)-1]
		numStr := s[:len(s)-1]
		num, err := strconv.Atoi(numStr)
		if err == nil && num > 0 {
			var d time.Duration
			switch unit {
			case 'w':
				d = time.Duration(num) * 7 * 24 * time.Hour
			case 'm':
				d = time.Duration(num) * 30 * 24 * time.Hour
			case 'y':
				d = time.Duration(num) * 365 * 24 * time.Hour
			default:
				return "", fmt.Errorf("invalid duration unit %q (use w, m, or y)", string(unit))
			}
			return time.Now().Add(-d).Format("2006-01-02"), nil
		}
	}

	return "", fmt.Errorf("invalid --since value: %q (use a year like 2025, duration like 8w, or date like 2024-01-15)", s)
}

// WriteEntryDir writes changelog entry files to the specified directory.
// Each version gets two files: <version>.md and <version>_overview.md.
// Existing overview files are preserved and their content will be used
// in subsequent changelog generations if --dir is specified.
func WriteEntryDir(dir string, cl *Changelog) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	for _, entry := range cl.Entries {
		versionFile := filepath.Join(dir, entry.Version+".md")
		content := FormatEntry(entry)
		if err := os.WriteFile(versionFile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", versionFile, err)
		}

		// Write or preserve overview file
		overviewFile := filepath.Join(dir, entry.Version+"_overview.md")
		if _, err := os.Stat(overviewFile); os.IsNotExist(err) {
			// Create empty overview file so users know they can edit it
			if err := os.WriteFile(overviewFile, []byte{}, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", overviewFile, err)
			}
		}
	}

	return nil
}

// LoadOverviewFiles reads overview files from the directory and returns
// a map of version -> overview text. Only returns versions that have
// non-empty overview files.
func LoadOverviewFiles(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	overviews := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Match *_overview.md but not just _overview.md
		if len(name) > 12 && name[len(name)-12:] == "_overview.md" {
			content, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			trimmed := strings.TrimSpace(string(content))
			if trimmed != "" {
				version := name[:len(name)-12] // strip _overview.md
				overviews[version] = trimmed
			}
		}
	}

	return overviews, nil
}

// buildEntry creates an Entry from a list of structured commits.
func buildEntry(version, date string, commits []git.CommitInfo) Entry {
	items := make(map[Category][]Item)
	for _, c := range commits {
		cat, cleaned, ok := categorizeCommit(c.Message)
		if !ok {
			continue
		}
		items[cat] = append(items[cat], Item{Hash: c.Short, Message: cleaned})
	}

	return Entry{
		Version: version,
		Date:    date,
		Items:   items,
	}
}

// tagMatchesVersion checks if a git tag matches a version string.
// Handles both root tags ("v1.0.0") and module-prefixed tags ("x/mod/v1.0.0").
func tagMatchesVersion(tag, modulePath, version string) bool {
	if modulePath == "." {
		return tag == version
	}
	// For module paths, check both "x/mod/v1.0.0" == version and
	// allow the version to be specified without prefix
	return tag == version || tag == modulePath+"/"+version
}
