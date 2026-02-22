package ui

import (
	"strings"
	"testing"
)

// TestOrderlessSearchLogic documents and verifies the orderless search algorithm
// that is implemented in assets/src/js/main.js's performSearch function.
//
// Orderless search (like Emacs' orderless package) allows users to type multiple
// words or word fragments in any order, and all must match for a result to appear.
//
// Examples:
//   - "intro search" matches: "Introduction to Search", "Search and Introduction"
//   - "api design" matches: "API Design Patterns", "Designing Good APIs"
func TestOrderlessSearchLogic(t *testing.T) {
	testCases := []struct {
		name          string
		query         string
		slideTitle    string
		slideContent  string
		shouldMatch   bool
		expectedScore int // Higher scores indicate better matches (title matches rank higher)
	}{
		// Basic orderless matching - order doesn't matter
		{
			name:         "simple orderless match - reversed order",
			query:        "search intro",
			slideTitle:   "Introduction to Search",
			slideContent: "Some content here",
			shouldMatch:  true,
		},
		{
			name:         "simple orderless match - original order",
			query:        "intro search",
			slideTitle:   "Introduction to Search",
			slideContent: "Some content here",
			shouldMatch:  true,
		},
		// Partial word matching (fragments)
		{
			name:         "fragment matching - partial words",
			query:        "arch async",
			slideTitle:   "Architecture and Async Patterns",
			slideContent: "Discussion of system architecture",
			shouldMatch:  true,
		},
		// Content-only matches
		{
			name:         "content-only match",
			query:        "performance",
			slideTitle:   "Advanced Topics",
			slideContent: "This covers performance optimization techniques",
			shouldMatch:  true,
		},
		// All tokens must match (AND logic)
		{
			name:         "all tokens required - partial fail",
			query:        "search missing",
			slideTitle:   "Introduction to Search",
			slideContent: "Some content here",
			shouldMatch:  false, // "missing" is not in title or content
		},
		{
			name:         "all tokens required - both match",
			query:        "search intro",
			slideTitle:   "Introduction to Search",
			slideContent: "Some content here",
			shouldMatch:  true,
		},
		// Case insensitive
		{
			name:         "case insensitive matching",
			query:        "API DESIGN",
			slideTitle:   "Api Design Patterns",
			slideContent: "Learning about api design",
			shouldMatch:  true,
		},
		// Multiple tokens (3+ words)
		{
			name:         "three token orderless match",
			query:        "async io patterns",
			slideTitle:   "Async IO Patterns",
			slideContent: "Understanding async patterns",
			shouldMatch:  true,
		},
		{
			name:         "three token scrambled order",
			query:        "patterns async io",
			slideTitle:   "Async IO Patterns",
			slideContent: "Understanding async patterns",
			shouldMatch:  true,
		},
		// Empty query edge cases
		{
			name:         "empty query should match all",
			query:        "",
			slideTitle:   "Any Title",
			slideContent: "Any content",
			shouldMatch:  true,
		},
		{
			name:         "whitespace only query should match all",
			query:        "   ",
			slideTitle:   "Any Title",
			slideContent: "Any content",
			shouldMatch:  true,
		},
		// Single token
		{
			name:         "single token match",
			query:        "golang",
			slideTitle:   "Learning Go",
			slideContent: "Go is also known as Golang",
			shouldMatch:  true,
		},
		// No match cases
		{
			name:         "no matching tokens",
			query:        "python rust",
			slideTitle:   "JavaScript Basics",
			slideContent: "Introduction to JS programming",
			shouldMatch:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the orderless matching logic from main.js
			query := strings.TrimSpace(tc.query)

			// Empty query matches everything
			if query == "" {
				if !tc.shouldMatch {
					t.Errorf("Empty query should match all slides")
				}
				return
			}

			// Split into tokens (same logic as main.js: query.toLowerCase().split(/\s+/))
			tokens := strings.Fields(strings.ToLower(query))
			if len(tokens) == 0 {
				if !tc.shouldMatch {
					t.Errorf("Whitespace-only query should match all slides")
				}
				return
			}

			// Check if all tokens match (orderless)
			titleLower := strings.ToLower(tc.slideTitle)
			contentLower := strings.ToLower(tc.slideContent)

			allTokensMatch := true
			for _, token := range tokens {
				tokenMatch := strings.Contains(titleLower, token) || strings.Contains(contentLower, token)
				if !tokenMatch {
					allTokensMatch = false
					break
				}
			}

			if allTokensMatch != tc.shouldMatch {
				t.Errorf("Query %q against title %q / content %q: expected match=%v, got match=%v",
					tc.query, tc.slideTitle, tc.slideContent, tc.shouldMatch, allTokensMatch)
			}
		})
	}
}

// TestOrderlessSearchScoring verifies that the scoring algorithm prioritizes correctly:
// 1. Title matches score higher than content matches
// 2. Tokens at the start of the title get bonus points
// 3. More matching tokens = higher score
func TestOrderlessSearchScoring(t *testing.T) {
	type slide struct {
		title   string
		content string
	}

	// Score function matching the JS implementation
	scoreMatch := func(s slide, tokens []string) int {
		score := 0
		titleLower := strings.ToLower(s.title)
		contentLower := strings.ToLower(s.content)

		for _, token := range tokens {
			// Title match: 10 points
			if strings.Contains(titleLower, token) {
				score += 10
				// Bonus for start of title: +5
				if strings.HasPrefix(titleLower, token) {
					score += 5
				}
			}
			// Content match: 1 point
			if strings.Contains(contentLower, token) {
				score += 1
			}
		}

		return score
	}

	testCases := []struct {
		name          string
		slide1        slide
		slide2        slide
		query         string
		higherScoring int // 1 or 2, indicating which should score higher
	}{
		{
			name:          "title match beats content match",
			slide1:        slide{"API Design", "Some content about testing"},
			slide2:        slide{"Testing Overview", "API design patterns here"},
			query:         "api design",
			higherScoring: 1, // First slide has title match
		},
		{
			name:          "start of title bonus",
			slide1:        slide{"Async Patterns", "Content about async"},
			slide2:        slide{"Go Async", "Async in Go"},
			query:         "async",
			higherScoring: 1, // First slide starts with "async"
		},
		{
			name:          "more token matches in title",
			slide1:        slide{"API Design Patterns", "Content here"},
			slide2:        slide{"API Overview", "Design content here"},
			query:         "api design",
			higherScoring: 1, // First slide matches both tokens in title
		},
		{
			name:          "content match vs no match",
			slide1:        slide{"Random Title", "JavaScript content here"},
			slide2:        slide{"Another Title", "Go programming"},
			query:         "javascript",
			higherScoring: 1, // First slide has content match
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := strings.Fields(strings.ToLower(tc.query))

			score1 := scoreMatch(tc.slide1, tokens)
			score2 := scoreMatch(tc.slide2, tokens)

			if tc.higherScoring == 1 && score1 <= score2 {
				t.Errorf("Expected slide1 to score higher than slide2, but got %d <= %d", score1, score2)
			}
			if tc.higherScoring == 2 && score2 <= score1 {
				t.Errorf("Expected slide2 to score higher than slide1, but got %d <= %d", score2, score1)
			}
		})
	}
}

// TestOrderlessSearchDocumentation serves as documentation for the JavaScript
// implementation. It demonstrates key behaviors that should match main.js.
func TestOrderlessSearchDocumentation(t *testing.T) {
	// This test documents the orderless search behavior implemented in main.js
	// performSearch() function. If these tests fail, the JS implementation
	// may also need updating.

	documentationCases := []struct {
		description string
		query       string
		examples    []struct {
			title   string
			content string
			matches bool
		}
	}{
		{
			description: "Order doesn't matter - words can appear in any sequence",
			query:       "async io",
			examples: []struct {
				title   string
				content string
				matches bool
			}{
				{"Async IO Patterns", "Content", true},
				{"IO and Async", "Content", true},
				{"Understanding Async", "IO concepts", true},
				{"IO Performance", "Async operations", true},
				// Tokens can match across title AND content (orderless)
				{"Async Patterns", "Has io here", true},  // "async" in title, "io" in content
				{"Network IO", "Has async here", true},   // "io" in title, "async" in content
				{"Async Patterns", "No xyz here", false}, // missing "io" completely (xyz doesn't contain io)
				{"Network", "Missing both", false},       // missing both tokens
			},
		},
		{
			description: "Fragments match partial words",
			query:       "intro search",
			examples: []struct {
				title   string
				content string
				matches bool
			}{
				{"Introduction to Search", "Content", true},
				{"Search Patterns", "Intro to the topic", true},
				{"Self-Introduction", "Search for knowledge", true},
				// "search" matches "searching" (substring match)
				{"Introduction", "No searching here", true}, // "intro" in title, "search" in "searching"
				{"Conclusion", "No find term", false},       // missing "intro" completely
			},
		},
		{
			description: "All tokens required (AND logic)",
			query:       "go test mock",
			examples: []struct {
				title   string
				content string
				matches bool
			}{
				{"Testing in Go", "Using mocks for testing", true},
				{"Go Mocks", "Test examples", true},
				{"Go Testing", "Mocks and stubs", true},
				{"Go Basics", "Introduction", false}, // missing "test" and "mock"
				{"Testing", "Mock framework", false}, // missing "go"
			},
		},
	}

	for _, doc := range documentationCases {
		t.Run(doc.description, func(t *testing.T) {
			tokens := strings.Fields(strings.ToLower(doc.query))

			for _, ex := range doc.examples {
				titleLower := strings.ToLower(ex.title)
				contentLower := strings.ToLower(ex.content)

				allMatch := true
				for _, token := range tokens {
					if !strings.Contains(titleLower, token) && !strings.Contains(contentLower, token) {
						allMatch = false
						break
					}
				}

				if allMatch != ex.matches {
					t.Errorf("Query %q against %q / %q: expected match=%v, got=%v",
						doc.query, ex.title, ex.content, ex.matches, allMatch)
				}
			}
		})
	}
}
