package linter

import (
	"strings"
	"testing"
)

func TestHeadingRule(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
		ruleID   string
	}{
		{
			name:     "valid heading",
			markdown: "# Awesome List\n\n## Contents\n\nContent here.\n",
			wantErr:  false,
			ruleID:   "awesome-heading",
		},
		{
			name:     "missing heading",
			markdown: "Some text without a heading.\n",
			wantErr:  true,
			ruleID:   "awesome-heading",
		},
		{
			name:     "first heading not level 1",
			markdown: "## Awesome List\n\nContent here.\n",
			wantErr:  true,
			ruleID:   "awesome-heading",
		},
		{
			name:     "first heading not title case",
			markdown: "# awesome list\n\nContent here.\n",
			wantErr:  true,
			ruleID:   "awesome-heading",
		},
		{
			name:     "multiple H1 headings",
			markdown: "# Awesome List\n\n# Another H1\n\nContent here.\n",
			wantErr:  true,
			ruleID:   "awesome-heading",
		},
		{
			name:     "HTML heading detected",
			markdown: "<h1>Awesome List</h1>\n\nContent here.\n",
			wantErr:  false,
			ruleID:   "awesome-heading",
		},
		{
			name:     "centered image as heading",
			markdown: "<div align=\"center\"><img src=\"logo.png\"></div>\n\nSome content here.\n",
			wantErr:  false,
			ruleID:   "awesome-heading",
		},
	}

	rule := &headingRule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseMarkdown([]byte(tt.markdown))
			results := rule.Check(doc, []byte(tt.markdown))
			hasErr := false
			for _, r := range results {
				if r.RuleID == tt.ruleID && r.Severity == SeverityError {
					hasErr = true
					break
				}
			}
			if hasErr != tt.wantErr {
				t.Errorf("headingRule.Check() hasErr = %v, wantErr = %v; results: %+v", hasErr, tt.wantErr, results)
			}
		})
	}
}

func TestBadgeRule(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
	}{
		{
			name:     "valid badge present",
			markdown: "# [![Awesome](https://awesome.re/badge.svg)](https://awesome.re)\n\nContent here.\n",
			wantErr:  false,
		},
		{
			name:     "badge missing",
			markdown: "# Awesome List\n\nContent here.\n",
			wantErr:  true,
		},
		{
			name:     "invalid badge source",
			markdown: "# [![Awesome](https://evil.com/badge.svg)](https://awesome.re)\n\nContent here.\n",
			wantErr:  true,
		},
		{
			name:     "badge in non-H1 heading",
			markdown: "# Awesome List\n\n## [![Awesome](https://awesome.re/badge.svg)](https://awesome.re)\n\nContent here.\n",
			wantErr:  true,
		},
		{
			name:     "badge with flat variant",
			markdown: "# [![Awesome](https://awesome.re/badge-flat.svg)](https://awesome.re)\n\nContent here.\n",
			wantErr:  false,
		},
		{
			name:     "badge on separate line after heading",
			markdown: "# Awesome List\n\n[![Awesome](https://awesome.re/badge.svg)](https://github.com/sindresorhus/awesome)\n\nContent here.\n",
			wantErr:  false,
		},
	}

	rule := &badgeRule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseMarkdown([]byte(tt.markdown))
			results := rule.Check(doc, []byte(tt.markdown))
			hasErr := false
			for _, r := range results {
				if r.RuleID == "awesome-badge" && r.Severity == SeverityError {
					hasErr = true
					break
				}
			}
			if hasErr != tt.wantErr {
				t.Errorf("badgeRule.Check() hasErr = %v, wantErr = %v; results: %+v", hasErr, tt.wantErr, results)
			}
		})
	}
}

func TestListItemRule(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
	}{
		{
			name:     "valid list item",
			markdown: "- [Example](https://example.com) - Description.\n\n  More text.\n",
			wantErr:  false,
		},
		{
			name:     "missing link URL",
			markdown: "- [Broken]() - Description.\n\n  More text.\n",
			wantErr:  true,
		},
		{
			name:     "missing link text",
			markdown: "- [](https://example.com) - Description.\n\n  More text.\n",
			wantErr:  true,
		},
		{
			name:     "no dash separator",
			markdown: "- [Example](https://example.com) Description.\n\n  More text.\n",
			wantErr:  true,
		},
		{
			name:     "invalid casing in description",
			markdown: "- [Example](https://example.com) - description starts lowercase.\n\n  More text.\n",
			wantErr:  true,
		},
		{
			name:     "missing punctuation",
			markdown: "- [Example](https://example.com) - Description without period\n\n  More text.\n",
			wantErr:  true,
		},
		{
			name:     "ordered list skipped",
			markdown: "1. [Example](https://example.com) - Description.\n\n  More text.\n",
			wantErr:  false,
		},
		{
			name:     "ellipsis punctuation",
			markdown: "- [Example](https://example.com) - Description...\n\n  More text.\n",
			wantErr:  false,
		},
		{
			name:     "exclamation mark punctuation",
			markdown: "- [Example](https://example.com) - Description!\n\n  More text.\n",
			wantErr:  false,
		},
		{
			name:     "all caps acronym in description",
			markdown: "- [Example](https://example.com) - ACPI is supported.\n\n  More text.\n",
			wantErr:  false,
		},
	}

	rule := &listItemRule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseMarkdown([]byte(tt.markdown))
			results := rule.Check(doc, []byte(tt.markdown))
			hasErr := false
			for _, r := range results {
				if r.RuleID == "awesome-list-item" && r.Severity == SeverityError {
					hasErr = true
					break
				}
			}
			if hasErr != tt.wantErr {
				t.Errorf("listItemRule.Check() hasErr = %v, wantErr = %v; results: %+v", hasErr, tt.wantErr, results)
			}
		})
	}
}

func TestTOCRule(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
	}{
		{
			name:     "no TOC",
			markdown: "# Awesome List\n\n## Section 1\n\nContent.\n\n## Section 2\n\nContent.\n",
			wantErr:  false,
		},
		{
			name:     "TOC first section",
			markdown: "# Awesome List\n\n## Contents\n\n- [Section 1](#section-1)\n\n## Section 1\n\nContent.\n",
			wantErr:  false,
		},
		{
			name:     "TOC not first section",
			markdown: "# Awesome List\n\n## Intro\n\nSome intro.\n\n## Contents\n\n- [Section 1](#section-1)\n\n## Section 1\n\nContent.\n",
			wantErr:  true,
		},
		{
			name:     "TOC with no content after",
			markdown: "# Awesome List\n\n## Contents\n\nJust a TOC.\n",
			wantErr:  true,
		},
	}

	rule := &tocRule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseMarkdown([]byte(tt.markdown))
			results := rule.Check(doc, []byte(tt.markdown))
			hasErr := false
			for _, r := range results {
				if r.RuleID == "awesome-toc" && r.Severity == SeverityError {
					hasErr = true
					break
				}
			}
			if hasErr != tt.wantErr {
				t.Errorf("tocRule.Check() hasErr = %v, wantErr = %v; results: %+v", hasErr, tt.wantErr, results)
			}
		})
	}
}

func TestNoCiBadgeRule(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
	}{
		{
			name:     "CI badge in title",
			markdown: "![Build Status](https://travis-ci.org/example.svg)\n\nContent here.\n",
			wantErr:  true,
		},
		{
			name:     "CI badge in URL",
			markdown: "![Example](https://circleci.com/example.svg)\n\nContent here.\n",
			wantErr:  true,
		},
		{
			name:     "non-CI badge",
			markdown: "![Example](https://example.com/image.svg)\n\nContent here.\n",
			wantErr:  false,
		},
		{
			name:     "Travis in title",
			markdown: "![Build](https://example.com/badge.svg \"Travis Build Status\")\n\nContent here.\n",
			wantErr:  true,
		},
		{
			name:     "CircleCI in URL",
			markdown: "![Example](https://circleci.com/gh/example.svg)\n\nContent here.\n",
			wantErr:  true,
		},
	}

	rule := &noCiBadgeRule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseMarkdown([]byte(tt.markdown))
			results := rule.Check(doc, []byte(tt.markdown))
			hasErr := false
			for _, r := range results {
				if r.RuleID == "awesome-no-ci-badge" && r.Severity == SeverityError {
					hasErr = true
					break
				}
			}
			if hasErr != tt.wantErr {
				t.Errorf("noCiBadgeRule.Check() hasErr = %v, wantErr = %v; results: %+v", hasErr, tt.wantErr, results)
			}
		})
	}
}

func TestDoubleLinkRule(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
	}{
		{
			name:     "duplicate link",
			markdown: "- [Example](https://example.com)\n- [Example Again](https://example.com)\n",
			wantErr:  true,
		},
		{
			name:     "unique links",
			markdown: "- [Example](https://example.com)\n- [Other](https://other.com)\n",
			wantErr:  false,
		},
		{
			name:     "empty link skipped",
			markdown: "- [Empty]()\n- [Empty]()\n",
			wantErr:  false,
		},
	}

	rule := &doubleLinkRule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseMarkdown([]byte(tt.markdown))
			results := rule.Check(doc, []byte(tt.markdown))
			hasErr := false
			for _, r := range results {
				if r.RuleID == "double-link" && r.Severity == SeverityError {
					hasErr = true
					break
				}
			}
			if hasErr != tt.wantErr {
				t.Errorf("doubleLinkRule.Check() hasErr = %v, wantErr = %v; results: %+v", hasErr, tt.wantErr, results)
			}
		})
	}
}

func TestContributingRule(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantWarn bool
	}{
		{
			name:     "contributing section found",
			markdown: "# Awesome List\n\n## Contributing\n\nHow to contribute.\n",
			wantWarn: true,
		},
		{
			name:     "contributing guidelines found",
			markdown: "# Awesome List\n\n## Contributing Guidelines\n\nHow to contribute.\n",
			wantWarn: true,
		},
		{
			name:     "no contributing section",
			markdown: "# Awesome List\n\n## Section 1\n\nContent.\n",
			wantWarn: false,
		},
	}

	rule := &contributingRule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseMarkdown([]byte(tt.markdown))
			results := rule.Check(doc, []byte(tt.markdown))
			hasWarn := false
			for _, r := range results {
				if r.RuleID == "awesome-contributing" && r.Severity == SeverityWarning {
					hasWarn = true
					break
				}
			}
			if hasWarn != tt.wantWarn {
				t.Errorf("contributingRule.Check() hasWarn = %v, wantWarn = %v; results: %+v", hasWarn, tt.wantWarn, results)
			}
		})
	}
}

func TestLicenseRule(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
	}{
		{
			name:     "license section found",
			markdown: "# Awesome List\n\n## License\n\nMIT License.\n",
			wantErr:  true,
		},
		{
			name:     "licence section found",
			markdown: "# Awesome List\n\n## Licence\n\nMIT Licence.\n",
			wantErr:  true,
		},
		{
			name:     "no license section",
			markdown: "# Awesome List\n\n## Section 1\n\nContent.\n",
			wantErr:  false,
		},
	}

	rule := &licenseRule{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseMarkdown([]byte(tt.markdown))
			results := rule.Check(doc, []byte(tt.markdown))
			hasErr := false
			for _, r := range results {
				if r.RuleID == "awesome-license" && r.Severity == SeverityError {
					hasErr = true
					break
				}
			}
			if hasErr != tt.wantErr {
				t.Errorf("licenseRule.Check() hasErr = %v, wantErr = %v; results: %+v", hasErr, tt.wantErr, results)
			}
		})
	}
}

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"hello", "Hello"},
		{"hello world", "Hello World"},
		{"the quick brown fox", "The Quick Brown Fox"},
		{"a tale of two cities", "A Tale of Two Cities"},
		{"hello and goodbye", "Hello and Goodbye"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := toTitleCase(tt.input); got != tt.want {
				t.Errorf("toTitleCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidCasing(t *testing.T) {
	tests := []struct {
		word string
		want bool
	}{
		{"Hello", true},
		{"hello", false},
		{"ACPI", true},
		{"HTML", true},
		{"", false},
		{"123", true},
		{"JSON", true},
		{"eBPF", true},
		{"macOS", true},
	}
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			if got := isValidCasing(tt.word); got != tt.want {
				t.Errorf("isValidCasing(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

func TestIsMinorWord(t *testing.T) {
	tests := []struct {
		word string
		want bool
	}{
		{"the", true},
		{"and", true},
		{"of", true},
		{"hello", false},
		{"World", false},
		{"THE", true},
	}
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			if got := isMinorWord(tt.word); got != tt.want {
				t.Errorf("isMinorWord(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"hello", "Hello"},
		{"h", "H"},
		{"hello world", "Hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := capitalizeFirst(tt.input); got != tt.want {
				t.Errorf("capitalizeFirst(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsCaseAllowListed(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"title", true},
		{"capital", true},
		{"hello", false},
		{"TITLE", true},
		{"Capital", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isCaseAllowListed(tt.input); got != tt.want {
				t.Errorf("isCaseAllowListed(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidBadgeURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://awesome.re", true},
		{"https://github.com/sindresorhus/awesome", true},
		{"https://github.com/sindresorhus/awesome#readme", true},
		{"https://example.com", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := isValidBadgeURL(tt.url); got != tt.want {
				t.Errorf("isValidBadgeURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestIsValidBadgeSourceURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://awesome.re/badge.svg", true},
		{"https://awesome.re/badge-flat.svg", true},
		{"https://awesome.re/badge-flat2.svg", true},
		{"https://evil.com/badge.svg", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := isValidBadgeSourceURL(tt.url); got != tt.want {
				t.Errorf("isValidBadgeSourceURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestResults(t *testing.T) {
	t.Run("HasErrors", func(t *testing.T) {
		r := &Results{Results: []Result{
			{RuleID: "test", Severity: SeverityWarning},
			{RuleID: "test2", Severity: SeverityError},
		}}
		if !r.HasErrors() {
			t.Error("HasErrors() = false, want true")
		}

		r2 := &Results{Results: []Result{
			{RuleID: "test", Severity: SeverityWarning},
		}}
		if r2.HasErrors() {
			t.Error("HasErrors() = true, want false")
		}
	})
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		s    Severity
		want string
	}{
		{SeverityError, "error"},
		{SeverityWarning, "warning"},
		{Severity(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("Severity.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLinter(t *testing.T) {
	t.Run("LintFile", func(t *testing.T) {
		l := New()
		results, err := l.LintFile("testdata/valid.md")
		if err != nil {
			t.Fatalf("LintFile() error = %v", err)
		}
		if results == nil {
			t.Fatal("LintFile() results = nil")
		}
	})

	t.Run("Lint with GitHub URL", func(t *testing.T) {
		l := New()
		_, err := l.Lint("https://github.com/user/repo")
		if err == nil {
			t.Error("Lint() expected error for GitHub URL")
		}
	})

	t.Run("Lint with file path", func(t *testing.T) {
		l := New()
		results, err := l.Lint("testdata/valid.md")
		if err != nil {
			t.Fatalf("Lint() error = %v", err)
		}
		if results == nil {
			t.Fatal("Lint() results = nil")
		}
	})

	t.Run("New creates linter with rules", func(t *testing.T) {
		l := New()
		if len(l.rules) == 0 {
			t.Error("New() created linter with no rules")
		}
	})
}

func TestWritePretty(t *testing.T) {
	t.Run("no issues", func(t *testing.T) {
		r := &Results{FilePath: "test.md"}
		var buf strings.Builder
		r.WritePretty(&buf)
		if !strings.Contains(buf.String(), "No issues found") {
			t.Errorf("unexpected output: %s", buf.String())
		}
	})

	t.Run("with issues", func(t *testing.T) {
		r := &Results{
			FilePath: "test.md",
			Results: []Result{
				{RuleID: "test-rule", Severity: SeverityError, Message: "error message", Line: 1, Column: 5},
				{RuleID: "test-rule2", Severity: SeverityWarning, Message: "warning message", Line: 2, Column: 3},
			},
		}
		var buf strings.Builder
		r.WritePretty(&buf)
		output := buf.String()
		if !strings.Contains(output, "1 errors, 1 warnings") {
			t.Errorf("unexpected output: %s", output)
		}
	})
}

func TestWriteJSON(t *testing.T) {
	r := &Results{
		FilePath: "test.md",
		Results: []Result{
			{RuleID: "test", Severity: SeverityError, Message: "test", Line: 1, Column: 1},
		},
	}
	var buf strings.Builder
	err := r.WriteJSON(&buf)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	if !strings.Contains(buf.String(), "test.md") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}
