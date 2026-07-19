package expand

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Expand performs brace expansion on a single word.
// Returns the expanded words. For words without braces, returns [word].
func Expand(word string) []string {
	results := expandBrace(word)
	if len(results) == 0 {
		return []string{word}
	}
	return results
}

// expandBrace recursively expands brace groups in s.
func expandBrace(s string) []string {
	start, end, ok := findBraces(s)
	if !ok {
		return []string{s}
	}

	prefix := s[:start]
	content := s[start+1 : end]
	suffix := s[end+1:]

	parts := parseBraceContent(content)
	if len(parts) == 0 {
		// {} → literal
		return []string{s}
	}

	var results []string
	for _, part := range parts {
		expanded := expandBrace(part)
		for _, e := range expanded {
			// Recursively expand the full result (handles adjacent groups like {a,b}{1,2})
			for _, r := range expandBrace(prefix + e + suffix) {
				results = append(results, r)
			}
		}
	}

	return results
}

// findBraces finds the first unescaped '{' and its matching '}', handling nesting.
// Returns the indices of '{' and '}', or ok=false if no valid brace group found.
func findBraces(s string) (start, end int, ok bool) {
	depth := 0
	braceStart := -1
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		// Skip ${...} - variable expansion, not brace expansion
		if ch == '$' && i+1 < len(s) && s[i+1] == '{' {
			// Skip until matching }
			i += 2 // skip ${
			depthIn := 1
			for i < len(s) && depthIn > 0 {
				if s[i] == '{' {
					depthIn++
				} else if s[i] == '}' {
					depthIn--
				}
				if depthIn > 0 {
					i++
				}
			}
			continue
		}

		if ch == '{' {
			if depth == 0 {
				braceStart = i
			}
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 && braceStart >= 0 {
				// Must have comma or range pattern to be a valid brace group
				inner := s[braceStart+1 : i]
				if isValidBraceContent(inner) {
					return braceStart, i, true
				}
				// Not a valid brace group, look for another
				braceStart = -1
			}
			if depth < 0 {
				return 0, 0, false
			}
		}
	}

	return 0, 0, false
}

// isValidBraceContent checks if the content between braces is a valid brace expression.
// Must contain a comma at depth 0, match a range pattern, or be a single element.
func isValidBraceContent(content string) bool {
	if len(content) == 0 {
		return false
	}

	// Check for comma at depth 0
	hasComma := false
	depth := 0
	escaped := false
	for i := 0; i < len(content); i++ {
		ch := content[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
		} else if ch == ',' && depth == 0 {
			hasComma = true
		}
	}

	if hasComma {
		return true
	}

	// Check for range pattern
	if isValidRange(content) {
		return true
	}

	// Single element like {x} — valid only if it doesn't look like an invalid range
	// e.g., {1..a} or {a..1} are not valid
	if strings.Contains(content, "..") {
		return false
	}

	return true
}

// isValidRange checks if the content matches a range pattern like "1..5" or "a..f" or "1..10..2".
func isValidRange(content string) bool {
	parts := strings.Split(content, "..")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}

	start := strings.TrimSpace(parts[0])
	end := strings.TrimSpace(parts[1])

	if len(start) == 0 || len(end) == 0 {
		return false
	}

	// Both must be numeric or both must be alphabetic
	startIsNum := isNumeric(start)
	endIsNum := isNumeric(end)
	startIsAlpha := isAlpha(start)
	endIsAlpha := isAlpha(end)

	if startIsNum && endIsNum {
		// Check for overflow
		_, err1 := strconv.ParseInt(start, 10, 64)
		_, err2 := strconv.ParseInt(end, 10, 64)
		return err1 == nil && err2 == nil
	}

	if startIsAlpha && endIsAlpha && len(start) == 1 && len(end) == 1 {
		return true
	}

	return false
}

// parseBraceContent parses the content of a brace group and returns the expanded parts.
// For lists, splits by comma. For ranges, generates the sequence.
func parseBraceContent(content string) []string {
	// Check for range first
	if isValidRange(content) {
		return expandRange(content)
	}

	// Split by comma at depth 0
	return splitByComma(content)
}

// splitByComma splits content by commas at depth 0, trimming whitespace.
func splitByComma(content string) []string {
	var parts []string
	depth := 0
	escaped := false
	start := 0

	for i := 0; i < len(content); i++ {
		ch := content[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
		} else if ch == ',' && depth == 0 {
			parts = append(parts, content[start:i])
			start = i + 1
		}
	}

	parts = append(parts, content[start:])
	return parts
}

// expandRange expands a range pattern like "1..5" or "a..f" or "1..10..2".
func expandRange(content string) []string {
	parts := strings.Split(content, "..")

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	var step int64 = 1
	if len(parts) == 3 && parts[2] != "" {
		s, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil || s <= 0 {
			return nil
		}
		step = s
	}

	// Check if numeric
	if isNumeric(startStr) && isNumeric(endStr) {
		return expandNumericRange(startStr, endStr, step)
	}

	// Check if alphabetic
	if isAlpha(startStr) && isAlpha(endStr) && len(startStr) == 1 && len(endStr) == 1 {
		return expandAlphaRange(rune(startStr[0]), rune(endStr[0]), step)
	}

	return nil
}

// expandNumericRange generates a numeric range with optional zero-padding.
func expandNumericRange(startStr, endStr string, step int64) []string {
	start, _ := strconv.ParseInt(startStr, 10, 64)
	end, _ := strconv.ParseInt(endStr, 10, 64)

	width := max(len(endStr), len(startStr))

	// Detect zero-padding: if either has a leading zero, pad to the wider width
	hasLeadingZero := (len(startStr) > 1 && startStr[0] == '0') || (len(endStr) > 1 && endStr[0] == '0')
	zeroPad := hasLeadingZero

	var results []string
	if start <= end {
		for i := start; i <= end; i += step {
			if zeroPad {
				results = append(results, fmt.Sprintf("%0*d", width, i))
			} else {
				results = append(results, strconv.FormatInt(i, 10))
			}
		}
	} else {
		for i := start; i >= end; i -= step {
			if zeroPad {
				results = append(results, fmt.Sprintf("%0*d", width, i))
			} else {
				results = append(results, strconv.FormatInt(i, 10))
			}
		}
	}

	return results
}

// expandAlphaRange generates an alphabetic range.
func expandAlphaRange(start, end rune, step int64) []string {
	var results []string
	if start <= end {
		for i := start; i <= end; i += rune(step) {
			results = append(results, string(i))
		}
	} else {
		for i := start; i >= end; i -= rune(step) {
			results = append(results, string(i))
		}
	}
	return results
}

// isNumeric checks if a string is a valid integer (optionally with leading zeros).
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// isAlpha checks if a string contains only alphabetic characters.
func isAlpha(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
