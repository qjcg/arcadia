package linter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	return l.lintFile(path, false)
}

// LintFileWithFix lints a single markdown file and applies auto-fixes.
func (l *Linter) LintFileWithFix(path string) (*Results, error) {
	return l.lintFile(path, true)
}

func (l *Linter) lintFile(path string, fix bool) (*Results, error) {
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
	doc := parseMarkdownWithDir(source, filepath.Dir(path))

	// Run each rule
	for _, rule := range l.rules {
		ruleResults := rule.Check(doc, source)
		results.Results = append(results.Results, ruleResults...)
	}

	// Apply fixes if requested — run in a loop until stable
	if fix {
		for {
			anyFix := false
			for _, rule := range l.rules {
				if !rule.Fixable() {
					continue
				}
				// Collect results for this rule from current lint
				var ruleResults []Result
				for _, res := range results.Results {
					if res.RuleID == rule.ID() {
						ruleResults = append(ruleResults, res)
					}
				}
				if len(ruleResults) > 0 {
					newSource := rule.Fix(doc, source, ruleResults)
					if !bytes.Equal(newSource, source) {
						source = newSource
						anyFix = true
					}
				}
			}
			if !anyFix {
				break
			}
			// Re-lint after fixes to catch newly introduced errors
			if err := os.WriteFile(path, source, 0o644); err != nil {
				return nil, fmt.Errorf("writing fixes to %s: %w", path, err)
			}
			doc = parseMarkdownWithDir(source, filepath.Dir(path))
			results.Results = nil
			for _, rule := range l.rules {
				ruleResults := rule.Check(doc, source)
				results.Results = append(results.Results, ruleResults...)
			}
		}
	}

	return results, nil
}

// Lint lints a file by path or GitHub URL.
func (l *Linter) Lint(path string) (*Results, error) {
	return l.lint(path, false)
}

// LintWithFix lints a file by path and applies auto-fixes.
func (l *Linter) LintWithFix(path string) (*Results, error) {
	return l.lint(path, true)
}

func (l *Linter) lint(path string, fix bool) (*Results, error) {
	// Check if it's a GitHub URL
	if strings.HasPrefix(path, "https://github.com/") || strings.HasPrefix(path, "http://github.com/") {
		if fix {
			return nil, fmt.Errorf("fix mode is not supported for GitHub URLs; please lint a local file")
		}
		return l.lintGitHubRepo(path)
	}

	if fix {
		return l.LintFileWithFix(path)
	}
	return l.LintFile(path)
}

func (l *Linter) lintGitHubRepo(repoURL string) (*Results, error) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "awesome-lint-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repo with depth 1 for speed
	cmd := exec.Command("git", "clone", "--depth", "1", "--", repoURL, tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("cloning repository %s: %w\n%s", repoURL, err, string(output))
	}

	// Find the readme file
	var readmePath string
	readmeCandidates := []string{"README.md", "readme.md", "Readme.md"}
	for _, name := range readmeCandidates {
		candidate := filepath.Join(tmpDir, name)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			readmePath = candidate
			break
		}
	}

	if readmePath == "" {
		return nil, fmt.Errorf("unable to find a valid readme file in %s", repoURL)
	}

	return l.LintFile(readmePath)
}

// Rule defines a single lint rule.
type Rule interface {
	// ID returns the rule identifier.
	ID() string
	// Check runs the rule against the parsed markdown document.
	Check(doc *MarkdownDoc, source []byte) []Result
	// Fixable returns true if this rule can auto-fix its findings.
	Fixable() bool
	// Fix applies auto-fixes to the source bytes for the given results.
	// Returns the modified source.
	Fix(doc *MarkdownDoc, source []byte, results []Result) []byte
}
