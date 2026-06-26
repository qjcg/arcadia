package scaffold

import (
	"strings"
)

// ExpandBraces expands brace patterns like "a/{b,c}/d" into ["a/b/d", "a/c/d"].
// Returns the input as a single-element slice if no braces are found.
func ExpandBraces(pattern string) []string {
	results := expandBracesRec(pattern)
	// Deduplicate while preserving order
	seen := make(map[string]bool)
	dedup := make([]string, 0, len(results))
	for _, r := range results {
		if !seen[r] {
			seen[r] = true
			dedup = append(dedup, r)
		}
	}
	return dedup
}

// SplitCommasRespectingBraces splits a string on commas that are not inside
// curly brace groups. Used to safely re-parse pflag slices that may have
// been broken apart by its comma-delimited string slice handling.
func SplitCommasRespectingBraces(s string) []string {
	return splitBraces(s)
}

func expandBracesRec(pattern string) []string {
	braceStart := strings.IndexByte(pattern, '{')
	if braceStart < 0 {
		return []string{pattern}
	}

	// Find matching closing brace, respecting nesting
	depth := 0
	braceEnd := -1
	for i := braceStart; i < len(pattern); i++ {
		switch pattern[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				braceEnd = i
				i = len(pattern)
			}
		}
	}

	if braceEnd < 0 {
		// Unmatched brace — treat as literal
		return []string{pattern}
	}

	prefix := pattern[:braceStart]
	middle := pattern[braceStart+1 : braceEnd]
	suffix := pattern[braceEnd+1:]

	// Split middle on unescaped commas, respecting nesting
	alternatives := splitBraces(middle)

	var results []string
	for _, alt := range alternatives {
		combined := prefix + alt + suffix
		results = append(results, expandBracesRec(combined)...)
	}
	return results
}

// splitBraces splits a brace middle on commas, respecting nested braces.
func splitBraces(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}
