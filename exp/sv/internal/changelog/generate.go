package changelog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	mastersemver "github.com/Masterminds/semver/v3"
	"github.com/qjcg/arcadia/exp/sv/internal/discovery"
	"github.com/qjcg/arcadia/exp/sv/internal/git"
	"github.com/qjcg/arcadia/exp/sv/internal/semver"
)

// Generate generates a Changelog for the given module.
// If from or to looks like a date (year, duration, or ISO date), it uses
// date-based tag generation (filtering tags by their date). Otherwise it
// treats from/to as version tags.
func Generate(root, modulePath, from, to string) (*Changelog, error) {
	fromDate, fromErr := parseSinceCheck(from)
	toDate, toErr := parseSinceCheck(to)

	if fromErr == nil || toErr == nil {
		return generateDateMode(root, modulePath, fromDate, toDate)
	}
	return generateVersionMode(root, modulePath, from, to)
}

func generateDateMode(root, modulePath, fromDate, toDate string) (*Changelog, error) {
	if fromDate == "" && toDate == "" {
		// Both are empty — no date filtering, same as version mode with no bounds
		return generateVersionMode(root, modulePath, "", "")
	}

	allTags, err := git.Tags(root, modulePath)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}

	// Build ascending tag list
	ascTags := make([]string, len(allTags))
	for i, t := range allTags {
		ascTags[len(ascTags)-1-i] = t
	}

	// Filter tags by date range
	startIdx := 0
	if fromDate != "" {
		startIdx = len(ascTags) // default: past end (no match)
		for i, t := range ascTags {
			tagDate, err := git.TagDate(root, t)
			if err == nil && tagDate >= fromDate {
				startIdx = i
				break
			}
		}
	}

	endIdx := len(ascTags) - 1
	if toDate != "" {
		endIdx = -1 // default: before start (no match)
		for i, t := range ascTags {
			tagDate, err := git.TagDate(root, t)
			if err != nil {
				continue
			}
			if tagDate > toDate {
				break // past the date range in ascending order
			}
			if i >= startIdx {
				endIdx = i
			}
		}
	}

	if startIdx <= endIdx {
		return generateFromTags(root, modulePath, ascTags, startIdx, endIdx, toDate == "")
	}

	// No tags fall in the date range
	if toDate == "" {
		// No explicit end boundary: show unreleased commits since fromDate
		commits, err := git.CommitsSinceDateDetail(root, fromDate, modulePath, submoduleExclusions(root, modulePath))
		if err != nil {
			return nil, fmt.Errorf("getting commits: %w", err)
		}
		if len(commits) == 0 {
			return &Changelog{}, nil
		}
		entry := buildEntry(unreleasedVersion(modulePath), "", commits)
		return &Changelog{Entries: []Entry{entry}}, nil
	}
	return &Changelog{}, nil
}

func generateVersionMode(root, modulePath, fromVersion, toVersion string) (*Changelog, error) {
	if fromVersion != "" || toVersion != "" {
		return GenerateFromGit(root, modulePath, fromVersion, toVersion)
	}
	return GenerateFromGit(root, modulePath, "", "")
}

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

	return generateFromTags(root, modulePath, ascTags, startIdx, endIdx, toVersion == "")
}

// generateFromTags builds changelog entries from a filtered range of tags.
// ascTags is the full ascending tag list; startIdx and endIdx define the range.
func generateFromTags(root, modulePath string, ascTags []string, startIdx, endIdx int, addUnreleased bool) (*Changelog, error) {
	tags := ascTags[startIdx : endIdx+1]
	excludePaths := submoduleExclusions(root, modulePath)

	// Filter out retracted versions so they don't generate entries or skew
	// the unreleased range. Retracted tags (like v1.0.0, v1.0.1, v1.0.2) are
	// often created at historically earlier commits than higher semver tags,
	// causing version-sorted tag lists to produce incorrect commit ranges.
	tags = filterNonRetracted(tags, root, modulePath)
	if len(tags) == 0 {
		return &Changelog{}, nil
	}

	var entries []Entry
	for i, tag := range tags {
		var commits []git.CommitInfo
		var err error

		// Determine previous tag for commit range using adjacent entries in the
		// filtered list, which is already sorted by version.
		if i > 0 {
			commits, err = git.CommitsBetweenDetail(root, tags[i-1], tag, modulePath, excludePaths)
		} else {
			commits, err = git.CommitsBetweenDetail(root, "", tag, modulePath, excludePaths)
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

	// Add unreleased (commits from newest non-retracted tag in range to HEAD)
	if addUnreleased {
		latestTag := tags[len(tags)-1]
		unreleasedCommits, err := git.CommitsBetweenDetail(root, latestTag, "", modulePath, excludePaths)
		if err == nil && len(unreleasedCommits) > 0 {
			entry := buildEntry(unreleasedVersion(modulePath), "", unreleasedCommits)
			entries = append(entries, entry)
		}
	}

	// Sort entries: newest first for output using semver comparison
	unrelVer := unreleasedVersion(modulePath)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Version == unrelVer {
			return true
		}
		if entries[j].Version == unrelVer {
			return false
		}
		// Strip module path prefix for proper semver parsing
		a := entries[i].Version
		b := entries[j].Version
		if modulePath != "." {
			a = strings.TrimPrefix(a, modulePath+"/")
			b = strings.TrimPrefix(b, modulePath+"/")
		}
		vi, errI := mastersemver.NewVersion(a)
		vj, errJ := mastersemver.NewVersion(b)
		if errI == nil && errJ == nil {
			return vi.GreaterThan(vj)
		}
		// Fall back to string comparison if parsing fails
		return entries[i].Version > entries[j].Version
	})

	return &Changelog{Entries: entries}, nil
}

// filterNonRetracted removes retracted versions from the tag list and returns
// the filtered slice. Retracted tags don't represent real releases and should
// not appear in the changelog.
func filterNonRetracted(tags []string, root, modulePath string) []string {
	goMod, err := git.ReadGoMod(root, modulePath)
	if err != nil {
		return tags
	}
	retractions := semver.ParseRetractions(goMod)
	if len(retractions) == 0 {
		return tags
	}

	filtered := make([]string, 0, len(tags))
	for _, t := range tags {
		// Extract version part: strip module path prefix
		versionPart := t
		if modulePath != "." && strings.HasPrefix(t, modulePath+"/") {
			versionPart = strings.TrimPrefix(t, modulePath+"/")
		}
		if !semver.IsVersionRetracted(versionPart, retractions) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// submoduleExclusions returns paths of submodules to exclude when generating
// changelog for the root module, so submodule commits don't leak into the root.
func submoduleExclusions(root, modulePath string) []string {
	if modulePath != "." && modulePath != "" {
		return nil
	}
	modules, err := discovery.FindModules(root)
	if err != nil {
		return nil
	}
	var exclusions []string
	for _, m := range modules {
		if m.Name != "." {
			exclusions = append(exclusions, m.Name)
		}
	}
	return exclusions
}

// unreleasedVersion returns the version string for unreleased entries.
// For root modules it's "unreleased"; for sub-modules it includes the module path.
func unreleasedVersion(modulePath string) string {
	if modulePath == "." || modulePath == "" {
		return "unreleased"
	}
	return modulePath + "/unreleased"
}

// parseSince parses a --from value that represents a date, duration, or year.
// Supports:
//   - "2025" (year, starting Jan 1)
//   - "8w" (weeks ago)
//   - "3m" (months ago)
//   - "1y" (years ago)
//   - ISO date string like "2024-01-15"
func parseSince(s string) (string, error) {
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

// parseSinceCheck is like parseSince but treats an empty string as "not a date"
// rather than an error, so empty from/to never trigger date-mode.
func parseSinceCheck(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("empty string")
	}
	return parseSince(s)
}

// WriteEntryDir writes changelog entry files to the specified directory.
// Each version gets two files: <version>.md and <version>_overview.md.
// Existing overview files are preserved and their content will be used
// in subsequent changelog generations if --dir is specified.
func WriteEntryDir(dir string, cl *Changelog) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Remove stale entry files (from previous runs) while preserving overview files
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			// Remove .md files that are not overview files
			if strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, "_overview.md") {
				os.Remove(filepath.Join(dir, name))
			}
		}
	}

	for _, entry := range cl.Entries {
		versionFile := filepath.Join(dir, entry.Version+".md")
		if err := os.MkdirAll(filepath.Dir(versionFile), 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", versionFile, err)
		}
		content := FormatEntry(entry)
		if err := os.WriteFile(versionFile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", versionFile, err)
		}

		// Write or preserve overview file
		overviewFile := filepath.Join(dir, entry.Version+"_overview.md")
		if err := os.MkdirAll(filepath.Dir(overviewFile), 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", overviewFile, err)
		}
		if _, err := os.Stat(overviewFile); os.IsNotExist(err) {
			// Create empty overview file so users know they can edit it
			if err := os.WriteFile(overviewFile, []byte{}, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", overviewFile, err)
			}
		}
	}

	return nil
}

// WriteChangelogFile writes a CHANGELOG.md file into the given module directory
// with the full changelog in keepachangelog format.
func WriteChangelogFile(moduleDir string, cl *Changelog) error {
	path := filepath.Join(moduleDir, "CHANGELOG.md")
	content := FormatChangelog(cl)
	return os.WriteFile(path, []byte(content), 0o644)
}

// LoadOverviewFiles reads overview files from the directory tree and returns
// a map of version -> overview text. Only returns versions that have
// non-empty overview files. Searches recursively to support module-prefixed
// versions (e.g. "x/mod/v1.0.0" stored under dir/x/mod/).
func LoadOverviewFiles(dir string) (map[string]string, error) {
	overviews := make(map[string]string)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		// Match *_overview.md but not just _overview.md
		if len(name) > 12 && name[len(name)-12:] == "_overview.md" {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			trimmed := strings.TrimSpace(string(content))
			if trimmed != "" {
				// Reconstruct version from relative path
				rel, err := filepath.Rel(dir, path)
				if err != nil {
					return nil
				}
				rel = filepath.ToSlash(rel)
				version := rel[:len(rel)-12] // strip _overview.md
				overviews[version] = trimmed
			}
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
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
