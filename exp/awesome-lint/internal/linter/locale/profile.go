// Package locale defines pluggable per-language formatting conventions.
//
// A Profile captures the punctuation, Table of Contents heading names, and
// casing rules of a language. Rules consult the profile detected for a
// document instead of hardcoding language-specific behavior, so new languages
// can be supported by registering another profile.
package locale

import (
	"slices"
	"strings"
	"unicode"
)

// Profile defines the formatting conventions of a language.
type Profile struct {
	// Name is a stable identifier such as "english" or "chinese".
	Name string
	// Scripts are Unicode ranges used to detect the language in a document.
	Scripts []*unicode.RangeTable
	// Priority breaks ties between profiles sharing a script; higher wins.
	Priority int
	// IsDefault marks the fallback profile. It never wins detection; it is
	// returned only when no other profile matches the document.
	IsDefault bool

	// ValidEndingPunct lists runes accepted as list-item description
	// termination in this language.
	ValidEndingPunct []rune
	// DefaultEndingPunct is appended when a description lacks one.
	DefaultEndingPunct rune
	// MapASCIIEnding maps ASCII punctuation to its localized equivalent used
	// by --fix (for example '.' -> '。').
	MapASCIIEnding map[rune]rune

	// ContentsHeadings lists the (lowercased) heading texts that identify a
	// Table of Contents in this language.
	ContentsHeadings []string

	// Case holds casing conventions; nil means the language has no word
	// casing, so both title-casing and description-casing checks are skipped.
	Case *CaseConfig
}

// CaseConfig describes the casing conventions of a language.
type CaseConfig struct {
	MinorWords  map[string]bool
	AllowListed map[string]bool
}

// HasValidEnding reports whether s ends with acceptable punctuation.
func (p *Profile) HasValidEnding(s string) bool {
	runes := []rune(s)
	if len(runes) == 0 {
		return false
	}
	last := runes[len(runes)-1]
	return slices.Contains(p.ValidEndingPunct, last)
}

// FixEndingPunctuation returns s with the trailing ASCII punctuation replaced
// by its localized equivalent, or the default ending punctuation appended.
func (p *Profile) FixEndingPunctuation(s string) string {
	runes := []rune(s)
	if len(runes) > 0 {
		if repl, ok := p.MapASCIIEnding[runes[len(runes)-1]]; ok {
			runes[len(runes)-1] = repl
			return string(runes)
		}
	}
	return s + string(p.DefaultEndingPunct)
}

// IsContentsHeading reports whether heading text identifies a Table of
// Contents in this language.
func (p *Profile) IsContentsHeading(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	return slices.Contains(p.ContentsHeadings, lower)
}

// AppliesTitleCase reports whether the language uses title casing on headings.
func (p *Profile) AppliesTitleCase() bool { return p.Case != nil }

// AppliesDescriptionCase reports whether the language enforces description
// casing.
func (p *Profile) AppliesDescriptionCase() bool { return p.Case != nil }

// TitleCase returns heading text normalized to the language's title case. For
// languages without a casing concept it returns the input unchanged.
func (p *Profile) TitleCase(s string) string {
	if p.Case == nil || s == "" {
		return s
	}
	words := strings.Fields(s)
	for i, w := range words {
		if i == 0 || i == len(words)-1 {
			words[i] = capitalizeFirst(w)
		} else if p.IsMinorWord(w) {
			words[i] = strings.ToLower(w)
		} else {
			words[i] = capitalizeFirst(w)
		}
	}
	return strings.Join(words, " ")
}

// IsMinorWord reports whether w is a minor (uncapitalized) title word.
func (p *Profile) IsMinorWord(w string) bool {
	if p.Case == nil {
		return false
	}
	return p.Case.MinorWords[strings.ToLower(w)]
}

// CaseAllowListed reports whether s is exempt from casing checks.
func (p *Profile) CaseAllowListed(s string) bool {
	if p.Case == nil {
		return true
	}
	return p.Case.AllowListed[strings.ToLower(s)]
}

// IsValidCasing reports whether word observes the language's casing rules.
// The check accepts full-uppercase, an uppercase first letter, or any word
// containing an uppercase letter. Languages without a casing concept always
// accept the word.
func (p *Profile) IsValidCasing(word string) bool {
	if p.Case == nil {
		return true
	}
	cleaned := strings.Map(func(r rune) rune {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return -1
		}
		return r
	}, word)

	if cleaned == "" {
		return false
	}

	if strings.ToUpper(cleaned) == cleaned {
		return true
	}

	runes := []rune(cleaned)
	if unicode.IsUpper(runes[0]) {
		return true
	}

	for _, r := range cleaned {
		if unicode.IsUpper(r) {
			return true
		}
	}

	return false
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
