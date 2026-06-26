package site

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gosimple/slug"
)

// TreeNode represents a page or section in the navigation tree.
type TreeNode struct {
	Title    string
	URL      string
	Children []TreeNode
}

// PageData holds the rendered content and metadata for a single page.
type PageData struct {
	Title    string
	Content  []byte
	HTML     []byte
	URL      string
	Children []PageData
	Draft    bool
	Order    int
}

// titleFromDirName converts a directory name to a display title.
// "services" → "Services", "my-section" → "My section"
func titleFromDirName(name string) string {
	if len(name) == 0 {
		return "Untitled"
	}
	// Replace hyphens/underscores with spaces, then title case the first letter
	result := strings.ReplaceAll(name, "-", " ")
	result = strings.ReplaceAll(result, "_", " ")
	// Upper case first letter (handle multi-byte safely with simple ASCII approach)
	b := []byte(result)
	if len(b) > 0 && b[0] >= 'a' && b[0] <= 'z' {
		b[0] = b[0] - 'a' + 'A'
	}
	return string(b)
}

// WalkContentDir recursively walks contentDir and returns a tree of PageData.
// Directories with an index.md or index.org become section nodes.
func WalkContentDir(contentDir string) ([]PageData, error) {
	return walkDir(contentDir, contentDir)
}

func walkDir(dir string, root string) ([]PageData, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var pages []PageData

	// Process files before directories so that index pages sort first.
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".md" && ext != ".org" {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fm, body := ParseFrontmatter(content)
		title := ""
		if fm != nil {
			title = fm.Title
		}
		if title == "" {
			title = DetectTitleFromContent(body)
		}
		var html []byte
		if ext == ".org" {
			html, err = RenderOrg(body, entry.Name())
		} else {
			html, err = RenderMarkdown(body)
		}
		if err != nil {
			return nil, err
		}
		draft := false
		order := 0
		if fm != nil {
			draft = fm.Draft
			order = fm.Order
		} else {
			order = DetectOrderFromContent(content)
		}
		rel, _ := filepath.Rel(root, path)
		baseName := strings.TrimSuffix(rel, ext)
		parts := strings.Split(baseName, string(filepath.Separator))
		for i, p := range parts {
			parts[i] = slug.Make(p)
		}
		url := strings.Join(parts, "/") + ".html"
		pages = append(pages, PageData{
			Title:   title,
			Content: body,
			HTML:    html,
			URL:     url,
			Draft:   draft,
			Order:   order,
		})
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		children, err := walkDir(path, root)
		if err != nil {
			return nil, err
		}
		if len(children) > 0 {
			// Use the directory name as the section title when no index
			// page provides one via frontmatter or heading.
			sectionTitle := titleFromDirName(entry.Name())
			firstURL := ""
			for _, c := range children {
				if c.URL == "" || strings.HasSuffix(c.URL, "/index.html") || strings.HasSuffix(c.URL, "/index.md") || strings.HasSuffix(c.URL, "/index.org") {
					if c.Title != "" {
						sectionTitle = c.Title
					}
					if c.URL != "" && firstURL == "" {
						firstURL = c.URL
					}
				}
			}
			// Fall back to first content-bearing child for URL
			if firstURL == "" {
				for _, c := range children {
					if c.URL != "" && len(c.HTML) > 0 {
						firstURL = c.URL
						break
					}
				}
			}
			if firstURL == "" {
				firstURL = children[0].URL
			}
			pages = append(pages, PageData{
				Title:    sectionTitle,
				URL:      firstURL,
				Children: children,
			})
		}
	}

	sort.SliceStable(pages, func(i, j int) bool {
		if pages[i].Order != pages[j].Order {
			return pages[i].Order < pages[j].Order
		}
		return pages[i].Title < pages[j].Title
	})

	return pages, nil
}

// BuildPageTree converts PageData into a flat list for rendering.
// It also returns the navigation TreeNode tree.
func BuildPageTree(root string) ([]PageData, *TreeNode, error) {
	pages, err := WalkContentDir(root)
	if err != nil {
		return nil, nil, err
	}

	// Flatten pages and build tree
	var flat []PageData
	navRoot := &TreeNode{Title: "Root"}

	for _, p := range pages {
		if p.Draft {
			continue
		}
		// Compute section URL from index page first, then content-bearing child
		var sectionURL string
		for _, c := range p.Children {
			if c.URL != "" && (strings.HasSuffix(c.URL, "/index.html") || strings.HasSuffix(c.URL, "/index.md") || strings.HasSuffix(c.URL, "/index.org")) {
				if idx := strings.LastIndex(c.URL, "/"); idx >= 0 {
					sectionURL = c.URL[:idx] + "/"
				} else {
					sectionURL = c.URL + "/"
				}
				break
			}
		}
		if sectionURL == "" {
			for _, c := range p.Children {
				if c.URL != "" && len(c.HTML) > 0 {
					if idx := strings.LastIndex(c.URL, "/"); idx >= 0 {
						sectionURL = c.URL[:idx] + "/"
					} else {
						sectionURL = c.URL + "/"
					}
					break
				}
			}
		}
		if sectionURL == "" {
			for _, c := range p.Children {
				if c.URL != "" {
					if idx := strings.LastIndex(c.URL, "/"); idx >= 0 {
						sectionURL = c.URL[:idx] + "/"
					} else {
						sectionURL = c.URL + "/"
					}
					break
				}
			}
		}
		flattenPage(p, sectionURL, &flat, navRoot)
	}

	return flat, navRoot, nil
}

func flattenPage(p PageData, prefixURL string, flat *[]PageData, navNode *TreeNode) {
	hasContent := len(p.HTML) > 0

	if hasContent {
		*flat = append(*flat, p)
		navNode.Children = append(navNode.Children, TreeNode{
			Title: p.Title,
			URL:   p.URL,
		})
	}

	if len(p.Children) > 0 {
		var sectionNode *TreeNode
		sectionURL := prefixURL

		if !hasContent {
			if sectionURL == "" {
				for _, c := range p.Children {
					if c.URL != "" {
						if idx := strings.LastIndex(c.URL, "/"); idx >= 0 {
							sectionURL = c.URL[:idx] + "/"
						} else {
							sectionURL = c.URL + "/"
						}
						break
					}
				}
			}
			sectionNode = &TreeNode{Title: p.Title, URL: sectionURL}
			// Reserve a slot — the fully built node will be written back
			// after children are populated (avoids copying an empty Children slice).
			nChildren := len(navNode.Children)
			navNode.Children = append(navNode.Children, TreeNode{})
			defer func() {
				navNode.Children[nChildren] = *sectionNode
			}()
		} else if sectionURL != "" {
			sectionNode = &navNode.Children[len(navNode.Children)-1]
			sectionNode.URL = sectionURL
			sectionNode.Children = []TreeNode{}
		}

		if sectionNode == nil {
			return
		}

		for _, child := range p.Children {
			if child.Draft {
				continue
			}

			if len(child.HTML) == 0 && len(child.Children) > 0 {
				// Nested section without index page — recurse
				flattenPage(child, "", flat, sectionNode)
				continue
			}

			childURL := child.URL
			if sectionURL != "" && !strings.HasPrefix(child.URL, sectionURL) && child.URL != "" {
				childURL = sectionURL + child.URL
			}
			sectionNode.Children = append(sectionNode.Children, TreeNode{
				Title: child.Title,
				URL:   childURL,
			})
			if len(child.HTML) > 0 {
				*flat = append(*flat, PageData{
					Title: child.Title,
					HTML:  child.HTML,
					URL:   childURL,
					Draft: child.Draft,
					Order: child.Order,
				})
			}
		}
	}
}
