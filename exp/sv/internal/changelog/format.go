package changelog

import (
	"fmt"
	"strings"
)

// FormatChangelog formats the full changelog in keepachangelog format.
func FormatChangelog(cl *Changelog) string {
	var b strings.Builder

	b.WriteString("# Changelog\n\n")
	b.WriteString("All notable changes to this project will be documented in this file.\n\n")
	b.WriteString("The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),\n")
	b.WriteString("and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n")

	for i, entry := range cl.Entries {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(formatEntryHeader(entry))
		if entry.Overview != "" {
			b.WriteString("\n" + entry.Overview + "\n")
		}
		b.WriteString(formatEntryItems(entry))
	}

	return b.String()
}

// FormatEntry formats a single version entry in keepachangelog format.
func FormatEntry(entry Entry) string {
	var b strings.Builder
	b.WriteString(formatEntryHeader(entry))
	if entry.Overview != "" {
		b.WriteString("\n" + entry.Overview + "\n")
	}
	b.WriteString(formatEntryItems(entry))
	return b.String()
}

func formatEntryHeader(entry Entry) string {
	if entry.Date != "" {
		return fmt.Sprintf("## [%s] - %s\n", entry.Version, entry.Date)
	}
	return fmt.Sprintf("## [%s]\n", entry.Version)
}

func formatEntryItems(entry Entry) string {
	var b strings.Builder
	for _, cat := range AllCategories {
		items, ok := entry.Items[cat]
		if !ok || len(items) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("\n### %s\n", string(cat)))
		for _, item := range items {
			b.WriteString(FormatItem(item.Hash, item.Message) + "\n")
		}
	}
	return b.String()
}
