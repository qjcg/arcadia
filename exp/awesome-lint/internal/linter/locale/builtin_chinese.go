package locale

import "unicode"

// registerBuiltinChinese registers the Chinese profile.
//
// Chinese uses the ideographic full stop (。U+3002) and full-width
// punctuation. It has no word casing, so Case is nil and casing checks are
// skipped.
func registerBuiltinChinese() {
	Register(&Profile{
		Name:     "chinese",
		Scripts:  []*unicode.RangeTable{unicode.Han},
		Priority: 1,

		ValidEndingPunct:   []rune{'。', '！', '？', '…'},
		DefaultEndingPunct: '。',
		MapASCIIEnding: map[rune]rune{
			'.': '。',
			'!': '！',
			'?': '？',
		},

		ContentsHeadings: []string{"目录", "目錄"},

		Case: nil,
	})
}
