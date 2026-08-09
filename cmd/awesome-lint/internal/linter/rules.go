package linter

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
		&noRepeatItemInDescriptionRule{},
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

func (r *headingRule) Fixable() bool { return true }

func (r *headingRule) Fix(doc *MarkdownDoc, source []byte, results []Result) []byte {
	_ = doc
	for _, res := range results {
		if res.RuleID != r.ID() {
			continue
		}
		if res.Message == "Main heading must be in title case" {
			// Find the heading line and apply title case
			lines := bytes.Split(source, []byte("\n"))
			if res.Line-1 < len(lines) {
				line := lines[res.Line-1]
				headingText := strings.TrimPrefix(string(line), "# ")
				titled := toTitleCase(headingText)
				if titled != headingText {
					newLine := "# " + titled
					lines[res.Line-1] = []byte(newLine)
					source = bytes.Join(lines, []byte("\n"))
				}
			}
		}
	}
	return source
}

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

func (r *badgeRule) Fixable() bool { return false }

func (r *badgeRule) Fix(_ *MarkdownDoc, source []byte, _ []Result) []byte { return source }

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

		hasBadge := r.checkNodeForBadge(n, doc, &results)
		if !hasBadge {
			if sibling := n.NextSibling(); sibling != nil && sibling.Kind() != ast.KindHeading {
				hasBadge = r.checkNodeForBadge(sibling, doc, &results)
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
func (r *badgeRule) checkNodeForBadge(n ast.Node, doc *MarkdownDoc, results *[]Result) bool {
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

func (r *listItemRule) Fixable() bool { return true }

func (r *listItemRule) Fix(doc *MarkdownDoc, source []byte, results []Result) []byte {
	_ = doc
	for _, res := range results {
		if res.RuleID != r.ID() {
			continue
		}
		lines := bytes.Split(source, []byte("\n"))
		if res.Line-1 >= len(lines) {
			continue
		}
		line := lines[res.Line-1]
		str := string(line)

		if res.Message == "List item description must start with valid casing" {
			// Find the " - " separator and capitalize the next word
			if idx := strings.Index(str, " - "); idx >= 0 {
				descStart := idx + 3
				rest := str[descStart:]
				fields := strings.Fields(rest)
				if len(fields) > 0 {
					first := fields[0]
					// Only auto-fix if the first word is all lowercase
					if strings.ToLower(first) == first && strings.ToUpper(first) != first {
						capitalized := capitalizeFirst(first)
						newLine := str[:descStart] + strings.Replace(rest, first, capitalized, 1)
						lines[res.Line-1] = []byte(newLine)
						source = bytes.Join(lines, []byte("\n"))
					}
				}
			}
		}

		if res.Message == "List item description must end with proper punctuation" {
			// Find the " - " separator and add proper end punctuation.
			if idx := strings.Index(str, " - "); idx >= 0 {
				descStart := idx + 3
				rest := str[descStart:]
				// Trim trailing whitespace
				rest = strings.TrimRight(rest, " \t")
				chinese := isChinese(doc)
				if rest != "" && !hasValidEndingSuffix(rest, chinese) {
					rest = fixEndingPunctuation(rest, chinese)
					newLine := str[:descStart] + rest
					lines[res.Line-1] = []byte(newLine)
					source = bytes.Join(lines, []byte("\n"))
				}
			}
		}
	}
	return source
}

func (r *listItemRule) Check(doc *MarkdownDoc, source []byte) []Result {
	var results []Result

	// Find lists to validate: skip the ToC list (the first list after Contents heading)
	lists := r.findListsToValidate(doc)

	for _, list := range lists {
		for c := list.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() != ast.KindListItem {
				continue
			}
			results = append(results, r.validateListItem(c, doc)...)
		}
	}

	return results
}

func (r *listItemRule) findListsToValidate(doc *MarkdownDoc) []ast.Node {
	// Find the Contents heading
	var tocNode ast.Node
	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindHeading {
			return ast.WalkContinue
		}
		h := n.(*ast.Heading)
		if h.Level == 2 {
			text := strings.TrimSpace(doc.TextOf(n))
			if strings.EqualFold(text, "contents") {
				tocNode = n
				return ast.WalkStop
			}
		}
		return ast.WalkContinue
	}))

	// If there's a ToC, find the first heading after it and validate lists after that
	var postTOCHeading ast.Node
	if tocNode != nil {
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
				postTOCHeading = n
				return ast.WalkStop
			}
			return ast.WalkContinue
		}))
	}

	// Collect all lists after the post-TOC heading (or all lists if no ToC)
	var lists []ast.Node
	started := tocNode == nil // if no ToC, start from beginning
	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}
		if n == tocNode {
			started = false
			return ast.WalkContinue
		}
		if n == postTOCHeading {
			started = true
			return ast.WalkContinue
		}
		if !started {
			return ast.WalkContinue
		}
		if n.Kind() == ast.KindList {
			list := n.(*ast.List)
			if !list.IsOrdered() {
				lists = append(lists, n)
			}
		}
		return ast.WalkContinue
	}))

	return lists
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

	// Skip items where the first child is text (not a link) — not a valid awesome list item
	firstChild := para.FirstChild()
	if firstChild != nil && firstChild.Kind() == ast.KindText {
		return results
	}

	linkNode, foundDash, hasContentAfterLink, descriptionNodes, enDashErr := r.findLinkAndDescription(para, doc)
	if enDashErr != nil {
		return enDashErr
	}

	results = append(results, r.validateLinkNode(linkNode, n, doc)...)
	if len(results) > 0 {
		return results
	}

	results = append(results, r.validateDescription(linkNode, descriptionNodes, foundDash, hasContentAfterLink, doc)...)
	return results
}

func (r *listItemRule) findLinkAndDescription(para ast.Node, doc *MarkdownDoc) (ast.Node, bool, bool, []ast.Node, []Result) {
	var linkNode ast.Node
	var descriptionNodes []ast.Node
	foundDash := false
	hasContentAfterLink := false

	enDash := "\u2013"
	emDash := "\u2014"

	// First pass: try to find a link followed by " - " separator
	var linkIndex int
	children := make([]ast.Node, 0)
	for c := para.FirstChild(); c != nil; c = c.NextSibling() {
		children = append(children, c)
	}

	for i, c := range children {
		isLink := c.Kind() == ast.KindLink
		if isLink && i < len(children)-1 {
			next := children[i+1]
			if next.Kind() == ast.KindText {
				text := string(next.(*ast.Text).Segment.Value(doc.Source))
				if strings.HasPrefix(text, " - ") {
					linkNode = c
					linkIndex = i
					foundDash = true
					descriptionNodes = append(descriptionNodes, next)
					for j := i + 2; j < len(children); j++ {
						descriptionNodes = append(descriptionNodes, children[j])
					}
					hasContentAfterLink = true
					return linkNode, foundDash, hasContentAfterLink, descriptionNodes, nil
				}
			}
		}
	}

	// Second pass: find first link-like node (fallback)
	for i, c := range children {
		isLink := c.Kind() == ast.KindLink
		if isLink && !foundDash {
			linkNode = c
			linkIndex = i
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
				for j := i + 1; j < len(children); j++ {
					descriptionNodes = append(descriptionNodes, children[j])
				}
				return linkNode, foundDash, hasContentAfterLink, descriptionNodes, nil
			}
			if (strings.HasPrefix(text, " "+enDash+" ") || strings.HasPrefix(text, " "+emDash+" ")) && linkNode != nil {
				line, col := doc.LineColOf(c)
				return linkNode, foundDash, hasContentAfterLink, descriptionNodes, []Result{{
					RuleID:   r.ID(),
					Severity: SeverityError,
					Message:  "List item link and description separated by invalid en-dash or em-dash",
					Line:     line,
					Column:   col,
				}}
			}
		}

		if foundDash {
			descriptionNodes = append(descriptionNodes, c)
		}
	}

	_ = linkIndex
	return linkNode, foundDash, hasContentAfterLink, descriptionNodes, nil
}

func (r *listItemRule) validateLinkNode(linkNode ast.Node, n ast.Node, doc *MarkdownDoc) []Result {
	if linkNode == nil {
		line, col := doc.LineColOf(n)
		return []Result{{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Invalid list item link",
			Line:     line,
			Column:   col,
		}}
	}

	// Support linkReference nodes (e.g., [foo] style references)
	if linkNode.Kind() == ast.KindLink {
		link := linkNode.(*ast.Link)
		linkURL := string(link.Destination)
		linkText := doc.TextOf(linkNode)

		if linkURL == "" {
			line, col := doc.LineColOf(linkNode)
			return []Result{{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Invalid list item link URL",
				Line:     line,
				Column:   col,
			}}
		}

		if strings.HasPrefix(linkURL, "#") {
			line, col := doc.LineColOf(linkNode)
			return []Result{{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Invalid list item link URL",
				Line:     line,
				Column:   col,
			}}
		}

		if linkText == "" {
			line, col := doc.LineColOf(linkNode)
			return []Result{{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Invalid list item link text",
				Line:     line,
				Column:   col,
			}}
		}
	}

	return nil
}

func (r *listItemRule) validateDescription(linkNode ast.Node, descriptionNodes []ast.Node, foundDash, hasContentAfterLink bool, doc *MarkdownDoc) []Result {
	var results []Result

	if !foundDash && hasContentAfterLink {
		line, col := doc.LineColOf(linkNode)
		return []Result{{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "List item link and description must be separated with a dash",
			Line:     line,
			Column:   col,
		}}
	}

	if len(descriptionNodes) == 0 {
		return nil
	}

	descText := ""
	for _, dn := range descriptionNodes {
		descText += doc.TextOf(dn)
	}

	descText = strings.TrimSpace(descText)
	if descText == "" {
		return nil
	}

	// Check for special cases: emoji-only or parenthetical-only descriptions
	if r.isSpecialCaseDescription(descText) {
		return nil
	}

	// Validate description node types
	for _, dn := range descriptionNodes {
		if !r.isValidDescriptionNodeType(dn) {
			line, col := doc.LineColOf(dn)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "List item description contains invalid markdown",
				Line:     line,
				Column:   col,
			})
			return results
		}
	}

	descContent := strings.TrimPrefix(descText, "- ")
	descContent = strings.TrimSpace(descContent)

	if descContent != "" {
		firstWord := strings.Fields(descContent)[0]
		if !isValidCasing(firstWord) && !r.startsWithSymbolPrefix(firstWord) {
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

	// Validate suffix node type
	lastNode := descriptionNodes[len(descriptionNodes)-1]
	if !r.isValidSuffixNodeType(lastNode) {
		line, col := doc.LineColOf(lastNode)
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "List item description must end with proper punctuation",
			Line:     line,
			Column:   col,
		})
		return results
	}

	if !hasValidEndingSuffix(descContent, isChinese(doc)) && !strings.HasSuffix(descContent, "...") {
		line, col := doc.LineColOf(descriptionNodes[len(descriptionNodes)-1])
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "List item description must end with proper punctuation",
			Line:     line,
			Column:   col,
		})
	}

	return results
}

func (r *listItemRule) isSpecialCaseDescription(text string) bool {
	// Emoji-only description: contains only emoji and spaces
	emojiOnly := regexp.MustCompile(`^[\p{S}\p{P}\s]+$`)
	if emojiOnly.MatchString(strings.TrimSpace(text)) {
		// Check if there's at least one symbol character (emoji)
		hasSymbol := false
		for _, r := range text {
			if unicode.Is(unicode.S, r) {
				hasSymbol = true
				break
			}
		}
		if hasSymbol {
			return true
		}
	}

	// Parenthetical-only description
	stripped := strings.TrimSpace(text)
	stripped = strings.TrimPrefix(stripped, "- ")
	stripped = strings.TrimSpace(stripped)

	if strings.HasPrefix(stripped, "(") && strings.HasSuffix(stripped, ")") {
		return true
	}

	return false
}

func (r *listItemRule) startsWithSymbolPrefix(word string) bool {
	if word == "" {
		return false
	}
	first := []rune(word)[0]
	// Check for common symbol prefixes
	if strings.ContainsAny(string(first), "/@#$~&%") {
		return true
	}
	// Check for emoji/symbol characters (Unicode Symbol categories)
	if unicode.Is(unicode.S, first) {
		return true
	}
	return false
}

// ideographicFullStop is the Chinese full stop (U+3002).
const ideographicFullStop = '。'

// isChinese reports whether the document contains Han (Chinese) characters.
func isChinese(doc *MarkdownDoc) bool {
	for _, r := range string(doc.Source) {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// isValidEnding reports whether r is acceptable end punctuation.
// For Chinese documents the ideographic full stop (。) and other full-width
// punctuation (！？…) are required instead of the ASCII forms (.!?).
func isValidEnding(r rune, chinese bool) bool {
	if chinese {
		return r == ideographicFullStop || r == '！' || r == '？' || r == '…'
	}
	return r == '.' || r == '!' || r == '?' || r == '…'
}

// hasValidEndingSuffix reports whether s ends with acceptable punctuation.
func hasValidEndingSuffix(s string, chinese bool) bool {
	runes := []rune(s)
	if len(runes) == 0 {
		return false
	}
	return isValidEnding(runes[len(runes)-1], chinese)
}

// fixEndingPunctuation returns s with proper end punctuation appended (or the
// trailing ASCII punctuation replaced) using the ideographic full stop (。)
// for Chinese documents and '.' for other languages.
func fixEndingPunctuation(s string, chinese bool) string {
	if chinese {
		runes := []rune(s)
		if len(runes) > 0 {
			switch runes[len(runes)-1] {
			case '.':
				runes[len(runes)-1] = ideographicFullStop
				return string(runes)
			case '!':
				runes[len(runes)-1] = '！'
				return string(runes)
			case '?':
				runes[len(runes)-1] = '？'
				return string(runes)
			}
		}
		return s + string(ideographicFullStop)
	}
	return s + "."
}

func (r *listItemRule) isValidDescriptionNodeType(n ast.Node) bool {
	switch n.Kind() {
	case ast.KindText,
		ast.KindEmphasis,
		ast.KindCodeSpan,
		ast.KindLink,
		ast.KindImage,
		ast.KindString:
		return true
	default:
		return false
	}
}

func (r *listItemRule) isValidSuffixNodeType(n ast.Node) bool {
	switch n.Kind() {
	case ast.KindText,
		ast.KindEmphasis,
		ast.KindCodeSpan,
		ast.KindLink,
		ast.KindImage,
		ast.KindString:
		return true
	default:
		return false
	}
}

// contributingRule checks that contributing.md exists and is not empty.
type contributingRule struct{}

func (r *contributingRule) ID() string { return "awesome-contributing" }

func (r *contributingRule) Fixable() bool { return false }

func (r *contributingRule) Fix(_ *MarkdownDoc, source []byte, _ []Result) []byte { return source }

func (r *contributingRule) Check(doc *MarkdownDoc, source []byte) []Result {
	_ = source
	var results []Result

	if doc.Dir == "" {
		return results
	}

	// Check for contributing.md in root or .github/ directory (case-insensitive)
	candidates := []struct {
		dir string
	}{
		{doc.Dir},
		{filepath.Join(doc.Dir, ".github")},
	}

	var foundFile string
	for _, c := range candidates {
		entries, err := os.ReadDir(c.dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.EqualFold(entry.Name(), "contributing.md") {
				foundFile = filepath.Join(c.dir, entry.Name())
				break
			}
		}
		if foundFile != "" {
			break
		}
	}

	if foundFile == "" {
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Missing file contributing.md",
			Line:     1,
			Column:   1,
		})
		return results
	}

	content, err := os.ReadFile(foundFile)
	if err == nil && len(strings.TrimSpace(string(content))) == 0 {
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "contributing.md file must not be empty",
			Line:     1,
			Column:   1,
		})
	}

	return results
}

// licenseRule ensures no license section exists in the readme.
type licenseRule struct{}

func (r *licenseRule) ID() string { return "awesome-license" }

func (r *licenseRule) Fixable() bool { return false }

func (r *licenseRule) Fix(_ *MarkdownDoc, source []byte, _ []Result) []byte { return source }

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

func (r *noCiBadgeRule) Fixable() bool { return false }

func (r *noCiBadgeRule) Fix(_ *MarkdownDoc, source []byte, _ []Result) []byte { return source }

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

func (r *doubleLinkRule) Fixable() bool { return false }

func (r *doubleLinkRule) Fix(_ *MarkdownDoc, source []byte, _ []Result) []byte { return source }

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

func (r *tocRule) Fixable() bool { return false }

func (r *tocRule) Fix(_ *MarkdownDoc, source []byte, _ []Result) []byte { return source }

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

	// Collect all headings after ToC
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
			if h.Level >= 2 {
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
		return results
	}

	// Build set of heading slugs for validation
	headingSlugs := make(map[string]ast.Node)
	headingsUsed := make(map[ast.Node]bool)
	for _, h := range headingsAfter {
		text := doc.TextOf(h)
		slug := gitHubSlug(text)
		headingSlugs[slug] = h
	}

	// Find the ToC list (first list after the Contents heading)
	var tocList ast.Node
	for c := tocNode.NextSibling(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindList {
			tocList = c
			break
		}
	}

	if tocList == nil {
		return results
	}

	// Validate ToC list items
	tocLinks := make(map[string]bool)
	ast.Walk(tocList, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindLink {
			return ast.WalkContinue
		}
		link := n.(*ast.Link)
		linkURL := string(link.Destination)
		linkText := strings.TrimSpace(doc.TextOf(n))

		// Check for duplicate ToC entries
		if tocLinks[linkURL] {
			line, col := doc.LineColOf(n)
			results = append(results, Result{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  fmt.Sprintf("Duplicate ToC entry: %s", linkText),
				Line:     line,
				Column:   col,
			})
		}
		tocLinks[linkURL] = true

		// Check if link is an anchor
		if strings.HasPrefix(linkURL, "#") {
			slug := linkURL[1:] // strip the #
			if _, exists := headingSlugs[slug]; !exists {
				line, col := doc.LineColOf(n)
				results = append(results, Result{
					RuleID:   r.ID(),
					Severity: SeverityError,
					Message:  fmt.Sprintf("ToC link points to non-existent heading: %s", linkText),
					Line:     line,
					Column:   col,
				})
			} else {
				// Mark heading as covered by ToC
				headingsUsed[headingSlugs[slug]] = true
			}
		}
		return ast.WalkContinue
	}))

	// Check for missing headings in ToC (only level 2 headings are required)
	deniedSections := map[string]bool{
		"contributing": true, "footnotes": true, "related lists": true,
		"related lists and projects": true, "license": true, "licence": true,
	}
	for _, h := range headingsAfter {
		heading := h.(*ast.Heading)
		if heading.Level != 2 {
			continue
		}
		if headingsUsed[h] {
			continue
		}
		text := strings.ToLower(strings.TrimSpace(doc.TextOf(h)))
		if deniedSections[text] {
			continue
		}
		line, col := doc.LineColOf(h)
		results = append(results, Result{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  fmt.Sprintf("Heading not covered by ToC: %s", doc.TextOf(h)),
			Line:     line,
			Column:   col,
		})
	}

	return results
}

// gitHubSlug generates a GitHub-style slug from heading text.
func gitHubSlug(text string) string {
	var result strings.Builder
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == ' ' {
			if r == ' ' {
				result.WriteRune('-')
			} else {
				result.WriteRune(r)
			}
		}
	}
	slug := strings.TrimRight(result.String(), "-")
	return slug
}

// spellCheckEntry defines a single spell-check rule with a compiled regex pattern.
type spellCheckEntry struct {
	pattern *regexp.Regexp
	correct string
}

// spellCheckRules returns all spell-check rules compiled from the upstream awesome-lint.
func spellCheckRules() []spellCheckEntry {
	return []spellCheckEntry{
		{regexp.MustCompile(`(?i)\bnode\.?js\b`), "Node.js"},
		{regexp.MustCompile(`(?i)\bjavascript\b`), "JavaScript"},
		{regexp.MustCompile(`(?i)\btypescript\b`), "TypeScript"},
		{regexp.MustCompile(`(?i)\bpython\b`), "Python"},
		{regexp.MustCompile(`(?i)\bc\+\+\b`), "C++"},
		{regexp.MustCompile(`(?i)\bc#\b`), "C#"},
		{regexp.MustCompile(`(?i)\bphp\b`), "PHP"},
		{regexp.MustCompile(`(?i)\bruby\b`), "Ruby"},
		{regexp.MustCompile(`(?i)\brust\b`), "Rust"},
		{regexp.MustCompile(`(?i)\bswift\b`), "Swift"},
		{regexp.MustCompile(`(?i)\bkotlin\b`), "Kotlin"},
		{regexp.MustCompile(`(?i)\bscala\b`), "Scala"},
		{regexp.MustCompile(`(?i)\bclojure\b`), "Clojure"},
		{regexp.MustCompile(`(?i)\bhaskell\b`), "Haskell"},
		{regexp.MustCompile(`(?i)\bjulia\b`), "Julia"},
		{regexp.MustCompile(`(?i)\bperl\b`), "Perl"},
		{regexp.MustCompile(`(?i)\bdart\b`), "Dart"},
		{regexp.MustCompile(`(?i)\bzig\b`), "Zig"},
		{regexp.MustCompile(`(?i)\bcrystal\b`), "Crystal"},
		{regexp.MustCompile(`(?i)\bvue\.?js\b`), "Vue.js"},
		{regexp.MustCompile(`(?i)\bvuejs\b`), "Vue.js"},
		{regexp.MustCompile(`(?i)\breact\b`), "React"},
		{regexp.MustCompile(`(?i)\bangular\b`), "Angular"},
		{regexp.MustCompile(`(?i)\bnext\.?js\b`), "Next.js"},
		{regexp.MustCompile(`(?i)\bnextjs\b`), "Next.js"},
		{regexp.MustCompile(`(?i)\bnuxt\.?js\b`), "Nuxt.js"},
		{regexp.MustCompile(`(?i)\bnuxtjs\b`), "Nuxt.js"},
		{regexp.MustCompile(`(?i)\bsvelte\b`), "Svelte"},
		{regexp.MustCompile(`(?i)\bsveltekit\b`), "SvelteKit"},
		{regexp.MustCompile(`(?i)\bember\.?js\b`), "Ember.js"},
		{regexp.MustCompile(`(?i)\bbackbone\.?js\b`), "Backbone.js"},
		{regexp.MustCompile(`(?i)\bexpress\.?js\b`), "Express.js"},
		{regexp.MustCompile(`(?i)\bfastify\b`), "Fastify"},
		{regexp.MustCompile(`(?i)\bkoa\.?js\b`), "Koa.js"},
		{regexp.MustCompile(`(?i)\bdjango\b`), "Django"},
		{regexp.MustCompile(`(?i)\bflask\b`), "Flask"},
		{regexp.MustCompile(`(?i)\bfastapi\b`), "FastAPI"},
		{regexp.MustCompile(`(?i)\bruby\s?on\s?rails\b`), "Ruby on Rails"},
		{regexp.MustCompile(`(?i)\brails\b`), "Rails"},
		{regexp.MustCompile(`(?i)\bspring boot\b`), "Spring Boot"},
		{regexp.MustCompile(`(?i)\blaravel\b`), "Laravel"},
		{regexp.MustCompile(`(?i)\bsymfony\b`), "Symfony"},
		{regexp.MustCompile(`(?i)\bcodeigniter\b`), "CodeIgniter"},
		{regexp.MustCompile(`(?i)\baspnet\b`), "ASP.NET"},
		{regexp.MustCompile(`(?i)\b\.net\b`), ".NET"},
		{regexp.MustCompile(`(?i)\bbootstrap\b`), "Bootstrap"},
		{regexp.MustCompile(`(?i)\btailwind css\b`), "Tailwind CSS"},
		{regexp.MustCompile(`(?i)\bmaterial-?ui\b`), "Material-UI"},
		{regexp.MustCompile(`(?i)\bant design\b`), "Ant Design"},
		{regexp.MustCompile(`(?i)\bchakra ui\b`), "Chakra UI"},
		{regexp.MustCompile(`(?i)\bpostgresql?\b`), "PostgreSQL"},
		{regexp.MustCompile(`(?i)\bmysql\b`), "MySQL"},
		{regexp.MustCompile(`(?i)\bmongo\s?db\b`), "MongoDB"},
		{regexp.MustCompile(`(?i)\bmongodb\b`), "MongoDB"},
		{regexp.MustCompile(`(?i)\bsqlite\b`), "SQLite"},
		{regexp.MustCompile(`(?i)\bredis\b`), "Redis"},
		{regexp.MustCompile(`(?i)\belasticsearch\b`), "Elasticsearch"},
		{regexp.MustCompile(`(?i)\bsolr\b`), "Solr"},
		{regexp.MustCompile(`(?i)\bcassandra\b`), "Cassandra"},
		{regexp.MustCompile(`(?i)\bcouchdb\b`), "CouchDB"},
		{regexp.MustCompile(`(?i)\bneo4j\b`), "Neo4j"},
		{regexp.MustCompile(`(?i)\binfluxdb\b`), "InfluxDB"},
		{regexp.MustCompile(`(?i)\bmariadb\b`), "MariaDB"},
		{regexp.MustCompile(`(?i)\bdynamodb\b`), "DynamoDB"},
		{regexp.MustCompile(`(?i)\bfirestore\b`), "Firestore"},
		{regexp.MustCompile(`(?i)\bsupabase\b`), "Supabase"},
		{regexp.MustCompile(`(?i)\bprisma\b`), "Prisma"},
		{regexp.MustCompile(`(?i)\bdocker\b`), "Docker"},
		{regexp.MustCompile(`(?i)\bkubernetes\b`), "Kubernetes"},
		{regexp.MustCompile(`(?i)\bk8s\b`), "Kubernetes"},
		{regexp.MustCompile(`(?i)\bterraform\b`), "Terraform"},
		{regexp.MustCompile(`(?i)\bansible\b`), "Ansible"},
		{regexp.MustCompile(`(?i)\bjenkins\b`), "Jenkins"},
		{regexp.MustCompile(`(?i)\bgitlab ci\b`), "GitLab CI"},
		{regexp.MustCompile(`(?i)\bgithub actions\b`), "GitHub Actions"},
		{regexp.MustCompile(`(?i)\btravis ci\b`), "Travis CI"},
		{regexp.MustCompile(`(?i)\bcircleci\b`), "CircleCI"},
		{regexp.MustCompile(`(?i)\baws\b`), "AWS"},
		{regexp.MustCompile(`(?i)\bamazon web services\b`), "Amazon Web Services"},
		{regexp.MustCompile(`(?i)\bmicrosoft azure\b`), "Microsoft Azure"},
		{regexp.MustCompile(`(?i)\bgoogle cloud\b`), "Google Cloud"},
		{regexp.MustCompile(`(?i)\bgcp\b`), "GCP"},
		{regexp.MustCompile(`(?i)\bdigitalocean\b`), "DigitalOcean"},
		{regexp.MustCompile(`(?i)\blinode\b`), "Linode"},
		{regexp.MustCompile(`(?i)\bvultr\b`), "Vultr"},
		{regexp.MustCompile(`(?i)\bheroku\b`), "Heroku"},
		{regexp.MustCompile(`(?i)\bvercel\b`), "Vercel"},
		{regexp.MustCompile(`(?i)\bnetlify\b`), "Netlify"},
		{regexp.MustCompile(`(?i)\bcloudflare\b`), "Cloudflare"},
		{regexp.MustCompile(`(?i)\bweb\s?assembly\b`), "WebAssembly"},
		{regexp.MustCompile(`(?i)\bwebassembly\b`), "WebAssembly"},
		{regexp.MustCompile(`(?i)\bwasm\b`), "WebAssembly"},
		{regexp.MustCompile(`(?i)\bgraph\s?ql\b`), "GraphQL"},
		{regexp.MustCompile(`(?i)\bgraphql\b`), "GraphQL"},
		{regexp.MustCompile(`(?i)\bwebsocket\b`), "WebSocket"},
		{regexp.MustCompile(`(?i)\bwebrtc\b`), "WebRTC"},
		{regexp.MustCompile(`(?i)\bpwa\b`), "PWA"},
		{regexp.MustCompile(`(?i)\bprogressive web app\b`), "Progressive Web App"},
		{regexp.MustCompile(`(?i)\bservice worker\b`), "Service Worker"},
		{regexp.MustCompile(`(?i)\bwebgl\b`), "WebGL"},
		{regexp.MustCompile(`(?i)\bwebgpu\b`), "WebGPU"},
		{regexp.MustCompile(`(?i)\bwebpack\b`), "Webpack"},
		{regexp.MustCompile(`(?i)\bvite\b`), "Vite"},
		{regexp.MustCompile(`(?i)\bparcel\b`), "Parcel"},
		{regexp.MustCompile(`(?i)\brollup\b`), "Rollup"},
		{regexp.MustCompile(`(?i)\besbuild\b`), "esbuild"},
		{regexp.MustCompile(`(?i)\bbabel\b`), "Babel"},
		{regexp.MustCompile(`(?i)\bswc\b`), "SWC"},
		{regexp.MustCompile(`(?i)\bturbopack\b`), "Turbopack"},
		{regexp.MustCompile(`(?i)\bjest\b`), "Jest"},
		{regexp.MustCompile(`(?i)\bvitest\b`), "Vitest"},
		{regexp.MustCompile(`(?i)\bmocha\b`), "Mocha"},
		{regexp.MustCompile(`(?i)\bchai\b`), "Chai"},
		{regexp.MustCompile(`(?i)\bjasmine\b`), "Jasmine"},
		{regexp.MustCompile(`(?i)\bcypress\b`), "Cypress"},
		{regexp.MustCompile(`(?i)\bplaywright\b`), "Playwright"},
		{regexp.MustCompile(`(?i)\bpuppeteer\b`), "Puppeteer"},
		{regexp.MustCompile(`(?i)\bselenium\b`), "Selenium"},
		{regexp.MustCompile(`(?i)\bwebdriver\b`), "WebDriver"},
		{regexp.MustCompile(`(?i)\bstorybook\b`), "Storybook"},
		{regexp.MustCompile(`(?i)\btensorflow\b`), "TensorFlow"},
		{regexp.MustCompile(`(?i)\bpytorch\b`), "PyTorch"},
		{regexp.MustCompile(`(?i)\bscikit-?learn\b`), "scikit-learn"},
		{regexp.MustCompile(`(?i)\bkeras\b`), "Keras"},
		{regexp.MustCompile(`(?i)\bpandas\b`), "Pandas"},
		{regexp.MustCompile(`(?i)\bnumpy\b`), "NumPy"},
		{regexp.MustCompile(`(?i)\bscipy\b`), "SciPy"},
		{regexp.MustCompile(`(?i)\bmatplotlib\b`), "Matplotlib"},
		{regexp.MustCompile(`(?i)\bseaborn\b`), "Seaborn"},
		{regexp.MustCompile(`(?i)\bjupyter\b`), "Jupyter"},
		{regexp.MustCompile(`(?i)\bopenai\b`), "OpenAI"},
		{regexp.MustCompile(`(?i)\bhugging face\b`), "Hugging Face"},
		{regexp.MustCompile(`(?i)\blangchain\b`), "LangChain"},
		{regexp.MustCompile(`(?i)\bllama\b`), "LLaMA"},
		{regexp.MustCompile(`(?i)\bchatgpt\b`), "ChatGPT"},
		{regexp.MustCompile(`(?i)\breact native\b`), "React Native"},
		{regexp.MustCompile(`(?i)\bflutter\b`), "Flutter"},
		{regexp.MustCompile(`(?i)\bxamarin\b`), "Xamarin"},
		{regexp.MustCompile(`(?i)\bionic\b`), "Ionic"},
		{regexp.MustCompile(`(?i)\bcordova\b`), "Cordova"},
		{regexp.MustCompile(`(?i)\bphonegap\b`), "PhoneGap"},
		{regexp.MustCompile(`(?i)\bnativescript\b`), "NativeScript"},
		{regexp.MustCompile(`(?i)\bexpo\b`), "Expo"},
		{regexp.MustCompile(`(?i)\bbitcoin\b`), "Bitcoin"},
		{regexp.MustCompile(`(?i)\bethereum\b`), "Ethereum"},
		{regexp.MustCompile(`(?i)\bblockchain\b`), "Blockchain"},
		{regexp.MustCompile(`(?i)\bsolidity\b`), "Solidity"},
		{regexp.MustCompile(`(?i)\bweb3\b`), "Web3"},
		{regexp.MustCompile(`(?i)\bnft\b`), "NFT"},
		{regexp.MustCompile(`(?i)\bdefi\b`), "DeFi"},
		{regexp.MustCompile(`(?i)\bdao\b`), "DAO"},
		{regexp.MustCompile(`(?i)\bmetamask\b`), "MetaMask"},
		{regexp.MustCompile(`(?i)\bopenzeppelin\b`), "OpenZeppelin"},
		{regexp.MustCompile(`(?i)\bunity\b`), "Unity"},
		{regexp.MustCompile(`(?i)\bunreal engine\b`), "Unreal Engine"},
		{regexp.MustCompile(`(?i)\bgodot\b`), "Godot"},
		{regexp.MustCompile(`(?i)\bblender\b`), "Blender"},
		{regexp.MustCompile(`(?i)\bopengl\b`), "OpenGL"},
		{regexp.MustCompile(`(?i)\bvulkan\b`), "Vulkan"},
		{regexp.MustCompile(`(?i)\bdirectx\b`), "DirectX"},
		{regexp.MustCompile(`(?i)\bffmpeg\b`), "FFmpeg"},
		{regexp.MustCompile(`(?i)\bimagemagick\b`), "ImageMagick"},
		{regexp.MustCompile(`(?i)\bgimp\b`), "GIMP"},
		{regexp.MustCompile(`(?i)\binkscape\b`), "Inkscape"},
		{regexp.MustCompile(`(?i)\bphotoshop\b`), "Photoshop"},
		{regexp.MustCompile(`(?i)\billustrator\b`), "Illustrator"},
		{regexp.MustCompile(`(?i)\bafter effects\b`), "After Effects"},
		{regexp.MustCompile(`(?i)\bpremiere pro\b`), "Premiere Pro"},
		{regexp.MustCompile(`(?i)\bfigma\b`), "Figma"},
		{regexp.MustCompile(`(?i)\bsketch\b`), "Sketch"},
		{regexp.MustCompile(`(?i)\binvision\b`), "InVision"},
		{regexp.MustCompile(`(?i)\bcanva\b`), "Canva"},
		{regexp.MustCompile(`(?i)\bstack overflow\b`), "Stack Overflow"},
		{regexp.MustCompile(`(?i)\byoutube\b`), "YouTube"},
		{regexp.MustCompile(`(?i)\bgithub\b`), "GitHub"},
		{regexp.MustCompile(`(?i)\bgitlab\b`), "GitLab"},
		{regexp.MustCompile(`(?i)\bbitbucket\b`), "Bitbucket"},
		{regexp.MustCompile(`(?i)\bsourcetree\b`), "SourceTree"},
		{regexp.MustCompile(`(?i)\bslack\b`), "Slack"},
		{regexp.MustCompile(`(?i)\bdiscord\b`), "Discord"},
		{regexp.MustCompile(`(?i)\bmicrosoft teams\b`), "Microsoft Teams"},
		{regexp.MustCompile(`(?i)\bzoom\b`), "Zoom"},
		{regexp.MustCompile(`(?i)\btrello\b`), "Trello"},
		{regexp.MustCompile(`(?i)\basana\b`), "Asana"},
		{regexp.MustCompile(`(?i)\bnotion\b`), "Notion"},
		{regexp.MustCompile(`(?i)\bobsidian\b`), "Obsidian"},
		{regexp.MustCompile(`(?i)\bevernote\b`), "Evernote"},
		{regexp.MustCompile(`(?i)\bonenote\b`), "OneNote"},
		{regexp.MustCompile(`(?i)\bgoogle docs\b`), "Google Docs"},
		{regexp.MustCompile(`(?i)\bgoogle sheets\b`), "Google Sheets"},
		{regexp.MustCompile(`(?i)\bgoogle drive\b`), "Google Drive"},
		{regexp.MustCompile(`(?i)\bmicrosoft office\b`), "Microsoft Office"},
		{regexp.MustCompile(`(?i)\boffice 365\b`), "Office 365"},
		{regexp.MustCompile(`(?i)\bsharepoint\b`), "SharePoint"},
		{regexp.MustCompile(`(?i)\bdropbox\b`), "Dropbox"},
		{regexp.MustCompile(`(?i)\bicloud\b`), "iCloud"},
		{regexp.MustCompile(`(?i)\bonedrive\b`), "OneDrive"},
		{regexp.MustCompile(`(?i)\bmacos\b`), "macOS"},
		{regexp.MustCompile(`(?i)\bwindows\b`), "Windows"},
		{regexp.MustCompile(`(?i)\blinux\b`), "Linux"},
		{regexp.MustCompile(`(?i)\bubuntu\b`), "Ubuntu"},
		{regexp.MustCompile(`(?i)\bdebian\b`), "Debian"},
		{regexp.MustCompile(`(?i)\bcentos\b`), "CentOS"},
		{regexp.MustCompile(`(?i)\bredhat\b`), "RedHat"},
		{regexp.MustCompile(`(?i)\bfedora\b`), "Fedora"},
		{regexp.MustCompile(`(?i)\barch linux\b`), "Arch Linux"},
		{regexp.MustCompile(`(?i)\balpine\b`), "Alpine"},
		{regexp.MustCompile(`(?i)\bfreebsd\b`), "FreeBSD"},
		{regexp.MustCompile(`(?i)\bopenbsd\b`), "OpenBSD"},
		{regexp.MustCompile(`(?i)\bnetbsd\b`), "NetBSD"},
		{regexp.MustCompile(`(?i)\bios\b`), "iOS"},
		{regexp.MustCompile(`(?i)\bandroid\b`), "Android"},
		{regexp.MustCompile(`(?i)\bwatchos\b`), "watchOS"},
		{regexp.MustCompile(`(?i)\btvos\b`), "tvOS"},
		{regexp.MustCompile(`(?i)\bipados\b`), "iPadOS"},
		{regexp.MustCompile(`(?i)\bvscode\b`), "VSCode"},
		{regexp.MustCompile(`(?i)\bvisual studio code\b`), "Visual Studio Code"},
		{regexp.MustCompile(`(?i)\bvisual studio\b`), "Visual Studio"},
		{regexp.MustCompile(`(?i)\bintellij\b`), "IntelliJ"},
		{regexp.MustCompile(`(?i)\bwebstorm\b`), "WebStorm"},
		{regexp.MustCompile(`(?i)\bpycharm\b`), "PyCharm"},
		{regexp.MustCompile(`(?i)\bphpstorm\b`), "PhpStorm"},
		{regexp.MustCompile(`(?i)\bgoland\b`), "GoLand"},
		{regexp.MustCompile(`(?i)\brubymine\b`), "RubyMine"},
		{regexp.MustCompile(`(?i)\bclion\b`), "CLion"},
		{regexp.MustCompile(`(?i)\brider\b`), "Rider"},
		{regexp.MustCompile(`(?i)\beclipse\b`), "Eclipse"},
		{regexp.MustCompile(`(?i)\bnetbeans\b`), "NetBeans"},
		{regexp.MustCompile(`(?i)\bxcode\b`), "Xcode"},
		{regexp.MustCompile(`(?i)\bandroid studio\b`), "Android Studio"},
		{regexp.MustCompile(`(?i)\bsublime text\b`), "Sublime Text"},
		{regexp.MustCompile(`(?i)\batom\b`), "Atom"},
		{regexp.MustCompile(`(?i)\bbrackets\b`), "Brackets"},
		{regexp.MustCompile(`(?i)\bvim\b`), "Vim"},
		{regexp.MustCompile(`(?i)\bneovim\b`), "Neovim"},
		{regexp.MustCompile(`(?i)\bemacs\b`), "Emacs"},
		{regexp.MustCompile(`(?i)\bnano\b`), "Nano"},
		{regexp.MustCompile(`(?i)\bgit\b`), "Git"},
		{regexp.MustCompile(`(?i)\bsvn\b`), "SVN"},
		{regexp.MustCompile(`(?i)\bmercurial\b`), "Mercurial"},
		{regexp.MustCompile(`(?i)\bpostman\b`), "Postman"},
		{regexp.MustCompile(`(?i)\binsomnia\b`), "Insomnia"},
		{regexp.MustCompile(`(?i)\bwireshark\b`), "Wireshark"},
	}
}

// spellCheckRule checks for common spelling mistakes.
type spellCheckRule struct{}

func (r *spellCheckRule) ID() string { return "awesome-spell-check" }

func (r *spellCheckRule) Fixable() bool { return true }

// shouldSkipSpellFix reports whether a spell-check match at the given 1-based
// column in line should be left untouched. It mirrors the contexts that Fix
// deliberately skips: text inside a URL and inside a link definition label.
// Check must agree so that reported issues are always fixable.
func shouldSkipSpellFix(lineStr string, col int, wrong string) bool {
	idx := col - 1
	if idx < 0 || idx+len(wrong) > len(lineStr) || lineStr[idx:idx+len(wrong)] != wrong {
		return true // Out of bounds or not at this exact position (nothing to fix)
	}
	// Skip if the match is inside a URL (between :// and a space/end)
	before := lineStr[:idx]
	if i := strings.LastIndex(before, "://"); i >= 0 {
		afterProtocol := before[i+3:]
		if !strings.ContainsAny(afterProtocol, " \t") {
			return true // Inside a URL, skip
		}
	}
	// Skip if the match is inside a definition label ([label]:)
	openBracket := strings.LastIndex(before, "[")
	if openBracket >= 0 {
		after := lineStr[idx+len(wrong):]
		if strings.HasPrefix(after, "]:") || strings.HasPrefix(after, "] ") {
			beforeBracket := before[:openBracket]
			if !strings.ContainsAny(beforeBracket, "[") {
				return true // Inside a definition label, skip
			}
		}
	}
	return false
}

func (r *spellCheckRule) Fix(_ *MarkdownDoc, source []byte, results []Result) []byte {
	// Apply fixes in reverse order to preserve positions
	for i := len(results) - 1; i >= 0; i-- {
		res := results[i]
		if res.RuleID != r.ID() {
			continue
		}
		lines := bytes.Split(source, []byte("\n"))
		if res.Line-1 >= len(lines) {
			continue
		}
		lineStr := string(lines[res.Line-1])
		var wrong, correct string
		if n, err := fmt.Sscanf(res.Message, "Text %q should be written as %q", &wrong, &correct); n == 2 && err == nil {
			if shouldSkipSpellFix(lineStr, res.Column, wrong) {
				continue
			}
			newLine := lineStr[:res.Column-1] + correct + lineStr[res.Column-1+len(wrong):]
			lines[res.Line-1] = []byte(newLine)
			source = bytes.Join(lines, []byte("\n"))
		}
	}
	return source
}

func (r *spellCheckRule) Check(doc *MarkdownDoc, source []byte) []Result {
	var results []Result
	rules := spellCheckRules()

	// Only scan prose text: Direct String/Text content of paragraphs, while
	// skipping link labels/destinations, code, and raw HTML. This prevents
	// flagging project names inside links (e.g. [obsidian-typst]) or URL hosts.
	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindText {
			return ast.WalkContinue
		}
		text := n.(*ast.Text)
		if text.Segment.IsEmpty() {
			return ast.WalkContinue
		}
		// Skip text that lives inside a link, image, or code span, plus any
		// text directly inside a heading is allowed but code blocks are not.
		for p := n.Parent(); p != nil; p = p.Parent() {
			switch p.Kind() {
			case ast.KindLink, ast.KindImage, ast.KindCodeSpan:
				return ast.WalkContinue
			case ast.KindFencedCodeBlock, ast.KindCodeBlock:
				return ast.WalkContinue
			}
		}

		seg := text.Segment
		segmentStart := seg.Start

		for _, entry := range rules {
			// Run the pattern against the text node segment only.
			for _, loc := range entry.pattern.FindAllIndex(source[seg.Start:seg.Stop], -1) {
				pos := segmentStart + loc[0]
				matched := string(source[pos : pos+(loc[1]-loc[0])])
				// Skip if the text is already in the correct form
				if matched == entry.correct {
					continue
				}
				line := bytes.Count(source[:pos], []byte("\n")) + 1
				col := pos - bytes.LastIndex(source[:pos], []byte("\n"))
				results = append(results, Result{
					RuleID:   r.ID(),
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("Text %q should be written as %q (if referring to the technology)", matched, entry.correct),
					Line:     line,
					Column:   col,
				})
			}
		}
		return ast.WalkContinue
	}))

	return results
}

// noRepeatItemInDescriptionRule checks that list item descriptions don't
// start by repeating the item name.
type noRepeatItemInDescriptionRule struct{}

func (r *noRepeatItemInDescriptionRule) ID() string { return "no-repeat-item-in-description" }

func (r *noRepeatItemInDescriptionRule) Fixable() bool { return false }

func (r *noRepeatItemInDescriptionRule) Fix(_ *MarkdownDoc, source []byte, _ []Result) []byte {
	return source
}

func (r *noRepeatItemInDescriptionRule) Check(doc *MarkdownDoc, source []byte) []Result {
	var results []Result

	ast.Walk(doc.Root, walkHelper(func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || n.Kind() != ast.KindListItem {
			return ast.WalkContinue
		}

		var para ast.Node
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() == ast.KindParagraph || c.Kind() == ast.KindTextBlock {
				para = c
				break
			}
		}
		if para == nil {
			return ast.WalkContinue
		}

		// Find the main link (first link in the paragraph)
		var mainLink ast.Node
		var linkIndex int
		for c := para.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() == ast.KindLink {
				// Skip links that contain images (badge links)
				hasImage := false
				for cc := c.FirstChild(); cc != nil; cc = cc.NextSibling() {
					if cc.Kind() == ast.KindImage {
						hasImage = true
						break
					}
				}
				if !hasImage {
					mainLink = c
					break
				}
			}
			linkIndex++
		}
		if mainLink == nil {
			return ast.WalkContinue
		}

		itemName := doc.TextOf(mainLink)
		if itemName == "" {
			return ast.WalkContinue
		}

		// Find the description text after the main link
		var descText string
		for c := mainLink.NextSibling(); c != nil; c = c.NextSibling() {
			if c.Kind() == ast.KindText {
				text := string(c.(*ast.Text).Segment.Value(source))
				// Look for " - " separator
				if _, after, ok := strings.Cut(text, " - "); ok {
					descText += after
				}
			}
		}
		descText = strings.TrimSpace(descText)

		if descText == "" {
			return ast.WalkContinue
		}

		// Check if description starts with item name (case-insensitive)
		lowerDesc := strings.ToLower(descText)
		lowerName := strings.ToLower(itemName)
		if strings.HasPrefix(lowerDesc, lowerName) {
			afterName := descText[len(lowerName):]
			// Check for word boundary to avoid false positives
			if afterName == "" || strings.Contains(" .,!?:-", string(afterName[0])) {
				line, col := doc.LineColOf(mainLink)
				results = append(results, Result{
					RuleID:   r.ID(),
					Severity: SeverityError,
					Message:  fmt.Sprintf("List item description must not start with the item name %q", itemName),
					Line:     line,
					Column:   col,
				})
			}
		}

		return ast.WalkContinue
	}))

	return results
}

// definitionCaseRule checks that definition labels are lowercase.
type definitionCaseRule struct{}

func (r *definitionCaseRule) ID() string { return "definition-case" }

func (r *definitionCaseRule) Fixable() bool { return true }

func (r *definitionCaseRule) Fix(doc *MarkdownDoc, source []byte, results []Result) []byte {
	_ = doc
	for _, res := range results {
		if res.RuleID != r.ID() {
			continue
		}
		lines := bytes.Split(source, []byte("\n"))
		if res.Line-1 < len(lines) {
			line := lines[res.Line-1]
			str := string(line)
			// Find the first [label] pattern
			if start := strings.Index(str, "["); start >= 0 {
				if end := strings.Index(str[start:], "]"); end >= 0 {
					label := str[start+1 : start+end]
					lower := strings.ToLower(label)
					if label != lower {
						newLine := str[:start+1] + lower + str[start+end:]
						lines[res.Line-1] = []byte(newLine)
						source = bytes.Join(lines, []byte("\n"))
					}
				}
			}
		}
	}
	return source
}

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
