package linter

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/ast"
)

// defaultRules returns all default lint rules.
func defaultRules() []Rule {
	return []Rule{
		&headingRule{},
		&badgeRule{},
		&listItemRule{},
		&contributingRule{},
		&licenseRule{},
		&noCiBadgeRule{},
		&doubleLinkRule{},
		&tocRule{},
		&spellCheckRule{},
		&definitionCaseRule{},
	}
}

// walkHelper is a helper to make ast.Walk closures cleaner.
// It returns a Walker that ignores the error return.
func walkHelper(fn func(n ast.Node, entering bool) ast.WalkStatus) ast.Walker {
	return func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		return fn(n, entering), nil
	}
}

// headingRule validates the main heading of the awesome list.
type headingRule struct{}

func (r *headingRule) ID() string { return "awesome-heading" }

func (r *headingRule) Check(doc *MarkdownDoc, source []byte) []Result {
	var results []Result
	headings := 0

	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if n.Kind() == ast.KindHTMLBlock || n.Kind() == ast.KindRawHTML {
			html := string(n.Lines().Value(source))
			htmlLower := strings.ToLower(html)
			if strings.Contains(htmlLower, "<h1") || strings.Contains(htmlLower, "<h2") ||
				strings.Contains(htmlLower, "<h3") || strings.Contains(htmlLower, "<h4") ||
				strings.Contains(htmlLower, "<h5") || strings.Contains(htmlLower, "<h6") {
				headings++
				return ast.WalkSkipChildren
			}
			if (strings.Contains(htmlLower, "align=\"center\"") || strings.Contains(htmlLower, "align='center'")) &&
				strings.Contains(htmlLower, "<img") {
				headings++
				return ast.WalkSkipChildren
			}
			return ast.WalkContinue
		}

		if n.Kind() != ast.KindHeading {
			return ast.WalkContinue
		}

		head := n.(*ast.Heading)
		headings++

		if headings == 1 && head.Level > 1 {
			line, col := doc.LineColOf(n)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Main list heading must be of depth 1",
				Line:     line,
				Column:   col,
			})
		}

		if headings == 1 {
			text := doc.TextOf(n)

			expected := toTitleCase(text)
			if text != "" && text != expected && !isCaseAllowListed(text) {
				line, col := doc.LineColOf(n)
				results = append(results, Result{
					RuleID:   r.ID(),
					Severity: SeverityError,
					Message:  "Main heading must be in title case",
					Line:     line,
					Column:   col,
				})
			}
		}

		if headings > 1 && head.Level == 1 {
			line, col := doc.LineColOf(n)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "List can only have one heading",
				Line:     line,
				Column:   col,
			})
		}

		return ast.WalkSkipChildren
	}))

	if headings == 0 {
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Missing main list heading",
			Line:     1,
			Column:   1,
		})
	}

	return results
}

// badgeRule validates the presence of the Awesome badge.
type badgeRule struct{}

func (r *badgeRule) ID() string { return "awesome-badge" }

func (r *badgeRule) Check(doc *MarkdownDoc, source []byte) []Result {
	var results []Result
	firstHeading := true

	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindHeading {
			return ast.WalkContinue
		}

		if !firstHeading {
			return ast.WalkContinue
		}
		firstHeading = false

		head := n.(*ast.Heading)
		if head.Level != 1 {
			return ast.WalkContinue
		}

		hasBadge := r.checkNodeForBadge(n, doc, source, &results)
		if !hasBadge {
			if sibling := n.NextSibling(); sibling != nil && sibling.Kind() != ast.KindHeading {
				hasBadge = r.checkNodeForBadge(sibling, doc, source, &results)
			}
		}

		if !hasBadge {
			line, col := doc.LineColOf(n)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Missing Awesome badge next to the main heading",
				Line:     line,
				Column:   col,
			})
		}

		return ast.WalkContinue
	}))

	return results
}

// checkNodeForBadge scans a node's children for a link containing an Awesome badge image.
func (r *badgeRule) checkNodeForBadge(n ast.Node, doc *MarkdownDoc, source []byte, results *[]Result) bool {
	hasBadge := false
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() != ast.KindLink {
			continue
		}
		link := c.(*ast.Link)
		url := string(link.Destination)

		if !isValidBadgeURL(url) {
			continue
		}

		for cc := c.FirstChild(); cc != nil; cc = cc.NextSibling() {
			if cc.Kind() == ast.KindImage {
				img := cc.(*ast.Image)
				imgURL := string(img.Destination)
				if !isValidBadgeSourceURL(imgURL) {
					line, col := doc.LineColOf(cc)
					*results = append(*results, Result{
						RuleID:   r.ID(),
						Severity: SeverityError,
						Message:  "Invalid badge source",
						Line:     line,
						Column:   col,
					})
					return hasBadge
				}
				hasBadge = true
			}
		}
	}
	return hasBadge
}

func isValidBadgeURL(url string) bool {
	return url == "https://awesome.re" ||
		url == "https://github.com/sindresorhus/awesome" ||
		url == "https://github.com/sindresorhus/awesome#readme"
}

func isValidBadgeSourceURL(url string) bool {
	return url == "https://awesome.re/badge.svg" ||
		url == "https://awesome.re/badge-flat.svg" ||
		url == "https://awesome.re/badge-flat2.svg"
}

// listItemRule validates list item structure.
type listItemRule struct{}

func (r *listItemRule) ID() string { return "awesome-list-item" }

func (r *listItemRule) Check(doc *MarkdownDoc, source []byte) []Result {
	var results []Result

	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindList {
			return ast.WalkContinue
		}

		list := n.(*ast.List)
		if list.IsOrdered() {
			return ast.WalkContinue
		}

		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() != ast.KindListItem {
				continue
			}
			results = append(results, r.validateListItem(c, doc)...)
		}

		return ast.WalkContinue
	}))

	return results
}

func (r *listItemRule) validateListItem(n ast.Node, doc *MarkdownDoc) []Result {
	var results []Result

	var para ast.Node
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindParagraph || c.Kind() == ast.KindTextBlock {
			para = c
			break
		}
	}
	if para == nil {
		return results
	}

	var linkNode ast.Node
	var descriptionNodes []ast.Node
	foundDash := false
	hasContentAfterLink := false

	enDash := "\u2013"
	emDash := "\u2014"

	for c := para.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindLink && !foundDash {
			linkNode = c
			continue
		}

		if linkNode != nil {
			hasContentAfterLink = true
		}

		if c.Kind() == ast.KindText {
			text := string(c.(*ast.Text).Segment.Value(doc.Source))
			if strings.HasPrefix(text, " - ") && linkNode != nil {
				foundDash = true
				descriptionNodes = append(descriptionNodes, c)
				continue
			}
			if (strings.HasPrefix(text, " "+enDash+" ") || strings.HasPrefix(text, " "+emDash+" ")) && linkNode != nil {
				line, col := doc.LineColOf(c)
				results = append(results, Result{
					RuleID:   r.ID(),
					Severity: SeverityError,
					Message:  "List item link and description separated by invalid en-dash or em-dash",
					Line:     line,
					Column:   col,
				})
				return results
			}
		}

		if foundDash {
			descriptionNodes = append(descriptionNodes, c)
		}
	}

	if linkNode == nil {
		line, col := doc.LineColOf(n)
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Invalid list item link",
			Line:     line,
			Column:   col,
		})
		return results
	}

	link := linkNode.(*ast.Link)
	linkURL := string(link.Destination)
	linkText := doc.TextOf(linkNode)

	if linkURL == "" {
		line, col := doc.LineColOf(linkNode)
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Invalid list item link URL",
			Line:     line,
			Column:   col,
		})
		return results
	}

	if strings.HasPrefix(linkURL, "#") {
		line, col := doc.LineColOf(linkNode)
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Invalid list item link URL",
			Line:     line,
			Column:   col,
		})
		return results
	}

	if linkText == "" {
		line, col := doc.LineColOf(linkNode)
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Invalid list item link text",
			Line:     line,
			Column:   col,
		})
		return results
	}

	if !foundDash && hasContentAfterLink {
		line, col := doc.LineColOf(linkNode)
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "List item link and description must be separated with a dash",
			Line:     line,
			Column:   col,
		})
		return results
	}

	if len(descriptionNodes) == 0 {
		return results
	}

	descText := ""
	for _, dn := range descriptionNodes {
		descText += doc.TextOf(dn)
	}

	descText = strings.TrimSpace(descText)
	if descText == "" {
		return results
	}

	descContent := strings.TrimPrefix(descText, "- ")
	descContent = strings.TrimSpace(descContent)

	if descContent != "" {
		firstWord := strings.Fields(descContent)[0]
		if !isValidCasing(firstWord) {
			line, col := doc.LineColOf(descriptionNodes[0])
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "List item description must start with valid casing",
				Line:     line,
				Column:   col,
			})
		}
	}

	lastChar := descContent[len(descContent)-1]
	if lastChar != '.' && lastChar != '!' && lastChar != '?' {
		if !strings.HasSuffix(descContent, "...") {
			line, col := doc.LineColOf(descriptionNodes[len(descriptionNodes)-1])
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "List item description must end with proper punctuation",
				Line:     line,
				Column:   col,
			})
		}
	}

	return results
}

// contributingRule validates the presence of a contributing.md file.
type contributingRule struct{}

func (r *contributingRule) ID() string { return "awesome-contributing" }

func (r *contributingRule) Check(doc *MarkdownDoc, source []byte) []Result {
	_ = source
	var results []Result

	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindHeading {
			return ast.WalkContinue
		}

		text := strings.ToLower(doc.TextOf(n))
		if text == "contributing" || text == "contributing guidelines" {
			line, col := doc.LineColOf(n)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityWarning,
				Message:  "Contributing section found in readme; should be in contributing.md",
				Line:     line,
				Column:   col,
			})
		}

		return ast.WalkContinue
	}))

	return results
}

// licenseRule ensures no license section exists in the readme.
type licenseRule struct{}

func (r *licenseRule) ID() string { return "awesome-license" }

func (r *licenseRule) Check(doc *MarkdownDoc, source []byte) []Result {
	_ = source
	var results []Result

	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindHeading {
			return ast.WalkContinue
		}

		text := strings.ToLower(doc.TextOf(n))
		if text == "license" || text == "licence" {
			line, col := doc.LineColOf(n)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Forbidden license section found",
				Line:     line,
				Column:   col,
			})
		}

		return ast.WalkContinue
	}))

	return results
}

// noCiBadgeRule ensures no CI badges are present.
type noCiBadgeRule struct{}

func (r *noCiBadgeRule) ID() string { return "awesome-no-ci-badge" }

func (r *noCiBadgeRule) Check(doc *MarkdownDoc, source []byte) []Result {
	_ = source
	var results []Result

	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindImage {
			return ast.WalkContinue
		}

		img := n.(*ast.Image)
		title := string(img.Title)
		url := string(img.Destination)

		if strings.Contains(strings.ToLower(title), "build status") ||
			strings.Contains(strings.ToLower(title), "travis") ||
			strings.Contains(strings.ToLower(title), "circleci") {
			line, col := doc.LineColOf(n)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Readme must not contain CI badge",
				Line:     line,
				Column:   col,
			})
		} else if strings.Contains(strings.ToLower(url), "travis") ||
			strings.Contains(strings.ToLower(url), "circleci") {
			line, col := doc.LineColOf(n)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Readme must not contain CI badge",
				Line:     line,
				Column:   col,
			})
		}

		return ast.WalkContinue
	}))

	return results
}

// doubleLinkRule detects duplicate links.
type doubleLinkRule struct{}

func (r *doubleLinkRule) ID() string { return "double-link" }

func (r *doubleLinkRule) Check(doc *MarkdownDoc, source []byte) []Result {
	var results []Result
	linkMap := make(map[string][]ast.Node)

	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindLink {
			return ast.WalkContinue
		}

		link := n.(*ast.Link)
		url := string(link.Destination)
		if url == "" {
			return ast.WalkContinue
		}

		// Skip links that appear inside a list item description (after the " - " separator)
		if isLinkInDescription(n, source) {
			return ast.WalkContinue
		}

		normalized := normalizeURL(url)
		linkMap[normalized] = append(linkMap[normalized], n)
		return ast.WalkContinue
	}))

	for _, nodes := range linkMap {
		if len(nodes) <= 1 {
			continue
		}
		for _, n := range nodes {
			line, col := doc.LineColOf(n)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  fmt.Sprintf("Duplicate link: %s", string(n.(*ast.Link).Destination)),
				Line:     line,
				Column:   col,
			})
		}
	}

	return results
}

// isLinkInDescription checks if a link node is inside a list item description
// (i.e., after the " - " separator).
func isLinkInDescription(n ast.Node, source []byte) bool {
	// Find the inline container (Paragraph or TextBlock) that holds this link
	parent := n.Parent()
	if parent.Kind() != ast.KindParagraph && parent.Kind() != ast.KindTextBlock {
		return false
	}

	// Check if this container is inside a list item
	listItem := parent.Parent()
	if listItem == nil || listItem.Kind() != ast.KindListItem {
		return false
	}

	// Walk siblings before this link to find a " - " separator
	for c := parent.FirstChild(); c != nil; c = c.NextSibling() {
		if c == n {
			break
		}
		if c.Kind() == ast.KindText {
			text := string(c.(*ast.Text).Segment.Value(source))
			if strings.Contains(text, " - ") {
				return true
			}
		}
	}

	return false
}

// normalizeURL normalizes a URL for duplicate comparison.
// It strips protocol, trailing slashes, and common index files.
func normalizeURL(url string) string {
	// Strip protocol
	if strings.HasPrefix(url, "https://") {
		url = url[len("https://"):]
	} else if strings.HasPrefix(url, "http://") {
		url = url[len("http://"):]
	}

	// Strip trailing slash
	url = strings.TrimSuffix(url, "/")

	// Strip common index files
	indexFiles := []string{"/index.html", "/index.htm", "/index.php", "/index.shtml"}
	for _, idx := range indexFiles {
		url = strings.TrimSuffix(url, idx)
	}

	return url
}

// tocRule validates the Table of Contents.
type tocRule struct{}

func (r *tocRule) ID() string { return "awesome-toc" }

func (r *tocRule) Check(doc *MarkdownDoc, source []byte) []Result {
	_ = source
	var results []Result

	var tocNode ast.Node
	beforeTOC := true
	headingsBefore := 0

	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}
		if n.Kind() == ast.KindHeading {
			h := n.(*ast.Heading)
			if h.Level == 2 {
				text := strings.TrimSpace(doc.TextOf(n))
				if strings.EqualFold(text, "contents") {
					tocNode = n
					beforeTOC = false
					return ast.WalkContinue
				}
			}
			if beforeTOC {
				headingsBefore++
			}
		}
		return ast.WalkContinue
	}))

	if tocNode == nil {
		return results
	}

	if headingsBefore > 1 {
		line, col := doc.LineColOf(tocNode)
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Table of Contents must be the first section",
			Line:     line,
			Column:   col,
		})
	}

	var headingsAfter []ast.Node
	afterTOC := false
	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}
		if n == tocNode {
			afterTOC = true
			return ast.WalkContinue
		}
		if !afterTOC {
			return ast.WalkContinue
		}
		if n.Kind() == ast.KindHeading {
			h := n.(*ast.Heading)
			if h.Level == 2 {
				headingsAfter = append(headingsAfter, n)
			}
		}
		return ast.WalkContinue
	}))

	if len(headingsAfter) == 0 {
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Missing content headers",
			Line:     1,
			Column:   1,
		})
	}

	return results
}

// spellCheckRule checks for common spelling mistakes.
type spellCheckRule struct{}

func (r *spellCheckRule) ID() string { return "awesome-spell-check" }

func (r *spellCheckRule) Check(doc *MarkdownDoc, source []byte) []Result {
	var results []Result
	spellings := map[string]string{
		"WASM": "WebAssembly",
	}
	for lineIdx, line := range strings.Split(string(source), "\n") {
		for wrong, correct := range spellings {
			idx := strings.Index(line, wrong)
			if idx >= 0 {
				results = append(results, Result{
					RuleID:   r.ID(),
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("Text %q should be written as %q (if referring to the technology)", wrong, correct),
					Line:     lineIdx + 1,
					Column:   idx + 1,
				})
			}
		}
	}
	return results
}

// definitionCaseRule checks that definition labels are lowercase.
type definitionCaseRule struct{}

func (r *definitionCaseRule) ID() string { return "definition-case" }

func (r *definitionCaseRule) Check(doc *MarkdownDoc, source []byte) []Result {
	var results []Result
	for lineIdx, line := range strings.Split(string(source), "\n") {
		// Match [label]: URL
		if strings.Contains(line, "]:") {
			start := strings.Index(line, "[")
			end := strings.Index(line, "]:")
			if start >= 0 && end > start+1 {
				label := line[start+1 : end]
				if label != strings.ToLower(label) {
					results = append(results, Result{
						RuleID:   r.ID(),
						Severity: SeverityError,
						Message:  "Unexpected uppercase characters in definition label, expected lowercase",
						Line:     lineIdx + 1,
						Column:   start + 1,
					})
				}
			}
		}
	}
	return results
}

func toTitleCase(s string) string {
	if s == "" {
		return s
	}

	words := strings.Fields(s)
	for i, w := range words {
		if i == 0 || i == len(words)-1 {
			words[i] = capitalizeFirst(w)
		} else if isMinorWord(w) {
			words[i] = strings.ToLower(w)
		} else {
			words[i] = capitalizeFirst(w)
		}
	}
	return strings.Join(words, " ")
}

func isMinorWord(w string) bool {
	minor := map[string]bool{
		"a": true, "an": true, "the": true,
		"and": true, "but": true, "or": true, "for": true, "nor": true,
		"on": true, "at": true, "to": true, "by": true, "from": true,
		"in": true, "of": true, "with": true, "as": true, "up": true,
		"is": true, "it": true, "vs": true,
	}
	return minor[strings.ToLower(w)]
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func isCaseAllowListed(s string) bool {
	allowListed := map[string]bool{
		"title": true, "capital": true,
	}
	return allowListed[strings.ToLower(s)]
}

func isValidCasing(word string) bool {
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

	hasUpper := false
	for _, r := range cleaned {
		if unicode.IsUpper(r) {
			hasUpper = true
			break
		}
	}
	if hasUpper {
		return true
	}

	return false
}
