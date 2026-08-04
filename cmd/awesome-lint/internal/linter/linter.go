package linter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Severity represents the severity level of a lint result.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	default:
		return "unknown"
	}
}

// Result represents a single lint finding.
type Result struct {
	RuleID   string   `json:"ruleId"`
	Severity Severity `json:"severity"`
	Line     int      `json:"line"`
	Column   int      `json:"column"`
	Message  string   `json:"message"`
}

// Results holds all lint results for a file.
type Results struct {
	FilePath string   `json:"filePath"`
	Results  []Result `json:"results"`
}

// HasErrors returns true if any result has error severity.
func (r *Results) HasErrors() bool {
	for _, result := range r.Results {
		if result.Severity == SeverityError {
			return true
		}
	}
	return false
}

// WritePretty writes results in a human-readable format.
func (r *Results) WritePretty(w io.Writer) {
	if len(r.Results) == 0 {
		fmt.Fprintf(w, "✔ No issues found in %s\n", r.FilePath)
		return
	}

	fmt.Fprintf(w, "%s:\n", r.FilePath)

	// Sort by line, then column
	sort.Slice(r.Results, func(i, j int) bool {
		if r.Results[i].Line != r.Results[j].Line {
			return r.Results[i].Line < r.Results[j].Line
		}
		return r.Results[i].Column < r.Results[j].Column
	})

	var errCount, warnCount int
	for _, res := range r.Results {
		mark := "✖"
		if res.Severity == SeverityWarning {
			mark = "⚠"
			warnCount++
		} else {
			errCount++
		}
		fmt.Fprintf(w, "  %s  %d:%d  %s  %s\n", mark, res.Line, res.Column, res.Message, res.RuleID)
	}

	fmt.Fprintf(w, "\n%d errors, %d warnings\n", errCount, warnCount)
}

// WriteJSON writes results as JSON.
func (r *Results) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// Linter is the main linting engine.
type Linter struct {
	rules []Rule
}

// New creates a new Linter with all default rules.
func New() *Linter {
	return &Linter{
		rules: defaultRules(),
	}
}

// LintFile lints a single markdown file.
func (l *Linter) LintFile(path string) (*Results, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	results := &Results{
		FilePath: filepath.Base(path),
	}

	// Check filename for irregular characters
	base := filepath.Base(path)
	if strings.Contains(base, "_") {
		results.Results = append(results.Results, Result{
			RuleID:   "no-file-name-irregular-characters",
			Severity: SeverityError,
			Line:     1,
			Column:   1,
			Message:  "Unexpected character _ in file name",
		})
	}

	// Parse markdown
	doc := parseMarkdown(source)

	// Run each rule
	for _, rule := range l.rules {
		ruleResults := rule.Check(doc, source)
		results.Results = append(results.Results, ruleResults...)
	}

	return results, nil
}

// Lint lints a file by path or GitHub URL.
func (l *Linter) Lint(path string) (*Results, error) {
	// Check if it's a GitHub URL
	if strings.HasPrefix(path, "https://github.com/") || strings.HasPrefix(path, "http://github.com/") {
		return l.lintGitHubRepo(path)
	}

	return l.LintFile(path)
}

func (l *Linter) lintGitHubRepo(_ string) (*Results, error) {
	return nil, fmt.Errorf("GitHub repository linting is not yet supported; please lint a local file")
}

// Rule defines a single lint rule.
type Rule interface {
	// ID returns the rule identifier.
	ID() string
	// Check runs the rule against the parsed markdown document.
	Check(doc *MarkdownDoc, source []byte) []Result
}
