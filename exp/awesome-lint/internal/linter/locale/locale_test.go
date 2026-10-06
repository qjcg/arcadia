package locale

import "testing"

func TestForDocDetection(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"english defaults when no other script", "A curated list of awesome things.\n", "english"},
		{"chinese detected by Han script", "一个精选列表。\n更多内容。\n", "chinese"},
		{"chinese detected alongside latin", "- [示例](https://example.com) - 这是描述。\n", "chinese"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ForDoc([]byte(tt.src)).Name; got != tt.want {
				t.Errorf("ForDoc(%q) = %q, want %q", tt.src, got, tt.want)
			}
		})
	}
}

func TestEnglishEnding(t *testing.T) {
	en := Default()
	if !en.HasValidEnding("Description.") {
		t.Error("english should accept '.' as ending")
	}
	if en.HasValidEnding("没有句号") {
		t.Error("english should not accept a Chinese ideographic stop")
	}
	if got := en.FixEndingPunctuation("Description"); got != "Description." {
		t.Errorf("FixEndingPunctuation = %q, want %q", got, "Description.")
	}
	if !en.AppliesTitleCase() || !en.AppliesDescriptionCase() {
		t.Error("english should apply title and description case")
	}
}

func TestChineseEnding(t *testing.T) {
	zh := ForDoc([]byte("这是一段中文。"))
	if !zh.HasValidEnding("这是描述。") {
		t.Error("chinese should accept ideographic full stop")
	}
	if zh.HasValidEnding("This is English.") {
		t.Error("chinese should not accept an ASCII period")
	}
	if got := zh.FixEndingPunctuation("这是描述."); got != "这是描述。" {
		t.Errorf("FixEndingPunctuation = %q, want %q", got, "这是描述。")
	}
	if got := zh.FixEndingPunctuation("这是描述"); got != "这是描述。" {
		t.Errorf("FixEndingPunctuation default = %q, want %q", got, "这是描述。")
	}
	if zh.AppliesTitleCase() || zh.AppliesDescriptionCase() {
		t.Error("chinese should skip casing checks")
	}
}

func TestContentsHeadings(t *testing.T) {
	en := Default()
	if !en.IsContentsHeading("Contents") {
		t.Error("english should recognize 'Contents'")
	}
	if en.IsContentsHeading("目录") {
		t.Error("english should not recognize '目录'")
	}
	zh := ForDoc([]byte("中文目录列表。"))
	if !zh.IsContentsHeading("目录") {
		t.Error("chinese should recognize '目录'")
	}
	if zh.IsContentsHeading("Contents") {
		t.Error("chinese should not recognize 'Contents'")
	}
	if !zh.IsContentsHeading(" 目录 ") {
		t.Error("chinese should trim heading text before matching")
	}
}
