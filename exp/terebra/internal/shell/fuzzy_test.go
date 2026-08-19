package shell

import (
	"testing"
)

func TestFuzzyMatchEmptyQuery(t *testing.T) {
	ok, score := fuzzyMatch("hello world", "")
	if !ok {
		t.Fatal("expected match for empty query")
	}
	if score != 0 {
		t.Fatalf("expected score 0, got %d", score)
	}
}

func TestFuzzyMatchExact(t *testing.T) {
	ok, score := fuzzyMatch("hello", "hello")
	if !ok {
		t.Fatal("expected match")
	}
	if score <= 0 {
		t.Fatalf("expected positive score, got %d", score)
	}
}

func TestFuzzyMatchCaseInsensitive(t *testing.T) {
	ok, _ := fuzzyMatch("Hello World", "hello")
	if !ok {
		t.Fatal("expected case-insensitive match")
	}
}

func TestFuzzyMatchSubsequence(t *testing.T) {
	ok, _ := fuzzyMatch("abcdef", "ace")
	if !ok {
		t.Fatal("expected subsequence match")
	}
}

func TestFuzzyMatchNoMatch(t *testing.T) {
	ok, score := fuzzyMatch("abc", "xyz")
	if ok {
		t.Fatal("expected no match")
	}
	if score != 0 {
		t.Fatalf("expected score 0, got %d", score)
	}
}

func TestFuzzyMatchWordStartBonus(t *testing.T) {
	// Match at word start should score higher than mid-word
	ok1, s1 := fuzzyMatch("foo bar", "b")
	ok2, s2 := fuzzyMatch("foo bar", "a")
	if !ok1 || !ok2 {
		t.Fatal("both should match")
	}
	if s1 <= s2 {
		t.Fatalf("expected word-start match (%d) to beat mid-word (%d)", s1, s2)
	}
}

func TestFuzzyMatchConsecutiveBonus(t *testing.T) {
	ok1, s1 := fuzzyMatch("abcdef", "ab")
	ok2, s2 := fuzzyMatch("abcdef", "af")
	if !ok1 || !ok2 {
		t.Fatal("both should match")
	}
	if s1 <= s2 {
		t.Fatalf("expected consecutive match (%d) to beat non-consecutive (%d)", s1, s2)
	}
}

func TestFuzzyModelFilterSortsByScore(t *testing.T) {
	filtered := filterEntries([]string{"zzz", "ba", "a"}, "a")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered entries, got %d", len(filtered))
	}
	// "a" (word-start match) should rank above "ba" (mid-word match)
	if filtered[0].entry != "a" {
		t.Fatalf("expected 'a' first, got %q", filtered[0].entry)
	}
}

func TestFuzzyModelFilterEmpty(t *testing.T) {
	filtered := filterEntries([]string{"x", "y"}, "")
	if len(filtered) != 2 {
		t.Fatalf("expected all entries with empty query, got %d", len(filtered))
	}
}

func TestFuzzyModelFilterNoMatches(t *testing.T) {
	filtered := filterEntries([]string{"abc"}, "zzz")
	if len(filtered) != 0 {
		t.Fatalf("expected no matches, got %d", len(filtered))
	}
}

func TestFilterEntriesCapsAtMaxResults(t *testing.T) {
	entries := make([]string, maxResults+10)
	for i := range entries {
		entries[i] = "item"
	}
	filtered := filterEntries(entries, "item")
	if len(filtered) != maxResults {
		t.Fatalf("expected %d results, got %d", maxResults, len(filtered))
	}
}

func TestHighlightMatchEmptyQuery(t *testing.T) {
	got := highlightMatch("hello", "")
	if got != "hello" {
		t.Fatalf("expected unchanged entry, got %q", got)
	}
}

func TestHighlightMatch(t *testing.T) {
	got := highlightMatch("hello", "ho")
	// Styled output should be longer than plain due to ANSI codes
	if len(got) <= len("hello") {
		t.Fatalf("expected styled output longer than plain, got %q", got)
	}
}

func TestHighlightMatchNoMatch(t *testing.T) {
	got := highlightMatch("hello", "xyz")
	if got != "hello" {
		t.Fatalf("expected unchanged entry, got %q", got)
	}
}
