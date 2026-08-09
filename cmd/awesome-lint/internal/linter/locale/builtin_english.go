package locale

import "unicode"

// registerBuiltinEnglish registers the English profile. It is the default and
// is always registered first.
func registerBuiltinEnglish() {
	Register(&Profile{
		Name:      "english",
		Scripts:   []*unicode.RangeTable{unicode.Latin},
		Priority:  0,
		IsDefault: true,

		ValidEndingPunct:   []rune{'.', '!', '?', '…'},
		DefaultEndingPunct: '.',
		MapASCIIEnding:     nil,

		ContentsHeadings: []string{"contents", "table of contents", "toc"},

		Case: &CaseConfig{
			MinorWords: map[string]bool{
				"a": true, "an": true, "the": true,
				"and": true, "but": true, "or": true, "for": true, "nor": true,
				"on": true, "at": true, "to": true, "by": true, "from": true,
				"in": true, "of": true, "with": true, "as": true, "up": true,
				"is": true, "it": true, "vs": true,
			},
			AllowListed: map[string]bool{
				"title": true, "capital": true,
			},
		},
	})
}
