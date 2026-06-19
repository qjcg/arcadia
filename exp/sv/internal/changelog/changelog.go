package changelog

import (
	"fmt"
	"regexp"
	"strings"
)

// Category represents a keepachangelog section category.
type Category string

const (
	Added      Category = "Added"
	Changed    Category = "Changed"
	Deprecated Category = "Deprecated"
	Removed    Category = "Removed"
	Fixed      Category = "Fixed"
	Security   Category = "Security"
)

// AllCategories is the order categories appear in keepachangelog output.
var AllCategories = []Category{Added, Changed, Deprecated, Removed, Fixed, Security}

// Item represents a single changelog entry item with its commit reference.
type Item struct {
	Hash    string // Short commit hash (7 chars)
	Message string // Cleaned commit message
}

// URLPrefix is the base URL for linking commit hashes.
// When set, items render as "- [hash](prefix/hash) - message".
var URLPrefix string

// FormatItem formats a changelog item with its commit hash.
// With URL prefix: "- [abc1234](https://example.com/abc1234) - message"
// Without: "- abc1234 - message"
func FormatItem(hash, message string) string {
	if URLPrefix != "" {
		return fmt.Sprintf("- [%s](%s%s) - %s", hash, URLPrefix, hash, message)
	}
	return fmt.Sprintf("- %s - %s", hash, message)
}

// Entry represents a single version's changelog entry.
type Entry struct {
	Version  string              // e.g., "v0.1.0" or "unreleased"
	Date     string              // ISO date string or empty for unreleased
	Overview string              // Commentary text placed before sections
	Items    map[Category][]Item // Categorized changelog items
}

// Changelog holds all entries for the full changelog.
type Changelog struct {
	Entries []Entry
}

// categorizeCommit parses a conventional commit message and returns its
// keepachangelog category, cleaned message body, and whether it should be included.
func categorizeCommit(msg string) (Category, string, bool) {
	isBreaking := strings.Contains(msg, "BREAKING CHANGE:") || strings.Contains(msg, "!:")

	commitType := extractType(msg)

	// Breaking changes always go in Changed regardless of type
	if isBreaking {
		return Changed, cleanMessage(msg), true
	}

	switch commitType {
	case "feat":
		return Added, cleanMessage(msg), true
	case "fix":
		return Fixed, cleanMessage(msg), true
	case "deprecated":
		return Deprecated, cleanMessage(msg), true
	case "remove", "removed":
		return Removed, cleanMessage(msg), true
	case "security":
		return Security, cleanMessage(msg), true
	case "docs", "chore", "refactor", "test", "style", "build", "ci", "perf":
		return Changed, cleanMessage(msg), true
	default:
		return "", "", false
	}
}

// extractType extracts the conventional commit type from a message.
// Handles formats: "type: msg", "type(scope): msg", "type!: msg", "type(scope)!: msg".
func extractType(msg string) string {
	re := regexp.MustCompile(`^([a-zA-Z]+)(\([^)]*\))?!?:\s`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) > 1 {
		return strings.ToLower(matches[1])
	}
	return ""
}

// cleanMessage removes the conventional commit type prefix and any BREAKING CHANGE footers.
func cleanMessage(msg string) string {
	// Strip BREAKING CHANGE prefix or footer
	if after, ok := strings.CutPrefix(msg, "BREAKING CHANGE:"); ok {
		msg = strings.TrimSpace(after)
		return msg
	}
	if idx := strings.Index(msg, "BREAKING CHANGE:"); idx > 0 {
		msg = strings.TrimSpace(msg[:idx])
	}
	// Strip type(scope)!: or type: prefix
	re := regexp.MustCompile(`^[a-zA-Z]+(\([^)]*\))?!?:\s*`)
	return re.ReplaceAllString(msg, "")
}
