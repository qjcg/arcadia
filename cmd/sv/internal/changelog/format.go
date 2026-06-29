package changelog

import (
	"fmt"
	"sort"
	"strings"
)

// FormatChangelog formats the full changelog in keepachangelog format.
func FormatChangelog(cl *Changelog) string {
	var b strings.Builder

	b.WriteString("# Changelog\n\n")
	b.WriteString("All notable changes to this project will be documented in this file.\n\n")
	b.WriteString("The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),\n")
	b.WriteString("and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n")

	b.WriteString(formatChangelog(cl, 1))

	return b.String()
}

// FormatMultiModuleChangelog formats a combined changelog for multiple modules.
// A single "# Changelog" header is written at the top, then each module appears
// under a "## Module: <path>" section with its version entries demoted one level.
func FormatMultiModuleChangelog(modules map[string]*Changelog) string {
	var b strings.Builder

	b.WriteString("# Changelog\n\n")
	b.WriteString("All notable changes to this project will be documented in this file.\n\n")
	b.WriteString("The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),\n")
	b.WriteString("and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n")

	// Sort module paths for deterministic output
	var paths []string
	for path := range modules {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for i, path := range paths {
		cl := modules[path]
		if len(cl.Entries) == 0 {
			continue
		}
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("## Module: %s\n\n", path))
		b.WriteString(formatChangelog(cl, 2))
	}

	return b.String()
}

// formatChangelog formats entries with the given heading level offset.
// level=1 produces "## [version]" / "### Added".
// level=2 produces "### [version]" / "#### Added".
func formatChangelog(cl *Changelog, level int) string {
	var b strings.Builder
	for i, entry := range cl.Entries {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(formatEntryHeader(entry, level))
		if entry.Overview != "" {
			b.WriteString("\n" + entry.Overview + "\n")
		}
		b.WriteString(formatEntryItems(entry, level))
	}
	return b.String()
}

// FormatEntry formats a single version entry in keepachangelog format.
func FormatEntry(entry Entry) string {
	var b strings.Builder
	b.WriteString(formatEntryHeader(entry, 1))
	if entry.Overview != "" {
		b.WriteString("\n" + entry.Overview + "\n")
	}
	b.WriteString(formatEntryItems(entry, 1))
	return b.String()
}

func formatEntryHeader(entry Entry, level int) string {
	prefix := strings.Repeat("#", level+1)
	if entry.Date != "" {
		return fmt.Sprintf("%s [%s] - %s\n", prefix, entry.Version, entry.Date)
	}
	return fmt.Sprintf("%s [%s]\n", prefix, entry.Version)
}

func formatEntryItems(entry Entry, level int) string {
	prefix := strings.Repeat("#", level+2)
	var b strings.Builder
	for _, cat := range AllCategories {
		items, ok := entry.Items[cat]
		if !ok || len(items) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("\n%s %s\n", prefix, string(cat)))
		for _, item := range items {
			b.WriteString(FormatItem(item.Hash, item.Message) + "\n")
		}
	}
	return b.String()
}
