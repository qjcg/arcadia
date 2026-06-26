package site

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
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
		}
		rel, _ := filepath.Rel(root, path)
		baseName := strings.TrimSuffix(rel, ext)
		url := baseName + ".html"
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
			var sectionTitle string
			if children[0].Title != "" {
				sectionTitle = children[0].Title
			}
			firstURL := children[0].URL
			for _, c := range children {
				if c.URL != "" {
					firstURL = c.URL
					break
				}
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
		// Compute section URL from first child's path
		var sectionURL string
		for _, c := range p.Children {
			if c.URL != "" {
				firstURL := c.URL
				if idx := strings.LastIndex(firstURL, "/"); idx >= 0 {
					sectionURL = firstURL[:idx] + "/"
				} else {
					sectionURL = firstURL + "/"
				}
				break
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
			navNode.Children = append(navNode.Children, *sectionNode)
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
